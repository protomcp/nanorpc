# NanoRPC is RPC for NanoPB

## Development

### Lint

```sh
go install github.com/golangci/golangci-lint/cmd/...@latest
golangci-lint run
```
or

```sh
make get
make lint
```

# Build

```sh
go get -v ./...
go generate -v ./...
go build -v ./...
```
or

```sh
make get
make build
```

## See also

* https://github.com/nanopb/nanopb
* https://github.com/amery/protogen

