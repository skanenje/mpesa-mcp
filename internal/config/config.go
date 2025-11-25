package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	ConsumerKey  string
	ConsumerSec  string
	BaseURL      string
	BusinessCode string
	Passkey      string
	CallbackURL  string
	AccountRef   string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	cfg := &Config{
		ConsumerKey:  os.Getenv("MPESA_CONSUMER_KEY"),
		ConsumerSec:  os.Getenv("MPESA_CONSUMER_SECRET"),
		BaseURL:      os.Getenv("BASE_URL"),
		BusinessCode: os.Getenv("BUSINESS_SHORTCODE"),
		Passkey:      os.Getenv("PASSKEY"),
		CallbackURL:  os.Getenv("CALLBACK_URL"),
		AccountRef:   os.Getenv("ACCOUNT_REFERENCE"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validate checks that all required configuration is present
func (c *Config) validate() error {
	if c.ConsumerKey == "" || c.ConsumerSec == "" {
		return fmt.Errorf("MPESA_CONSUMER_KEY and MPESA_CONSUMER_SECRET are required")
	}
	if c.BaseURL == "" {
		return fmt.Errorf("BASE_URL is required")
	}
	if c.BusinessCode == "" || c.Passkey == "" {
		return fmt.Errorf("BUSINESS_SHORTCODE and PASSKEY are required")
	}
	return nil
}
