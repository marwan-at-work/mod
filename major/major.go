package major

import (
	"flag"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/marwan-at-work/vgop/modfile"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

var tagNum = flag.Int("t", 0, "upgrade to a specific tag version instead of incrementally")

// Run upgrades or downgrades a module path and
// all of its dependencies.
func Run() error {
	flag.Parse()
	op := getOperation()
	modFile, err := getModFile()
	if err != nil {
		return errors.Wrap(err, "could not get go.mod file")
	}
	modName := modFile.Module.Mod.Path
	var newModPath string
	switch op {
	case "upgrade":
		newModPath = getNext(modName)
	case "downgrade":
		newModPath = getPrevious(modName)
	}

	c := &packages.Config{Mode: packages.LoadSyntax, Tests: true}
	pkgs, err := packages.Load(c, "./...")
	if err != nil {
		return errors.Wrap(err, "could not load package")
	}

	for _, p := range pkgs {
		err = updateImportPath(p, modName, newModPath)
		if err != nil {
			return err
		}
	}
	modFile.Module.Syntax.Token[1] = newModPath
	bts, err := modFile.Format()
	if err != nil {
		return errors.Wrap(err, "could not format go.mod file with new import path")
	}
	err = ioutil.WriteFile("go.mod", bts, 0660)
	if err != nil {
		return errors.Wrap(err, "could not rewrite go.mod file")
	}

	return nil
}

func getOperation() string {
	args := flag.Args()
	if len(args) != 1 {
		log.Fatal("Use: mod upgrade|downgrade")
	}

	op := args[0]
	if op != "upgrade" && op != "downgrade" {
		log.Fatal("unknown command " + op)
	}

	return op
}

func getNext(s string) string {
	ss := strings.Split(s, "/")
	num, isMajor := versionSuffix(ss)
	if !isMajor {
		if *tagNum != 0 {
			return s + "/v" + strconv.Itoa(*tagNum)
		}
		return s + "/v2"
	}

	newV := num + 1
	if *tagNum != 0 {
		newV = *tagNum
	}
	return strings.Join(ss[:len(ss)-1], "/") + "/v" + strconv.Itoa(newV)
}

func versionSuffix(ss []string) (int, bool) {
	last := ss[len(ss)-1]
	if !strings.HasPrefix(last, "v") {
		return 0, false
	}

	numStr := last[1:]
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, false
	}

	return num, true
}

func getPrevious(s string) string {
	ss := strings.Split(s, "/")
	num, isMajor := versionSuffix(ss)
	if !isMajor {
		return s
	}

	if num == 2 {
		return strings.Join(ss[:len(ss)-1], "/")
	}

	newV := num - 1
	return strings.Join(ss[:len(ss)-1], "/") + "/v" + strconv.Itoa(newV)
}

func updateImportPath(p *packages.Package, old, new string) error {
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

func getModFile() (*modfile.File, error) {
	bts, err := ioutil.ReadFile("go.mod")
	if err != nil {
		return nil, errors.Wrap(err, "could not open go.mod file")
	}
	dir, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err, "could not get CWD")
	}
	f, err := modfile.Parse(filepath.Join(dir, "go.mod"), bts, nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse go.mod file")
	}
	return f, nil
}
