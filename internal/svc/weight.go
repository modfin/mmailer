package svc

import (
	"math/rand"
	"github.com/modfin/mmailer"
	"sync"
)

type weightService struct {
	mmailer.Service
	weight uint
}

// Eg. 2 services A with weight 9 and B with Weight 1
// for every 100 messages on average 90 will be sent to A and 10 to B
func WithWeight(weight uint, service mmailer.Service) mmailer.Service {
	return &weightService{
		Service: service,
		weight:  weight,
	}
}

func SelectRoundRobin() mmailer.SelectStrategy {
	var i int64
	var mu sync.Mutex
	return func(services []mmailer.Service) mmailer.Service {
		mu.Lock()
		defer mu.Unlock()
		defer func() { i += 1 }()
		l := len(services)
		return services[i%int64(l)]
	}
}

func SelectWeighted(services []mmailer.Service) mmailer.Service {
	var ws []*weightService
	var sum uint
	for _, s := range services {
		w, ok := s.(*weightService)
		if ok {
			sum += w.weight
			ws = append(ws, w)
		}
	}
	if len(ws) == 0 {
		return mmailer.SelectRandom(services)
	}
	rand.Shuffle(len(ws), func(i, j int) {
		ws[i], ws[j] = ws[j], ws[i]
	})

	r := int(rand.Int31n(int32(sum))) + 1

	for _, s := range ws {
		r -= int(s.weight)
		if r <= 0 {
			return s
		}
	}
	return ws[len(ws)-1]
}
