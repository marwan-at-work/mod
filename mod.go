package mod

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/mod/modfile"
)

// GetModFile returns an AST of the given directory's go.mod file
// and returns err if file is not found.
func GetModFile(dir string) (*modfile.File, error) {
	bts, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return nil, fmt.Errorf("could not open go.mod file: %w", err)
	}
	f, err := modfile.Parse(filepath.Join(dir, "go.mod"), bts, nil)
	if err != nil {
		return nil, fmt.Errorf("could not parse go.mod file: %w", err)
	}
	return f, nil
}
