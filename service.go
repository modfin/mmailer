package mmailer

import (
	"context"
	"net/mail"
	"strings"

	"github.com/modfin/henry/slicez"
)

type Service interface {
	Name() string
	CanSend(email Email) bool
	Send(ctx context.Context, email Email) (res []Response, err error)
	UnmarshalPosthook(body []byte) ([]Posthook, error)
}

type ServiceApiKey struct {
	Service string
	ApiKey
}

type ApiKey struct {
	Domain string
	Key    string
	Props  map[string]string
}

const ApiKeyAnyDomain = ""

func KeyByEmailDomain(apiKeys []ApiKey, emailFrom string) (ApiKey, bool) {
	domain := ""
	if from, err := mail.ParseAddress(emailFrom); err == nil {
		parts := strings.Split(from.Address, "@")
		if len(parts) == 2 {
			d := strings.ToLower(strings.TrimSpace(parts[1]))
			if d != "" {
				domain = d
			}
		}
	}
	domainKey, ok := slicez.Find(apiKeys, func(e ApiKey) bool {
		return domain != "" && e.Domain == domain
	})
	if ok {
		return domainKey, true
	}
	return slicez.Find(apiKeys, func(e ApiKey) bool {
		return e.Domain == ApiKeyAnyDomain
	})
}
