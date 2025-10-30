package svc

import (
	"context"
	"errors"

	"github.com/modfin/mmailer"
	"github.com/modfin/mmailer/internal/logger"
)

func RetryEach(ctx context.Context, s mmailer.Service, e mmailer.Email, services []mmailer.Service) (res []mmailer.Response, err error) {
	res, err = s.Send(ctx, e)
	if err == nil {
		return res, nil
	}

	var errs []error
	for _, ss := range services {
		ctx := logger.AddToLogContext(ctx, "fallback_service", ss.Name())
		logger.WarnCtx(ctx, "err sending mail, retrying with fallback", "error", err)
		ctx = logger.AddToLogContext(ctx, "service", ss.Name())
		res, err = ss.Send(ctx, e)
		if err == nil {
			return res, nil
		}
		errs = append(errs, err)
	}
	return nil, errors.Join(err)
}

func RetryOneOther(ctx context.Context, s mmailer.Service, e mmailer.Email, services []mmailer.Service) (res []mmailer.Response, err error) {
	res, err = s.Send(ctx, e)
	if err == nil {
		return res, nil
	}
	for _, ss := range services {
		if s.Name() == ss.Name() {
			continue
		}
		ctx := logger.AddToLogContext(ctx, "fallback_service", ss.Name())
		logger.WarnCtx(ctx, "err sending mail, retrying with fallback", "error", err)
		ctx = logger.AddToLogContext(ctx, "service", ss.Name())
		return ss.Send(ctx, e)
	}
	return nil, err
}

func RetrySame(ctx context.Context, s mmailer.Service, e mmailer.Email, services []mmailer.Service) (res []mmailer.Response, err error) {
	res, err = s.Send(ctx, e)
	if err == nil {
		return res, nil
	}
	return s.Send(ctx, e)
}
