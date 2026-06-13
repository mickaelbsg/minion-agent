package storage

import (
    "os"
    "path/filepath"
    "testing"
)

func TestStorageCreatesDBDir(t *testing.T) {
    dir := t.TempDir()
    dbPath := filepath.Join(dir, "nonexistent", "minion.db")
    s, err := New(dbPath)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    defer s.DB.Close()
    if _, err := os.Stat(filepath.Dir(dbPath)); os.IsNotExist(err) {
        t.Fatalf("directory %s was not created", filepath.Dir(dbPath))
    }
}
