package bundler

import (
	"os"
	"path/filepath"
	"testing"
)

func Test_PrepareBundleDirs(t *testing.T) {
	tmpDirName, err := os.MkdirTemp("", "Test_PrepareBundleDirs_*")
	if err != nil {
		t.Errorf("cannot create tmpDirName")
		return
	}
	defer func() { _ = os.RemoveAll(tmpDirName) }()

	res, err := PrepareBundleDirs(tmpDirName)
	if err != nil {
		t.Errorf("err %s", err.Error())
		return
	}

	expected, _ := filepath.Abs(tmpDirName)
	if res != expected {
		t.Errorf("want %s, got %s", expected, res)
		return
	}
}
