package svc

import (
	"context"
	"errors"
	"fmt"
	"github.com/modfin/mmailer"
)

func RetryEach(ctx context.Context, s mmailer.Service, e mmailer.Email, services []mmailer.Service) (res []mmailer.Response, err error) {
	res, err = s.Send(ctx, e)
	if err == nil {
		return res, nil
	}

	var acc string = err.Error()
	for _, ss := range services {
		res, err = ss.Send(ctx, e)
		if err == nil {
			return res, nil
		}
		acc = fmt.Sprintf("%s: %s", err.Error(), acc)
	}
	return nil, errors.New(acc)
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
