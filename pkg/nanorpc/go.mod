module github.com/amery/nanorpc/pkg/nanorpc

go 1.21.9

require (
	darvaza.org/core v0.14.3
	github.com/amery/nanorpc/pkg/nanopb v0.0.0
)

require google.golang.org/protobuf v1.34.2

require (
	golang.org/x/net v0.27.0 // indirect
	golang.org/x/text v0.16.0 // indirect
)

replace github.com/amery/nanorpc/pkg/nanopb => ../nanopb
