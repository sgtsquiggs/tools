# Go Tools

[![PkgGoDev](https://pkg.go.dev/badge/github.com/sgtsquiggs/tools)](https://pkg.go.dev/github.com/sgtsquiggs/tools)

This repository holds the source for various packages and tools that support
the Go programming language.

## Tools

### structtagger

`structtagger` is a code generator that reads Go struct definitions and emits
string constants derived from struct field tags.

#### Install

```bash
go install github.com/sgtsquiggs/tools/cmd/structtagger@latest
```

#### Usage

```bash
# Generate JSON tag constants for type Car in the current package
structtagger -type=Car -tag=json

# Generate constants for multiple types and tags
structtagger -type=Car,Bike -tag=json,bson
```

Typically used with `go generate`:

```go
//go:generate structtagger -type=Car -tag=json

type Car struct {
	Make  string `json:"make"`
	Model string `json:"model"`
}
```

This produces `car_structtags.go`:

```go
const (
	CarMakeJson  = "make"
	CarModelJson = "model"
)
```

#### Supported Tags

- `json`, `bson`, `yaml`, `toml`, `redis`, `protobuf`

#### Behavior

- Unexported fields, `_` placeholders, and `XXX_` protobuf-private fields are skipped.
- Fields tagged with `-` are skipped.
- Redis tags are silently omitted when missing; other missing tags fall back to the field name.

## Development

```bash
make test   # run tests
make build  # build structtagger
make lint   # run linters
make install # install locally
```
