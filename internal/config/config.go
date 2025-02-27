package config

import (
	"os"
)

// GetAPIHost returns the API host URL from environment or default
func GetAPIHost() string {
	return getEnv("PINATA_API_HOST", "api.pinata.cloud")
}

// GetUploadsHost returns the uploads host URL from environment or default
func GetUploadsHost() string {
	return getEnv("PINATA_UPLOADS_HOST", "uploads.pinata.cloud")
}

// Helper function to get environment variable with fallback
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}
