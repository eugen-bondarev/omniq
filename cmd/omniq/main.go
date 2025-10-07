package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type JobInfo struct {
	Name       string
	Fields     []FieldInfo
	DepType    string
	DepPackage string
}

type FieldInfo struct {
	Name string
	Type string
}

type GenerationData struct {
	Package   string
	Jobs      []JobInfo
	DepType   string
	DepImport string
}

const generatedCodeTemplate = `package {{.Package}}

import (
	"github.com/eugen-bondarev/omniq"
{{if .DepImport}}	"{{.DepImport}}"{{end}}
)

// Jobs
{{range .Jobs}}func (j *{{.Name}}) Type() string {
	return "{{.Name}}"
}

{{end}}{{range .Jobs}}func (j *{{.Name}}) GetIDContainer() *omniq.WithID {
	return &j.WithID
}

{{end}}{{range .Jobs}}func New{{.Name}}(id string, data map[string]any) *{{.Name}} {
	return &{{.Name}}{
		WithID: omniq.WithID{
			ID: id,
		},{{range .Fields}}
		{{.Name}}: data["{{.Name}}"].({{.Type}}),{{end}}
	}
}

{{end}}// Registry
type JobFactory struct{}

func (f *JobFactory) Instantiate(t string, id string, data map[string]any) omniq.Job[{{.DepType}}] {
	var j omniq.Job[{{.DepType}}]
	switch t {
{{range .Jobs}}	case "{{.Name}}":
		j = New{{.Name}}(id, data)
{{end}}	default:
		panic("Unknown job type: " + t)
	}
	return j
}
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <jobs_file>\n", os.Args[0])
		os.Exit(1)
	}

	jobsFile := os.Args[1]

	// Parse the jobs file
	jobs, packageName, depImport, err := parseJobsFile(jobsFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing jobs file: %v\n", err)
		os.Exit(1)
	}

	if len(jobs) == 0 {
		fmt.Fprintf(os.Stderr, "No job structs found in %s\n", jobsFile)
		os.Exit(1)
	}

	// Verify all jobs have the same dependency type
	firstDepType := jobs[0].DepType
	for _, job := range jobs[1:] {
		if job.DepType != firstDepType {
			fmt.Fprintf(os.Stderr, "Error: Job %s has dependency type %s, but %s has %s. All jobs must have the same dependency type.\n",
				job.Name, job.DepType, jobs[0].Name, firstDepType)
			os.Exit(1)
		}
	}

	// Generate the code
	data := GenerationData{
		Package:   packageName,
		Jobs:      jobs,
		DepType:   firstDepType,
		DepImport: depImport,
	}

	tmpl, err := template.New("generated").Parse(generatedCodeTemplate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing template: %v\n", err)
		os.Exit(1)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing template: %v\n", err)
		os.Exit(1)
	}

	// Format the generated code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting generated code: %v\n", err)
		os.Exit(1)
	}

	// Write to jobs_gen.go in the same directory
	dir := filepath.Dir(jobsFile)
	outputFile := filepath.Join(dir, "jobs_gen.go")

	if err := os.WriteFile(outputFile, formatted, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing generated file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated %s\n", outputFile)
}

func parseJobsFile(filename string) ([]JobInfo, string, string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, "", "", err
	}

	var jobs []JobInfo
	packageName := node.Name.Name
	depImportPath := ""

	// Find imports to determine dependency package
	for _, imp := range node.Imports {
		if imp.Path.Value != "" {
			path := strings.Trim(imp.Path.Value, "\"")
			if strings.Contains(path, "/deps") {
				depImportPath = path
			}
		}
	}

	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			// Check if this struct has a Run method
			runMethod := findRunMethod(node, typeSpec.Name.Name)
			if runMethod == nil {
				continue // Not a job struct
			}

			// Extract dependency type from Run method
			depType, err := extractDependencyType(runMethod)
			if err != nil {
				return nil, "", "", fmt.Errorf("error extracting dependency type from %s.Run: %v", typeSpec.Name.Name, err)
			}

			// Extract fields (excluding WithID)
			var fields []FieldInfo
			for _, field := range structType.Fields.List {
				// Skip embedded WithID field
				if len(field.Names) == 0 {
					continue
				}

				for _, name := range field.Names {
					if name.Name == "WithID" {
						continue
					}
					fieldType := exprToString(field.Type)
					fields = append(fields, FieldInfo{
						Name: name.Name,
						Type: fieldType,
					})
				}
			}

			jobs = append(jobs, JobInfo{
				Name:       typeSpec.Name.Name,
				Fields:     fields,
				DepType:    depType,
				DepPackage: depImportPath,
			})
		}
	}

	return jobs, packageName, depImportPath, nil
}

func findRunMethod(node *ast.File, structName string) *ast.FuncDecl {
	for _, decl := range node.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Recv == nil || funcDecl.Name.Name != "Run" {
			continue
		}

		if len(funcDecl.Recv.List) == 0 {
			continue
		}

		// Check if receiver is our struct
		recvType := funcDecl.Recv.List[0].Type
		if starExpr, ok := recvType.(*ast.StarExpr); ok {
			if ident, ok := starExpr.X.(*ast.Ident); ok && ident.Name == structName {
				return funcDecl
			}
		}
	}
	return nil
}

func extractDependencyType(funcDecl *ast.FuncDecl) (string, error) {
	if funcDecl.Type.Params == nil || len(funcDecl.Type.Params.List) == 0 {
		return "", fmt.Errorf("Run method has no parameters")
	}

	// Expecting exactly one parameter (the dependency)
	if len(funcDecl.Type.Params.List) != 1 {
		return "", fmt.Errorf("Run method should have exactly one parameter")
	}

	param := funcDecl.Type.Params.List[0]
	return exprToString(param.Type), nil
}

func exprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return exprToString(e.X) + "." + e.Sel.Name
	case *ast.StarExpr:
		return "*" + exprToString(e.X)
	case *ast.ArrayType:
		return "[]" + exprToString(e.Elt)
	case *ast.MapType:
		return "map[" + exprToString(e.Key) + "]" + exprToString(e.Value)
	default:
		return fmt.Sprintf("%T", e)
	}
}
