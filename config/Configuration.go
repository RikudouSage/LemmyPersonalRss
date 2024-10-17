package config

import (
	"strconv"
	"syscall"
)

type Configuration struct {
	Instance     string
	Port         int
	DatabasePath *string
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

	GlobalConfiguration = &Configuration{
		Instance:     getEnvOrPanic("INSTANCE"),
		Port:         port,
		DatabasePath: dbPath,
	}
}
