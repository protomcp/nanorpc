module github.com/amery/nanorpc

go 1.21.9

require (
	github.com/amery/nanorpc/pkg/generator v0.0.0-00010101000000-000000000000
	github.com/amery/protogen v0.3.11
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/cobra v1.7.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	google.golang.org/protobuf v1.35.1 // indirect
)

replace (
	github.com/amery/nanorpc/pkg/generator => ./pkg/generator
	github.com/amery/nanorpc/pkg/nanopb => ./pkg/nanopb
	github.com/amery/nanorpc/pkg/nanorpc => ./pkg/nanorpc
)
