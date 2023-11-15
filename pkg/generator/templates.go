package generator

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"text/template"
)

// WithTemplates loads embedded templates/**.gotmpl into
// an existing [template.Template]
func (gen *Generator) WithTemplates(root *template.Template, templates fs.FS) error {
	if gen.t != nil {
		return errors.New("templates already attached")
	}

	if root == nil {
		root = template.New("")
	}

	dir, err := fs.Sub(templates, "templates")
	if err != nil {
		return err
	}

	t, err := root.ParseFS(dir, "*.gotmpl")
	if err != nil {
		return err
	}

	gen.t = t
	return nil
}

// T find a template, combines it with given data, and
// renders it onto the given Writer
func (gen *Generator) T(name string, out io.Writer, data any) error {
	var t *template.Template

	if gen.t != nil {
		t = gen.t.Lookup(name + ".gotmpl")
	}

	if t == nil {
		return &template.ExecError{
			Name: name,
			Err:  fmt.Errorf("template %q not found", name),
		}
	}

	return t.Execute(out, data)
}
