package bundler

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	srcF, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcF.Close()

	dstF, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstF.Close()
	_, err = io.Copy(dstF, srcF)
	return err
}

// ComputeFileHash computes a hash for the content of a file
func ComputeFileHash(fname string) (string, error) {
	f, err := os.Open(fname)
	if err != nil {
		return "", err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}
	hb := sha256.Sum256(b)
	return hex.EncodeToString(hb[:]), nil
}

// CopyDir copies one dir from source to dest
// ignoring symlinks, and setting a default mode
func CopyDir(src, dst string) error {
	absSrc, err := filepath.Abs(src)
	if err != nil {
		return err
	}
	lenAbsSrc := len(absSrc)
	absDst, err := filepath.Abs(dst)
	if err != nil {
		return err
	}

	// check that dst is not a subdir of srs to avoid recursive
	// infinite copy
	if strings.HasPrefix(absDst, absSrc) {
		return fmt.Errorf("destination is under source directory")
	}

	err = filepath.Walk(absSrc,
		func(path string, info os.FileInfo, err error) error {
			if !strings.HasPrefix(path, absSrc) {
				// avoiding symlinks ?
				return nil
			}
			if err != nil {
				return err
			}
			relPath := path[lenAbsSrc:]
			dst := filepath.Join(absDst, relPath)
			if info.IsDir() {
				if err := os.MkdirAll(dst, info.Mode()); err != nil {
					return err
				}
			} else if info.Mode().IsRegular() {
				if err := CopyFile(path, dst); err != nil {
					return err
				}
			}
			return nil
		})
	return err
}
