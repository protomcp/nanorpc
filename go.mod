module github.com/amery/nanorpc

go 1.21

replace (
	github.com/amery/nanorpc/pkg/generator => ./pkg/generator
	github.com/amery/nanorpc/pkg/nanopb => ./pkg/nanopb
	github.com/amery/nanorpc/pkg/nanorpc => ./pkg/nanorpc
)
