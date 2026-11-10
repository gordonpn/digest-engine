// Package config loads the configuration supported by the current runtime.
package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port            int
	ShutdownTimeout time.Duration
}

// Load reads environment configuration. Errors identify fields but never echo values.
func Load() (Config, error) {
	c := Config{Port: 8080, ShutdownTimeout: 10 * time.Second}
	if value, set := os.LookupEnv("PORT"); set {
		port, err := strconv.Atoi(value)
		if err != nil || port < 1 || port > 65535 {
			return Config{}, errors.New("PORT must be an integer between 1 and 65535")
		}
		c.Port = port
	}
	if value, set := os.LookupEnv("DIGEST_ENGINE_SHUTDOWN_TIMEOUT"); set {
		duration, err := time.ParseDuration(value)
		if err != nil || duration <= 0 {
			return Config{}, errors.New("DIGEST_ENGINE_SHUTDOWN_TIMEOUT must be a positive duration")
		}
		c.ShutdownTimeout = duration
	}
	return c, nil
}
