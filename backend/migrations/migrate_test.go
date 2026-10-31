package migrations

import (
	"io/fs"
	"strings"
	"testing"
)

func TestEmbeddedMigrations(t *testing.T) {
	entries, err := fs.ReadDir(FS, ".")
	if err != nil {
		t.Fatalf("failed to read embedded migrations dir: %v", err)
	}

	var upCount, downCount int
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasSuffix(name, ".up.sql") {
			upCount++
		} else if strings.HasSuffix(name, ".down.sql") {
			downCount++
		}
	}

	expectedPairs := 5
	if upCount != expectedPairs {
		t.Errorf("expected %d up migration files, got %d", expectedPairs, upCount)
	}

	if downCount != expectedPairs {
		t.Errorf("expected %d down migration files, got %d", expectedPairs, downCount)
	}
}
