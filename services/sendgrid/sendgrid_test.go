package sendgrid

import (
	"reflect"
	"testing"

	"github.com/modfin/mmailer"
	"github.com/modfin/mmailer/services"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func TestSendgridConfigurer_ApplyConfig(t *testing.T) {
	t.Run("Valid EU Pool", func(t *testing.T) {
		message, originalMessage := setupMessage()

		configurer := SendgridConfigurer{}
		configItems := []mmailer.ConfigItem{
			{Service: "", Key: mmailer.IpPool, Value: "sg_eu"},
		}

		services.ApplyConfig("sendgrid", configItems, configurer, message)

		if message.IPPoolID != "sg_eu" {
			t.Errorf("Expected IP Pool to be 'sg_eu', got %v", message.IPPoolID)
		}

		if reflect.DeepEqual(message, originalMessage) {
			t.Errorf("Expected message to be changed for valid pool")
		}
	})

	t.Run("Valid US Pool", func(t *testing.T) {
		message, originalMessage := setupMessage()

		configurer := SendgridConfigurer{}
		configItems := []mmailer.ConfigItem{
			{Service: "", Key: mmailer.IpPool, Value: "sg_us"},
		}

		services.ApplyConfig("sendgrid", configItems, configurer, message)

		if message.IPPoolID != "sg_us" {
			t.Errorf("Expected IP Pool to be 'sg_us', got %v", message.IPPoolID)
		}

		if reflect.DeepEqual(message, originalMessage) {
			t.Errorf("Expected message to be changed for valid pool")
		}
	})

	t.Run("Invalid Pool", func(t *testing.T) {
		message, originalMessage := setupMessage()

		configurer := SendgridConfigurer{}
		configItems := []mmailer.ConfigItem{
			{Service: "", Key: mmailer.IpPool, Value: "invalid_pool"},
		}

		services.ApplyConfig("sendgrid", configItems, configurer, message)

		if message.IPPoolID != "initial_pool" {
			t.Errorf("Expected IP Pool to remain 'initial_pool', got %v", message.IPPoolID)
		}

		if !reflect.DeepEqual(message, originalMessage) {
			t.Errorf("Expected message to be changed for valid pool")
		}
	})

	t.Run("Service-Specific Config", func(t *testing.T) {
		message, originalMessage := setupMessage()

		configurer := SendgridConfigurer{}
		configItems := []mmailer.ConfigItem{
			{Service: "sendgrid", Key: mmailer.IpPool, Value: "sg_eu"},
			{Service: "mailjet", Key: mmailer.IpPool, Value: "mj_pool"}, // Should be ignored
		}

		services.ApplyConfig("sendgrid", configItems, configurer, message)

		if message.IPPoolID != "sg_eu" {
			t.Errorf("Expected IP Pool to be 'sg_eu', got %v", message.IPPoolID)
		}

		if reflect.DeepEqual(message, originalMessage) {
			t.Errorf("Expected message to be changed for valid pool")
		}
	})

	t.Run("Global and Service-Specific Config", func(t *testing.T) {
		message, originalMessage := setupMessage()

		configurer := SendgridConfigurer{}
		configItems := []mmailer.ConfigItem{
			{Service: "", Key: mmailer.IpPool, Value: "sg_us"},         // Global config
			{Service: "sendgrid", Key: mmailer.IpPool, Value: "sg_eu"}, // Service-specific (should override)
		}

		services.ApplyConfig("sendgrid", configItems, configurer, message)

		// Service-specific config should be applied last (overriding global)
		if message.IPPoolID != "sg_eu" {
			t.Errorf("Expected final IP Pool to be 'sg_eu', got %v", message.IPPoolID)
		}

		if reflect.DeepEqual(message, originalMessage) {
			t.Errorf("Expected message to be changed for valid pool")
		}
	})

	t.Run("Vendor Config No-op", func(t *testing.T) {
		message, originalMessage := setupMessage()

		configurer := SendgridConfigurer{}
		configItems := []mmailer.ConfigItem{
			{Service: "", Key: mmailer.Vendor, Value: "sendgrid"},
		}

		services.ApplyConfig("sendgrid", configItems, configurer, message)

		// Vendor config should be no-op
		if message.IPPoolID != "initial_pool" {
			t.Errorf("Vendor config should be no-op, IP Pool should remain 'initial_pool', got %v", message.IPPoolID)
		}

		if !reflect.DeepEqual(message, originalMessage) {
			t.Errorf("Expected message to remain unchanged for invalid pool")
		}
	})

	t.Run("Unknown Config Key", func(t *testing.T) {
		message, originalMessage := setupMessage()

		configurer := SendgridConfigurer{}
		configItems := []mmailer.ConfigItem{
			{Service: "", Key: "X-Unknown", Value: "unknown_value"},
		}

		services.ApplyConfig("sendgrid", configItems, configurer, message)

		// Unknown config should be ignored
		if message.IPPoolID != "initial_pool" {
			t.Errorf("Unknown config should be ignored, IP Pool should remain 'initial_pool', got %v", message.IPPoolID)
		}

		if !reflect.DeepEqual(message, originalMessage) {
			t.Errorf("Expected message to remain unchanged for invalid pool")
		}
	})

	t.Run("Empty Config", func(t *testing.T) {
		message, originalMessage := setupMessage()

		configurer := SendgridConfigurer{}

		services.ApplyConfig("sendgrid", []mmailer.ConfigItem{}, configurer, message)

		// Should remain unchanged
		if message.IPPoolID != "initial_pool" {
			t.Errorf("Expected IP Pool to remain 'initial_pool', got %v", message.IPPoolID)
		}

		if !reflect.DeepEqual(message, originalMessage) {
			t.Errorf("Expected message to remain unchanged for invalid pool")
		}
	})
}

func TestSendgridConfigurer_SetIpPool(t *testing.T) {
	configurer := SendgridConfigurer{}

	t.Run("Valid EU Pool Direct", func(t *testing.T) {
		message, originalMessage := setupMessage()

		configurer.SetIpPool("sg_eu", message)

		if message.IPPoolID != "sg_eu" {
			t.Errorf("Expected IP Pool to be 'sg_eu', got %v", message.IPPoolID)
		}

		if reflect.DeepEqual(message, originalMessage) {
			t.Errorf("Expected message to be changed for valid pool")
		}
	})

	t.Run("Valid US Pool Direct", func(t *testing.T) {
		message, originalMessage := setupMessage()

		configurer.SetIpPool("sg_us", message)

		if message.IPPoolID != "sg_us" {
			t.Errorf("Expected IP Pool to be 'sg_us', got %v", message.IPPoolID)
		}

		if reflect.DeepEqual(message, originalMessage) {
			t.Errorf("Expected message to be changed for valid pool")
		}
	})

	t.Run("Invalid Pool Direct", func(t *testing.T) {
		message, originalMessage := setupMessage()

		configurer.SetIpPool("invalid_pool", message)

		// Should remain unchanged for invalid pool
		if message.IPPoolID != "initial_pool" {
			t.Errorf("Expected IP Pool to remain 'initial_pool' for invalid pool, got %v", message.IPPoolID)
		}

		if !reflect.DeepEqual(message, originalMessage) {
			t.Errorf("Expected message to remain unchanged for invalid pool")
		}
	})
}

func setupMessage() (message *mail.SGMailV3, originalMessage *mail.SGMailV3) {
	message = mail.NewSingleEmail(
		mail.NewEmail("test", "test@example.com"),
		"Test Subject",
		nil,
		"Test Text",
		"<p>Test HTML</p>",
	)
	message.SetIPPoolID("initial_pool")
	originalMessage = mail.NewSingleEmail(
		mail.NewEmail("test", "test@example.com"),
		"Test Subject",
		nil,
		"Test Text",
		"<p>Test HTML</p>",
	)
	originalMessage.SetIPPoolID("initial_pool")
	return
}
