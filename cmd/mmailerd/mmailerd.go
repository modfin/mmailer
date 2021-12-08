package main

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/modfin/mmailer"
	"github.com/modfin/mmailer/internal/config"
	"github.com/modfin/mmailer/internal/svc"
	"github.com/modfin/mmailer/services/generic"
	"github.com/modfin/mmailer/services/mailjet"
	"github.com/modfin/mmailer/services/mandrill"
	"github.com/modfin/mmailer/services/sendgrid"
	"github.com/labstack/echo-contrib/jaegertracing"
	"github.com/labstack/echo-contrib/prometheus"
	"io/ioutil"
	"log"
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

	loadServices()

	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	if config.Get().Metrics {
		p := prometheus.NewPrometheus("echo", nil)
		p.Use(e)
	}

	if config.Get().Tracing {
		closer := jaegertracing.New(e, func(e echo.Context) bool {
			switch e.Path() {
			case "/ping":
				return true
			case "/metrics":
				return false
			}
			return false
		})
		defer closer.Close()
	}

	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	e.POST("/send", func(c echo.Context) error {
		ctx := c.Request().Context()
		key := c.QueryParam("key")
		if subtle.ConstantTimeCompare([]byte(key), []byte(config.Get().APIKey)) == 0 {
			return c.String(http.StatusUnauthorized, "not authorized")
		}

		b, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			log.Println("[err] ", err)
			return c.String(http.StatusInternalServerError, "could not read body")
		}

		mail := mmailer.NewEmail()
		err = json.Unmarshal(b, &mail)
		if err != nil {
			log.Println("[err] ", err)
			return c.String(http.StatusInternalServerError, "could unmarshal json")
		}

		res, err := facade.Send(ctx, mail, c.Request().Header.Get("X-Service"))
		if err != nil {
			log.Println("[err] ", err)
			return c.String(http.StatusInternalServerError, "could not send email")
		}
		return c.JSON(http.StatusOK, res)
	})

	e.POST("/posthook", func(c echo.Context) error {
		key := c.QueryParam("key")
		if subtle.ConstantTimeCompare([]byte(key), []byte(config.Get().PosthookKey)) == 0 {
			return c.String(http.StatusUnauthorized, "not authorized")
		}
		if len(config.Get().PosthookForward) == 0 {
			return c.String(http.StatusOK, "ok")
		}

		resp, err := facade.UnmarshalPosthook(c.Request())
		if err != nil {
			log.Println("[err] ", err)
			return c.String(http.StatusOK, "ok")
		}
		go func(hook []mmailer.Posthook) {
			data, err := json.Marshal(hook)
			if err != nil {
				log.Println("[err] could not marshal json, ", err)
				return
			}

			buf := bytes.NewBuffer(data)
			resp, err := http.DefaultClient.Post(config.Get().PosthookForward, "application/json", buf)
			if err != nil {
				log.Println("[err] could not post, ", err)
				return
			}
			_ = resp.Body.Close()
		}(resp)

		return c.String(http.StatusOK, "ok")
	})

	fmt.Println()
	fmt.Printf(">       Send mail by a HTTP POST %s/send?key=%s\n", config.Get().PublicURL, config.Get().APIKey)
	fmt.Printf("> Posthooks will be forwarded to %s\n\n", config.Get().PosthookForward)
	fmt.Println("Starting server on " + config.Get().HttpInterface)
	start(e)
}

func start(e *echo.Echo) {
	term := make(chan os.Signal)
	signal.Notify(term, syscall.SIGTERM)
	go func() {
		<-term
		defer close(term)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		_ = e.Shutdown(ctx)
	}()
	fmt.Println(e.Start(config.Get().HttpInterface))
	fmt.Println("Shutting down server")
	<-term
	fmt.Println("Terminating")
}

func loadServices() {
	if len(config.Get().Services) == 0 {
		log.Fatal("Services has to be provide")
	}

	var strategyName = strings.ToLower(config.Get().SelectStrategy)
	var selects mmailer.SelectStrategy
	fmt.Printf("Select Strategy: ")
	switch strategyName {
	case "weighted":
		fmt.Println("Weighted")
		selects = svc.SelectWeighted
	case "roundrobin":
		fmt.Println("RoundRobin")
		selects = svc.SelectRoundRobin()
	case "random":
		fallthrough
	default:
		fmt.Println("Random")
		selects = mmailer.SelectRandom
	}

	var retry mmailer.RetryStrategy
	fmt.Printf("Retry Strategy:  ")
	switch strings.ToLower(config.Get().RetryStrategy) {
	case "oneother":
		fmt.Println("OneOther")
		retry = svc.RetryOneOther
	case "each":
		fmt.Println("Each")
		retry = svc.RetryEach
	case "same":
		fmt.Println("Same")
		retry = svc.RetrySame
	case "none":
		fallthrough
	default:
		fmt.Println("None")
		retry = mmailer.RetryNone
	}

	var services []mmailer.Service
	fmt.Println("Services:")
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
			if config.Get().Tracing {
				s = svc.WithTracing(s)
			}
			if config.Get().Metrics {
				s = svc.WithMetric(s)
			}
			if weighted {
				s = svc.WithWeight(weight, s)
			}
			return s
		}

		switch strings.ToLower(parts[0]) {
		case "mailjet":
			if len(parts) != 3 {
				log.Println("mailjet api string is not valid,", s)
				continue
			}
			fmt.Printf(" -  Mailjet: add the following posthook url %s/posthook?key=%s&service=mailjet\n", config.Get().PublicURL, config.Get().PosthookKey)

			services = append(services, decorate(mailjet.New(parts[1], parts[2])))
		case "mandrill":
			if len(parts) != 2 {
				log.Println("mandrill api string is not valid,", s)
				continue
			}
			fmt.Printf(" - Mandrill: add the following posthook url %s/posthook?key=%s&service=mandrill\n", config.Get().PublicURL, config.Get().PosthookKey)
			services = append(services, decorate(mandrill.New(parts[1])))
		case "sendgrid":
			if len(parts) != 2 {
				log.Println("sendgrid api string is not valid,", s)
				continue
			}
			fmt.Printf(" - Sendgrid: add the following posthook url %s/posthook?key=%s&service=sendgrid\n", config.Get().PublicURL, config.Get().PosthookKey)
			services = append(services, decorate(sendgrid.New(parts[1])))
		case "generic":
			u, err := url.Parse(strings.Join(parts[1:], ":"))
			if err != nil{
				log.Println("[Err] could not parse url, ", parts[1], " expected smtp://user:pass@host.com:port" )
				continue
			}
			fmt.Printf(" - Generic: posthooks are not implmented, adding %s\n", u.String())
			services = append(services, decorate(generic.New(u)))
		}
	}

	if len(services) == 0 {
		log.Fatal("At least one valid service has to be provide")
	}

	facade = mmailer.New(selects, retry, services...)
}
