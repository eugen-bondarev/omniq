package deps

type Dependencies struct {
	Mailer Mailer
}

type Mailer interface {
	SendEmail(to, subject, body string) error
	SendTemplateEmail(to, template string, data map[string]any) error
}
