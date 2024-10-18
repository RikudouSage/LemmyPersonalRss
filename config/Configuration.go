package config

import (
	"strconv"
	"syscall"
	"time"
)

type Configuration struct {
	Instance      string
	Port          int
	DatabasePath  *string
	CacheDuration time.Duration
	Logging       bool
}

var GlobalConfiguration *Configuration

func getEnvOrPanic(name string) string {
	value, ok := syscall.Getenv(name)
	if !ok {
		panic("environment variable " + name + " not found")
	}

	return value
}

func getEnvOrDefault(name string, defaultValue string) string {
	value, ok := syscall.Getenv(name)
	if !ok {
		return defaultValue
	}

	return value
}

func init() {
	portStr := getEnvOrDefault("PORT", "8080")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		panic(err)
	}

	dbPathRaw := getEnvOrDefault("DATABASE_PATH", "")
	var dbPath *string
	if dbPathRaw != "" {
		dbPath = &dbPathRaw
	}

	cacheDurationStr := getEnvOrDefault("CACHE_DURATION", "300")
	cacheDuration, err := strconv.Atoi(cacheDurationStr)
	if err != nil {
		panic(err)
	}

	loggingStr := getEnvOrDefault("LOGGING", "true")
	logging, err := strconv.ParseBool(loggingStr)
	if err != nil {
		panic(err)
	}

	GlobalConfiguration = &Configuration{
		Instance:      getEnvOrDefault("INSTANCE", ""),
		Port:          port,
		DatabasePath:  dbPath,
		CacheDuration: time.Duration(cacheDuration) * time.Second,
		Logging:       logging,
	}
}
