package mod

import (
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/marwan-at-work/vgop/modfile"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
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

// LoadPackagesIn returns a generator that yields package structs, or an error
// if there was a problem loading the package in the given directory
func LoadPackagesIn(dir string, tests bool) (<-chan *packages.Package, error) {
	c := &packages.Config{
		Mode:  packages.LoadSyntax,
		Tests: tests,
		Dir:   dir,
	}

	pkgs, err := packages.Load(c, "./...")
	if err != nil {
		return nil, errors.Wrap(err, dir)
	}

	gen := make(chan *packages.Package)

	go func() {
		for _, p := range pkgs {
			gen <- p
		}

		close(gen)
	}()

	return gen, nil
}

// UpdateImportPath takes a package p, iterates its imports and if import 'old'
// is found, replace it with import 'new'
func UpdateImportPath(p *packages.Package, old, new string) error {
	for _, syn := range p.Syntax {
		var rewritten bool

		for _, i := range syn.Imports {
			imp := strings.Replace(i.Path.Value, `"`, ``, 2)
			if strings.HasPrefix(imp, old) {
				newImp := strings.Replace(imp, old, new, 1)
				rewrote := astutil.RewriteImport(p.Fset, syn, imp, newImp)
				if rewrote {
					rewritten = true
				}
			}
		}

		if !rewritten {
			continue
		}

		goFileName := p.Fset.File(syn.Pos()).Name()

		f, err := os.Create(goFileName)
		if err != nil {
			return errors.Wrapf(err, "could not create go file %v", goFileName)
		}
		err = format.Node(f, p.Fset, syn)
		f.Close()
		if err != nil {
			return errors.Wrapf(err, "could not rewrite go file %v", goFileName)
		}
	}

	return nil
}
