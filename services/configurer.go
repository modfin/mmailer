package services

import (
	"fmt"
	"github.com/modfin/henry/slicez"
	"github.com/modfin/mmailer"
	"math"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type Configurer[T any] interface {
	SetIpPool(poolId string, message T)
}

var warmupStartDate = "2025-09-14"

var (
	euLimiter     *rate.Limiter
	limiterMutex  sync.RWMutex
	lastCreatedAt time.Time
)

func refreshEuLimiter(now time.Time) *rate.Limiter {

	limiterMutex.Lock()
	defer limiterMutex.Unlock()

	fmt.Println("### REFRESH EU LIMITER BEGIN ###")

	// double-check after acquiring write lock
	if euLimiter == nil || now.Sub(lastCreatedAt) >= 24*time.Hour {
		start, err := time.ParseInLocation(time.DateOnly, warmupStartDate, time.UTC)
		if err != nil {
			fmt.Println("could not parse warmup start date", err)
			return nil
		}
		warmupDay := now.Sub(start).Hours() / 24

		// formula loosely based on schedule from:
		// https://www.twilio.com/docs/sendgrid/ui/sending-email/warming-up-an-ip-address#automated-ip-warmup-hourly-send-schedule
		// divided by 3, since we use 3 mmailer instances
		mailsPerHour := int(20*math.Pow(1.25, warmupDay)) / 3

		fmt.Println("[WARMUP_DAY]", warmupDay, "[MAILS_PER_HOUR]", mailsPerHour)
		euLimiter = rate.NewLimiter(rate.Limit(mailsPerHour)*rate.Every(time.Hour), mailsPerHour)
		lastCreatedAt = now
	}

	fmt.Println("### REFRESH EU LIMITER COMPLETE ###")

	return euLimiter
}

func allowEuIp() bool {
	now := time.Now()

	limiterMutex.RLock()
	defer limiterMutex.RUnlock()

	// check if we need to recreate the limiter (daily or first time)
	if euLimiter == nil || now.Sub(lastCreatedAt) >= 24*time.Hour {
		go refreshEuLimiter(now)
		// don't block while refreshing, but assume false until done
		return false
	}
	return euLimiter.Allow()
}

func ApplyConfig[T any](service string, conf []mmailer.ConfigItem, configurer Configurer[T], m T) {
	fmt.Println("[APPLYING CONFIG]")
	conf = slicez.Filter(conf, func(ci mmailer.ConfigItem) bool {
		return ci.Service == "" || ci.Service == service
	})

	for _, c := range conf {
		switch c.Key {
		case mmailer.IpPool:
			if allowEuIp() {
				fmt.Println("applying IpPool: ", c.Value)
				configurer.SetIpPool(c.Value, m)
			} else {
				fmt.Println("wanted to apply IpPool (blocked by warmup): ", c.Value)
			}
		case mmailer.Vendor:
			// no op, maybe we should just remove this item in mmailerd when we read it
		default:
			fmt.Println("[SKIPPING BAD CONFIG KEY]", c.Key)
		}
	}
}
