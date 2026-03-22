// Package config owns startup configuration parsing and validation for backend-api.
package config

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// TelemetryMode controls which OTLP delivery path the service is configured for.
type TelemetryMode string

const (
	// TelemetryModeDisabled keeps the Phase 2 skeleton runnable before observability wiring lands.
	TelemetryModeDisabled TelemetryMode = "disabled"
	// TelemetryModeDirectOTLP targets the direct Cloud Run exporter path.
	TelemetryModeDirectOTLP TelemetryMode = "direct_otlp"
	// TelemetryModeCollectorGateway targets the future GKE collector gateway path.
	TelemetryModeCollectorGateway TelemetryMode = "collector_gateway"
)

// Config captures the validated environment contract needed to boot the API runtime.
type Config struct {
	AppEnv      string
	LogLevel    slog.Level
	HTTPPort    int
	DatabaseURL string

	AuthIssuerURL *url.URL
	AuthAudience  string

	Telemetry TelemetryConfig
}

// TelemetryConfig stores the observability-related environment contract even though
// exporter wiring is deferred to a later task.
type TelemetryConfig struct {
	Mode         TelemetryMode
	OTLPEndpoint *url.URL
	OTLPHeaders  string
}

// LoadFromEnv parses the required environment variables and returns all validation
// failures together so startup exits with one actionable error message.
func LoadFromEnv() (Config, error) {
	var problems []string
	cfg := Config{}

	cfg.AppEnv = requireNonEmptyEnv("APP_ENV", &problems)
	cfg.LogLevel = parseLogLevel(requireNonEmptyEnv("LOG_LEVEL", &problems), &problems)
	cfg.HTTPPort = parsePort(requireNonEmptyEnv("HTTP_PORT", &problems), &problems)
	cfg.DatabaseURL = parseRequiredURLString("DATABASE_URL", &problems)
	cfg.AuthIssuerURL = parseAbsoluteURL("AUTH_ISSUER_URL", requireNonEmptyEnv("AUTH_ISSUER_URL", &problems), &problems)
	cfg.AuthAudience = parseAudience("AUTH_AUDIENCE", requireNonEmptyEnv("AUTH_AUDIENCE", &problems), &problems)
	cfg.Telemetry.Mode = parseTelemetryMode(requireNonEmptyEnv("OTEL_MODE", &problems), &problems)

	endpointValue, endpointExists := lookupEnv("OTEL_EXPORTER_OTLP_ENDPOINT", &problems)
	headersValue, headersExists := lookupEnv("OTEL_EXPORTER_OTLP_HEADERS", &problems)
	if endpointExists && endpointValue != "" {
		cfg.Telemetry.OTLPEndpoint = parseAbsoluteURL("OTEL_EXPORTER_OTLP_ENDPOINT", endpointValue, &problems)
	}
	if headersExists {
		cfg.Telemetry.OTLPHeaders = headersValue
		if headersValue != "" {
			validateOTLPHeaders(headersValue, &problems)
		}
	}

	if cfg.Telemetry.Mode != TelemetryModeDisabled {
		if endpointValue == "" {
			problems = append(problems, "OTEL_EXPORTER_OTLP_ENDPOINT cannot be empty when OTEL_MODE is enabled")
		}
		if headersValue == "" {
			problems = append(problems, "OTEL_EXPORTER_OTLP_HEADERS cannot be empty when OTEL_MODE is enabled")
		}
	}

	if len(problems) > 0 {
		return Config{}, fmt.Errorf("invalid environment configuration:\n- %s", strings.Join(problems, "\n- "))
	}

	return cfg, nil
}

func lookupEnv(name string, problems *[]string) (string, bool) {
	value, exists := os.LookupEnv(name)
	if !exists {
		*problems = append(*problems, fmt.Sprintf("%s is required", name))
		return "", false
	}

	return strings.TrimSpace(value), true
}

func requireNonEmptyEnv(name string, problems *[]string) string {
	value, exists := lookupEnv(name, problems)
	if !exists {
		return ""
	}
	if value == "" {
		*problems = append(*problems, fmt.Sprintf("%s cannot be empty", name))
	}
	return value
}

func parseLogLevel(value string, problems *[]string) slog.Level {
	switch strings.ToLower(value) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "":
		return slog.LevelInfo
	default:
		*problems = append(*problems, fmt.Sprintf("LOG_LEVEL must be one of debug, info, warn, or error, got %q", value))
		return slog.LevelInfo
	}
}

func parsePort(value string, problems *[]string) int {
	if value == "" {
		return 0
	}

	port, err := strconv.Atoi(value)
	if err != nil || port < 1 || port > 65535 {
		*problems = append(*problems, fmt.Sprintf("HTTP_PORT must be an integer between 1 and 65535, got %q", value))
		return 0
	}

	return port
}

func parseRequiredURLString(name string, problems *[]string) string {
	value := requireNonEmptyEnv(name, problems)
	if value == "" {
		return ""
	}

	if parseAbsoluteURL(name, value, problems) == nil {
		return ""
	}

	return value
}

func parseAbsoluteURL(name, value string, problems *[]string) *url.URL {
	if value == "" {
		return nil
	}

	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		*problems = append(*problems, fmt.Sprintf("%s must be an absolute URL, got %q", name, value))
		return nil
	}

	return parsed
}

func parseAudience(name, value string, problems *[]string) string {
	if value == "" {
		return ""
	}

	if parseAbsoluteURL(name, value, problems) == nil {
		return ""
	}

	return value
}

func parseTelemetryMode(value string, problems *[]string) TelemetryMode {
	switch TelemetryMode(strings.ToLower(value)) {
	case TelemetryModeDisabled:
		return TelemetryModeDisabled
	case TelemetryModeDirectOTLP:
		return TelemetryModeDirectOTLP
	case TelemetryModeCollectorGateway:
		return TelemetryModeCollectorGateway
	case "":
		return TelemetryModeDisabled
	default:
		*problems = append(*problems, fmt.Sprintf("OTEL_MODE must be one of disabled, direct_otlp, or collector_gateway, got %q", value))
		return TelemetryModeDisabled
	}
}

func validateOTLPHeaders(value string, problems *[]string) {
	pairs := strings.Split(value, ",")
	for _, pair := range pairs {
		header := strings.TrimSpace(pair)
		if header == "" {
			*problems = append(*problems, "OTEL_EXPORTER_OTLP_HEADERS cannot contain empty header entries")
			return
		}

		name, rawValue, found := strings.Cut(header, "=")
		if !found || strings.TrimSpace(name) == "" || strings.TrimSpace(rawValue) == "" {
			*problems = append(*problems, fmt.Sprintf("OTEL_EXPORTER_OTLP_HEADERS entries must use key=value format, got %q", header))
			return
		}
	}
}
