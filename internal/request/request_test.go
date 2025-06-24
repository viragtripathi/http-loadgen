package request

import (
	"testing"
	"text/template"
)

func TestRenderTemplate(t *testing.T) {
	funcs := template.FuncMap{"uuid": func() string { return "test-id" }}
	data := map[string]any{"name": "test"}
	out, err := renderTemplate("test", "hello {{ .name }}", data, funcs)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if out != "hello test" {
		t.Errorf("unexpected output: %s", out)
	}
}
