package replace

import (
	"fmt"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/marwan-at-work/mod"
	"github.com/pkg/errors"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

func Run(dir, oldModName string, newModName string) error {
	var modFile *modfile.File
	modFile, err := mod.GetModFile(dir)
	if err != nil {
		return errors.Wrap(err, "could not get go.mod file")
	}

	// loading the packages
	c := &packages.Config{Mode: packages.NeedName | packages.NeedFiles, Tests: true, Dir: dir}
	pkgs, err := packages.Load(c, "./...")
	if err != nil {
		return errors.Wrap(err, "could not load package")
	}

	// Starting the processing
	ids := map[string]struct{}{}
	files := map[string]struct{}{}
	for _, p := range pkgs {
		if _, ok := ids[p.ID]; ok {
			continue
		}
		ids[p.ID] = struct{}{}
		err = updateImportPath(p, oldModName, newModName, files)
		if err != nil {
			return err
		}
	}

	// check if this new entry is already in the mode file
	var alreadyExists bool
	for _, req := range modFile.Require {
		if req.Mod.Path == newModName {
			alreadyExists = true
			break
		}
	}

	// drop the entry from gomod
	if alreadyExists {
		err = modFile.DropRequire(oldModName)
		if err != nil {
			return fmt.Errorf("error dropping %q: %w", oldModName, err)
		}
	} else {
		for _, req := range modFile.Require {
			if req.Mod.Path == oldModName {
				req.Mod.Path = newModName
				req.Mod.Version = "latest"
				req.Syntax.Token[1] = newModName
				break
			}
		}
		modFile.SetRequire(modFile.Require)
	}

	// Tidy and finish the run
	bts, err := modFile.Format()
	if err != nil {
		return errors.Wrap(err, "could not format go.mod file with new import path")
	}
	err = os.WriteFile(filepath.Join(dir, "go.mod"), bts, 0660)
	if err != nil {
		return errors.Wrap(err, "could not rewrite go.mod file")
	}
	fmt.Println("running go mod tidy...")
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()

	return nil
}

func updateImportPath(p *packages.Package, old, new string, files map[string]struct{}) error {

	goFileNames := append(p.GoFiles, p.IgnoredFiles...)
	for _, goFileName := range goFileNames {

		if _, ok := files[goFileName]; ok {
			continue
		}
		files[goFileName] = struct{}{}

		fset := token.NewFileSet()
		parsed, err := parser.ParseFile(fset, goFileName, nil, parser.ParseComments)
		if err != nil {
			return errors.Wrapf(err, "could not parse go file %v", goFileName)
		}

		var rewritten bool
		for _, i := range parsed.Imports {
			imp := strings.Replace(i.Path.Value, `"`, ``, 2)
			if strings.HasPrefix(imp, fmt.Sprintf("%s", old)) || imp == old {
				newImp := strings.Replace(imp, old, new, 1)
				rewrote := astutil.RewriteImport(fset, parsed, imp, newImp)
				if rewrote {
					rewritten = true
				}
			}
		}
		if !rewritten {
			continue
		}

		f, err := os.Create(goFileName)
		if err != nil {
			return errors.Wrapf(err, "could not create go file %v", goFileName)
		}
		err = format.Node(f, fset, parsed)
		f.Close()
		if err != nil {
			return errors.Wrapf(err, "could not rewrite go file %v", goFileName)
		}
	}

	return nil
}
