package main_test

import (
	"flag"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

var (
	cleanup    = flag.Bool("cleanup", false, "remove any files/directories created during testing")
	moduleName = flag.String("name", "test_module", "name of the module created for testing")
)

var tmpls = []struct {
	name string
	pkg  string
	path string
}{
	{
		"main",
		"main",
		"./",
	},
	{
		"pkg_a_1",
		"pkg_a",
		"pkg_a/",
	},
	{
		"pkg_a_2",
		"pkg_a",
		"pkg_a/",
	},
	{
		"pkg_b_a_1",
		"pkg_b_a",
		"pkg_a/pkg_b_a/",
	},
}

// mkworkspace returns a teardown function, the tmp workspace root dir and an error if any
func mkworkspace(dir string) (func(), *string, error) {
	if err := os.Mkdir(dir, os.ModePerm); err != nil {
		return nil, nil, err
	}

	return func() { os.RemoveAll(dir) }, &dir, nil
}

// writesrc creates .go source files from templates in the directory ./template/
func writesrc(modroot, oldimport, modname string) error {
	for _, srctmpl := range tmpls {
		pkgpath := filepath.Join(modroot, srctmpl.path)

		if _, err := os.Stat(pkgpath); os.IsNotExist(err) {
			os.MkdirAll(pkgpath, os.ModePerm)
		}

		tmpl, err := template.ParseFiles(filepath.Join("template/", srctmpl.name+".tmpl"))
		if err != nil {
			return err
		}

		tmplfd, err := os.Create(filepath.Join(modroot, srctmpl.path, srctmpl.name+".go"))
		if err != nil {
			return err
		}
		defer tmplfd.Close()

		var args = map[string]string{
			"PackageName": srctmpl.pkg,
			"ImportPath":  oldimport,
			"ModuleName":  modname,
		}

		if err := tmpl.Execute(tmplfd, args); err != nil {
			return err
		}
	}

	return nil
}

// gomodinit will initialise the directory root as a go module
func gomodinit(modname, root string) error {
	c := exec.Command("go", "mod", "init", modname)

	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Dir = root

	err := c.Run()
	if err != nil {
		return err
	}

	return nil
}

func TestForkRewrite(t *testing.T) {
	modImportPathNew := "github.com/new/" + *moduleName
	modImportPathOld := "github.com/old/" + *moduleName

	fmt.Println("old module:", modImportPathOld)
	fmt.Println("new module:", modImportPathNew)

	teardown, dir, err := mkworkspace("test_module")
	if err != nil {
		if teardown != nil {
			teardown()
		}
		t.Fatal(err)
	}

	var root string

	if dir == nil {
		t.Fatal("the test module root directory is nil")
	}

	root = *dir

	if *cleanup {
		defer teardown()
	}

	if err := writesrc(root, modImportPathOld, *moduleName); err != nil {
		t.Fatal(err)
	}

	if err := gomodinit(modImportPathNew, root); err != nil {
		t.Fatal(err)
	}
}
