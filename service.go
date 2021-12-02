package mmailer

import (
	"context"
)

type Service interface {
	Name() string
	Send(ctx context.Context, email Email) (res []Response, err error)
	UnmarshalPosthook(body []byte) ([]Posthook, error)
}
