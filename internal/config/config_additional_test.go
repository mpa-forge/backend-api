package config

import (
	"strings"
	"testing"
)

func TestLoadDatabaseURLFromEnvRejectsEmptySplitValues(t *testing.T) {
	testCases := []struct {
		name        string
		emptyKey    string
		wantProblem string
	}{
		{name: "host", emptyKey: "DB_HOST", wantProblem: "DB_HOST cannot be empty"},
		{name: "name", emptyKey: "DB_NAME", wantProblem: "DB_NAME cannot be empty"},
		{name: "user", emptyKey: "DB_USER", wantProblem: "DB_USER cannot be empty"},
		{name: "password", emptyKey: "DB_PASSWORD", wantProblem: "DB_PASSWORD cannot be empty"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("DATABASE_URL", "")
			t.Setenv("DB_HOST", "/cloudsql/example-project:europe-west1:platform-rc-db")
			t.Setenv("DB_NAME", "platform_rc")
			t.Setenv("DB_USER", "api_user")
			t.Setenv("DB_PASSWORD", "secret")
			t.Setenv(tc.emptyKey, "")

			var problems []string
			databaseURL := LoadDatabaseURLFromEnv(&problems)
			if databaseURL != "" {
				t.Fatalf("LoadDatabaseURLFromEnv() = %q, want empty string", databaseURL)
			}

			message := strings.Join(problems, "\n")
			if !strings.Contains(message, tc.wantProblem) {
				t.Fatalf("LoadDatabaseURLFromEnv() problems = %q, want substring %q", message, tc.wantProblem)
			}
		})
	}
}

func TestComposePostgresURLNormalizesDatabaseNameAndCloudSQLSocketHost(t *testing.T) {
	testCases := []struct {
		name    string
		dbHost  string
		dbName  string
		wantDSN string
	}{
		{
			name:    "tcp host with leading slash in db name",
			dbHost:  "10.10.0.4:5432",
			dbName:  "/platform_rc",
			wantDSN: "postgres://api_user:secret@10.10.0.4:5432/platform_rc?sslmode=disable",
		},
		{
			name:    "cloud sql socket with leading slash in db name",
			dbHost:  "/cloudsql/example-project:europe-west1:platform-rc-db",
			dbName:  "/platform_rc",
			wantDSN: "postgres://api_user:secret@/platform_rc?host=%2Fcloudsql%2Fexample-project%3Aeurope-west1%3Aplatform-rc-db&sslmode=disable",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := composePostgresURL(tc.dbHost, tc.dbName, "api_user", "secret")
			if got != tc.wantDSN {
				t.Fatalf("composePostgresURL() = %q, want %q", got, tc.wantDSN)
			}
		})
	}
}
