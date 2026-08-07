# crdgen

A Go library that turns a JSON Schema — a Helm chart's `values.schema.json` — into a
Kubernetes CustomResourceDefinition, by generating a throwaway Go module and running
controller-gen over it.

[![Test and coverage](https://github.com/krateo-platformops/crdgen/actions/workflows/test.yaml/badge.svg)](https://github.com/krateo-platformops/crdgen/actions/workflows/test.yaml)

## What is this

The origin library of the Krateo composition-CRD generation lineage: `crdgen.Generate`
transpiles a chart's `values.schema.json` into Go API types (kubebuilder markers
included), then shells out to `go mod tidy` + `controller-gen` to emit the CRD
manifest. The composition engine in production today consumes the successor
implementation of this package — the `crdgen` package inside
[plumbing](https://github.com/krateo-platformops/plumbing) (a direct, exec-free
transpiler); this repo remains the standalone library and the lineage's design anchor.
Full picture: [docs/index.md](docs/index.md).

## Install

```sh
go get github.com/krateo-platformops/crdgen@v0.6.0
```

Runtime prerequisites: a Go toolchain on `PATH` and module-proxy network access —
`Generate` executes `go mod tidy` and `go run … controller-gen` in a temp workdir.
Minimal code: [docs/usage.md](docs/usage.md).

## Configure

See [docs/configuration.md](docs/configuration.md). The whole surface is small:

| Setting | Default | Effect |
|---|---|---|
| `Options.Verbose` | `false` | Debug logging of the generation pipeline to stderr. |
| `Options.Managed` | `false` | Adds the provider-runtime `ConditionedStatus` + `failedObjectRef` to `.status` and the READY printer column. |
| `CRDGEN_CLEAN_WORKDIR` (env) | unset | Set to any value to KEEP the temp workdir after generation (debugging); unset removes it. |

## Examples

In [`examples/`](examples/), each with preconditions and the one command
([docs/examples.md](docs/examples.md)):

- [generate-crd](examples/generate-crd/README.md) — generate a composition CRD from a
  sample chart `values.schema.json` and print it: `go run .`.

## Docs

- [docs/index.md](docs/index.md) — the map of the bundle.
- [docs/overview.md](docs/overview.md) — the four-stage pipeline: parse → transpile to
  Go types → codegen → controller-gen exec.
- [docs/usage.md](docs/usage.md) — `go get` + minimal code, runtime prerequisites.
- [docs/configuration.md](docs/configuration.md) — options, the one env var, ambient
  Go-toolchain configuration.
- [docs/api.md](docs/api.md) — the exported Go API and the schema→CRD mapping rules;
  where the production pipeline (plumbing crdgen + core-provider version lifecycle,
  including the `vacuum` storage version) diverges.
- [docs/examples.md](docs/examples.md) — the runnable examples.
- [docs/release.md](docs/release.md) — tag-only release convention (Go module proxy,
  no OCI artifact).
- [docs/log.md](docs/log.md) — curated history.
- [docs/llms.txt](docs/llms.txt) — the version-pinned agent index.

## Develop & release

`go build ./... && go test ./...` (unit tests; the root end-to-end tests are behind
`-tags integration` and exec the real toolchain). Releasing is pushing a `vX.Y.Z` tag —
see [docs/release.md](docs/release.md).
