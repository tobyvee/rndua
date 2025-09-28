# rndua

`rndua` is a cross-platform Go utility for generating random User-Agent strings. This project includes a Makefile for building binaries targeting multiple operating systems and architectures.

## Features

- Generate random User-Agent strings
- Cross-compilation support for Linux, macOS (Darwin), and Windows
- Builds for both amd64 and arm64 architectures

## Building

To build all binaries for supported platforms, run:

```sh
make all
```

Or build for a specific platform/architecture:

```sh
make build-linux           # Linux amd64
make build-linux-aarch64   # Linux arm64
make build-darwin          # macOS amd64
make build-darwin-aarch64  # macOS arm64
make build-win             # Windows amd64
make build-win-aarch64     # Windows arm64
```

## Requirements

- Go (pre-installed in this dev container)
- GNU Make

## Usage

After building, run the binary for your platform:

```sh
./rndua-linux-amd64
```

## Development

This project is designed for use in a dev container running Debian GNU/Linux 12 (bookworm) with Go, Node.js, and other common tools pre-installed.