package web

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin/render"
)

const templatesDir = "html_templates"
const includesDir = "inc"

// HTMLWrapper is a "dumb" wrapper over the multitemplate
// implementation, that on every request reloads the templates:
// so it should not be used in production, but is useful to
// use it while developing, to have live updates from the loaded
// templates.
type HTMLWrapper struct {
	r        render.HTMLRender
	dirPaths []string
}

// NewHTMLRender creates a
func NewHTMLRender(dirPaths ...string) *HTMLWrapper {
	r := NewMultiRenderEngineFromDirs(dirPaths...)
	if r == nil {
		return nil
	}
	return &HTMLWrapper{
		r:        r,
		dirPaths: dirPaths,
	}
}

// Instance wraps the call to `Instance` for a fresh new
// MultiRenderEngine from the stored dirPaths in HTMLWrapper
func (h *HTMLWrapper) Instance(a string, b interface{}) render.Render {
	h.r = NewMultiRenderEngineFromDirs(h.dirPaths...)
	return h.r.Instance(a, b)
}

// Templates contains map of pages to be used as templates
// with a list of files that are used as "fragments" to
// be used by those templates.
type Templates struct {
	Pages    map[string]string
	Includes []string
}

// Collect searches a give `dirName` directory for files
// under directories called `html_templates` for main
// templates (and files under `html_templates/inc` dir for
// additional fragments to be used in main teamplates).
func (t *Templates) Collect(dirName string) {
	_ = filepath.Walk(dirName, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.Mode().IsDir() {
			return nil
		}

		pathComponents := strings.Split(path, string(os.PathSeparator))
		nComps := len(pathComponents)
		for idx, name := range pathComponents {
			if name == templatesDir {
				if idx == nComps-1 {
					return nil
				}
				if pathComponents[idx+1] == includesDir {
					t.Includes = append(t.Includes, path)
					continue
				}
				pageName := strings.Join(pathComponents[idx+1:], string(os.PathSeparator))
				t.Pages[pageName] = path
			}
		}
		return nil
	})
}

// NewMultiRenderEngine creates a render engine from the
// Templates struct.
// Currently, it only provides a `trans` function to templates
// to be able to translate content.
func NewMultiRenderEngine(t *Templates) render.HTMLRender {
	r := multitemplate.NewRenderer()

	funcs := template.FuncMap{
		"trans": func(i interface{}) string {
			s, ok := i.(string)
			if !ok {
				return "NO STRING"
			}
			return s
		},
	}

	for name, fpath := range t.Pages {
		l := make([]string, 0, len(t.Includes)+1)
		l = append(l, fpath)
		l = append(l, t.Includes...)

		// check if the includes templates can be parsed
		_, er := template.ParseFiles(l...)
		if er != nil {
			panic(fmt.Sprintf("cannot parse files: %s \n", er))
		}
		r.AddFromFilesFuncs(name, funcs, l...)
	}
	return r
}

// NewMultiRenderEngineFromDirs scan a list of directories to
// search for templates under directories named `html_templates`
// (and template fragments under `html_templates/inc` dirs).
func NewMultiRenderEngineFromDirs(dirPaths ...string) render.HTMLRender {
	t := Templates{
		Pages:    make(map[string]string, 64),
		Includes: make([]string, 0, 128),
	}
	// collect all the files
	for _, dp := range dirPaths {
		t.Collect(dp)
	}
	return NewMultiRenderEngine(&t)
}
