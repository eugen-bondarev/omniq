package services

import (
	"log"
)

// MockSMTPService is a simple implementation for demonstration
type MockSMTPService struct{}

func (s *MockSMTPService) SendEmail(to, subject, htmlBody string) error {
	// In a real implementation, this would use SMTP
	log.Printf("📧 SMTP: Sending email\n  To: %s\n  Subject: %s\n  Body: %s", to, subject, htmlBody)
	return nil
}

func (s *MockSMTPService) SendTemplateEmail(to, template string, data map[string]interface{}) error {
	// In a real implementation, this would render a template and send
	log.Printf("📧 SMTP: Sending template email\n  To: %s\n  Template: %s\n  Data: %+v", to, template, data)
	return nil
}

// MockDatabaseService is a simple implementation for demonstration
type MockDatabaseService struct{}

func (d *MockDatabaseService) Query(query string, args ...interface{}) (interface{}, error) {
	log.Printf("💾 DB: Executing query: %s with args: %+v", query, args)
	return map[string]interface{}{"result": "mock_data"}, nil
}

func (d *MockDatabaseService) Execute(query string, args ...interface{}) error {
	log.Printf("💾 DB: Executing command: %s with args: %+v", query, args)
	return nil
}
