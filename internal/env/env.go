package env

import "os"

func GetEnv(envName, defaultValue string) string {
	val := os.Getenv(envName)
	if val != "" {
		return val
	}
	return defaultValue
}
