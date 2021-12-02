package svc

import (
	"context"
	"mfn/mmailer"
	"mfn/mmailer/internal/tracing"
)

func WithTracing(service mmailer.Service) mmailer.Service {
	return &tracingService{
		service,
	}
}

type tracingService struct {
	mmailer.Service
}

func (m *tracingService) Send(ctx context.Context, email mmailer.Email) (res []mmailer.Response, err error) {
	var span *tracing.Span
	name := "Mail transfer: " + m.Name()
	ctx, span = tracing.Start(ctx, name)
	defer span.Done()

	return m.Service.Send(ctx, email)
}
