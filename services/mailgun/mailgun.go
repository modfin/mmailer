package mailgun

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/mailgun/mailgun-go/v5"
	"github.com/mailgun/mailgun-go/v5/events"
	"github.com/mailgun/mailgun-go/v5/mtypes"
	"github.com/modfin/henry/slicez"
	"github.com/modfin/mmailer"
	"github.com/modfin/mmailer/internal/logger"
	"github.com/modfin/mmailer/services"
)

type Mailgun struct {
	client *mailgun.Client
	confer services.Configurer[*mailgun.PlainMessage]
}

func New(apiKey string, webhookSigningKey string) *Mailgun {
	mg := &Mailgun{
		client: mailgun.NewMailgun(apiKey),
		confer: configurer{},
	}
	mg.client.SetWebhookSigningKey(webhookSigningKey)
	_ = mg.client.SetAPIBase(mailgun.APIBaseEU)
	return mg
}

func (m *Mailgun) Name() string {
	return "mailgun"
}

func (m *Mailgun) CanSend(e mmailer.Email) bool {
	for _, a := range e.Attachments {
		// TODO Can't find the option to set attachment content type in the mailgun api, can it be fixed?
		if a.ContentType == "" || a.ContentType == "application/octet-stream" {
			logger.Warn(
				"mailgun: unsupported attachment content-type",
				"email", e.To,
				"content_type", a.ContentType,
				"filename", a.Name,
			)
			return false
		}
	}
	if !strings.HasSuffix(e.From.Email, "strictlog.modfin.se") {
		return false
	}
	return true
}

func (m *Mailgun) Send(ctx context.Context, e mmailer.Email) ([]mmailer.Response, error) {
	from, err := mail.ParseAddress(e.From.String())
	if err != nil {
		return nil, fmt.Errorf("mailgun: failed to parse email: %w", err)
	}
	parts := strings.Split(from.Address, "@")
	domain, _ := slicez.Last(parts)
	if domain == "" {
		return nil, fmt.Errorf("mailgun: failed to get email domain: %v", from.Address)
	}

	to := slicez.Map(e.To, func(a mmailer.Address) string {
		return a.String()
	})

	msg := mailgun.NewMessage(domain, from.String(), e.Subject, e.Text, to...)
	services.ApplyConfig(m.Name(), e.ServiceConfig, m.confer, msg)

	for _, cc := range e.Cc {
		msg.AddCC(cc.String())
	}
	for k, v := range e.Headers {
		msg.AddHeader(k, v)
	}
	if strings.TrimSpace(e.Html) != "" {
		msg.SetHTML(e.Html)
	}

	for _, a := range e.Attachments {
		b, err := base64.StdEncoding.DecodeString(a.Content)
		if err != nil {
			return nil, fmt.Errorf("mailgun: failed to decode attachment: %w", err)
		}
		msg.AddBufferAttachment(a.Name, b)
	}

	resp, err := m.client.Send(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("mailgun: failed to send email: %w", err)
	}
	if resp.ID == "" {
		return nil, fmt.Errorf("mailgun: failed to send email: %s", resp.Message)
	}
	return []mmailer.Response{
		{
			Service: m.Name(),

			// We get raw Message-Id header here, ex <1761578515891502624.8555910852031586141@strictlog.modfin.se>
			// but in the webhook, the MessageID field doesn't contain the angle brackets.
			MessageId: strings.Trim(resp.ID, "<>"),
		},
	}, nil
}

func (m *Mailgun) UnmarshalPosthook(body []byte) ([]mmailer.Posthook, error) {
	var webhook mtypes.WebhookPayload
	if err := jsoniter.Unmarshal(body, &webhook); err != nil {
		return nil, err
	}
	verified, err := m.client.VerifyWebhookSignature(webhook.Signature)
	if err != nil {
		return nil, fmt.Errorf("mailgun: failed to verify signature: %w", err)
	}
	if !verified {
		return nil, errors.New("mailgun: failed to verify signature")
	}
	event, err := events.ParseEvent(webhook.EventData)
	if err != nil {
		return nil, fmt.Errorf("mailgun: failed to parse event: %w", err)
	}

	b, _ := jsoniter.MarshalIndent(event, "", "  ")
	fmt.Println(string(b))

	h := mmailer.Posthook{
		Service:   m.Name(),
		EventId:   event.GetID(),
		Timestamp: event.GetTimestamp(),
	}

	switch e := event.(type) {
	case *events.Accepted:
		h.Event = mmailer.EventProcessed
		h.MessageId = e.Message.Headers.MessageID
		h.Email = e.Recipient

	case *events.Delivered:
		h.Event = mmailer.EventDelivered
		h.MessageId = e.Message.Headers.MessageID
		h.Email = e.Recipient

	case *events.Opened:
		h.Event = mmailer.EventOpen
		h.MessageId = e.Message.Headers.MessageID
		h.Email = e.Recipient

	case *events.Failed:
		switch e.Severity {
		case "permanent":
			h.Event = mmailer.EventBounce

		case "temporary":
			h.Event = mmailer.EventDeferred
			if e.Reason == "suppress-bounce" {
				h.Event = mmailer.EventDropped
			}
		}

		h.MessageId = e.Message.Headers.MessageID
		h.Email = e.Recipient
		h.Info = fmt.Sprintf("%s; %d; %s", e.Reason, e.DeliveryStatus.Code, e.DeliveryStatus.Description)
	default:
		logger.Warn(fmt.Sprintf("received unsupported webhook event: %s", h.Event))
		return nil, nil
	}

	return []mmailer.Posthook{h}, nil
}

type configurer struct{}

func (s configurer) SetIpPool(poolId string, message *mailgun.PlainMessage) {
	// TODO?
}

func (s configurer) DisableTracking(message *mailgun.PlainMessage) {
	message.SetTracking(false)
}
