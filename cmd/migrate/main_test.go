package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestMainFailsWhenSplitDatabaseConfigIsIncomplete(t *testing.T) {
	t.Parallel()

	result := runMigrateMain(t, []string{"up"}, map[string]string{
		"DATABASE_URL": "",
		"DB_HOST":      "/cloudsql/example-project:europe-west1:platform-rc-db",
		"DB_NAME":      "platform_rc",
	})

	if result.exitCode == 0 {
		t.Fatal("exitCode = 0, want non-zero")
	}

	for _, want := range []string{
		"invalid database configuration:",
		"DB_USER is required",
		"DB_PASSWORD is required",
	} {
		if !strings.Contains(result.stderr, want) {
			t.Fatalf("stderr = %q, want substring %q", result.stderr, want)
		}
	}
}

func TestMainShowsUsageForUnknownCommandAfterLoadingSplitDatabaseConfig(t *testing.T) {
	t.Parallel()

	result := runMigrateMain(t, []string{"bogus"}, map[string]string{
		"DATABASE_URL": "",
		"DB_HOST":      "/cloudsql/example-project:europe-west1:platform-rc-db",
		"DB_NAME":      "platform_rc",
		"DB_USER":      "api_user",
		"DB_PASSWORD":  "secret",
	})

	if result.exitCode == 0 {
		t.Fatal("exitCode = 0, want non-zero")
	}

	want := "usage: go run ./cmd/migrate [up|down|seed|prepare]"
	if !strings.Contains(result.stderr, want) {
		t.Fatalf("stderr = %q, want substring %q", result.stderr, want)
	}
}

func TestMainHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	args := os.Getenv("MIGRATE_TEST_ARGS")
	if args == "" {
		os.Exit(0)
	}

	os.Args = append([]string{"migrate"}, strings.Split(args, "\x00")...)
	main()
}

type migrateMainResult struct {
	exitCode int
	stderr   string
}

func runMigrateMain(t *testing.T, args []string, env map[string]string) migrateMainResult {
	t.Helper()

	cmdArgs := []string{"-test.run=TestMainHelperProcess", "--"}
	cmd := exec.Command(os.Args[0], cmdArgs...)
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1", "MIGRATE_TEST_ARGS="+strings.Join(args, "\x00"))
	for key, value := range env {
		cmd.Env = append(cmd.Env, key+"="+value)
	}

	output, err := cmd.CombinedOutput()
	result := migrateMainResult{
		stderr: string(output),
	}

	if err == nil {
		return result
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		result.exitCode = exitErr.ExitCode()
		return result
	}

	t.Fatalf("runMigrateMain() error = %v", err)
	return migrateMainResult{}
}
