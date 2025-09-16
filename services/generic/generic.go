package generic

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/modfin/mmailer"
	"github.com/modfin/mmailer/internal/logger"
	"github.com/modfin/mmailer/internal/smtpx"
	"github.com/modfin/mmailer/services"
	"net/smtp"
	"net/url"
	"os"
	"strings"
)

// make generic implement mmailer.Service interface by implementing Name and Send methods
type Generic struct {
	smtpUrl *url.URL
	confer  services.Configurer[*smtpx.Message]
}

func New(smtpUrl *url.URL) *Generic {
	if smtpUrl.Port() == "" {
		smtpUrl.Host = smtpUrl.Host + ":25"
	}
	return &Generic{
		smtpUrl: smtpUrl,
		confer:  GenericConfigurer{},
	}
}

func (g *Generic) Name() string {
	return fmt.Sprintf("Generic smtp %s", g.smtpUrl.Host)
}

func (g *Generic) Send(ctx context.Context, email mmailer.Email) (res []mmailer.Response, err error) {
	message := smtpx.NewMessage()
	for k, v := range email.Headers {
		message.SetHeader(k, v)
	}

	message.SetHeader("From", email.From.String())
	ctx = logger.AddToLogContext(ctx, "from", email.From.String())

	var recp []string
	if len(email.To) > 0 {
		var tos []string
		for _, t := range email.To {
			tos = append(tos, t.String())
			recp = append(recp, t.Email)
		}
		message.SetHeader("To", strings.Join(tos, ", "))
	}
	if len(email.Cc) > 0 {
		var tos []string
		for _, t := range email.To {
			tos = append(tos, t.String())
			recp = append(recp, t.Email)
		}
		message.SetHeader("To", strings.Join(tos, ", "))
	}
	message.SetHeader("Subject", email.Subject)

	if len(email.Text) > 0 {
		message.SetBody("text/plain", email.Text)
	}
	if len(email.Html) > 0 {
		message.SetBody("text/html", email.Html)
	}

	var files []*os.File
	if len(email.Attachments) > 0 {
		// create a new temp file based on attachment content
		for _, a := range email.Attachments {
			// create a new temp file based on attachment content
			f, err := os.Create("/tmp/" + a.Name)
			if err != nil {
				logger.ErrorCtx(ctx, err, "could not create temp file")
				continue
			}

			b64decoded, err := base64.StdEncoding.DecodeString(a.Content)
			if err != nil {
				logger.ErrorCtx(ctx, err, "could not decode base64 content")
				continue
			}
			_, err = f.Write(b64decoded)
			if err != nil {
				logger.ErrorCtx(ctx, err, "could not write to temp file")
				continue
			}
			files = append(files, f)
			message.Attach(f.Name())
		}
	}

	defer func() {
		for _, f := range files {
			f := f
			err := f.Close()
			if err != nil {
				logger.ErrorCtx(ctx, err, "could not close temp file: "+f.Name())
			}
			err = os.Remove(f.Name())
			if err != nil {
				logger.ErrorCtx(ctx, err, "could not remove temp file: "+f.Name())
			}
		}
	}()

	var auth smtp.Auth = nil

	user := g.smtpUrl.User.Username()
	pass, ok := g.smtpUrl.User.Password()
	if ok {
		auth = smtp.CRAMMD5Auth(user, pass)
	}

	msgId := uuid.NewString()
	message.SetHeader("Message-ID", msgId)
	msg, err := message.Bytes()
	if err != nil {
		return nil, err
	}
	err = smtp.SendMail(g.smtpUrl.Host, auth, email.From.Email, recp, msg)
	if err != nil {
		return nil, err
	}

	var resps []mmailer.Response
	for _, e := range email.To {
		resps = append(resps, mmailer.Response{
			Service:   g.Name(),
			MessageId: msgId,
			Email:     e.Email,
		})
	}

	return resps, nil
}

func (m *Generic) UnmarshalPosthook(body []byte) ([]mmailer.Posthook, error) {
	return nil, errors.New("generic smtp does not have post hooks")
}

type GenericConfigurer struct{}

func (s GenericConfigurer) SetIpPool(poolId string, message *smtpx.Message) {
	// no op
}
