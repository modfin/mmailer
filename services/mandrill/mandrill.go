package mandrill

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/keighl/mandrill"
	"github.com/modfin/mmailer"
	"github.com/modfin/mmailer/services"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Mandrill struct {
	apiKey string
	confer services.Configurer[*mandrill.Message]
}

var httpClient = http.Client{Timeout: 60 * time.Second}

func (m *Mandrill) newClient() *mandrill.Client {
	c := mandrill.ClientWithKey(m.apiKey)
	c.HTTPClient = &httpClient
	return c
}

func New(apiKey string) *Mandrill {
	return &Mandrill{
		apiKey: apiKey,
		confer: MandrillConfigurer{},
	}
}

func (m *Mandrill) Name() string {
	return "mandrill"
}

func (m *Mandrill) Send(_ context.Context, email mmailer.Email) (res []mmailer.Response, err error) {
	message := &mandrill.Message{}

	services.ApplyConfig(m.Name(), email.ServiceConfig, m.confer, message)

	for _, a := range email.To {
		message.AddRecipient(a.Email, a.Name, "to")
	}

	for _, a := range email.Cc {
		message.AddRecipient(a.Email, a.Name, "cc")
	}

	if len(email.Attachments) > 0 {
		for _, a := range email.Attachments {
			message.Attachments = append(message.Attachments, &mandrill.Attachment{
				Name:    a.Name,
				Content: a.Content,
				Type:    a.ContentType,
			})
		}
	}

	message.Headers = email.Headers
	message.FromName = email.From.Name
	message.FromEmail = email.From.Email
	message.Subject = email.Subject
	message.Text = email.Text
	message.HTML = email.Html
	message.Async = true

	responses, err := m.newClient().MessagesSend(message)
	if err != nil {
		return nil, fmt.Errorf("%s: %s", m.Name(), err)
	}

	for _, r := range responses {
		res = append(res, mmailer.Response{
			Service:   m.Name(),
			MessageId: r.Id,
			Email:     r.Email,
		})
	}
	return res, nil

}

type posthook struct {
	ID    string `json:"_id,omitempty"`
	Event string `json:"event,omitempty"`
	Ts    int    `json:"ts"`
	Msg   struct {
		ID       string        `json:"_id"`
		Version  string        `json:"_version"`
		Clicks   []interface{} `json:"clicks"`
		Email    string        `json:"email"`
		Metadata struct {
			UserID int `json:"user_id"`
		} `json:"metadata"`
		Opens      []interface{} `json:"opens"`
		Sender     string        `json:"sender"`
		SMTPEvents []struct {
			DestinationIP string `json:"destination_ip"`
			Diag          string `json:"diag"`
			Size          int    `json:"size"`
			SourceIP      string `json:"source_ip"`
			Ts            int    `json:"ts"`
			Type          string `json:"type"`
		} `json:"smtp_events"`
		State             string   `json:"state"`
		Subject           string   `json:"subject"`
		Tags              []string `json:"tags"`
		Ts                int      `json:"ts"`
		BgtoolsCode       int      `json:"bgtools_code"`
		BounceDescription string   `json:"bounce_description"`
		Diag              string   `json:"diag"`
	} `json:"msg,omitempty"`
	IP       string `json:"ip,omitempty"`
	Location struct {
		City         string  `json:"city"`
		Country      string  `json:"country"`
		CountryShort string  `json:"country_short"`
		Latitude     float64 `json:"latitude"`
		Longitude    float64 `json:"longitude"`
		PostalCode   string  `json:"postal_code"`
		Region       string  `json:"region"`
		Timezone     string  `json:"timezone"`
	} `json:"location,omitempty"`
	UserAgent       string `json:"user_agent,omitempty"`
	UserAgentParsed struct {
		Mobile       bool   `json:"mobile"`
		OsCompany    string `json:"os_company"`
		OsCompanyURL string `json:"os_company_url"`
		OsFamily     string `json:"os_family"`
		OsIcon       string `json:"os_icon"`
		OsName       string `json:"os_name"`
		OsURL        string `json:"os_url"`
		Type         string `json:"type"`
		UaCompany    string `json:"ua_company"`
		UaCompanyURL string `json:"ua_company_url"`
		UaFamily     string `json:"ua_family"`
		UaIcon       string `json:"ua_icon"`
		UaName       string `json:"ua_name"`
		UaURL        string `json:"ua_url"`
		UaVersion    string `json:"ua_version"`
	} `json:"user_agent_parsed,omitempty"`
	URL    string `json:"url,omitempty"`
	Action string `json:"action,omitempty"`
	Reject struct {
		CreatedAt   string `json:"created_at"`
		Detail      string `json:"detail"`
		Email       string `json:"email"`
		Expired     bool   `json:"expired"`
		ExpiresAt   string `json:"expires_at"`
		LastEventAt string `json:"last_event_at"`
		Reason      string `json:"reason"`
		Sender      string `json:"sender"`
		Subaccount  string `json:"subaccount"`
	} `json:"reject,omitempty"`
	Type  string `json:"type,omitempty"`
	Entry struct {
		CreatedAt string `json:"created_at"`
		Detail    string `json:"detail"`
		Email     string `json:"email"`
	} `json:"entry,omitempty"`
}

func (m *Mandrill) UnmarshalPosthook(body []byte) ([]mmailer.Posthook, error) {

	vals, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, err
	}
	jsonString := vals.Get("mandrill_events")

	var hooks []posthook
	err = json.Unmarshal([]byte(jsonString), &hooks)
	if err != nil {
		return nil, err
	}

	var res []mmailer.Posthook
	for _, h := range hooks {
		if h.ID == "" {
			continue
		}

		var event mmailer.PosthookEvent
		var info string
		switch strings.ToLower(h.Event) {
		case "send":
			event = mmailer.EventDelivered
		case "deferral":
			event = mmailer.EventDeferred
			for _, e := range h.Msg.SMTPEvents {
				info += e.Diag + ";"
			}
		case "hard_bounce":
			event = mmailer.EventBounce
			info = "hard_bounce; " + h.Msg.BounceDescription

		case "soft_bounce": // https://mailchimp.com/help/soft-vs-hard-bounces/
			// Mandrill will convert soft bounces to hard if soft bounces persist over time
			event = mmailer.EventDeferred
			info = "deferral; soft_bounce; " + h.Msg.BounceDescription
		case "open":
			event = mmailer.EventOpen
		case "click":
			event = mmailer.EventClick
		case "spam":
			event = mmailer.EventSpam
		case "unsub":
			event = mmailer.EventUnsubscribe
		case "reject":
			event = mmailer.EventDropped
		default:
			event = mmailer.EventUnknown
			info = h.Event
		}
		res = append(res, mmailer.Posthook{
			Service:   m.Name(),
			MessageId: h.ID,
			Email:     h.Msg.Email,
			Event:     event,
			Info:      info,
		})
	}

	return res, nil
}

type MandrillConfigurer struct{}

func (s MandrillConfigurer) SetIpPool(poolId string, message *mandrill.Message) {
	// no op
}
