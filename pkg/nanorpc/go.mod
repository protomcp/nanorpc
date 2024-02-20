module github.com/amery/nanorpc/pkg/nanorpc

go 1.21.9

require (
	github.com/amery/nanorpc/pkg/nanopb v0.0.0-00010101000000-000000000000
	google.golang.org/protobuf v1.34.1
)

replace github.com/amery/nanorpc/pkg/nanopb => ../nanopb
