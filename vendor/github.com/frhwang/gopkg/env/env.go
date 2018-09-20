package env

import "os"

// GetOrDefault returns a value from environment variable,
// but returns a default value if a variable isn't exist
func GetOrDefault(name, defaultValue string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return defaultValue
}
