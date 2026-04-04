package config

import (
	"strings"
	"testing"
)

func TestLoadFromEnvSuccessWithDisabledTelemetry(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("HTTP_PORT", "8080")
	t.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/platform_blueprint?sslmode=disable")
	t.Setenv("AUTH_ISSUER_URL", "https://issuer.example.com")
	t.Setenv("AUTH_AUDIENCE", "https://api.example.com")
	t.Setenv("OTEL_MODE", "disabled")
	t.Setenv("OBS_TELEMETRY_PROFILE", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_HEADERS", "")

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv() error = %v", err)
	}

	if cfg.HTTPPort != 8080 {
		t.Fatalf("HTTPPort = %d, want 8080", cfg.HTTPPort)
	}
	if cfg.Telemetry.Mode != TelemetryModeDisabled {
		t.Fatalf("Telemetry.Mode = %q, want %q", cfg.Telemetry.Mode, TelemetryModeDisabled)
	}
	if cfg.Telemetry.Profile != "balanced" {
		t.Fatalf("Telemetry.Profile = %q, want %q", cfg.Telemetry.Profile, "balanced")
	}
}

func TestLoadFromEnvReportsMissingAndInvalidValues(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("LOG_LEVEL", "verbose")
	t.Setenv("HTTP_PORT", "0")
	t.Setenv("DATABASE_URL", "not-a-url")
	t.Setenv("AUTH_ISSUER_URL", "https://issuer.example.com")
	t.Setenv("AUTH_AUDIENCE", "")
	t.Setenv("OTEL_MODE", "direct_otlp")
	t.Setenv("OBS_TELEMETRY_PROFILE", "verbose")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_HEADERS", "")

	_, err := LoadFromEnv()
	if err == nil {
		t.Fatal("LoadFromEnv() error = nil, want validation error")
	}

	message := err.Error()
	for _, want := range []string{
		"LOG_LEVEL must be one of",
		"HTTP_PORT must be an integer",
		"DATABASE_URL must be an absolute URL",
		"AUTH_AUDIENCE cannot be empty",
		"OBS_TELEMETRY_PROFILE must be one of balanced, cost, or debug",
		"OTEL_EXPORTER_OTLP_ENDPOINT cannot be empty when OTEL_MODE is enabled",
		"OTEL_EXPORTER_OTLP_HEADERS cannot be empty when OTEL_MODE is enabled",
	} {
		if !strings.Contains(message, want) {
			t.Fatalf("LoadFromEnv() error = %q, want substring %q", message, want)
		}
	}
}
