package clientcli

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

func successOrDie[T any](value T, err error) T {
	if err != nil {
		log.Fatal(err)
	}

	return value
}

func GetenvWithFallback(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func GetenvBool(key string, fallback bool) (bool, error) {
	if raw := os.Getenv(key); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return false, fmt.Errorf("parsing %s environment variable: %w", key, err)
		}

		return parsed, nil
	}

	return fallback, nil
}

func MustGetenvBool(key string, fallback bool) bool {
	return successOrDie(GetenvBool(key, fallback))
}

func GetenvDuration(key string, fallback time.Duration) (time.Duration, error) {
	if raw := os.Getenv(key); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return 0, fmt.Errorf("parsing %s environment variable: %s", key, err)
		}

		return parsed, nil
	}

	return fallback, nil
}

func MustGetenvDuration(key string, fallback time.Duration) time.Duration {
	return successOrDie(GetenvDuration(key, fallback))
}
