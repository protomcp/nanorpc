package main

import (
	"embed"

	"github.com/amery/nanorpc/pkg/generator"
)

//go:generate mkdir -p templates

//go:embed templates/**.gotmpl
var templates embed.FS

func applyTemplates(gen *generator.Generator) error {
	return gen.WithTemplates(nil, &templates)
}
