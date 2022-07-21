package main // import "hello-zip"

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func compress(source, target string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			header.Name = strings.TrimPrefix(path, source+"\\")
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		if header.Name != baseDir && header.Name != baseDir+"/" {
			writer, err := archive.CreateHeader(header)
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

			_, err = io.Copy(writer, file)
			return err
		}

		return nil
	})

	return err
}

func decompress(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	if dest == "" || dest == "." || dest == "."+string(os.PathSeparator) {
		dest = ""
	}

	os.MkdirAll(dest, 0755)

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		// Check for ZipSlip (Directory traversal)
		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", path)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}

func jobMain() {
	seekPath := "samples"
	outputName := seekPath + ".zip"
	inputName := "archive.zip"
	destination := ""

	switch len(os.Args) {
	case 1:
		err := compress(seekPath, outputName)
		if err != nil {
			panic(err)
		}
	case 2:
		info, err := os.Stat(os.Args[1])
		if err != nil {
			panic(err)
		}

		if info.IsDir() {
			seekPath = os.Args[1]
			err := compress(seekPath, outputName)
			if err != nil {
				panic(err)
			}
		} else {
			inputName = os.Args[1]
			filename := filepath.Base(inputName)
			destination = strings.TrimSuffix(filename, filepath.Ext(filename))
			err := decompress(inputName, destination)
			if err != nil {
				panic(err)
			}
		}
	default:
		log.Println("Usage: hello-zip [directory (=compress) or filename (=decompress)]")
	}
}

func main() {
	jobMain()
}
