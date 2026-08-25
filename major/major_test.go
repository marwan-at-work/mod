package major

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var upDownCases = []struct {
	name  string
	input string
	next  string
	prev  string
}{
	{
		"non major",
		"github.com/marwan-at-work/mod",
		"github.com/marwan-at-work/mod/v2",
		"github.com/marwan-at-work/mod",
	},
	{
		"major",
		"mod/v2",
		"mod/v3",
		"mod",
	},
	{
		"double digit",
		"bitbucket.org/usr/pkg/sub/v45",
		"bitbucket.org/usr/pkg/sub/v46",
		"bitbucket.org/usr/pkg/sub/v44",
	},
	{
		"ninety nine",
		"mod/sub/v99",
		"mod/sub/v100",
		"mod/sub/v98",
	},
	{
		"gopkg.in v0",
		"gopkg.in/yaml.v0",
		"gopkg.in/yaml.v1",
		"gopkg.in/yaml.v0",
	},
	{
		"gopkg.in v1",
		"gopkg.in/yaml.v1",
		"gopkg.in/yaml.v2",
		"gopkg.in/yaml.v0",
	},
	{
		"gopkg.in v2",
		"gopkg.in/yaml.v2",
		"gopkg.in/yaml.v3",
		"gopkg.in/yaml.v1",
	},
}

func TestGetNext(t *testing.T) {
	for _, tc := range upDownCases {
		t.Run(tc.name, func(t *testing.T) {
			sep := getSeparator(tc.input)
			next := getNext(0, tc.input, sep)
			if next != tc.next {
				t.Fatalf("expected getNext to return %v but got %v", tc.next, next)
			}
			prev := getPrevious(tc.input, sep)
			if prev != tc.prev {
				t.Fatalf("expected getPrevious to return %v but got %v", tc.prev, prev)
			}
		})
	}
}

var versionSuffixCases = []struct {
	name    string
	input   []string
	output  int
	isMajor bool
}{
	{"non major", []string{"mod"}, 0, false},
	{"major", []string{"mod", "v2"}, 2, true},
	{"major double digit", []string{"mod", "v39"}, 39, true},
	{"major incorrect number", []string{"mod", "vxyz"}, 0, false},
}

func TestVersionSuffix(t *testing.T) {
	for _, tc := range versionSuffixCases {
		t.Run(tc.name, func(t *testing.T) {
			output, isMajor := versionSuffix(tc.input)
			if tc.output != output {
				t.Fatalf("expected output to be %v but got %v", tc.output, output)
			}
			if tc.isMajor != isMajor {
				t.Fatalf("expected major boolean to be %v but got %v", tc.isMajor, isMajor)
			}
		})
	}
}

func TestRewriteFile(t *testing.T) {
	const src = `package p

import (
	root "example.com/foo/armnetwork/v9"
	sub "example.com/foo/armnetwork/v9/fake"
	other "example.com/foo/armnetworkextra/v9"
	unrelated "example.com/bar/v9"
)

var _ = root.X
var _ = sub.X
var _ = other.X
var _ = unrelated.X
`
	dir := t.TempDir()
	file := filepath.Join(dir, "p.go")
	if err := os.WriteFile(file, []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}

	old := "example.com/foo/armnetwork/v9"
	new := "example.com/foo/armnetwork/v10"
	if err := rewriteFile(file, old, new, "/"); err != nil {
		t.Fatalf("rewriteFile returned error: %v", err)
	}

	got, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	gotStr := string(got)

	// The module path and its sub-path must be rewritten.
	for _, want := range []string{
		`"example.com/foo/armnetwork/v10"`,
		`"example.com/foo/armnetwork/v10/fake"`,
	} {
		if !strings.Contains(gotStr, want) {
			t.Fatalf("expected rewritten import %s, got:\n%s", want, gotStr)
		}
	}
	// A path that merely shares a prefix, and an unrelated path, must be
	// left untouched (segment-boundary matching, not substring matching).
	for _, unchanged := range []string{
		`"example.com/foo/armnetworkextra/v9"`,
		`"example.com/bar/v9"`,
	} {
		if !strings.Contains(gotStr, unchanged) {
			t.Fatalf("expected untouched import %s, got:\n%s", unchanged, gotStr)
		}
	}
	// No stale /v9 armnetwork import should remain.
	if strings.Contains(gotStr, `"example.com/foo/armnetwork/v9"`) {
		t.Fatalf("stale /v9 import left behind:\n%s", gotStr)
	}
}

func TestSyncVendorNoVendorDir(t *testing.T) {
	// With no vendor directory present, syncVendor must be a no-op and must
	// not attempt to run `go mod vendor`.
	dir := t.TempDir()
	if err := syncVendor(dir); err != nil {
		t.Fatalf("syncVendor with no vendor dir returned error: %v", err)
	}
}
