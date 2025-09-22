package mailgun

import (
	"context"
	"github.com/mailgun/mailgun-go"
	"github.com/modfin/mmailer"
)

type Mailgun struct {
	client *mailgun.MailgunImpl
}

func New(domain, apiKey, posthookUrl string) (*Mailgun, error) {
	mg := &Mailgun{
		client: mailgun.NewMailgun(domain, apiKey),
	}
	// TODO: provide some unique id here, like mmailer-slog for slog deployment?
	err := mg.client.CreateWebhook("mmailer-id", posthookUrl)
	if err != nil {
		return nil, err
	}
	return mg, nil
}

func (m *Mailgun) Name() string {
	return "mailgun"
}

func (m *Mailgun) Send(_ context.Context, email mmailer.Email) ([]mmailer.Response, error) {
	// TODO
	return nil, nil
}

func (m *Mailgun) UnmarshalPosthook(body []byte) ([]mmailer.Posthook, error) {
	// TODO
	// mailgun.Event{}
	return nil, nil
}
