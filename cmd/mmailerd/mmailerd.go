package main

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/modfin/mmailer"
	"github.com/modfin/mmailer/internal/config"
	"github.com/modfin/mmailer/internal/logger"
	"github.com/modfin/mmailer/internal/svc"
	"github.com/modfin/mmailer/services/brev"
	"github.com/modfin/mmailer/services/generic"
	"github.com/modfin/mmailer/services/mailjet"
	"github.com/modfin/mmailer/services/mandrill"
	"github.com/modfin/mmailer/services/sendgrid"
	"io/ioutil"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var facade *mmailer.Facade

func main() {
	handler := &logger.ContextHandler{
		slog.NewJSONHandler(os.Stdout, nil),
	}
	logger.InitializeLogger(slog.New(handler))
	loadServices()

	e := echo.New()
	ePub := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	if config.Get().Metrics {
		p := prometheus.NewPrometheus("echo", nil)
		p.Use(e)
	}

	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "mmailer pong")
	})

	e.POST("/send", func(c echo.Context) error {
		ctx := c.Request().Context()
		logger.InfoCtx(ctx, "Received send email request")
		key := c.QueryParam("key")
		if subtle.ConstantTimeCompare([]byte(key), []byte(config.Get().APIKey)) == 0 {
			return c.String(http.StatusUnauthorized, "not authorized")
		}

		b, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			logger.ErrorCtx(ctx, err, "could not read body")
			return c.String(http.StatusInternalServerError, "could not read body")
		}

		mail := mmailer.NewEmail()
		err = json.Unmarshal(b, &mail)
		if err != nil {
			logger.ErrorCtx(ctx, err, "could unmarshal json")
			return c.String(http.StatusInternalServerError, "could unmarshal json")
		}

		if len(strings.TrimSpace(config.Get().FromDomainOverride)) > 0 {
			parts := strings.Split(mail.From.Email, "@")
			if len(parts) != 2 {
				logger.WarnCtx(ctx, fmt.Sprintf("couldn't parse from-adress: %s", mail.From.Email))
				return c.String(http.StatusBadRequest, "couldn't parse from-adress")
			}
			parts[1] = strings.TrimSpace(config.Get().FromDomainOverride)
			mail.From.Email = strings.Join(parts, "@")
		}
		preferredService := c.QueryParam("X-Service")
		if len(preferredService) > 0 {
			ctx = logger.AddToLogContext(ctx, "preferredService", preferredService)
		}
		res, err := facade.Send(ctx, mail, c.Request().Header.Get("X-Service"))
		if err != nil {
			logger.ErrorCtx(ctx, err, "could not send email")
			return c.String(http.StatusInternalServerError, "could not send email")
		}
		return c.JSON(http.StatusOK, res)
	})

	ePub.POST("/posthook", func(c echo.Context) error {
		key := c.QueryParam("key")
		if subtle.ConstantTimeCompare([]byte(key), []byte(config.Get().PosthookKey)) == 0 {
			return c.String(http.StatusUnauthorized, "not authorized")
		}

		resp, err := facade.UnmarshalPosthook(c.Request())
		if err != nil {
			logger.Error(err, "could not unmarshal posthook")
			return c.String(http.StatusOK, "ok")
		}
		logger.Info(fmt.Sprintf("Posthook: %+v", resp))
		if len(config.Get().PosthookForward) == 0 {
			logger.Info("no forwarding posthook configured, ignoring")
			return c.String(http.StatusOK, "ok")
		}
		go func(hook []mmailer.Posthook) {
			data, err := json.Marshal(hook)
			if err != nil {
				logger.Error(err, "could not marshal json")
				return
			}

			buf := bytes.NewBuffer(data)
			resp, err := http.DefaultClient.Post(config.Get().PosthookForward, "application/json", buf)
			if err != nil {
				logger.Error(err, "could forward posthook")
				return
			}
			_ = resp.Body.Close()
		}(resp)

		return c.String(http.StatusOK, "ok")
	})

	logger.Info(fmt.Sprintf("Send mail by a HTTP POST %s/send?key=%s\n", config.Get().PublicURL, config.Get().APIKey))
	logger.Info(fmt.Sprintf("Posthooks will be forwarded to %s", config.Get().PosthookForward))
	logger.Info("Starting server on " + config.Get().HttpInterface)

	go start(ePub, config.Get().PublicHttpInterface)
	start(e, config.Get().HttpInterface)
	logger.Info("Terminating application")
}

func start(e *echo.Echo, address string) {
	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	go func() {
		if err := e.Start(address); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error(err, "Could not start server, shutting down...")
			os.Exit(1)
		}
	}()

	<-term
	logger.Info(fmt.Sprintf("Got kill signal, shutting down server %v ....", address))
	defer close(term)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = e.Shutdown(ctx)
	logger.Info(fmt.Sprintf("Shut down server %v ..", address))
}

func loadServices() {
	if len(config.Get().Services) == 0 {
		logger.Error(errors.New("no service has been provided"), "shutting down due to no service defined")
		os.Exit(1)
	}

	var strategyName = strings.ToLower(config.Get().SelectStrategy)
	var selects mmailer.SelectStrategy
	switch strategyName {
	case "weighted":
		logger.Info("Select Strategy: Weighted")
		selects = svc.SelectWeighted
	case "roundrobin":
		logger.Info("Select Strategy: RoundRobin")
		selects = svc.SelectRoundRobin()
	case "random":
		fallthrough
	default:
		logger.Info("Select Strategy: Random")
		selects = mmailer.SelectRandom
	}

	var retry mmailer.RetryStrategy
	switch strings.ToLower(config.Get().RetryStrategy) {
	case "oneother":
		logger.Info("Retry Strategy: OneOther")
		retry = svc.RetryOneOther
	case "each":
		logger.Info("Retry Strategy: Each")
		retry = svc.RetryEach
	case "same":
		logger.Info("Retry Strategy: Same")
		retry = svc.RetrySame
	case "none":
		fallthrough
	default:
		logger.Info("Retry Strategy: None")
		retry = mmailer.RetryNone
	}

	var services []mmailer.Service
	logger.Info("Services:")
	var weighted = strategyName == "weighted"
	for _, s := range config.Get().Services {
		parts := strings.Split(s, ":")

		var weight uint
		if weighted {
			var weightStr string
			weightStr, parts = parts[0], parts[1:]
			weightInt, err := strconv.ParseInt(weightStr, 10, 32)
			if err != nil {
				log.Fatal("could not pars service weight", err)
			}
			weight = uint(weightInt)
		}

		decorate := func(s mmailer.Service) mmailer.Service {
			if len(config.Get().AllowListFilter) > 0 {
				logger.Info(fmt.Sprintf("using allow list filter: %v", config.Get().AllowListFilter))
				s = svc.WithAllowListFilter(s, config.Get().AllowListFilter)
			}
			if config.Get().Metrics {
				s = svc.WithMetric(s)
			}
			if weighted {
				s = svc.WithWeight(weight, s)
			}
			return s
		}

		posthookUrl := fmt.Sprintf("%s/posthook?key=%s&service=%s", config.Get().PublicURL, config.Get().PosthookKey, strings.ToLower(parts[0]))

		switch strings.ToLower(parts[0]) {
		case "mailjet":
			if len(parts) != 3 {
				logger.Warn(fmt.Sprintf("mailjet api string is not valid, %s", s))
				continue
			}
			logger.Info(fmt.Sprintf(" -  Mailjet: add the following posthook url %s", posthookUrl))

			services = append(services, decorate(mailjet.New(parts[1], parts[2])))
		case "mandrill":
			if len(parts) != 2 {
				logger.Warn("mandrill api string is not valid,", s)
				continue
			}
			logger.Info(fmt.Sprintf(" - Mandrill: add the following posthook url %s", posthookUrl))
			services = append(services, decorate(mandrill.New(parts[1])))
		case "sendgrid":
			if len(parts) != 2 {
				logger.Warn("sendgrid api string is not valid,", s)
				continue
			}
			logger.Info(fmt.Sprintf(" - Sendgrid: add the following posthook url %s", posthookUrl))
			services = append(services, decorate(sendgrid.New(parts[1])))
		case "brev":
			brev, err := brev.New(parts[1:], posthookUrl)
			if err != nil {
				logger.Warn("brev api string is not valid,", s)
				continue
			}
			logger.Info(fmt.Sprintf(" - Brev: add the following posthook url %s", posthookUrl))
			services = append(services, decorate(brev))
		case "generic":
			u, err := url.Parse(strings.Join(parts[1:], ":"))
			if err != nil {
				logger.Info(fmt.Sprintf("[Err] could not parse url, %s, expected smtp://user:pass@host:port", parts[1]))
				continue
			}

			logger.Info(fmt.Sprintf(" - Generic: add the following posthook url %s", posthookUrl))
			services = append(services, decorate(generic.New(u)))
		}

	}

	if len(services) == 0 {
		log.Fatal("No valid services has to be provide")
	}

	facade = mmailer.New(selects, retry, services...)
}
