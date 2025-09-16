package services

import (
	"fmt"
	"github.com/modfin/henry/slicez"
	"github.com/modfin/mmailer"
)

type Configurer[T any] interface {
	SetIpPool(poolId string, message T)
}

func ApplyConfig[T any](service string, conf []mmailer.ConfigItem, configurer Configurer[T], m T) {
	fmt.Println("[APPLYING CONFIG]")
	conf = slicez.Filter(conf, func(ci mmailer.ConfigItem) bool {
		return ci.Service == "" || ci.Service == service
	})

	for _, c := range conf {
		switch c.Key {
		case mmailer.IpPool:
			fmt.Println("applying IpPool: ", c.Value)
			configurer.SetIpPool(c.Value, m)
		case mmailer.Vendor:
			// no op, maybe we should just remove this item in mmailerd when we read it
		default:
			fmt.Println("[SKIPPING BAD CONFIG KEY]", c.Key)
		}
	}
}
