package mailgun

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

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
	apiKeys           []mmailer.ApiKey
	webhookSigningKey string
	confer            services.Configurer[*mailgun.PlainMessage]
}

func New(apiKeys []mmailer.ApiKey, webhookSigningKey string) *Mailgun {
	mg := &Mailgun{
		apiKeys:           apiKeys,
		webhookSigningKey: webhookSigningKey,
		confer:            configurer{},
	}
	return mg
}

func (m *Mailgun) Name() string {
	return "mailgun"
}

func (m *Mailgun) CanSend(e mmailer.Email) bool {
	for _, a := range e.Attachments {
		// TODO Can't find the option to set attachment content type in the mailgun api, can it be fixed?
		if a.ContentType != "" && a.ContentType != "application/octet-stream" {
			logger.Warn(
				"mailgun: unsupported attachment content-type",
				"email", e.To,
				"content_type", a.ContentType,
				"filename", a.Name,
			)
			return false
		}
	}
	_, ok := mmailer.KeyByEmailDomain(m.apiKeys, e.From.Email)
	return ok
}

func (m *Mailgun) newClient(addr string) (*mailgun.Client, error) {
	k, ok := mmailer.KeyByEmailDomain(m.apiKeys, addr)
	if !ok {
		return nil, errors.New("mailgun: no api key found for " + addr)
	}
	client := mailgun.NewMailgun(k.Key)
	if k.Props != nil && k.Props["region"] == "eu" {
		err := client.SetAPIBase(mailgun.APIBaseEU)
		if err != nil {
			return nil, fmt.Errorf("failed to set EU region")
		}
	}
	return client, nil
}

func (m *Mailgun) Send(ctx context.Context, e mmailer.Email) ([]mmailer.Response, error) {
	from, err := mail.ParseAddress(e.From.String())
	if err != nil {
		return nil, fmt.Errorf("mailgun: failed to parse email: %w", err)
	}
	client, err := m.newClient(from.Address)
	if err != nil {
		return nil, fmt.Errorf("mailgun: failed to create client: %w", err)
	}
	to := slicez.Map(e.To, func(a mmailer.Address) string {
		return a.String()
	})
	parts := strings.Split(from.Address, "@")
	domain, _ := slicez.Last(parts)
	if domain == "" {
		return nil, fmt.Errorf("mailgun: failed to get email domain: %v", from.Address)
	}
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

	resp, err := client.Send(ctx, msg)
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
	client := mailgun.NewMailgun("") // api key is not used for VerifyWebhookSignature
	client.SetWebhookSigningKey(m.webhookSigningKey)
	verified, err := client.VerifyWebhookSignature(webhook.Signature)
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
		h.Info = infoString(false, "", "", e.DeliveryStatus)

	case *events.Opened:
		h.Event = mmailer.EventOpen
		h.MessageId = e.Message.Headers.MessageID
		h.Email = e.Recipient

	case *events.Failed:
		switch e.Severity {
		case "permanent":
			h.Event = mmailer.EventBounce
			if e.Reason == "suppress-bounce" {
				h.Event = mmailer.EventDropped
			}
		case "temporary":
			h.Event = mmailer.EventDeferred
		}
		h.MessageId = e.Message.Headers.MessageID
		h.Email = e.Recipient
		h.Info = infoString(true, e.Reason, e.Severity, e.DeliveryStatus)

	default:
		logger.Warn(fmt.Sprintf("received unsupported webhook event: %s", h.Event))
		return nil, nil
	}

	return []mmailer.Posthook{h}, nil
}

func infoString(fail bool, reason, severity string, st events.DeliveryStatus) string {
	latency := time.Duration(st.SessionSeconds * float64(time.Second)).Truncate(time.Millisecond)

	var words []string
	if st.Code != 0 {
		words = append(words, fmt.Sprintf("%d", st.Code))
	}
	if st.EnhancedCode != "" {
		words = append(words, st.EnhancedCode)
	}
	if !fail {
		words = append(words, st.Message)
	}
	if st.MxHost != "" {
		words = append(words, st.MxHost)
	}
	if latency > time.Millisecond {
		words = append(words, latency.String())
	}
	if reason != "" {
		words = append(words, reason)
	}
	if severity != "" {
		words = append(words, severity)
	}

	var flags []string
	if st.AttemptNo > 0 {
		flags = append(flags, fmt.Sprintf("attempt:%d", st.AttemptNo))
	}
	if st.Utf8 != nil && *st.Utf8 {
		flags = append(flags, "utf8")
	}
	if st.TLS != nil && *st.TLS {
		flags = append(flags, "tls")
	}
	if st.CertificateVerified != nil && *st.CertificateVerified {
		flags = append(flags, "certificate-verified")
	}
	if len(flags) > 0 {
		words = append(words, fmt.Sprintf("[%s]", strings.Join(flags, ", ")))
	}

	msg := strings.Join(words, " ")
	if fail {
		msg += ": " + st.Message
	}
	return msg
}

type configurer struct{}

func (s configurer) SetIpPool(poolId string, message *mailgun.PlainMessage) {
	// TODO?
}

func (s configurer) DisableTracking(message *mailgun.PlainMessage) {
	message.SetTracking(false) // untested
}
