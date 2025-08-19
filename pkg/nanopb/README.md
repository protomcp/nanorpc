# NanoPB Go Package

<!-- cspell:ignore Petteri Aimonen -->

[![Go Reference][godoc-badge]][godoc-url]
[![Go Report Card][goreportcard-badge]][goreportcard-url]
[![codecov][codecov-badge]][codecov-url]

This package provides Go bindings for the NanoPB protocol buffer definitions.

## Overview

The `nanopb` package contains the generated Go code from the NanoPB protocol
buffer definitions. It provides the necessary types and options for configuring
NanoPB code generation when working with Protocol Buffers in embedded systems.

## Purpose

This package contains the generated Go code from the NanoPB protocol buffer
definitions. NanoPB is a small code-size Protocol Buffers implementation
designed for embedded systems and microcontrollers.

The protocol buffer definitions in this package are used to configure code
generation options for the NanoPB C library. These options control memory
allocation strategies, field sizes, and other code generation preferences
for resource-constrained environments.

## Source

The protocol buffer definitions are sourced from the
[NanoPB](https://github.com/nanopb/nanopb) project, version 0.4.7 or later.

## Usage

This package is primarily used internally by the NanoRPC project to support
protocol buffer options when generating code for embedded systems.

## Licence

This package includes code derived from the NanoPB project and is licensed
under the zlib licence. See the [LICENSE](LICENSE) file for full details.

The original NanoPB project is:

- Copyright © 2011 Petteri Aimonen <jpa@nanopb.mail.kapsi.fi>

This Go package is:

- Copyright © 2023-2025 Apptly Software Ltd <oss@apptly.co>

[godoc-badge]: https://pkg.go.dev/badge/protomcp.org/nanorpc/pkg/nanopb.svg
[godoc-url]: https://pkg.go.dev/protomcp.org/nanorpc/pkg/nanopb
[goreportcard-badge]: https://goreportcard.com/badge/protomcp.org/nanorpc/pkg/nanopb
[goreportcard-url]: https://goreportcard.com/report/protomcp.org/nanorpc/pkg/nanopb
[codecov-badge]: https://codecov.io/gh/protomcp/nanorpc/graph/badge.svg?flag=nanopb
[codecov-url]: https://codecov.io/gh/protomcp/nanorpc?flag=nanopb
