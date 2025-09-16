package svc

import (
	"context"
	"github.com/modfin/mmailer"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) Name() string {
	return "mock service"
}

func (m *MockService) UnmarshalPosthook(body []byte) ([]mmailer.Posthook, error) {
	args := m.Called(body)
	return args.Get(0).([]mmailer.Posthook), args.Error(1)
}

func (m *MockService) Send(ctx context.Context, email mmailer.Email) ([]mmailer.Response, error) {
	args := m.Called(ctx, email)
	return args.Get(0).([]mmailer.Response), args.Error(1)
}

func TestWithAllowListFilter_AllRecipientsAllowed(t *testing.T) {
	mockService := new(MockService)
	allowList := []string{"allowed@example.com", "@example.com"}
	filter := WithAllowListFilter(mockService, allowList)

	email := mmailer.Email{
		To: []mmailer.Address{
			{Email: "allowed@example.com"},
			{Email: "user@example.com"},
		},
	}

	mockService.On("Send", mock.Anything, email).Return([]mmailer.Response{}, nil)

	res, err := filter.Send(context.Background(), email)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	mockService.AssertCalled(t, "Send", mock.Anything, email)
}

func TestWithAllowListFilter_SomeRecipientsBlacklisted(t *testing.T) {
	mockService := new(MockService)
	allowList := []string{"allowed@example.com"}
	filter := WithAllowListFilter(mockService, allowList)

	email := mmailer.Email{
		To: []mmailer.Address{
			{Email: "allowed@example.com"},
			{Email: "blacklisted@example.com"},
		},
	}

	expectedEmail := mmailer.Email{
		To: []mmailer.Address{
			{Email: "allowed@example.com"},
		},
	}

	mockService.On("Send", mock.Anything, expectedEmail).Return([]mmailer.Response{}, nil)

	res, err := filter.Send(context.Background(), email)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	mockService.AssertCalled(t, "Send", mock.Anything, expectedEmail)
}

func TestWithAllowListFilter_AllRecipientsBlacklisted(t *testing.T) {
	mockService := new(MockService)
	allowList := []string{"allowed@example.com"}
	filter := WithAllowListFilter(mockService, allowList)

	email := mmailer.Email{
		To: []mmailer.Address{
			{Email: "blacklisted@example.com"},
		},
	}

	res, err := filter.Send(context.Background(), email)

	assert.NoError(t, err)
	assert.Empty(t, res)
	mockService.AssertNotCalled(t, "Send", mock.Anything, mock.Anything)
}

func TestWithAllowListFilter_EmptyAllowList(t *testing.T) {
	mockService := new(MockService)
	allowList := []string{}
	filter := WithAllowListFilter(mockService, allowList)

	email := mmailer.Email{
		To: []mmailer.Address{
			{Email: "anyone@example.com"},
		},
	}

	mockService.On("Send", mock.Anything, email).Return([]mmailer.Response{}, nil)

	res, err := filter.Send(context.Background(), email)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	mockService.AssertCalled(t, "Send", mock.Anything, email)
}
