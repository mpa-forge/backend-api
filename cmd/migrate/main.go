package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/mpa-forge/backend-api/internal/config"
	"github.com/mpa-forge/backend-api/internal/database"
)

func main() {
	if len(os.Args) != 2 {
		fail("usage: go run ./cmd/migrate [up|down|seed|prepare]")
	}

	var problems []string
	databaseURL := config.LoadDatabaseURLFromEnv(&problems)
	if databaseURL == "" {
		fail(fmt.Sprintf("invalid database configuration:\n- %s", strings.Join(problems, "\n- ")))
	}

	migrator := database.NewMigrator(databaseURL)
	ctx := context.Background()

	var err error
	switch os.Args[1] {
	case "up":
		err = migrator.Up(ctx)
	case "down":
		err = migrator.Down(ctx)
	case "seed":
		err = migrator.Seed(ctx)
	case "prepare":
		err = migrator.Prepare(ctx)
	default:
		fail("usage: go run ./cmd/migrate [up|down|seed|prepare]")
	}

	if err != nil {
		fail(err.Error())
	}
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
