//go:build ignore
// +build ignore

package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	seekable "github.com/SaveTheRbtz/zstd-seekable-format-go"
	"github.com/klauspost/compress/zstd"
)

var (
	argVerbose = flag.Bool("v", false, "verbose")
	argLevel   = flag.Int("l", 1, "zstd compression level")
	argList    = flag.Bool("t", false, "list files in archive")
)

// create a zip file, encapsulated within seekable zstd, ready to be used with zipstd package

func main() {
	flag.Parse()

	if flag.NArg() < 2 {
		fmt.Println("usage: create [-v] [-l level] [-t] archive.zip.zst file1 file2 ...")
		return
	}
	archivePath := flag.Arg(0)

	if *argList {
		list(archivePath)
		return
	}

	filePaths := flag.Args()[1:]

	// create archive
	f, err := os.Create(archivePath)
	if err != nil {
		log.Fatal(err)
	}

	enc, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.EncoderLevelFromZstd(*argLevel)))
	if err != nil {
		log.Fatal(err)
	}
	defer enc.Close()

	sw, err := seekable.NewWriter(f, enc)
	if err != nil {
		log.Fatal(err)
	}
	defer sw.Close()

	// create zip writer in w
	zipw := zip.NewWriter(sw)
	defer zipw.Close()

	// add files
	for _, path := range filePaths {
		if *argVerbose {
			log.Printf("adding %s", path)
		}
		err = addFile(zipw, path)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func list(archivePath string) {
	f, err := os.Open(archivePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	// get f size
	fi, err := f.Stat()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(archivePath, "size:", fi.Size())

	dec, err := zstd.NewReader(nil)
	if err != nil {
		log.Fatal(err)
	}
	defer dec.Close()
	sr, err := seekable.NewReader(f, dec)
	if err != nil {
		log.Fatal(err)
	}
	defer sr.Close()

	// find the size of the zip file
	n, err := sr.Seek(0, io.SeekEnd)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("seek to end:", n)
	// rewind
	_, err = sr.Seek(0, io.SeekStart)
	if err != nil {
		log.Fatal(err)
	}

	zipr, err := zip.NewReader(sr, n)
	if err != nil {
		log.Fatal("ziperr:", err)
	}

	// list files
	for _, f := range zipr.File {
		log.Printf("%s", f.Name)
	}
}

func addFile(zipw *zip.Writer, path string) error {
	// add the file in the zip as "zip.Store" (no compression) to let zstd do its job
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	fheader := &zip.FileHeader{
		Name:   path,
		Method: zip.Store,
	}
	w, err := zipw.CreateHeader(fheader)
	if err != nil {
		log.Fatal(err)
	}
	_, err = io.Copy(w, f)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
