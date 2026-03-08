package main

import (
	"os"

	"github.com/joho/godotenv"
)

const Version = "0.1.0"

func main() {
	err := godotenv.Load()
	if err != nil {
		Warn("Error loading .env file")
	}

	db_url := os.Getenv("DB_URL")
	service_key := os.Getenv("SERVICE_KEY")
	manager_url := os.Getenv("MANAGER_URL")

	if db_url == "" || service_key == "" || manager_url == "" {
		Error("Missing required environment variables. Required variables: DB_URL, SERVICE_KEY, MANAGER_URL")
		return
	}
}
