package mmailer

import (
	"context"
	"errors"
	"fmt"
	"github.com/modfin/mmailer/internal/logger"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
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
	if len(f.Services) == 0 {
		return nil, errors.New("facade no services to use")
	}

	var service Service

	// If service is specified
	if len(preferredService) > 0 {
		preferredService = strings.ToLower(preferredService)
		for _, s := range f.Services {
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
		service = strategy(f.Services)
	}

	if service == nil {
		return nil, errors.New("selected service does not have a mailer associated with it")
	}

	retry := f.Retry
	if retry == nil {
		retry = RetryNone
	}

	ctx = logger.AddToLogContext(ctx, "service", service.Name())
	logger.InfoCtx(ctx, fmt.Sprintf("Sending mail to %v through %s at [%v]", email.To, service.Name(), time.Now().String()))
	return retry(ctx, service, email, f.Services)
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
