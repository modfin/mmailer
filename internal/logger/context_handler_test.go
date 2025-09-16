package logger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"log/slog"
)

func TestAddToLogContext_NewContextWithAttr(t *testing.T) {
	ctx := context.Background()
	attr := slog.Attr{Key: "key1", Value: slog.StringValue("value1")}

	newCtx := AddToLogContext(ctx, attr.Key, attr.Value)
	attrs, ok := newCtx.Value(logFields).([]slog.Attr)

	assert.True(t, ok)
	assert.Len(t, attrs, 1)
	assert.Equal(t, attr, attrs[0])
}

func TestAddToLogContext_UpdateExistingAttr(t *testing.T) {
	ctx := context.WithValue(context.Background(), logFields, []slog.Attr{
		{
			Key:   "key1",
			Value: slog.StringValue("value1"),
		},
		{
			Key:   "key2",
			Value: slog.StringValue("value2"),
		},
	})

	newCtx := AddToLogContext(ctx, "key1", "new_value")
	attrs, ok := newCtx.Value(logFields).([]slog.Attr)

	expected := []slog.Attr{
		{
			Key:   "key1",
			Value: slog.StringValue("new_value"),
		},
		{
			Key:   "key2",
			Value: slog.StringValue("value2"),
		},
	}

	assert.True(t, ok)
	assert.Len(t, attrs, 2)
	assert.Equal(t, expected, attrs)
}

func TestAddToLogContext_AddNewAttrToExistingContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), logFields, []slog.Attr{{Key: "key1", Value: slog.StringValue("value1")}})
	attr := slog.Attr{Key: "key2", Value: slog.StringValue("value2")}

	newCtx := AddToLogContext(ctx, attr.Key, attr.Value)
	attrs, ok := newCtx.Value(logFields).([]slog.Attr)

	assert.True(t, ok)
	assert.Len(t, attrs, 2)
	assert.Equal(t, slog.Attr{Key: "key1", Value: slog.StringValue("value1")}, attrs[0])
	assert.Equal(t, attr, attrs[1])
}

func TestAddToLogContext_NilParentContext(t *testing.T) {
	attr := slog.Attr{Key: "key1", Value: slog.StringValue("value1")}

	newCtx := AddToLogContext(nil, attr.Key, attr.Value)
	attrs, ok := newCtx.Value(logFields).([]slog.Attr)

	assert.True(t, ok)
	assert.Len(t, attrs, 1)
	assert.Equal(t, attr, attrs[0])
}
