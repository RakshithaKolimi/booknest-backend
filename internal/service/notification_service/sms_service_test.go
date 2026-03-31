package notification_service

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sns"

	"booknest/internal/domain"
)

func TestSNSSMSSendSMSSuccess(t *testing.T) {
	httpClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.Method != http.MethodPost {
				t.Fatalf("expected POST request, got %s", req.Method)
			}

			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("read request body: %v", err)
			}
			if !bytes.Contains(body, []byte("%2B919999999999")) {
				t.Fatalf("expected encoded phone number in request body, got %s", string(body))
			}
			if !bytes.Contains(body, []byte("OTP")) {
				t.Fatalf("expected message content in request body, got %s", string(body))
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Content-Type": []string{"text/xml"},
				},
				Body: io.NopCloser(bytes.NewBufferString(
					`<PublishResponse xmlns="http://sns.amazonaws.com/doc/2010-03-31/"><PublishResult><MessageId>msg-123</MessageId></PublishResult><ResponseMetadata><RequestId>req-123</RequestId></ResponseMetadata></PublishResponse>`,
				)),
			}, nil
		}),
	}

	client := sns.NewFromConfig(aws.Config{
		Region:      "ap-south-1",
		Credentials: credentials.NewStaticCredentialsProvider("key", "secret", ""),
		HTTPClient:  httpClient,
	})

	provider := NewSNSSMS(client)
	if err := provider.SendSMS("+919999999999", "Your OTP is 123456"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestSNSSMSSendSMSFailure(t *testing.T) {
	httpClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Header: http.Header{
					"Content-Type": []string{"text/xml"},
				},
				Body: io.NopCloser(bytes.NewBufferString(
					`<ErrorResponse xmlns="http://sns.amazonaws.com/doc/2010-03-31/"><Error><Type>Sender</Type><Code>InvalidParameter</Code><Message>Bad phone number</Message></Error><RequestId>req-123</RequestId></ErrorResponse>`,
				)),
			}, nil
		}),
	}

	client := sns.NewFromConfig(aws.Config{
		Region:      "ap-south-1",
		Credentials: credentials.NewStaticCredentialsProvider("key", "secret", ""),
		HTTPClient:  httpClient,
	})

	provider := NewSNSSMS(client)
	if err := provider.SendSMS("+919999999999", "Your OTP is 123456"); err == nil {
		t.Fatalf("expected error, got nil")
	}
}

var _ domain.SMSProvider = (*SNSSMS)(nil)
