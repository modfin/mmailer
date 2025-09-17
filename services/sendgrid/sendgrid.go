package sendgrid

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/modfin/mmailer"
	"github.com/modfin/mmailer/internal/logger"
	"github.com/modfin/mmailer/services"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"strings"
)

type Sendgrid struct {
	apiKey string
	confer services.Configurer[*mail.SGMailV3]
}

func (m *Sendgrid) newClient() *sendgrid.Client {
	return sendgrid.NewSendClient(m.apiKey)
}

func New(apiKey string) *Sendgrid {
	return &Sendgrid{
		apiKey: apiKey,
		confer: SendgridConfigurer{},
	}
}

func (m *Sendgrid) Name() string {
	return "sendgrid"
}

func (m *Sendgrid) Send(_ context.Context, email mmailer.Email) (res []mmailer.Response, err error) {
	from := mail.NewEmail(email.From.Name, email.From.Email)

	message := mail.NewSingleEmail(from, email.Subject, nil, email.Text, email.Html)

	// TODO: unsure about this one
	message.SetIPPoolID(pool_sg_us)

	services.ApplyConfig(m.Name(), email.ServiceConfig, m.confer, message)

	for k, v := range email.Headers {
		if k == "Reply-To" {
			message.SetReplyTo(&mail.Email{
				Address: v,
			})
		} else {
			message.SetHeader(k, v)
		}
	}

	if len(email.Attachments) > 0 {
		for _, a := range email.Attachments {
			message.AddAttachment(&mail.Attachment{
				Content:     a.Content,
				Filename:    a.Name,
				Type:        a.ContentType,
				Disposition: "attachment",
			})
		}
	}
	// Hm.. With multiple TO or CC, only one message id is returned corresponging to
	// Message-ID header. Which is reasonable, but make things hard to track.
	// Adding multiple personalization might be a better way, to have it act like other vendors.
	message.Personalizations[0].To = nil
	for _, a := range email.To {
		message.Personalizations[0].AddTos(&mail.Email{
			Name:    a.Name,
			Address: a.Email,
		})
	}
	for _, a := range email.Cc {
		message.Personalizations[0].AddCCs(&mail.Email{
			Name:    a.Name,
			Address: a.Email,
		})
	}

	response, err := m.newClient().Send(message)
	if err != nil {
		return nil, fmt.Errorf("%s: %s", m.Name(), err)
	}
	if response.StatusCode > 299 {
		return nil, fmt.Errorf("%s: %s", m.Name(), fmt.Errorf("%+v", response))
	}

	for _, id := range response.Headers["X-Message-Id"] {
		res = append(res, mmailer.Response{
			Service:   m.Name(),
			MessageId: id,
		})
	}

	return res, nil

}

type posthook struct {
	Email       string   `json:"email"`
	Timestamp   int      `json:"timestamp"`
	SMTPID      string   `json:"smtp-id"`
	Event       string   `json:"event"`
	Category    []string `json:"category"`
	SgEventID   string   `json:"sg_event_id"`
	SgMessageID string   `json:"sg_message_id"`
	Response    string   `json:"response,omitempty"`
	Attempt     string   `json:"attempt,omitempty"`
	Useragent   string   `json:"useragent,omitempty"`
	IP          string   `json:"ip,omitempty"`
	URL         string   `json:"url,omitempty"`
	Reason      string   `json:"reason,omitempty"`
	Status      string   `json:"status,omitempty"`
	AsmGroupID  int      `json:"asm_group_id,omitempty"`
}

func (m *Sendgrid) UnmarshalPosthook(body []byte) ([]mmailer.Posthook, error) {
	var hooks []posthook
	err := json.Unmarshal(body, &hooks)
	if err != nil {
		return nil, err
	}
	var res []mmailer.Posthook
	for _, h := range hooks {
		if h.SgMessageID == "" {
			continue
		}
		var event mmailer.PosthookEvent
		var info string
		switch strings.ToLower(h.Event) {
		case "delivered":
			event = mmailer.EventDelivered
			info = h.Response
		case "deferred":
			event = mmailer.EventDeferred
			info = h.Response
		case "open":
			event = mmailer.EventOpen
		case "click":
			event = mmailer.EventClick
		case "bounce":
			event = mmailer.EventBounce
			info = h.Reason
		case "dropped":
			event = mmailer.EventDropped
			info = h.Reason
		case "spamreport":
			event = mmailer.EventSpam
		case "unsubscribe":
			event = mmailer.EventUnsubscribe
		case "group_unsubscribe":
			event = mmailer.EventUnsubscribe
		default:
			logger.Warn(fmt.Sprintf("received unsupported webhook event: %s", h.Event))
			event = mmailer.EventUnknown
			info = h.Event

		}

		parts := strings.Split(h.SgMessageID, ".")
		messageId := parts[0]

		res = append(res, mmailer.Posthook{
			Service:   m.Name(),
			MessageId: messageId,
			Email:     h.Email,
			Event:     event,
			Info:      info,
		})
	}
	return res, nil
}

const pool_sg_us = "sg_us"
const pool_sg_eu = "sg_eu"

type SendgridConfigurer struct{}

func (s SendgridConfigurer) SetIpPool(poolId string, message *mail.SGMailV3) {
	if poolId == pool_sg_us || poolId == pool_sg_eu {
		fmt.Println("[temporary log]", "[ip pool]", poolId)
		message.SetIPPoolID(poolId)
		return
	}
}
