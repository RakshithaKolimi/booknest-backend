package notification_service

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ses"

	"booknest/internal/domain"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestSESEmailSendEmailSuccess(t *testing.T) {
	httpClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.Method != http.MethodPost {
				t.Fatalf("expected POST request, got %s", req.Method)
			}

			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("read request body: %v", err)
			}
			if !bytes.Contains(body, []byte("user%40example.com")) {
				t.Fatalf("expected encoded destination email in request body, got %s", string(body))
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Content-Type": []string{"text/xml"},
				},
				Body: io.NopCloser(bytes.NewBufferString(
					`<SendEmailResponse xmlns="http://ses.amazonaws.com/doc/2010-12-01/"><SendEmailResult><MessageId>msg-123</MessageId></SendEmailResult><ResponseMetadata><RequestId>req-123</RequestId></ResponseMetadata></SendEmailResponse>`,
				)),
			}, nil
		}),
	}

	client := ses.NewFromConfig(aws.Config{
		Region:      "ap-south-1",
		Credentials: credentials.NewStaticCredentialsProvider("key", "secret", ""),
		HTTPClient:  httpClient,
	})

	provider := NewSESEmail(client, "noreply@booknest.test")
	result, err := provider.SendEmail("user@example.com", "Subject", "<p>Hello</p>")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Provider != domain.EmailNotificationProviderSES || result.MessageID != "msg-123" {
		t.Fatalf("unexpected send result: %+v", result)
	}
}

func TestSESEmailSendEmailFailure(t *testing.T) {
	httpClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Header: http.Header{
					"Content-Type": []string{"text/xml"},
				},
				Body: io.NopCloser(bytes.NewBufferString(
					`<ErrorResponse xmlns="http://ses.amazonaws.com/doc/2010-12-01/"><Error><Type>Sender</Type><Code>MessageRejected</Code><Message>Rejected</Message></Error><RequestId>req-123</RequestId></ErrorResponse>`,
				)),
			}, nil
		}),
	}

	client := ses.NewFromConfig(aws.Config{
		Region:      "ap-south-1",
		Credentials: credentials.NewStaticCredentialsProvider("key", "secret", ""),
		HTTPClient:  httpClient,
	})

	provider := NewSESEmail(client, "noreply@booknest.test")
	result, err := provider.SendEmail("user@example.com", "Subject", "<p>Hello</p>")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if result.Provider != domain.EmailNotificationProviderSES {
		t.Fatalf("expected SES provider in error result, got %+v", result)
	}
}

var _ domain.EmailProvider = (*SESEmail)(nil)
