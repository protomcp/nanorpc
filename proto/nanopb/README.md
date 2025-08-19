# NanoPB Protocol Buffer Definitions

This directory contains the Protocol Buffer definitions from the
[NanoPB](https://github.com/nanopb/nanopb) project.

## Overview

This directory provides the NanoPB protocol buffer definitions that are used
to configure code generation options for embedded systems. These definitions
are essential for controlling how NanoPB generates C code for
resource-constrained environments.

## About NanoPB

NanoPB is a small code-size Protocol Buffers implementation in ANSI C. It is
especially suitable for use in microcontrollers and memory-restricted systems.
The project is maintained at <https://jpa.kapsi.fi/nanopb/>.

## Version

The `nanopb.proto` file in this directory is from NanoPB version 0.4.7 or later
(compatible through 0.4.9.1). This is determined by the presence of the
`fallback_type` field (field number 29) which was introduced in version 0.4.7.

## Modifications

The following modifications have been made to the original file:

1. **Added Go package option**: The line `option go_package =
   "protomcp.org/nanorpc/pkg/nanopb";` was added to specify the Go import
   path. This eliminates the need for manual import path mappings when
   generating Go code.

All other content remains identical to the upstream version.

## Purpose

This proto file defines custom options that control how NanoPB generates code
for embedded systems. These options allow fine-tuning of:

- Memory allocation strategies (static vs dynamic).
- Field sizes and array limits.
- Code generation preferences.
- Struct packing and alignment.
- Enum handling and naming conventions.

## Usage

The options defined in this file are used as extensions to standard Protocol
Buffer field, message, and file options. They control how the NanoPB code
generator creates C structures and serialisation code for resource-constrained
environments.

## Licence

The original NanoPB project is distributed under the zlib licence. See the
[NanoPB repository](https://github.com/nanopb/nanopb) for full licence details.
