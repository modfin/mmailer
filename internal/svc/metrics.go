package svc

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"mfn/mmailer"
)

var mailSend = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: "mmailer",
	Subsystem: "service",
	Name:      "send_status_count",
	Help:      "The total number of emails sent",
}, []string{"name", "status"})

var mailSendTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "mmailer",
	Subsystem: "service",
	Name:      "send",
}, []string{"name"})

var mailPosthook = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: "mmailer",
	Subsystem: "service",
	Name:      "posthook_status_count",
	Help:      "The total number of emails sent",
}, []string{"name", "status"})

func WithMetric(service mmailer.Service) mmailer.Service {
	return &metricService{
		service,
	}
}

type metricService struct {
	mmailer.Service
}

func (m *metricService) Send(ctx context.Context, email mmailer.Email) (res []mmailer.Response, err error) {
	name := m.Name()
	timer := prometheus.NewTimer(mailSendTime.WithLabelValues(name))
	defer func() {
		timer.ObserveDuration()
		if err == nil {
			mailSend.WithLabelValues(name, "success").Inc()
			return
		}
		mailSend.WithLabelValues(name, "error").Inc()
	}()
	return m.Service.Send(ctx, email)
}
func (m *metricService) UnmarshalPosthook(body []byte) (p []mmailer.Posthook, err error) {
	name := m.Name()
	defer func() {
		if err == nil {
			mailPosthook.WithLabelValues(name, "success").Inc()
			return
		}
		mailPosthook.WithLabelValues(name, "error").Inc()
	}()

	return m.Service.UnmarshalPosthook(body)
}
