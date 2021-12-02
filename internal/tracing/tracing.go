package tracing

import (
	"context"
	"encoding/json"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

type Span struct {
	hasErr bool
	ctx    context.Context
	span   opentracing.Span
}

func (s *Span) LogString(key, value string) {
	s.span.LogFields(log.String(key, value))
}
func (s *Span) LogInt(key string, value int) {
	s.span.LogFields(log.Int(key, value))
}
func (s *Span) LogFloat(key string, value float64) {
	s.span.LogFields(log.Float64(key, value))
}

func (s *Span) LogJson(key string, value interface{}) {
	data, _ := json.Marshal(value)
	s.LogString(key, string(data))
}
func (s *Span) LogAny(key string, value interface{}) {
	s.span.LogFields(log.Object(key, value))
}

func (s *Span) LogError(err error) {
	if err != nil && !s.hasErr {
		s.hasErr = true
		s.span = s.span.SetTag("error", true)
	}
	if err != nil {
		s.LogString("error", err.Error())
	}
}

func (s *Span) DoneWith(err error) {
	s.LogError(err)
	s.Done()
}

func (s *Span) Done() {
	if s.ctx.Err() != nil {
		s.span = s.span.SetTag("context", "canceled")
		s.LogString("context", s.ctx.Err().Error())
	}
	s.span.Finish()
}

func Start(parent context.Context, name string) (ctx context.Context, span *Span) {
	parentSpan := opentracing.SpanFromContext(parent)
	sp := opentracing.StartSpan(name,
		opentracing.ChildOf(parentSpan.Context()))
	sp.SetTag("name", name)
	parent = opentracing.ContextWithSpan(parent, sp)
	return parent, &Span{span: sp, ctx: parent}
}
