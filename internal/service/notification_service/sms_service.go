package notification_service

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type SNSSMS struct {
	client *sns.Client
}

func NewSNSSMS(client *sns.Client) *SNSSMS {
	return &SNSSMS{
		client: client,
	}
}

func (s *SNSSMS) SendSMS(to string, message string) error {
	_, err := s.client.Publish(context.TODO(), &sns.PublishInput{
		Message:     &message,
		PhoneNumber: &to,
	})
	return err
}
