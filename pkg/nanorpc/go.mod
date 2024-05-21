module github.com/amery/nanorpc/pkg/nanorpc

go 1.21.9

require (
	darvaza.org/core v0.13.1
	github.com/amery/nanorpc/pkg/nanopb v0.0.0-00010101000000-000000000000
	google.golang.org/protobuf v1.34.1
)

require (
	golang.org/x/net v0.20.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)

replace github.com/amery/nanorpc/pkg/nanopb => ../nanopb
