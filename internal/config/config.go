package config

import (
	"strings"
	"sync"

	"github.com/caarlos0/env/v6"
	"github.com/modfin/henry/slicez"
	"github.com/modfin/mmailer/internal/logger"
)

type AppConfig struct {
	PublicURL   string `env:"PUBLIC_URL" envDefault:"http://example.com/path/to/mmailer"`
	APIKey      string `env:"API_KEY"`
	PosthookKey string `env:"POSTHOOK_KEY"`
	Metrics     bool   `env:"METRICS" envDefault:"true"`

	HttpInterface       string `env:"HTTP_IFACE" envDefault:":8081"`
	PublicHttpInterface string `env:"PUBLIC_HTTP_IFACE" envDefault:":8080"`

	FromDomainOverride string `env:"FROM_DOMAIN_OVERRIDE"`

	Services             []string `env:"SERVICES" envSeparator:"\n"`
	ServiceIpPoolConfig  []string `env:"SERVICE_IP_POOL_CONFIG" envSeparator:"\n"`
	ServiceDomainApiKeys []string `env:"SERVICE_DOMAIN_API_KEYS" envSeparator:"\n"`

	RetryStrategy  string `env:"RETRY_STRATEGY"`
	SelectStrategy string `env:"SELECT_STRATEGY"`

	PosthookForward string   `env:"POSTHOOK_FORWARD"`
	Environment     string   `env:"ENVIRONMENT" envDefault:"DEVELOPMENT"`
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

func (a *AppConfig) IsDev() bool {
	return a.Environment == "DEVELOPMENT"
}

func (a *AppConfig) GetServiceIpPoolConfig(service string) []string {
	filteredPoolConfigs := slicez.Filter(a.ServiceIpPoolConfig, func(s string) bool {
		return strings.HasPrefix(s, service)
	})
	var poolNames []string
	for _, poolCfg := range filteredPoolConfigs {
		cfgParts := strings.Split(poolCfg, ":")
		if len(cfgParts) == 2 {
			poolNames = append(poolNames, cfgParts[1])
		}
	}
	return poolNames
}
