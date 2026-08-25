package major

import (
	"fmt"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"

	"github.com/marwan-at-work/mod"
)

// Run upgrades or downgrades a module path and
// all of its dependencies.
func Run(dir, op, modName string, tag int) error {

	client := true
	var modFile *modfile.File
	modFile, err := mod.GetModFile(dir)
	if err != nil {
		return fmt.Errorf("could not get go.mod file: %w", err)
	}
	if modName == "" {
		client = false
		modName = modFile.Module.Mod.Path
	}
	separator := getSeparator(modName)
	var newModPath string
	switch op {
	case "upgrade":
		newModPath = getNext(tag, modName, separator)
	case "downgrade":
		newModPath = getPrevious(modName, separator)
	}
	// Force -mod=mod so package loading does not fail on a repository that
	// vendors its dependencies. During a major upgrade the go.mod already
	// points at the new /vN path while vendor/modules.txt still records the
	// old one; under the default -mod=vendor this inconsistency makes
	// `go list` error out and no imports get rewritten. -mod=mod ignores the
	// vendor directory for loading, which is all we need to enumerate and
	// rewrite the source files.
	c := &packages.Config{
		Mode:       packages.NeedName | packages.NeedFiles,
		Tests:      true,
		Dir:        dir,
		BuildFlags: []string{"-mod=mod"},
	}
	pkgs, err := packages.Load(c, "./...")
	if err != nil {
		return fmt.Errorf("could not load package: %w", err)
	}
	ids := map[string]struct{}{}
	files := map[string]struct{}{}
	for _, p := range pkgs {
		if _, ok := ids[p.ID]; ok {
			continue
		}
		ids[p.ID] = struct{}{}
		err = updateImportPath(p, modName, newModPath, separator, files)
		if err != nil {
			return err
		}
	}
	if client {
		// It would be nicer to use modFile.AddRequire
		// and modFile.DropRequire, but they do not
		// preserve block location.
		var majorExists bool
		for _, req := range modFile.Require {
			if req.Mod.Path == newModPath {
				majorExists = true
				break
			}
		}
		if majorExists {
			return nil
		} else {
			for _, req := range modFile.Require {
				if req.Mod.Path == modName {
					req.Mod.Path = newModPath
					req.Mod.Version = "latest"
					break
				}
			}
			modFile.SetRequire(modFile.Require)
		}
	} else {
		modFile.Module.Syntax.Token[1] = newModPath
	}
	bts, err := modFile.Format()
	if err != nil {
		return fmt.Errorf("could not format go.mod file with new import path: %w", err)
	}
	err = os.WriteFile(filepath.Join(dir, "go.mod"), bts, 0660)
	if err != nil {
		return fmt.Errorf("could not rewrite go.mod file: %w", err)
	}
	if client {
		fmt.Println("running go mod tidy...")
		cmd := exec.Command("go", "mod", "tidy")
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return syncVendor(dir)
}

// syncVendor re-syncs the vendor directory after a major upgrade. Loading was
// done with -mod=mod, so the vendor tree still records the pre-upgrade module
// path and would leave the repository in an inconsistent state. If there is no
// vendor directory there is nothing to do.
func syncVendor(dir string) error {
	if _, err := os.Stat(filepath.Join(dir, "vendor")); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	fmt.Println("running go mod vendor...")
	cmd := exec.Command("go", "mod", "vendor")
	cmd.Dir = dir
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

func isGopkgin(s string) bool {
	return strings.HasPrefix(s, "gopkg.in/")
}

func getSeparator(s string) string {
	if isGopkgin(s) {
		return "."
	}
	return "/"
}

func getNext(tagNum int, s, sep string) string {
	ss := strings.Split(s, sep)
	num, isMajor := versionSuffix(ss)
	if !isMajor {
		if tagNum != 0 {
			return s + sep + "v" + strconv.Itoa(tagNum)
		}
		return s + sep + "v2"
	}

	newV := num + 1
	if tagNum != 0 {
		newV = tagNum
	}
	return strings.Join(ss[:len(ss)-1], sep) + sep + "v" + strconv.Itoa(newV)
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

func getPrevious(s, sep string) string {
	ss := strings.Split(s, sep)
	num, isMajor := versionSuffix(ss)
	if !isMajor || isGopkgin(s) && num == 0 {
		return s
	}

	if num == 2 && !isGopkgin(s) {
		return strings.Join(ss[:len(ss)-1], sep)
	}

	newV := num - 1
	return strings.Join(ss[:len(ss)-1], sep) + sep + "v" + strconv.Itoa(newV)
}

func updateImportPath(p *packages.Package, old, new, sep string, files map[string]struct{}) error {

	goFileNames := append(p.GoFiles, p.IgnoredFiles...)
	for _, goFileName := range goFileNames {
		// The file list can include other files such as C files that cannot be parsed,
		// skip anything that isn't a .go file.
		if !strings.HasSuffix(goFileName, ".go") {
			continue
		}

		if _, ok := files[goFileName]; ok {
			continue
		}
		files[goFileName] = struct{}{}

		if err := rewriteFile(goFileName, old, new, sep); err != nil {
			return err
		}
	}

	return nil
}

// rewriteFile parses a single .go file and rewrites any import whose path is
// old or is prefixed by old+sep to the corresponding new path.
func rewriteFile(goFileName, old, new, sep string) error {
	fset := token.NewFileSet()
	parsed, err := parser.ParseFile(fset, goFileName, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("could not parse go file %v: %w", goFileName, err)
	}

	var rewritten bool
	oldPrefix := old + sep
	for _, i := range parsed.Imports {
		imp, err := strconv.Unquote(i.Path.Value)
		if err != nil {
			continue
		}
		if strings.HasPrefix(imp, oldPrefix) || imp == old {
			// Only rewrite the leading module-path segment so that a
			// module path appearing as a non-leading substring is left
			// untouched.
			newImp := new + imp[len(old):]
			rewrote := astutil.RewriteImport(fset, parsed, imp, newImp)
			if rewrote {
				rewritten = true
			}
		}
	}
	if !rewritten {
		return nil
	}

	f, err := os.Create(goFileName)
	if err != nil {
		return fmt.Errorf("could not create go file %v: %w", goFileName, err)
	}
	err = format.Node(f, fset, parsed)
	f.Close()
	if err != nil {
		return fmt.Errorf("could not rewrite go file %v: %w", goFileName, err)
	}

	return nil
}
