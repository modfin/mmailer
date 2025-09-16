package svc

import (
	"context"
	"fmt"
	"github.com/modfin/henry/slicez"
	"github.com/modfin/mmailer"
	"github.com/modfin/mmailer/internal/logger"
	"strings"
)

type allowListFilter struct {
	mmailer.Service
	allowListFilter []string
}

func WithAllowListFilter(service mmailer.Service, allowList []string) mmailer.Service {
	return &allowListFilter{
		service, allowList,
	}
}

func (a *allowListFilter) Send(ctx context.Context, email mmailer.Email) (res []mmailer.Response, err error) {

	var filteredRecipients []mmailer.Address
	var blacklistedRecipients []mmailer.Address
	if len(a.allowListFilter) > 0 {
		for _, to := range email.To {
			//to := to
			parts := strings.Split(to.Email, "@")
			allowedDomain := ""
			if len(parts) > 1 {
				allowedDomain = parts[1]
			}
			if slicez.Contains(a.allowListFilter, to.Email) || slicez.Contains(a.allowListFilter, fmt.Sprintf("@%s", allowedDomain)) {
				filteredRecipients = append(filteredRecipients, to)
			} else {
				blacklistedRecipients = append(blacklistedRecipients, to)
			}
		}
		if len(filteredRecipients) == 0 {
			logger.WarnCtx(ctx, fmt.Sprintf("No recipients left after allow list filter: %v", a.allowListFilter))
			return []mmailer.Response{}, nil
		}
		email.To = filteredRecipients
	}
	if len(blacklistedRecipients) > 0 {
		logger.InfoCtx(ctx, fmt.Sprintf("Will not send email to %d recipient(s), using allow list filter: %v", len(blacklistedRecipients), a.allowListFilter))
	}

	return a.Service.Send(ctx, email)
}
