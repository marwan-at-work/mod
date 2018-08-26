package major

import "testing"

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
}

func TestGetNext(t *testing.T) {
	for _, tc := range upDownCases {
		t.Run(tc.name, func(t *testing.T) {
			next := getNext(tc.input)
			if next != tc.next {
				t.Fatalf("expected getNext to return %v but got %v", tc.next, next)
			}
			prev := getPrevious(tc.input)
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
