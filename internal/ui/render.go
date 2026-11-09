package ui

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"

	"github.com/fremenkiel/stdout.me/ui"
)

type Renderer struct {
	pageCache	map[string]*template.Template
	base			*template.Template
}

func NewRenderer() *Renderer {
	base := template.Must(template.New("").ParseFS(ui.Files, "html/layouts/*.html", "html/components/*.html"))

	cache := map[string]*template.Template{}
	pages, _ := fs.Glob(ui.Files, "html/pages/*.html")

	for _, page := range pages {
		name := filepath.Base(page)
		tmpl := template.Must(base.Clone())
		tmpl = template.Must(tmpl.ParseFS(ui.Files, page))
		cache[name] = tmpl
	}

	return &Renderer{pageCache: cache, base: base}
}

func (r *Renderer) RenderPage(w http.ResponseWriter, name string, data any) error {
	var buf bytes.Buffer

	filename := fmt.Sprintf("%s.html", name)

	temp, ok := r.pageCache[filename]
	if !ok {
		return fmt.Errorf("template %s not found", name)
	}

	if err := temp.ExecuteTemplate(&buf, filename, data); err != nil {
		return err
	}

	_, err := buf.WriteTo(w)
	return err
}

func (r *Renderer) RenderFragment(w http.ResponseWriter, name string, data any) error {
	var buf bytes.Buffer

	if err := r.base.ExecuteTemplate(&buf, name, data); err != nil {
		return err
	}

	_, err := buf.WriteTo(w)
	return err
}
