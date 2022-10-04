package zipstd

import (
	"io/fs"
	"strings"
	"testing"
)

func TestOpen(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *FS
		wantErr bool
	}{
		{"fake", args{"fake"}, nil, true},
		{"simple.zip.zst", args{"testdata/simple.zip.zst"}, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Open(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Open() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				return
			}

			var f fs.FS = got

			// list files in f
			err = fs.WalkDir(f, ".", func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				i, err := d.Info()
				if err != nil {
					return err
				}
				t.Logf("path: %s (%d bytes), type=%s", path, i.Size(), d.Type())
				if d.Type().IsRegular() && (strings.HasSuffix(path, ".txt") || strings.HasSuffix(path, ".md")) {
					content, err := fs.ReadFile(f, path)
					if err != nil {
						return err
					}
					t.Logf("content: %s", content)
				}

				return nil
			})
			if err != nil {
				t.Errorf("WalkDir() error = %v", err)
			}

			if err := got.Close(); err != nil {
				t.Errorf("Close() error = %v", err)
			}
		})
	}
}
