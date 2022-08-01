package bundler

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path"
	"path/filepath"
	"testing"
)

// createFile creates a file and returns the sha256 content hash
func createFile(fname string, content string) (string, error) {
	bcontent := []byte(content)
	sha := sha256.Sum256(bcontent)
	hash := hex.EncodeToString(sha[:])
	err := os.WriteFile(fname, []byte(bcontent), 0666)
	if err != nil {
		return hash, err
	}
	return hash, nil
}

func TestCopyFileAndHash(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	content := "this is a test"
	srcFile := path.Join(srcDir, "foo.txt")
	dstFile := path.Join(dstDir, "foo.txt")

	hash, err := createFile(srcFile, content)
	if err != nil {
		t.Errorf("cannot create test file")
		return
	}

	CopyFile(srcFile, dstFile)
	readHash, err := ComputeFileHash(dstFile)
	if err != nil {
		t.Errorf("error happened computing hash for %s : %s", dstFile, err)
		return
	}

	if readHash != hash {
		t.Errorf("copied file hash does not match")
		return
	}
}

func TestCopyDir(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// create subfolders
	os.MkdirAll(filepath.Join(srcDir, "a", "b", "c"), 0777)
	_, _ = createFile(filepath.Join(srcDir, "a", "foo.txt"), "foo")
	_, _ = createFile(filepath.Join(srcDir, "a", "b", "bar.txt"), "bar")
	_, _ = createFile(filepath.Join(srcDir, "a", "b", "c", "doah.txt"), "doah")

	// check that we cannot copy into a destination file that is
	// inside the source
	err := CopyDir(filepath.Join(srcDir, "a", "b"),
		filepath.Join(srcDir, "a", "b", "c"))
	if err == nil {
		t.Errorf("should not be able to copy into source subdir")
		return
	}

	err = CopyDir(filepath.Join(srcDir, "a", "b", "c"),
		filepath.Join(dstDir, "a", "b", "c"))
	if err != nil {
		t.Errorf("cannot copy to a diferent directory")
		return
	}

	err = CopyDir(filepath.Join(srcDir, "a", "b", "c"),
		filepath.Join(srcDir, "a", "b", "d"))
	if err != nil {
		t.Errorf("cannot copy to a sibling directory")
		return
	}
}
