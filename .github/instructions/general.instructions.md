---
applyTo: "**"
---

# xaligo — General Coding Guidelines

## Project Overview

`xaligo` is a Go CLI tool that converts a Vue-style custom DSL (`.xal` files) into Excalidraw JSON files,
using Vuetify-style grid / padding / margin / container layout.
It also provides an `add service` command for appending AWS service icons to existing `.excalidraw` files.

## Module Information

```
module: github.com/ryo-arima/xaligo
Go:     1.22
Key dependencies: github.com/spf13/cobra v1.8.1
                  gopkg.in/yaml.v3 v3.0.1
```

## Directory Structure

```
xaligo/
├── cmd/
│   └── main.go                   # Entry point
├── pkg/
│   ├── command.go                # Root cobra command (wires subcommands)
│   └── controller/
│       ├── render.go             # xaligo render <input.xal> -o output.excalidraw
│       ├── init.go               # xaligo init -o <dir>  (generates sample.xal)
│       ├── version.go            # xaligo version
│       └── add.go                # xaligo add service [flags]
├── internal/
│   ├── model/
│   │   └── ast.go                # DSL AST: Document, Node
│   ├── parser/
│   │   └── parser.go             # XML-based DSL parser
│   ├── layout/
│   │   └── layout.go             # Vuetify-style layout engine
│   ├── excalidraw/
│   │   └── scene.go              # Excalidraw JSON builder (for render command)
│   ├── entity/
│   │   ├── scene.go              # Scene struct (for add command)
│   │   └── service.go            # ServiceEntry struct
│   ├── repository/
│   │   ├── builder.go            # MakeText / MakeImage element builders
│   │   ├── scene.go              # ReadScene / WriteScene
│   │   ├── icon.go               # SvgToDataURL / FileID / LoadFromCSV / SVGBGColor
│   │   └── service_list.go       # ReadServiceList (CSV/TXT parser)
│   └── config/
│       └── config.go             # Config struct + findProjectRoot + etc/resources/aws/app.yaml loading
├── examples/
│   └── sample.xal               # Sample DSL file
├── etc/
│   └── resources/
│       └── aws/
│           ├── app.yaml         # Path / legend size settings (defaults apply when absent)
│           ├── service-catalog.csv  # Full SVG icon catalog
│           ├── svg/             # AWS icon SVGs (Architecture-Service/Group/Resource/Category-Icons)
│           └── templates/
│               ├── excalidraw/  # Per-AWS-group-tag templates (.excalidraw)
│               └── xal/         # Per-AWS-group-tag templates (.xal)
├── scripts/
│   ├── gen_service_catalog.py   # Regenerate service-catalog.csv
│   └── gen_group_templates.py   # Regenerate etc/resources/aws/templates/{excalidraw,xal}/
├── Makefile
├── go.mod / go.sum
└── README.md
```

## Architecture Guidelines

- **cmd → pkg/command.go → pkg/controller/ → internal/**: Keep dependencies unidirectional.
- `internal/` packages are only referenced from `pkg/`; never directly from `cmd/`.
- Each `controller` file exports an `Init<Cmd>Cmd() *cobra.Command` factory function, registered in `pkg/command.go`.
- Business logic stays in `internal/`; cobra flag handling is the responsibility of the `controller` layer.

## Coding Conventions

- Follow standard Go `gofmt` / `golint` style.
- Package names are lowercase single words (e.g., `controller`, `repository`, `entity`).
- Wrap errors with `fmt.Errorf("<context>: %w", err)` and return them to the caller.
- Do not use `panic`. Always return errors as `error`.
- Represent Excalidraw elements as `map[string]interface{}` (for compatibility with the existing format).

## Configuration File (etc/resources/aws/app.yaml)

Loaded from `etc/resources/aws/app.yaml` at the project root (directory containing `go.mod`).
When absent, all defaults are used — the file is not required.

```yaml
paths:
  asset_package:        etc/resources/aws/svg
  service_catalog_csv:  etc/resources/aws/service-catalog.csv
  output_frames:        output/aws-frames
legend:
  offset_x:  120
  offset_y:  0
  icon_size: 32
  font_size: 12
item:
  icon_size: 48   # default max icon size for <item> elements (px). Overridable with <frame item-size="N">
```

## CLI Command Reference

| Command | Description |
|---|---|
| `xaligo render <file.xal> -o <out.excalidraw>` | Convert DSL to Excalidraw JSON |
| `xaligo init [-o <dir>]` | Generate `sample.xal` |
| `xaligo version` | Print version |
| `xaligo add service --name <name> --file <file>` | Add a single AWS service icon |
| `xaligo add service --list <csv> --file <file>` | Bulk-add AWS service icons |
| `xaligo generate xal --clouds N --accounts N --regions N --azs N --az-layout grid\|staggered --subnets N --spacing vertical\|horizontal\|both --start top\|left --paper A4 --orientation portrait\|landscape -o out.xal` | Generate a .xal for an AWS infrastructure hierarchy |
| `xaligo generate excalidraw --xal <file.xal> -o <out.excalidraw> --services <csv>` | Convert .xal to .excalidraw with service legend |

## Build & Test

```bash
make build   # build .bin/xaligo
make run     # examples/sample.xal → output/sample.excalidraw
make clean   # remove .bin/ and output/
go test ./...            # run all tests
go build ./...           # check for build errors
```
