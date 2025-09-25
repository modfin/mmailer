package mailjet

import (
	"context"
	"encoding/json"
	"fmt"
	mj "github.com/mailjet/mailjet-apiv3-go/v3"
	"github.com/modfin/mmailer"
	"github.com/modfin/mmailer/services"
	"strings"
)

type Mailjet struct {
	apiKeyPublic  string
	apiKeyPrivate string
	confer        services.Configurer[*mj.MessagesV31]
}

var bannedHeaders = map[string]struct{}{
	"from":                          {},
	"sender":                        {},
	"subject":                       {},
	"to":                            {},
	"cc":                            {},
	"bcc":                           {},
	"return-path":                   {},
	"delivered-to":                  {},
	"dkim-signature":                {},
	"domainkey-status":              {},
	"received-spf":                  {},
	"authentication-results":        {},
	"received":                      {},
	"x-mailjet-prio":                {},
	"x-mailjet-debug":               {},
	"user-agent":                    {},
	"x-mailer":                      {},
	"x-mj-customid":                 {},
	"x-mj-eventpayload":             {},
	"x-mj-vars":                     {},
	"x-mj-templateerrordeliver":     {},
	"x-mj-templateerrorreporting":   {},
	"x-mj-templatelanguage":         {},
	"x-mailjet-trackopen":           {},
	"x-mailjet-trackclick":          {},
	"x-mj-templateid":               {},
	"x-mj-workflowid":               {},
	"x-feedback-id":                 {},
	"x-mailjet-segmentation":        {},
	"list-id":                       {},
	"x-mj-mid":                      {},
	"x-mj-errormessage":             {},
	"date":                          {},
	"x-csa-complaints":              {},
	"message-id":                    {},
	"x-mailjet-campaign":            {},
	"x-mj-statisticscontactslistid": {},
}

func (m *Mailjet) newClient() *mj.Client {
	return mj.NewMailjetClient(m.apiKeyPublic, m.apiKeyPrivate)
}

func New(apiKeyPublic, apiKeyPrivate string) *Mailjet {
	return &Mailjet{
		apiKeyPublic:  apiKeyPublic,
		apiKeyPrivate: apiKeyPrivate,
		confer:        MailjetConfigurer{},
	}
}
func (m *Mailjet) Name() string {
	return "mailjet"
}

func (m *Mailjet) Send(_ context.Context, email mmailer.Email) (res []mmailer.Response, err error) {
	message := mj.InfoMessagesV31{
		Headers: map[string]interface{}{},
		From: &mj.RecipientV31{
			Email: email.From.Email,
			Name:  email.From.Name,
		},
		Subject:  email.Subject,
		TextPart: email.Text,
		HTMLPart: email.Html,
	}

	for k, v := range email.Headers {
		_, exists := bannedHeaders[strings.ToLower(k)]
		if exists {
			continue
		}
		message.Headers[k] = v
	}

	var to mj.RecipientsV31
	for _, a := range email.To {
		to = append(to, mj.RecipientV31{
			Email: a.Email,
			Name:  a.Name,
		})
	}

	var cc mj.RecipientsV31
	for _, a := range email.Cc {
		cc = append(cc, mj.RecipientV31{
			Email: a.Email,
			Name:  a.Name,
		})
	}
	message.To = &to
	message.Cc = &cc

	messages := mj.MessagesV31{Info: []mj.InfoMessagesV31{message}}

	services.ApplyConfig(m.Name(), email.ServiceConfig, m.confer, &messages)

	response, err := m.newClient().SendMailV31(&messages)
	if err != nil {
		return nil, fmt.Errorf("%s: %s", m.Name(), err)
	}

	for _, rr := range response.ResultsV31 {
		for _, r := range rr.To {
			res = append(res, mmailer.Response{
				Service:   m.Name(),
				MessageId: r.MessageUUID,
				Email:     r.Email,
			})
		}
	}

	return res, nil

}

type posthook struct {
	Event          string `json:"event"`
	Time           int    `json:"time"`
	MessageID      int64  `json:"MessageID"`
	MessageGUID    string `json:"Message_GUID"`
	Email          string `json:"email"`
	MjCampaignID   int    `json:"mj_campaign_id"`
	MjContactID    int64  `json:"mj_contact_id"`
	Customcampaign string `json:"customcampaign"`
	IP             string `json:"ip"`
	Geo            string `json:"geo"`
	Agent          string `json:"agent"`
	CustomID       string `json:"CustomID"`
	Payload        string `json:"Payload"`
	Comment        string `json:"comment"`
	Error          string `json:"error"`
	Source         string `json:"source"`
}

func (m *Mailjet) UnmarshalPosthook(body []byte) ([]mmailer.Posthook, error) {
	var hooks []posthook
	// With mailjet you can select not to group events.
	if len(body) > 0 && body[0] == '{' {
		body = append(body, ']')
		body = append([]byte{'['}, body...)
	}

	err := json.Unmarshal(body, &hooks)
	if err != nil {
		return nil, err
	}
	var res []mmailer.Posthook
	for _, h := range hooks {
		if h.MessageGUID == "" {
			continue
		}

		var event mmailer.PosthookEvent
		var info string
		switch strings.ToLower(h.Event) {
		case "sent":
			event = mmailer.EventDelivered
		case "open":
			event = mmailer.EventOpen
		case "click":
			event = mmailer.EventClick
		case "bounce":
			event = mmailer.EventBounce
			info = h.Error + "; " + h.Comment
		case "blocked":
			event = mmailer.EventDropped
			info = h.Error
		case "spam":
			event = mmailer.EventSpam
			info = h.Source
		case "unsub":
			event = mmailer.EventUnsubscribe
		default:
			event = "unknown"
		}

		res = append(res, mmailer.Posthook{
			Service:   m.Name(),
			MessageId: h.MessageGUID,
			Email:     h.Email,
			Event:     event,
			Info:      info,
		})
	}
	return res, nil
}

type MailjetConfigurer struct{}

func (s MailjetConfigurer) SetIpPool(poolId string, message *mj.MessagesV31) {
	// no op
}

func (s MailjetConfigurer) DisableTracking(message *mj.MessagesV31) {}
