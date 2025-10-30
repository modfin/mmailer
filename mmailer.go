package mmailer

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/modfin/henry/slicez"
	"github.com/modfin/mmailer/internal/logger"
)

type Facade struct {
	Services  []Service
	Selecting SelectStrategy
	Retry     RetryStrategy
}

func New(selecting SelectStrategy, retry RetryStrategy, services ...Service) *Facade {
	return &Facade{
		Services:  services,
		Selecting: selecting,
		Retry:     retry,
	}
}

func (f *Facade) Send(ctx context.Context, email Email, preferredService string) (res []Response, err error) {

	services := slicez.Filter(f.Services, func(s Service) bool {
		return s.CanSend(email)
	})

	if len(services) == 0 {
		return nil, errors.New("facade no services to use")
	}

	var service Service

	// If service is specified
	if len(preferredService) > 0 {
		preferredService = strings.ToLower(preferredService)
		for _, s := range services {
			if s.Name() == preferredService {
				service = s
				break
			}
		}
	}

	// Regular selection strategy
	if service == nil {
		strategy := f.Selecting
		if strategy == nil {
			strategy = SelectRandom
		}
		service = strategy(services)
	}

	if service == nil {
		return nil, errors.New("selected service does not have a mailer associated with it")
	}

	retry := f.Retry
	if retry == nil {
		retry = RetryNone
	}

	to := slicez.Map(email.To, func(a Address) string {
		return a.Email
	})

	ctx = logger.AddToLogContext(ctx, "service", service.Name())
	ctx = logger.AddToLogContext(ctx, "addresses", to)
	logger.InfoCtx(ctx, "sending mail")
	return retry(ctx, service, email, services)
}

func (f *Facade) UnmarshalPosthook(r *http.Request) (res []Posthook, err error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	name := strings.ToLower(r.URL.Query().Get("service"))
	for _, s := range f.Services {
		if s.Name() == name {
			return s.UnmarshalPosthook(body)
		}
	}
	return nil, errors.New("could not find a service to unmarshal posthook to")
}
