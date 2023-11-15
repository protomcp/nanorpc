package generator

import (
	"errors"

	"github.com/amery/protogen/pkg/protogen"
)

// Generator is a proto generator for NanoRPC
type Generator struct {
	p *protogen.Plugin
}

func (*Generator) init() error {
	return nil
}

// NewGenerator assembles a nanorpc protoc generator
func NewGenerator(p *protogen.Plugin) (*Generator, error) {
	if p == nil {
		return nil, errors.New("protogen generator missing")
	}

	gen := &Generator{p: p}
	if err := gen.init(); err != nil {
		return nil, err
	}

	return gen, nil
}
