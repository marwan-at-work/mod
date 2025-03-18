package mod

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestGetModFile(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(`
module example.com/foo

require example.com/whatever v1.2.3

tool example.com/whatever/cmd/thing

go 1.23.4
`), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	mod, err := GetModFile(dir)
	if err != nil {
		t.Fatal(err)
	}

	if mod.Module.Mod.Path != "example.com/foo" {
		t.Fatalf("unexpected module path: %v", mod.Module.Mod.Path)
	}
}

func TestGetModFileErrors(t *testing.T) {
	t.Run("DoesNotExist", func(t *testing.T) {
		_, err := GetModFile(t.TempDir())
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("SyntaxError", func(t *testing.T) {
		dir := t.TempDir()
		err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("syntax error"), 0o644)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := GetModFile(dir); err == nil {
			t.Fatal("expected error")
		}
	})
}
