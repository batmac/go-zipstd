// implements FS interface from io/fs package for seekable zstd compressed zip files
package zipstd

import (
	"archive/zip"
	"io"
	"io/fs"
	"os"

	seekable "github.com/SaveTheRbtz/zstd-seekable-format-go"
	"github.com/klauspost/compress/zstd"
)

// implement the fs interface
type FS struct {
	zipr     *zip.Reader
	zstddec  *zstd.Decoder
	seekable seekable.Reader
	source   io.ReadCloser
}

// Open implements fs.FS
func (z *FS) Open(name string) (fs.File, error) {
	return z.zipr.Open(name)
}

// Close frees resources associated with the FS file
func (z *FS) Close() error {
	if err := z.seekable.Close(); err != nil {
		return err
	}
	z.zstddec.Close()
	if err := z.source.Close(); err != nil {
		return err
	}

	return nil
}

// Open opens a seekable zstd compressed zip file and returns a fs.FS (.zip.zst)
// zip files should use the "Store" compression method to let zstd do the compression,
// you can use the zipstd/cmd/create.go tool to create such a zip file.
func Open(path string) (*FS, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	dec, err := zstd.NewReader(nil)
	if err != nil {
		return nil, err
	}
	sr, err := seekable.NewReader(f, dec)
	if err != nil {
		return nil, err
	}

	// find the size of the zip file
	n, err := sr.Seek(0, io.SeekEnd)
	if err != nil {
		sr.Close()
		dec.Close()
		f.Close()
		return nil, err
	}
	// rewind
	_, err = sr.Seek(0, io.SeekStart)
	if err != nil {
		sr.Close()
		dec.Close()
		f.Close()
		return nil, err
	}

	zipr, err := zip.NewReader(sr, n)
	if err != nil {
		sr.Close()
		dec.Close()
		f.Close()
		return nil, err
	}

	return &FS{
		zipr:     zipr,
		zstddec:  dec,
		seekable: sr,
		source:   f,
	}, nil
}
