package db

import (
	"path/filepath"
	"testing"
)

func TestCreateDatabase(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	db, err := CreateDatabase(path)
	if err != nil {
		t.Fatalf("CreateDatabase failed: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("db.Ping failed: %v", err)
	}
}
