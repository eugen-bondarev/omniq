package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
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

type AddJobData struct {
	JobName   string
	RunParams string
}

// parseJobsDirectory scans all .go files in a directory and extracts job information
func parseJobsDirectory(jobsDir string) ([]JobInfo, string, string, string, error) {
	// Find all .go files in the directory, excluding generate.go and *_gen.go
	files, err := filepath.Glob(filepath.Join(jobsDir, "*.go"))
	if err != nil {
		return nil, "", "", "", err
	}

	var jobFiles []string
	for _, file := range files {
		basename := filepath.Base(file)
		if basename == "generate.go" || strings.HasSuffix(basename, "_gen.go") {
			continue
		}
		jobFiles = append(jobFiles, file)
	}

	if len(jobFiles) == 0 {
		return nil, "", "", "", fmt.Errorf("no job files found in directory %s", jobsDir)
	}

	var allJobs []JobInfo
	var packageName string
	var commonDepType string
	var depImportPath string

	// Parse each job file
	for _, filename := range jobFiles {
		jobs, pkgName, fileDepType, depImport, err := parseJobsFile(filename)
		if err != nil {
			return nil, "", "", "", fmt.Errorf("error parsing %s: %v", filename, err)
		}

		if packageName == "" {
			packageName = pkgName
		} else if packageName != pkgName {
			return nil, "", "", "", fmt.Errorf("package name mismatch: %s has package %s, but expected %s", filename, pkgName, packageName)
		}

		if commonDepType == "" && len(jobs) > 0 {
			commonDepType = fileDepType
			depImportPath = depImport
		}

		// Verify all jobs in this file have the same dependency type
		for _, job := range jobs {
			if job.DepType != commonDepType {
				return nil, "", "", "", fmt.Errorf("dependency type mismatch: job %s has type %s, but expected %s",
					job.Name, job.DepType, commonDepType)
			}
		}

		allJobs = append(allJobs, jobs...)
	}

	return allJobs, packageName, commonDepType, depImportPath, nil
}

// parseJobsFile parses a single Go file and extracts job information
func parseJobsFile(filename string) ([]JobInfo, string, string, string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, "", "", "", err
	}

	var jobs []JobInfo
	packageName := node.Name.Name
	var depImportPath string
	var commonDepType string

	// First pass: find imports to build import map
	importMap := make(map[string]string) // local name -> full path
	for _, imp := range node.Imports {
		if imp.Path.Value != "" {
			path := strings.Trim(imp.Path.Value, "\"")
			var localName string
			if imp.Name != nil {
				localName = imp.Name.Name
			} else {
				// Extract package name from path
				parts := strings.Split(path, "/")
				localName = parts[len(parts)-1]
			}
			importMap[localName] = path
		}
	}

	// Second pass: find job structs and their Run methods
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
			depType, err := extractDependencyType(runMethod, importMap)
			if err != nil {
				return nil, "", "", "", fmt.Errorf("error extracting dependency type from %s.Run: %v", typeSpec.Name.Name, err)
			}

			// Set common dependency type and import
			if commonDepType == "" {
				commonDepType = depType
				// Find the import path for this dependency type
				if strings.Contains(depType, ".") {
					parts := strings.Split(depType, ".")
					if len(parts) >= 2 {
						pkgName := parts[0]
						if importPath, ok := importMap[pkgName]; ok {
							depImportPath = importPath
						}
					}
				}
			} else if depType != commonDepType {
				return nil, "", "", "", fmt.Errorf("dependency type mismatch in file %s: job %s has type %s, but expected %s",
					filename, typeSpec.Name.Name, depType, commonDepType)
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

	return jobs, packageName, commonDepType, depImportPath, nil
}

// findRunMethod finds the Run method for a given struct name
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

// extractDependencyType extracts the dependency type from a Run method
func extractDependencyType(funcDecl *ast.FuncDecl, importMap map[string]string) (string, error) {
	if funcDecl.Type.Params == nil || len(funcDecl.Type.Params.List) == 0 {
		// No parameters means no dependencies - use struct{}
		return "struct{}", nil
	}

	// Expecting exactly one parameter (the dependency)
	if len(funcDecl.Type.Params.List) != 1 {
		return "", fmt.Errorf("Run method should have zero or one parameter")
	}

	param := funcDecl.Type.Params.List[0]
	depType := exprToString(param.Type)

	// Resolve the full type name including package if needed
	if strings.Contains(depType, ".") {
		parts := strings.Split(depType, ".")
		if len(parts) >= 2 {
			pkgName := parts[0]
			if _, ok := importMap[pkgName]; ok {
				// Type is already fully qualified with package name
				return depType, nil
			}
		}
	}

	return depType, nil
}

// exprToString converts an AST expression to its string representation
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
	case *ast.StructType:
		return "struct{}"
	default:
		return fmt.Sprintf("%T", e)
	}
}

// detectExistingJobDependency detects the dependency type used by existing jobs in a directory
func detectExistingJobDependency(jobsDir string) (string, string, error) {
	jobs, _, depType, depImport, err := parseJobsDirectory(jobsDir)
	if err != nil {
		// If parsing fails, it might be because there are no jobs yet
		return "struct{}", "", nil
	}

	if len(jobs) == 0 {
		return "struct{}", "", nil
	}

	// If the detected dependency type is not struct{}, use it
	if depType != "struct{}" {
		return depType, depImport, nil
	}

	return "struct{}", "", nil
}
