package services

import (
	"fmt"
	"github.com/modfin/henry/slicez"
	"github.com/modfin/mmailer"
	"github.com/modfin/mmailer/internal/logger"
)

type Configurer[T any] interface {
	SetIpPool(poolId string, message T)
	DisableTracking(message T)
}

func ApplyConfig[T any](service string, conf []mmailer.ConfigItem, configurer Configurer[T], m T) {
	logger.Info("Applying config")
	conf = slicez.Filter(conf, func(ci mmailer.ConfigItem) bool {
		return ci.Service == "" || ci.Service == service
	})

	for _, c := range conf {
		switch c.Key {
		case mmailer.IpPool:
			logger.Info(fmt.Sprintf("applying IpPool: %s", c.Value))
			configurer.SetIpPool(c.Value, m)
		case mmailer.Vendor:
			// no op, maybe we should just remove this item in mmailerd when we read it
		case mmailer.DisableTracking:
			logger.Info("disabling tracking")
			configurer.DisableTracking(m)
		default:
			logger.Warn(fmt.Sprintf("skipping bad config key: %s", c.Key))
		}
	}
}
