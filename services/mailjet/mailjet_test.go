package mailjet

import (
	"reflect"
	"testing"

	mj "github.com/mailjet/mailjet-apiv3-go/v3"
	"github.com/modfin/mmailer"
	"github.com/modfin/mmailer/services"
)

func TestMailjetConfigurer_ApplyConfig(t *testing.T) {
	t.Run("IP Pool No-op", func(t *testing.T) {
		message, originalMessage := setupMessage()

		configurer := MailjetConfigurer{}
		configItems := []mmailer.ConfigItem{
			{Service: "", Key: mmailer.IpPool, Value: "any_pool"},
		}

		services.ApplyConfig("mailjet", configItems, configurer, message)

		// Mailjet configurer is no-op, so we verify the entire message is unchanged
		if !reflect.DeepEqual(message, originalMessage) {
			t.Errorf("Mailjet message should remain unchanged after ApplyConfig")
		}
	})

	t.Run("Service-Specific Config No-op", func(t *testing.T) {
		message, originalMessage := setupMessage()

		configurer := MailjetConfigurer{}
		configItems := []mmailer.ConfigItem{
			{Service: "mailjet", Key: mmailer.IpPool, Value: "mailjet_pool"},
			{Service: "sendgrid", Key: mmailer.IpPool, Value: "sg_pool"}, // Should be ignored
		}

		services.ApplyConfig("mailjet", configItems, configurer, message)

		if !reflect.DeepEqual(message, originalMessage) {
			t.Errorf("Mailjet message should remain unchanged after service-specific config")
		}
	})

	t.Run("Vendor Config No-op", func(t *testing.T) {
		message, originalMessage := setupMessage()

		configurer := MailjetConfigurer{}
		configItems := []mmailer.ConfigItem{
			{Service: "", Key: mmailer.Vendor, Value: "mailjet"},
		}

		services.ApplyConfig("mailjet", configItems, configurer, message)

		if !reflect.DeepEqual(message, originalMessage) {
			t.Errorf("Mailjet message should remain unchanged for vendor config")
		}
	})

	t.Run("Empty Config", func(t *testing.T) {
		message, originalMessage := setupMessage()

		configurer := MailjetConfigurer{}

		services.ApplyConfig("mailjet", []mmailer.ConfigItem{}, configurer, message)

		if !reflect.DeepEqual(message, originalMessage) {
			t.Errorf("Mailjet message should remain unchanged with empty config")
		}
	})
}

func TestMailjetConfigurer_SetIpPool(t *testing.T) {
	configurer := MailjetConfigurer{}

	t.Run("SetIpPool No-op", func(t *testing.T) {
		message, originalMessage := setupMessage()

		// Call SetIpPool directly - should be no-op
		configurer.SetIpPool("any_pool", message)

		if !reflect.DeepEqual(message, originalMessage) {
			t.Errorf("Mailjet message should remain unchanged after SetIpPool call")
		}
	})
}

func setupMessage() (message *mj.MessagesV31, originalMessage *mj.MessagesV31) {
	message = &mj.MessagesV31{
		Info: []mj.InfoMessagesV31{
			{
				From: &mj.RecipientV31{
					Email: "test@example.com",
					Name:  "Test",
				},
				Subject:  "Test Subject",
				TextPart: "Test body",
			},
		},
	}

	// Create a deep copy for comparison
	originalMessage = &mj.MessagesV31{
		Info: []mj.InfoMessagesV31{
			{
				From: &mj.RecipientV31{
					Email: "test@example.com",
					Name:  "Test",
				},
				Subject:  "Test Subject",
				TextPart: "Test body",
			},
		},
	}
	return
}
