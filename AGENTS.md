# Agent Notes

Single Go module (`github.com/sgtsquiggs/tools`) with one command: `cmd/structtagger`.

## Development

- **Test:** `make test` (or `go test ./...`)
- **Build:** `make build` (or `go build ./cmd/structtagger`)
- **Lint:** `make lint` (requires `golangci-lint`)
- **Install:** `make install` (or `go install ./cmd/structtagger`)

## Release

- Automated via GoReleaser (`.goreleaser.yml`) on pushes to `v*` tags.
- `CGO_ENABLED=0`; targets `darwin/linux/windows` on `amd64/arm64` (no `windows/arm64`).
- Draft releases are created automatically (`.github/workflows/tag.yml`).

## Tool Behavior

- `structtagger` is a code generator that reads Go source via `golang.org/x/tools/go/packages` and writes `<type>_structtags.go`.
- Supported tags are hardcoded: `json`, `bson`, `yaml`, `toml`, `protobuf`, `redis`.
- With no positional args it defaults to the current directory package.
- Unexported fields, `_` placeholders, and `XXX_` protobuf-private fields are skipped.
- Redis tags are silently omitted when missing; other missing tags fall back to the field name.
