module protomcp.org/nanorpc

go 1.23.0

replace (
	protomcp.org/nanorpc/pkg/generator => ./pkg/generator
	protomcp.org/nanorpc/pkg/nanopb => ./pkg/nanopb
	protomcp.org/nanorpc/pkg/nanorpc => ./pkg/nanorpc
)
