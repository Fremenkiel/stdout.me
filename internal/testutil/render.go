package testutil

import (
	"bytes"
	"net/http"

	"github.com/fremenkiel/stdout.me/internal/ui"
)

type Renderer struct {
	renderer	*ui.Renderer
}

func NewRenderer() *Renderer {
	return &Renderer{
		renderer: ui.NewRenderer(),
	}
}

func (r *Renderer) RenderPage(w http.ResponseWriter, name string, data any) error {
	var buf bytes.Buffer

	buf.WriteString("page:")
	buf.WriteString(name)

	if _, err := w.Write(buf.Bytes()); err != nil {
		return err
	}

	if err := r.renderer.RenderPage(w, name, data); err != nil {
		return err 
	}

	return nil
}

func (r *Renderer) RenderFragment(w http.ResponseWriter, name string, data any) error {
	var buf bytes.Buffer

	buf.WriteString("fragment:")
	buf.WriteString(name)

	if _, err := w.Write(buf.Bytes()); err != nil {
		return err
	}

	if err := r.renderer.RenderFragment(w, name, data); err != nil {
		return err
	}

	return nil
}
