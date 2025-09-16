package generic

import (
	"reflect"
	"testing"

	"github.com/modfin/mmailer"
	"github.com/modfin/mmailer/internal/smtpx"
	"github.com/modfin/mmailer/services"
)

func TestGenericConfigurer_ApplyConfig(t *testing.T) {
	t.Run("IP Pool No-op", func(t *testing.T) {
		message, originalMessage := setupMessage()

		configurer := GenericConfigurer{}
		configItems := []mmailer.ConfigItem{
			{Service: "", Key: mmailer.IpPool, Value: "any_pool"},
		}

		services.ApplyConfig("generic", configItems, configurer, message)

		if !reflect.DeepEqual(message, originalMessage) {
			t.Errorf("Brev message should remain unchanged after SetIpPool call")
		}
	})

	t.Run("Service-Specific Config No-op", func(t *testing.T) {
		message, originalMessage := setupMessage()

		configurer := GenericConfigurer{}
		configItems := []mmailer.ConfigItem{
			{Service: "generic", Key: mmailer.IpPool, Value: "generic_pool"},
			{Service: "sendgrid", Key: mmailer.IpPool, Value: "sg_pool"}, // Should be ignored
		}

		services.ApplyConfig("generic", configItems, configurer, message)

		if !reflect.DeepEqual(message, originalMessage) {
			t.Errorf("Brev message should remain unchanged after SetIpPool call")
		}
	})

	t.Run("Vendor Config No-op", func(t *testing.T) {
		message, originalMessage := setupMessage()

		configurer := GenericConfigurer{}
		configItems := []mmailer.ConfigItem{
			{Service: "", Key: mmailer.Vendor, Value: "generic"},
		}

		services.ApplyConfig("generic", configItems, configurer, message)

		if !reflect.DeepEqual(message, originalMessage) {
			t.Errorf("Brev message should remain unchanged after SetIpPool call")
		}
	})

	t.Run("Empty Config", func(t *testing.T) {
		message, originalMessage := setupMessage()

		configurer := GenericConfigurer{}

		services.ApplyConfig("generic", []mmailer.ConfigItem{}, configurer, message)

		if !reflect.DeepEqual(message, originalMessage) {
			t.Errorf("Brev message should remain unchanged after SetIpPool call")
		}
	})
}

func TestGenericConfigurer_SetIpPool(t *testing.T) {
	configurer := GenericConfigurer{}

	t.Run("SetIpPool No-op", func(t *testing.T) {
		message, originalMessage := setupMessage()

		// Call SetIpPool directly - should be no-op
		configurer.SetIpPool("any_pool", message)

		if !reflect.DeepEqual(message, originalMessage) {
			t.Errorf("Brev message should remain unchanged after SetIpPool call")
		}
	})
}

func setupMessage() (message *smtpx.Message, originalMessage *smtpx.Message) {
	message = smtpx.NewMessage()
	message.SetHeader("Subject", "Test Subject")
	message.SetHeader("From", "test@example.com")

	originalMessage = smtpx.NewMessage()
	originalMessage.SetHeader("Subject", "Test Subject")
	originalMessage.SetHeader("From", "test@example.com")
	return
}
