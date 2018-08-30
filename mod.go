package mod

import (
	"io/ioutil"
	"path/filepath"

	"github.com/marwan-at-work/vgop/modfile"
	"github.com/pkg/errors"
)

// GetModFile returns an AST of the given directory's go.mod file
// and returns err if file is not found.
func GetModFile(dir string) (*modfile.File, error) {
	bts, err := ioutil.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return nil, errors.Wrap(err, "could not open go.mod file")
	}
	f, err := modfile.Parse(filepath.Join(dir, "go.mod"), bts, nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse go.mod file")
	}
	return f, nil
}
