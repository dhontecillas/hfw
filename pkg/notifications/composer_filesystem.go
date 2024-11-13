package notifications

import (
	"bytes"
	html_template "html/template"
	"os"
	"path"
	"strings"
	"text/template"
)

// FileSystemComposer is notification template renderer
type FileSystemComposer struct {
	templatesDir string
}

// NewFileSystemComposer creates a new FileSystemComposer
func NewFileSystemComposer(templatesDir string) *FileSystemComposer {
	return &FileSystemComposer{
		templatesDir: templatesDir,
	}
}

// Render renders a notifications using the provided data
func (r *FileSystemComposer) Render(notification string, data map[string]interface{}, carrier string) (*ContentSet, error) {
	lang := "en"
	if l, ok := data["lang"].(string); ok {
		lang = l
	}

	cs := ContentSet{
		HTMLs: make(map[string]string, 2),
		Texts: make(map[string]string, 2),
	}

	notifDir := path.Join(r.templatesDir, notification, carrier, lang)

	finfos, err := os.ReadDir(notifDir)
	if err != nil {
		return nil, err
	}

	for _, finfo := range finfos {
		if finfo.IsDir() {
			continue
		}
		fname := finfo.Name()
		tmplName := strings.Split(fname, ".")[0]
		content, err := os.ReadFile(path.Join(notifDir, fname))
		if err != nil {
			return nil, err
		}
		var b bytes.Buffer
		if strings.Contains(fname, ".html.") {
			t, err := html_template.New(notification).Parse(string(content))
			if err != nil {
				return nil, err
			}
			err = t.Execute(&b, data)
			if err != nil {
				return nil, err
			}
			cs.HTMLs[tmplName] = b.String()
		} else {
			t, err := template.New(notification).Parse(string(content))
			if err != nil {
				return nil, err
			}
			err = t.Execute(&b, data)
			if err != nil {
				return nil, err
			}
			cs.Texts[tmplName] = b.String()
		}
	}
	return &cs, nil
}
