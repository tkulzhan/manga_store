package helpers

import (
	"fmt"
	"manga_store/internal/logger"
	"os"
	"strconv"
)

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	logger.Warn(fmt.Sprintf("Could not find "+"%s in env. Returning fallback", key))
	return fallback
}

func GetEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		v, err := strconv.Atoi(value)
		if err != nil {
			return 0
		}
		return int(v)
	}
	logger.Warn(fmt.Sprintf("%s not found in environment variables", key))
	return fallback
}

func GetEnvBool(key string, fallback bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		v, err := strconv.ParseBool(value)
		if err != nil {
			return false
		}
		return v
	}
	logger.Warn(fmt.Sprintf("%s not found in environment variables", key))
	return fallback
}
