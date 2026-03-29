package notification_service

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"

	"booknest/internal/domain"
)

type SESEmail struct {
	client *ses.Client
	from   string
}

func NewSESEmail(client *ses.Client, from string) *SESEmail {
	return &SESEmail{
		client: client,
		from:   from,
	}
}

func (s *SESEmail) SendEmail(
	to string,
	subject string,
	body string,
) (domain.EmailSendResult, error) {

	input := &ses.SendEmailInput{
		Source: &s.from,
		Destination: &types.Destination{
			ToAddresses: []string{to},
		},
		Message: &types.Message{
			Subject: &types.Content{
				Data: &subject,
			},
			Body: &types.Body{
				Html: &types.Content{
					Data: &body,
				},
			},
		},
	}

	output, err := s.client.SendEmail(
		context.TODO(),
		input,
	)
	if err != nil {
		return domain.EmailSendResult{
			Provider: domain.EmailNotificationProviderSES,
		}, err
	}

	rawResponse, marshalErr := json.Marshal(output)
	if marshalErr != nil {
		rawResponse = []byte("{}")
	}

	result := domain.EmailSendResult{
		Provider: domain.EmailNotificationProviderSES,
		Response: string(rawResponse),
	}
	if output.MessageId != nil {
		result.MessageID = *output.MessageId
	}

	return result, nil
}
