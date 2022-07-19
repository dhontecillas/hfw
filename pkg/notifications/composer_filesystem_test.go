package notifications

import (
	"os"
	"path/filepath"
	"testing"
)

func Test_ReadTemplate(t *testing.T) {
	tmplDir := "./templates"
	curDir, _ := os.Getwd()
	tmplDir = filepath.Join(curDir, tmplDir)
	fsr := NewFileSystemComposer(tmplDir)
	cs, err := fsr.Render("users_requestregistration",
		map[string]interface{}{
			"lang":             "en",
			"url":              "foo/bar",
			"activation_token": "foobar",
		}, "email")
	if err != nil {
		t.Errorf("Err: %s", err.Error())
		return
	}
	if len(cs.Texts) != 2 {
		t.Errorf("Expected 2 text templates")
		return
	}

	if len(cs.HTMLs) != 1 {
		t.Errorf("Expected 1 html template")
		return
	}
}
