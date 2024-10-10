module github.com/amery/nanorpc/pkg/nanorpc

go 1.21.9

require (
	darvaza.org/core v0.15.1
	darvaza.org/slog v0.5.12
	darvaza.org/slog/handlers/discard v0.4.15
	darvaza.org/x/config v0.3.8
	darvaza.org/x/fs v0.3.3 // indirect
	darvaza.org/x/net v0.3.6
	github.com/amery/defaults v0.1.0 // indirect
	github.com/amery/nanorpc/pkg/nanopb v0.0.0
)

require google.golang.org/protobuf v1.35.1

require (
	github.com/gabriel-vasile/mimetype v1.4.5 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.22.1 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	golang.org/x/crypto v0.28.0 // indirect
	golang.org/x/net v0.30.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
)

replace github.com/amery/nanorpc/pkg/nanopb => ../nanopb
