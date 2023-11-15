// Package main implements a C99 generator for nanopb
package main

import (
	"io"
	"log"
	"os"

	"github.com/amery/protogen/pkg/protogen"
	"github.com/amery/protogen/pkg/protogen/plugin"

	"github.com/amery/nanorpc/pkg/generator"
)

func generate(p *protogen.Plugin) error {
	gen, err := generator.NewGenerator(p)
	if err != nil {
		return err
	}

	return applyTemplates(gen)
}

func run(in io.ReadCloser, out io.WriteCloser) error {
	opts := protogen.Options{
		Stdin:  in,
		Stdout: out,
	}

	return opts.Run(generate)
}

var rootCmd = plugin.MustRoot(&plugin.Config{
	RunE: run,
})

func main() {
	err := rootCmd.Execute()

	switch e := err.(type) {
	case plugin.ExitCoder:
		os.Exit(e.ExitCode())
	case nil:
		os.Exit(0)
	default:
		log.Fatal(err)
	}
}
