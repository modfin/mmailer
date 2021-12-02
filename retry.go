package mmailer

import (
	"context"
)

type RetryStrategy func(cxt context.Context, serviceToUse Service, email Email, backupServices []Service) (res []Response, err error)

func RetryNone(ctx context.Context, s Service, e Email, _ []Service) (res []Response, err error) {
	return s.Send(ctx, e)
}
