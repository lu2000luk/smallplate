package plate

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const ServiceType = "kv"

type Config struct {
	DBURL            string
	ServiceID        string
	ServiceKey       string
	ManagerURL       string
	PublicURL        string
	HTTPAddr         string
	AuthCacheSize    int
	RedisOpTimeout   time.Duration
	ShutdownTimeout  time.Duration
	PingInterval     time.Duration
	PongMissTimeout  time.Duration
	WriteTimeout     time.Duration
	ReconnectDelay   time.Duration
	MaxDialRetries   int
	PubSubBufferSize int
}

func LoadConfig() (Config, error) {
	cfg := Config{
		DBURL:            strings.TrimSpace(os.Getenv("DB_URL")),
		ServiceID:        strings.TrimSpace(os.Getenv("SERVICE_ID")),
		ServiceKey:       strings.TrimSpace(os.Getenv("SERVICE_KEY")),
		ManagerURL:       strings.TrimSpace(os.Getenv("MANAGER_URL")),
		PublicURL:        strings.TrimSpace(os.Getenv("PUBLIC_URL")),
		HTTPAddr:         envString("HTTP_ADDR", ":3400"),
		AuthCacheSize:    envInt("AUTH_CACHE_SIZE", 1000),
		RedisOpTimeout:   envDuration("REDIS_OP_TIMEOUT", 5*time.Second),
		ShutdownTimeout:  envDuration("SHUTDOWN_TIMEOUT", 15*time.Second),
		PingInterval:     envDuration("MANAGER_PING_INTERVAL", 30*time.Second),
		PongMissTimeout:  envDuration("MANAGER_PONG_TIMEOUT", 30*time.Second),
		WriteTimeout:     envDuration("MANAGER_WRITE_TIMEOUT", 10*time.Second),
		ReconnectDelay:   envDuration("MANAGER_RETRY_DELAY", time.Second),
		MaxDialRetries:   envInt("MANAGER_MAX_DIAL_RETRIES", 20),
		PubSubBufferSize: envInt("PUBSUB_BUFFER_SIZE", 128),
	}

	var missing []string
	if cfg.DBURL == "" {
		missing = append(missing, "DB_URL")
	}
	if cfg.ServiceID == "" {
		missing = append(missing, "SERVICE_ID")
	}
	if cfg.ServiceKey == "" {
		missing = append(missing, "SERVICE_KEY")
	}
	if cfg.ManagerURL == "" {
		missing = append(missing, "MANAGER_URL")
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	if cfg.AuthCacheSize <= 0 {
		cfg.AuthCacheSize = 1000
	}
	if cfg.MaxDialRetries <= 0 {
		cfg.MaxDialRetries = 20
	}
	if cfg.PubSubBufferSize <= 0 {
		cfg.PubSubBufferSize = 128
	}
	return cfg, nil
}

func (c Config) ManagerWSURL() (string, string, error) {
	base := strings.TrimSpace(c.ManagerURL)
	if base == "" {
		return "", "", fmt.Errorf("manager url is empty")
	}

	var u url.URL
	if strings.HasPrefix(base, "ws://") || strings.HasPrefix(base, "wss://") {
		parsed, err := url.Parse(base)
		if err != nil {
			return "", "", err
		}
		u = *parsed
	} else {
		u = url.URL{Scheme: "ws", Host: base}
	}

	u.Path = "/__service"
	values := url.Values{
		"id": []string{c.ServiceID},
		"t":  []string{ServiceType},
		"k":  []string{c.ServiceKey},
	}
	if c.PublicURL != "" {
		values.Set("u", c.PublicURL)
	}
	u.RawQuery = values.Encode()

	masked := u
	maskedValues := masked.Query()
	maskedValues.Set("k", "REDACTED")
	masked.RawQuery = maskedValues.Encode()

	return u.String(), masked.String(), nil
}

func envString(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}
