package main

import (
	"embed"

	"github.com/amery/nanorpc/pkg/generator"
)

//go:embed templates/**.gotmpl
var templates embed.FS

func applyTemplates(gen *generator.Generator) error {
	return gen.WithTemplates(nil, &templates)
}
