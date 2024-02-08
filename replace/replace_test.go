package replace

import (
	"os"
	"path"
	"testing"

	cp "github.com/otiai10/copy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	type args struct {
		dir        string
		oldModName string
		newModName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		// list of file we want to validate for the replacement
		wantCheckFile []string
	}{
		{
			name: "example1",
			args: args{
				dir:        "testdata/example1",
				oldModName: "github.com/go-playground/webhooks/v6",
				newModName: "github.com/xnok/webhooks/v6",
			},
			wantErr: false,
			wantCheckFile: []string{
				"go.mod",
				"example.go",
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				dir := t.TempDir()
				err := cp.Copy(tt.args.dir, dir)
				require.NoError(t, err)

				if err := Run(dir, tt.args.oldModName, tt.args.newModName); (err != nil) != tt.wantErr {
					assert.NoError(t, err)
					for i := range tt.wantCheckFile {
						fileCheck(t, path.Join(dir, tt.wantCheckFile[i]), tt.args.newModName, tt.args.oldModName)
					}
				}
			},
		)
	}
}

func fileCheck(t *testing.T, file, new, old string) {
	t.Helper()
	// read the whole file at once
	b, err := os.ReadFile(file)
	if err != nil {
		panic(err)
	}
	s := string(b)
	// //check whether s contains substring text
	assert.Contains(t, s, new)
	assert.NotContains(t, s, old)
}
