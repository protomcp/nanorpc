# NanoRPC is RPC for NanoPB
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Famery%2Fnanorpc.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Famery%2Fnanorpc?ref=badge_shield)


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



## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Famery%2Fnanorpc.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Famery%2Fnanorpc?ref=badge_large)