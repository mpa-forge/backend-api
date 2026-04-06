package main

import (
	"context"
	"testing"

	"github.com/mpa-forge/backend-api/internal/config"
)

func TestNewObservabilityRuntimeUsesSharedConfig(t *testing.T) {
	runtime, err := newObservabilityRuntime(context.Background(), config.Config{
		AppEnv: "local",
		Telemetry: config.TelemetryConfig{
			Mode:    config.TelemetryModeDisabled,
			Profile: "cost",
		},
	})
	if err != nil {
		t.Fatalf("newObservabilityRuntime() error = %v", err)
	}
	defer func() {
		if err := runtime.Shutdown(context.Background()); err != nil {
			t.Fatalf("Shutdown() error = %v", err)
		}
	}()

	metadata := runtime.Metadata()
	if metadata.ServiceName != "backend-api" {
		t.Fatalf("ServiceName = %q, want %q", metadata.ServiceName, "backend-api")
	}
	if metadata.Environment != "local" {
		t.Fatalf("Environment = %q, want %q", metadata.Environment, "local")
	}
	if metadata.Profile != "cost" {
		t.Fatalf("Profile = %q, want %q", metadata.Profile, "cost")
	}
	if metadata.Mode != "disabled" {
		t.Fatalf("Mode = %q, want %q", metadata.Mode, "disabled")
	}
	policy := runtime.Policy()
	if policy.TraceSampleRatio != 1 {
		t.Fatalf("TraceSampleRatio = %v, want 1 for local disabled runtime", policy.TraceSampleRatio)
	}
	if policy.HighLatencyThreshold <= 0 {
		t.Fatalf("HighLatencyThreshold = %v, want positive threshold", policy.HighLatencyThreshold)
	}
	if !policy.DropSuccessfulRequestDurationMetrics {
		t.Fatal("DropSuccessfulRequestDurationMetrics = false, want true for cost profile")
	}
}

func TestBuildVersionFallsBackToDev(t *testing.T) {
	if got := buildVersion(); got == "" {
		t.Fatal("buildVersion() returned empty string, want non-empty version")
	}
}
