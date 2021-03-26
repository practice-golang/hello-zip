package main // import "hello-zip"

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Wk struct {
	wr *zip.Writer
}

func (wk Wk) Walker(path string, info os.FileInfo, err error) error {
	fmt.Printf("Crawling: %#v\n", path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return nil
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	f, err := wk.wr.Create(path)
	if err != nil {
		return err
	}

	_, err = io.Copy(f, file)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	searchPath := "./samples"
	outputName := "output.zip"

	file, err := os.Create(outputName)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	walker := func(path string, info os.FileInfo, err error) error {
		fwr := Wk{
			wr: zip.NewWriter(file),
		}
		defer fwr.wr.Close()

		err = fwr.Walker(path, info, err)
		return err
	}

	err = filepath.Walk(searchPath, walker)
	if err != nil {
		panic(err)
	}
}
