package fork

import (
	"github.com/marwan-at-work/mod"
	"github.com/pkg/errors"
)

// Run will overwrite oldImportPath in all src files with the module name found in dir
func Run(dir, oldImportPath string) error {
	mf, err := mod.GetModFile(dir)
	if err != nil {
		return errors.Wrap(err, "could not get go.mod file")
	}

	pkgs, err := mod.LoadPackagesIn(dir, true)
	if err != nil {
		return errors.Wrap(err, "could not load package syntax")
	}

	for p := range pkgs {
		err = mod.UpdateImportPath(p, oldImportPath, mf.Module.Mod.Path)
		if err != nil {
			return errors.Wrap(err, "could not update import path")
		}
	}

	return nil
}
