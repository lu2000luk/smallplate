package main

import (
	"net/url"
	"os"
	"os/signal"

	"github.com/joho/godotenv"
)

const Version = "0.1.0"

func main() {

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	err := godotenv.Load()
	if err != nil {
		Warn("Error loading .env file (ignore if vars are set in the env)")
	}

	db_url := os.Getenv("DB_URL")
	service_id := os.Getenv("SERVICE_ID")
	service_key := os.Getenv("SERVICE_KEY")
	manager_url := os.Getenv("MANAGER_URL")

	if db_url == "" || service_id == "" || service_key == "" || manager_url == "" {
		Error("Missing required environment variables. Required variables: DB_URL, SERVICE_ID, SERVICE_KEY, MANAGER_URL")
		Log("Variables configuration guide:")
		Log("DB_URL: The URL of the Redis/Valkey (or compatible)")
		Log("SERVICE_ID: The unique identifier for this service (do not change this ever to avoid data loss)")
		Log("SERVICE_KEY: The secret key to connect to the manager")
		Log("MANAGER_URL: The URL of the manager service (e.g. manager.example.com), no protocol, no path")
		return
	}

	u := url.URL{
		Scheme: "ws",
		Host:   manager_url,
		Path:   "/__service",
		RawQuery: url.Values{
			"id": []string{service_id},
			"t":  []string{"kv"},
			"k":  []string{service_key},
		}.Encode(),
	}
	u_masked := url.URL{
		Scheme: "ws",
		Host:   u.Host,
		Path:   u.Path,
		RawQuery: url.Values{
			"id": []string{service_id},
			"t":  []string{"kv"},
			"k":  []string{"REDACTED"},
		}.Encode(),
	}
	Log("Connecting to", u_masked.String())
}
