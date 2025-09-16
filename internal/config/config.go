package config

import (
	"github.com/caarlos0/env/v6"
	"github.com/modfin/mmailer/internal/logger"
	"sync"
)

type AppConfig struct {
	PublicURL   string `env:"PUBLIC_URL" envDefault:"http://example.com/path/to/mmailer"`
	APIKey      string `env:"API_KEY"`
	PosthookKey string `env:"POSTHOOK_KEY"`
	Metrics     bool   `env:"METRICS" envDefault:"true"`

	HttpInterface       string `env:"HTTP_IFACE" envDefault:":8081"`
	PublicHttpInterface string `env:"PUBLIC_HTTP_IFACE" envDefault:":8080"`

	FromDomainOverride string `env:"FROM_DOMAIN_OVERRIDE"`

	Services []string `env:"SERVICES" envSeparator:"\n"`

	RetryStrategy  string `env:"RETRY_STRATEGY"`
	SelectStrategy string `env:"SELECT_STRATEGY"`

	PosthookForward string   `env:"POSTHOOK_FORWARD"`
	Enviroment      string   `env:"ENVIRONMENT" envDefault:"DEVELOPMENT"`
	AllowListFilter []string `env:"ALLOW_LIST" envSeparator:"," envDefault:"@modularfinance.se"`
}

var (
	once sync.Once
	cfg  AppConfig
)

func Get() *AppConfig {
	once.Do(func() {
		cfg = AppConfig{}
		if err := env.Parse(&cfg); err != nil {
			logger.Error(err, "Couldn't parse AppConfig from env")
			panic(err)
		}
	})
	return &cfg
}

func (a AppConfig) IsDev() bool {
	return a.Enviroment == "DEVELOPMENT"
}
