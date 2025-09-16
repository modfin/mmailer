package config

import (
	"github.com/caarlos0/env/v6"
	"log"
	"sync"
)

type AppConfig struct {
	PublicURL   string `env:"PUBLIC_URL" envDefault:"http://example.com/path/to/mmailer"`
	APIKey      string `env:"API_KEY"`
	PosthookKey string `env:"POSTHOOK_KEY"`
	Metrics     bool   `env:"METRICS" envDefault:"true"`

	HttpInterface string `env:"HTTP_IFACE" envDefault:":8080"`

	Services []string `env:"SERVICES" envSeparator:"\n"`

	RetryStrategy  string `env:"RETRY_STRATEGY"`
	SelectStrategy string `env:"SELECT_STRATEGY"`

	PosthookForward string `env:"POSTHOOK_FORWARD"`

	Whitelist []string `env:"WHITE_LIST" envSeparator:" "`
}

var (
	once sync.Once
	cfg  AppConfig
)

func Get() *AppConfig {
	once.Do(func() {
		cfg = AppConfig{}
		if err := env.Parse(&cfg); err != nil {
			log.Panic("Couldn't parse AppConfig from env: ", err)
		}
	})
	return &cfg
}
