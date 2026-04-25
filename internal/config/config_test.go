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
	t.Setenv("GRAFANA_CLOUD_INSTANCE_ID", "")
	t.Setenv("GRAFANA_OTLP_INGEST_TOKEN", "")

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
	if cfg.Telemetry.InstanceID != "" {
		t.Fatalf("Telemetry.InstanceID = %q, want empty string", cfg.Telemetry.InstanceID)
	}
	if cfg.Telemetry.IngestToken != "" {
		t.Fatalf("Telemetry.IngestToken = %q, want empty string", cfg.Telemetry.IngestToken)
	}
}

func TestLoadFromEnvSuccessWithEnabledTelemetry(t *testing.T) {
	t.Setenv("APP_ENV", "rc")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("HTTP_PORT", "8080")
	t.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/platform_blueprint?sslmode=disable")
	t.Setenv("AUTH_ISSUER_URL", "https://issuer.example.com")
	t.Setenv("AUTH_AUDIENCE", "https://api.example.com")
	t.Setenv("OTEL_MODE", "direct_otlp")
	t.Setenv("OBS_TELEMETRY_PROFILE", "cost")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "https://otlp.example.com/otlp")
	t.Setenv("GRAFANA_CLOUD_INSTANCE_ID", "1546554")
	t.Setenv("GRAFANA_OTLP_INGEST_TOKEN", "secret-token")

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv() error = %v", err)
	}

	if cfg.Telemetry.Mode != TelemetryModeDirectOTLP {
		t.Fatalf("Telemetry.Mode = %q, want %q", cfg.Telemetry.Mode, TelemetryModeDirectOTLP)
	}
	if cfg.Telemetry.Profile != "cost" {
		t.Fatalf("Telemetry.Profile = %q, want %q", cfg.Telemetry.Profile, "cost")
	}
	if cfg.Telemetry.OTLPEndpoint == nil || cfg.Telemetry.OTLPEndpoint.String() != "https://otlp.example.com/otlp" {
		t.Fatalf("Telemetry.OTLPEndpoint = %v, want https://otlp.example.com/otlp", cfg.Telemetry.OTLPEndpoint)
	}
	if cfg.Telemetry.InstanceID != "1546554" {
		t.Fatalf("Telemetry.InstanceID = %q, want %q", cfg.Telemetry.InstanceID, "1546554")
	}
	if cfg.Telemetry.IngestToken != "secret-token" {
		t.Fatalf("Telemetry.IngestToken = %q, want %q", cfg.Telemetry.IngestToken, "secret-token")
	}
}

func TestLoadFromEnvComposesDatabaseURLFromSplitTCPEnv(t *testing.T) {
	setValidRuntimeEnv(t)
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DB_HOST", "10.10.0.4:5432")
	t.Setenv("DB_NAME", "platform_rc")
	t.Setenv("DB_USER", "api_user")
	t.Setenv("DB_PASSWORD", "secret/pass")

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv() error = %v", err)
	}

	want := "postgres://api_user:secret%2Fpass@10.10.0.4:5432/platform_rc?sslmode=disable"
	if cfg.DatabaseURL != want {
		t.Fatalf("DatabaseURL = %q, want %q", cfg.DatabaseURL, want)
	}
}

func TestLoadFromEnvComposesDatabaseURLFromCloudSQLSocket(t *testing.T) {
	setValidRuntimeEnv(t)
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DB_HOST", "/cloudsql/example-project:europe-west1:platform-rc-db")
	t.Setenv("DB_NAME", "platform_rc")
	t.Setenv("DB_USER", "api_user")
	t.Setenv("DB_PASSWORD", "secret")

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv() error = %v", err)
	}

	want := "postgres://api_user:secret@/platform_rc?host=%2Fcloudsql%2Fexample-project%3Aeurope-west1%3Aplatform-rc-db&sslmode=disable"
	if cfg.DatabaseURL != want {
		t.Fatalf("DatabaseURL = %q, want %q", cfg.DatabaseURL, want)
	}
}

func TestLoadDatabaseURLFromEnvUsesDatabaseURLWhenSplitValuesAreBlank(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/platform_blueprint?sslmode=disable")
	t.Setenv("DB_HOST", "")
	t.Setenv("DB_NAME", "")
	t.Setenv("DB_USER", "")
	t.Setenv("DB_PASSWORD", "")

	var problems []string
	databaseURL := LoadDatabaseURLFromEnv(&problems)
	if len(problems) > 0 {
		t.Fatalf("LoadDatabaseURLFromEnv() problems = %v, want none", problems)
	}

	want := "postgres://postgres:postgres@localhost:5432/platform_blueprint?sslmode=disable"
	if databaseURL != want {
		t.Fatalf("LoadDatabaseURLFromEnv() = %q, want %q", databaseURL, want)
	}
}

func TestLoadDatabaseURLFromEnvPrefersSplitValuesWhenConfigured(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/platform_blueprint?sslmode=disable")
	t.Setenv("DB_HOST", "/cloudsql/example-project:europe-west1:platform-rc-db")
	t.Setenv("DB_NAME", "platform_rc")
	t.Setenv("DB_USER", "api_user")
	t.Setenv("DB_PASSWORD", "secret")

	var problems []string
	databaseURL := LoadDatabaseURLFromEnv(&problems)
	if len(problems) > 0 {
		t.Fatalf("LoadDatabaseURLFromEnv() problems = %v, want none", problems)
	}

	want := "postgres://api_user:secret@/platform_rc?host=%2Fcloudsql%2Fexample-project%3Aeurope-west1%3Aplatform-rc-db&sslmode=disable"
	if databaseURL != want {
		t.Fatalf("LoadDatabaseURLFromEnv() = %q, want %q", databaseURL, want)
	}
}

func TestLoadDatabaseURLFromEnvIgnoresInvalidDatabaseURLWhenSplitValuesAreConfigured(t *testing.T) {
	t.Setenv("DATABASE_URL", "not-a-url")
	t.Setenv("DB_HOST", "10.10.0.4:5432")
	t.Setenv("DB_NAME", "platform_rc")
	t.Setenv("DB_USER", "api_user")
	t.Setenv("DB_PASSWORD", "secret")

	var problems []string
	databaseURL := LoadDatabaseURLFromEnv(&problems)
	if len(problems) > 0 {
		t.Fatalf("LoadDatabaseURLFromEnv() problems = %v, want none", problems)
	}

	want := "postgres://api_user:secret@10.10.0.4:5432/platform_rc?sslmode=disable"
	if databaseURL != want {
		t.Fatalf("LoadDatabaseURLFromEnv() = %q, want %q", databaseURL, want)
	}
}

func TestLoadDatabaseURLFromEnvReportsMissingSplitValues(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DB_HOST", "/cloudsql/example-project:europe-west1:platform-rc-db")
	t.Setenv("DB_NAME", "platform_rc")

	var problems []string
	databaseURL := LoadDatabaseURLFromEnv(&problems)
	if databaseURL != "" {
		t.Fatalf("LoadDatabaseURLFromEnv() = %q, want empty string", databaseURL)
	}

	message := strings.Join(problems, "\n")
	for _, want := range []string{
		"DB_USER is required",
		"DB_PASSWORD is required",
	} {
		if !strings.Contains(message, want) {
			t.Fatalf("LoadDatabaseURLFromEnv() problems = %q, want substring %q", message, want)
		}
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
	t.Setenv("GRAFANA_CLOUD_INSTANCE_ID", "")
	t.Setenv("GRAFANA_OTLP_INGEST_TOKEN", "")

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
		"GRAFANA_CLOUD_INSTANCE_ID cannot be empty when OTEL_MODE is enabled",
		"GRAFANA_OTLP_INGEST_TOKEN cannot be empty when OTEL_MODE is enabled",
	} {
		if !strings.Contains(message, want) {
			t.Fatalf("LoadFromEnv() error = %q, want substring %q", message, want)
		}
	}
}

func setValidRuntimeEnv(t *testing.T) {
	t.Helper()

	t.Setenv("APP_ENV", "rc")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("HTTP_PORT", "8080")
	t.Setenv("AUTH_ISSUER_URL", "https://issuer.example.com")
	t.Setenv("AUTH_AUDIENCE", "https://api.example.com")
	t.Setenv("OTEL_MODE", "disabled")
	t.Setenv("OBS_TELEMETRY_PROFILE", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("GRAFANA_CLOUD_INSTANCE_ID", "")
	t.Setenv("GRAFANA_OTLP_INGEST_TOKEN", "")
}
