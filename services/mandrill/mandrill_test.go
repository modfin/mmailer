package mandrill

import (
	"reflect"
	"testing"

	"github.com/keighl/mandrill"
	"github.com/modfin/mmailer"
	"github.com/modfin/mmailer/services"
)

func TestMandrillConfigurer_ApplyConfig(t *testing.T) {
	t.Run("IP Pool No-op", func(t *testing.T) {
		message, originalMessage := setupMessage()

		configurer := MandrillConfigurer{}
		configItems := []mmailer.ConfigItem{
			{Service: "", Key: mmailer.IpPool, Value: "any_pool"},
		}

		services.ApplyConfig("mandrill", configItems, configurer, message)

		// Mandrill configurer is no-op, so we verify the entire message is unchanged
		if !reflect.DeepEqual(message, originalMessage) {
			t.Errorf("Mandrill message should remain unchanged after ApplyConfig")
		}
	})

	t.Run("Service-Specific Config No-op", func(t *testing.T) {
		message, originalMessage := setupMessage()

		configurer := MandrillConfigurer{}
		configItems := []mmailer.ConfigItem{
			{Service: "mandrill", Key: mmailer.IpPool, Value: "mandrill_pool"},
			{Service: "sendgrid", Key: mmailer.IpPool, Value: "sg_pool"}, // Should be ignored
		}

		services.ApplyConfig("mandrill", configItems, configurer, message)

		if !reflect.DeepEqual(message, originalMessage) {
			t.Errorf("Mandrill message should remain unchanged after service-specific config")
		}
	})

	t.Run("Vendor Config No-op", func(t *testing.T) {
		message, originalMessage := setupMessage()

		configurer := MandrillConfigurer{}
		configItems := []mmailer.ConfigItem{
			{Service: "", Key: mmailer.Vendor, Value: "mandrill"},
		}

		services.ApplyConfig("mandrill", configItems, configurer, message)

		if !reflect.DeepEqual(message, originalMessage) {
			t.Errorf("Mandrill message should remain unchanged for vendor config")
		}
	})

	t.Run("Empty Config", func(t *testing.T) {
		message, originalMessage := setupMessage()

		configurer := MandrillConfigurer{}

		services.ApplyConfig("mandrill", []mmailer.ConfigItem{}, configurer, message)

		if !reflect.DeepEqual(message, originalMessage) {
			t.Errorf("Mandrill message should remain unchanged with empty config")
		}
	})
}

func TestMandrillConfigurer_SetIpPool(t *testing.T) {
	configurer := MandrillConfigurer{}

	t.Run("SetIpPool No-op", func(t *testing.T) {
		message, originalMessage := setupMessage()

		// Call SetIpPool directly - should be no-op
		configurer.SetIpPool("any_pool", message)

		if !reflect.DeepEqual(message, originalMessage) {
			t.Errorf("Mandrill message should remain unchanged after SetIpPool call")
		}
	})
}

func setupMessage() (msg *mandrill.Message, originalMsg *mandrill.Message) {
	msg = &mandrill.Message{
		Subject:   "Test Subject",
		Text:      "Test body",
		FromEmail: "test@example.com",
	}
	originalMsg = &mandrill.Message{
		Subject:   "Test Subject",
		Text:      "Test body",
		FromEmail: "test@example.com",
	}
	return
}
