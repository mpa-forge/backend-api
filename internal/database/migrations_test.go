package database

import (
	"io/fs"
	"sort"
	"testing"
)

func TestEmbeddedSchemaFilesIncludeMigrationsAndSeeds(t *testing.T) {
	migrations, err := fs.Glob(schemaFiles, migrationsDir+"/*.sql")
	if err != nil {
		t.Fatalf("Glob migrations error = %v", err)
	}
	seeds, err := fs.Glob(schemaFiles, seedsDir+"/*.sql")
	if err != nil {
		t.Fatalf("Glob seeds error = %v", err)
	}

	sort.Strings(migrations)
	sort.Strings(seeds)

	if len(migrations) == 0 {
		t.Fatal("expected at least one embedded migration file")
	}
	if len(seeds) == 0 {
		t.Fatal("expected at least one embedded seed file")
	}
}
