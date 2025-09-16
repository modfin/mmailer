package brev

import (
	"reflect"
	"testing"

	brevpkg "github.com/modfin/brev"
	"github.com/modfin/mmailer"
	"github.com/modfin/mmailer/services"
)

func TestBrevConfigurer_ApplyConfig(t *testing.T) {
	t.Run("IP Pool No-op", func(t *testing.T) {
		message, originalMessage := setupMessage()

		configurer := BrevConfigurer{}
		configItems := []mmailer.ConfigItem{
			{Service: "", Key: mmailer.IpPool, Value: "any_pool"},
		}

		services.ApplyConfig("brev", configItems, configurer, message)

		// Brev configurer is no-op, so we verify the entire message is unchanged
		if !reflect.DeepEqual(message, originalMessage) {
			t.Errorf("Brev message should remain unchanged after ApplyConfig")
		}
	})

	t.Run("Service-Specific Config No-op", func(t *testing.T) {
		message, originalMessage := setupMessage()

		configurer := BrevConfigurer{}
		configItems := []mmailer.ConfigItem{
			{Service: "brev", Key: mmailer.IpPool, Value: "brev_pool"},
			{Service: "sendgrid", Key: mmailer.IpPool, Value: "sg_pool"}, // Should be ignored
		}

		services.ApplyConfig("brev", configItems, configurer, message)

		if !reflect.DeepEqual(message, originalMessage) {
			t.Errorf("Brev message should remain unchanged after service-specific config")
		}
	})

	t.Run("Vendor Config No-op", func(t *testing.T) {
		message, originalMessage := setupMessage()

		configurer := BrevConfigurer{}
		configItems := []mmailer.ConfigItem{
			{Service: "", Key: mmailer.Vendor, Value: "brev"},
		}

		services.ApplyConfig("brev", configItems, configurer, message)

		if !reflect.DeepEqual(message, originalMessage) {
			t.Errorf("Brev message should remain unchanged for vendor config")
		}
	})

	t.Run("Empty Config", func(t *testing.T) {
		message, originalMessage := setupMessage()

		configurer := BrevConfigurer{}

		services.ApplyConfig("brev", []mmailer.ConfigItem{}, configurer, message)

		if !reflect.DeepEqual(message, originalMessage) {
			t.Errorf("Brev message should remain unchanged with empty config")
		}
	})
}

func TestBrevConfigurer_SetIpPool(t *testing.T) {
	configurer := BrevConfigurer{}

	t.Run("SetIpPool No-op", func(t *testing.T) {
		message, originalMessage := setupMessage()

		// Call SetIpPool directly - should be no-op
		configurer.SetIpPool("any_pool", message)

		if !reflect.DeepEqual(message, originalMessage) {
			t.Errorf("Brev message should remain unchanged after SetIpPool call")
		}
	})
}

func setupMessage() (message *brevpkg.Email, originalMessage *brevpkg.Email) {
	message = brevpkg.NewEmail()
	message.Subject = "Test Subject"
	message.From = brevpkg.Address{Name: "Test User", Email: "test@example.com"}
	message.Text = "Test body"

	originalMessage = brevpkg.NewEmail()
	originalMessage.Subject = message.Subject
	originalMessage.From = message.From
	originalMessage.Text = message.Text

	return
}
