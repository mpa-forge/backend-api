package main

import (
	"context"
	"fmt"
	"os"

	"github.com/mpa-forge/backend-api/internal/database"
)

func main() {
	if len(os.Args) != 2 {
		fail("usage: go run ./cmd/migrate [up|down|seed|prepare]")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		fail("DATABASE_URL is required")
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
