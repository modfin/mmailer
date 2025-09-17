package brev

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/modfin/brev"
	"github.com/modfin/mmailer"
	"github.com/modfin/mmailer/services"
)

type Brev struct {
	client brev.Client
	confer services.Configurer[*brev.Email]
}

func New(configParts []string, posthookUrl string) (*Brev, error) {

	if len(configParts) < 2 {
		return nil, errors.New("brev: missing domain or key")
	}
	domain := configParts[0]
	key := configParts[1]
	configParts = configParts[2:]

	if len(configParts)%2 != 0 {
		return nil, errors.New("brev: invalid number of key parts")
	}

	var servers []brev.TargetServer
	for i := 0; i < len(configParts); i += 2 {
		c := brev.TargetServer{
			Host:    configParts[i],
			MxCNAME: configParts[i+1],
		}
		fmt.Printf(" - Brev: added config: %s (%s)\n", c.Host, c.MxCNAME)
		c.Host = "https://" + c.Host
		servers = append(servers, c)
	}
	if len(servers) == 0 {
		return nil, errors.New("brev: no servers")
	}

	b, err := brev.New(key, domain, posthookUrl, servers, 0)
	if err != nil {
		return nil, err
	}
	return &Brev{
		client: b,
		confer: BrevConfigurer{},
	}, nil
}

func (b *Brev) Name() string {
	return "brev"
}

func (b *Brev) Send(ctx context.Context, m mmailer.Email) (res []mmailer.Response, err error) {
	if b.client == nil {
		return nil, errors.New("brev: cant send, missing client")
	}
	bm := brev.NewEmail()
	bm.Subject = m.Subject
	bm.From = brev.Address{
		Name:  m.From.Name,
		Email: m.From.Email,
	}
	for _, t := range m.To {
		bm.To = append(bm.To, brev.Address{
			Name:  t.Name,
			Email: t.Email,
		})
	}
	for _, c := range m.Cc {
		bm.To = append(bm.Cc, brev.Address{
			Name:  c.Name,
			Email: c.Email,
		})
	}
	bm.HTML = m.Html
	bm.Text = m.Text
	for h, v := range m.Headers {
		bm.Headers[h] = []string{v}
	}

	if len(bm.Attachments) > 0 {
		for _, a := range m.Attachments {
			bm.Attachments = append(bm.Attachments, brev.Attachment{
				Filename:    a.Name,
				Content:     a.Content,
				ContentType: a.ContentType,
			})
		}
	}

	services.ApplyConfig(b.Name(), m.ServiceConfig, b.confer, bm)

	r, err := b.client.Send(ctx, bm)
	if err != nil {
		return nil, err
	}
	fmt.Println("[brev] got message_id:", r.MessageId)
	return []mmailer.Response{{
		Service:   b.Name(),
		MessageId: r.MessageId,
		//Email: TODO: ???
	}}, nil
}

func (b *Brev) UnmarshalPosthook(body []byte) ([]mmailer.Posthook, error) {
	var hook brev.Posthook
	err := json.Unmarshal(body, &hook)
	if err != nil {
		return nil, err
	}
	var ev mmailer.PosthookEvent
	switch hook.Event {
	case brev.EventDelivered:
		ev = mmailer.EventDelivered
	case brev.EventDeferred:
		ev = mmailer.EventDeferred
	case brev.EventBounce:
		ev = mmailer.EventBounce
	case brev.EventFailed: // TODO: when triggered?
		ev = mmailer.EventUnknown
	case brev.EventDropped:
		ev = mmailer.EventDropped
	case brev.EventSpam:
		ev = mmailer.EventSpam
	case brev.EventUnsubscribe:
		ev = mmailer.EventUnsubscribe
	default:
		ev = mmailer.EventUnknown
	}
	return []mmailer.Posthook{{
		Service:   b.Name(),
		MessageId: hook.MessageId,
		Email:     "",
		Event:     ev,
		Info:      hook.Info,
	}}, nil
}

type BrevConfigurer struct{}

func (s BrevConfigurer) SetIpPool(poolId string, message *brev.Email) {
	// no op
}
