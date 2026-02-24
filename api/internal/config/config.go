package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Addr        string
	DatabaseURL string

	AutoMigrate bool

	JWTSecret        string
	JWTAccessTTL     time.Duration
	DevAllowAllCORS  bool
	AllowedCORSOrigin string

	ShippingFlatINR int
	TaxRateBps      int // basis points, e.g. 1800 = 18%

	RazorpayKeyID        string
	RazorpayKeySecret    string
	RazorpayWebhookSecret string
}

func Load() (Config, error) {
	var c Config

	c.Addr = envOr("API_ADDR", ":8080")
	c.DatabaseURL = os.Getenv("DATABASE_URL")
	c.AutoMigrate = envBool("AUTO_MIGRATE", false)

	c.JWTSecret = os.Getenv("JWT_SECRET")
	c.JWTAccessTTL = envDuration("JWT_ACCESS_TTL", 24*time.Hour)

	c.DevAllowAllCORS = envBool("DEV_ALLOW_ALL_CORS", true)
	c.AllowedCORSOrigin = envOr("ALLOWED_CORS_ORIGIN", "")

	c.ShippingFlatINR = envInt("SHIPPING_FLAT_INR", 0)
	c.TaxRateBps = envInt("TAX_RATE_BPS", 0)

	c.RazorpayKeyID = os.Getenv("RAZORPAY_KEY_ID")
	c.RazorpayKeySecret = os.Getenv("RAZORPAY_KEY_SECRET")
	c.RazorpayWebhookSecret = os.Getenv("RAZORPAY_WEBHOOK_SECRET")

	if c.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if c.JWTSecret == "" {
		return Config{}, errors.New("JWT_SECRET is required")
	}
	return c, nil
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envBool(k string, def bool) bool {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func envInt(k string, def int) int {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func envDuration(k string, def time.Duration) time.Duration {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}

