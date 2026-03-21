package runtime

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"
)

var validEnvironments = map[string]struct{}{
	"local":       {},
	"development": {},
	"test":        {},
	"staging":     {},
	"production":  {},
}

type Config struct {
	AppEnv      string
	LogLevel    slog.Level
	HTTPPort    int
	DatabaseURL string
}

func LoadConfigFromEnv() (Config, error) {
	var cfg Config
	var validationErrs []string

	appEnv := strings.TrimSpace(os.Getenv("APP_ENV"))
	if appEnv == "" {
		validationErrs = append(validationErrs, "APP_ENV is required")
	} else {
		if _, ok := validEnvironments[appEnv]; !ok {
			validationErrs = append(validationErrs, "APP_ENV must be one of: local, development, test, staging, production")
		}
		cfg.AppEnv = appEnv
	}

	logLevelRaw := strings.TrimSpace(os.Getenv("LOG_LEVEL"))
	if logLevelRaw == "" {
		validationErrs = append(validationErrs, "LOG_LEVEL is required")
	} else {
		level, err := parseLogLevel(logLevelRaw)
		if err != nil {
			validationErrs = append(validationErrs, err.Error())
		} else {
			cfg.LogLevel = level
		}
	}

	httpPortRaw := strings.TrimSpace(os.Getenv("HTTP_PORT"))
	if httpPortRaw == "" {
		validationErrs = append(validationErrs, "HTTP_PORT is required")
	} else {
		port, err := strconv.Atoi(httpPortRaw)
		if err != nil || port < 1 || port > 65535 {
			validationErrs = append(validationErrs, "HTTP_PORT must be an integer between 1 and 65535")
		} else {
			cfg.HTTPPort = port
		}
	}

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		validationErrs = append(validationErrs, "DATABASE_URL is required")
	} else {
		if err := validateDatabaseURL(databaseURL); err != nil {
			validationErrs = append(validationErrs, err.Error())
		} else {
			cfg.DatabaseURL = databaseURL
		}
	}

	if len(validationErrs) > 0 {
		return Config{}, errors.New(strings.Join(validationErrs, "; "))
	}

	return cfg, nil
}

func parseLogLevel(value string) (slog.Level, error) {
	switch strings.ToLower(value) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("LOG_LEVEL must be one of: debug, info, warn, error")
	}
}

func validateDatabaseURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("DATABASE_URL must be a valid URL: %w", err)
	}

	if parsed.Scheme == "" || parsed.Host == "" {
		return errors.New("DATABASE_URL must include a scheme and host")
	}

	return nil
}
