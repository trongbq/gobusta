package engine

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}
	return err == nil, nil
}

func copyDir(src, dest string) error {
	// Make sure source directory exists
	srcStat, err := os.Stat(src)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err != nil {
		return errors.New(fmt.Sprintf("Can not copy from %v to %v, source does not exist", src, dest, src))
	}

	// Check destination is valid, create new if it does not exist
	_, err = os.Stat(dest)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err != nil {
		// Destination does not exists, create it
		err = os.Mkdir(dest, srcStat.Mode())
		if err != nil {
			return err
		}
	}
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		log.Printf("Copy file %v", path[len(src):])
		return copyFile(path, filepath.Join(dest, path[len(src):]))
	})
}

func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		log.Printf("Can not open source file %v", src)
		return err
	}

	// Make sure to create parent directory first before creating the file
	os.MkdirAll(filepath.Dir(dest), os.ModePerm)

	out, err := os.Create(dest)
	if err != nil {
		log.Printf("Can not create destination file %v", dest)
		return err
	}

	_, err = io.Copy(out, in)
	if err != nil {
		log.Printf("Can not copy file from %v to %v", src, dest)
		return err
	}

	err = in.Close()
	if err != nil {
		return err
	}
	return out.Close()
}
