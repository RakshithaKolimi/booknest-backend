package notification_service

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
)

const (
	templateKeyVerificationEmail = "verification_email"
	templateKeyPasswordReset     = "password_reset"
	templateKeyOrderReceipt      = "order_receipt"
)

//go:embed templates/*.html
var notificationTemplatesFS embed.FS

var notificationTemplateFiles = map[string]string{
	templateKeyVerificationEmail: "templates/verification_email.html",
	templateKeyPasswordReset:     "templates/password_reset.html",
	templateKeyOrderReceipt:      "templates/order_receipt.html",
}

func renderTemplate(templateKey string, data any) (string, error) {
	templateFile, ok := notificationTemplateFiles[templateKey]
	if !ok {
		return "", fmt.Errorf("notification template %q not found", templateKey)
	}

	tmpl, err := template.ParseFS(notificationTemplatesFS, templateFile)
	if err != nil {
		return "", fmt.Errorf("parse notification template %q: %w", templateKey, err)
	}

	var output bytes.Buffer
	if err := tmpl.Execute(&output, data); err != nil {
		return "", fmt.Errorf("render notification template %q: %w", templateKey, err)
	}

	return output.String(), nil
}
