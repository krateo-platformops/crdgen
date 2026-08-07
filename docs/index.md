---
type: Library
title: crdgen — index
description: The map of the crdgen doc bundle — the standalone Go library that transpiles a chart's values.schema.json into a Kubernetes CRD via generated Go types + controller-gen.
resource: github.com/krateo-platformops/crdgen
tags: [crdgen, library, crd, json-schema, controller-gen]
timestamp: 2026-08-07T00:00:00Z
---

# crdgen

A **Go library** with one job: turn a JSON Schema — in Krateo, a Helm chart's
`values.schema.json` — into a Kubernetes CustomResourceDefinition. It does this the
controller-author's way: it transpiles the schema into real Go API types annotated
with kubebuilder markers, materializes them as a throwaway Go module in a temp
directory, and runs `controller-gen` over that module to emit the CRD manifest. What
you get back is exactly the CRD `controller-gen` would produce for a hand-written
type — deepcopy-safe, structural, `crdVersions=v1`.

This repo is the **origin of the Krateo composition-CRD generation lineage**. The
composition engine in production today ([core-provider](https://github.com/krateo-platformops/core-provider))
and [oasgen-provider](https://github.com/krateo-platformops/oasgen-provider) consume
the successor implementation — the `crdgen` package inside
[plumbing](https://github.com/krateo-platformops/plumbing), a direct
JSON-Schema→structural-OpenAPI transpiler with no toolchain exec. This library remains
standalone and consumable as a Go module; [api.md](./api.md) documents both its own
mapping rules and where the production pipeline diverges (int-or-string handling, the
`vacuum` storage-version design).

## The bundle (start here)

- [overview](./overview.md) — the four-stage pipeline: JSON-Schema parse → transpile
  to Go structs → jennifer codegen with kubebuilder markers → `go mod tidy` +
  `controller-gen` exec; the workdir lifecycle and the spec digest.
- [usage](./usage.md) — `go get` + minimal code; the runtime prerequisites (Go
  toolchain on PATH, module-proxy access) and what they cost.
- [configuration](./configuration.md) — the whole config surface: two `Options`
  knobs, the `CRDGEN_CLEAN_WORKDIR` env var, and the ambient Go-toolchain
  configuration the subprocesses inherit.
- [api](./api.md) — the exported Go API (`Options`, `JsonSchemaGetter`, `Result`,
  `Generate`) and the complete schema→CRD mapping rules of this library, followed by
  the mapping + version-lifecycle rules of the production composition pipeline
  (plumbing crdgen, core-provider `AppendVersion`/`RemoveStaleVersions` and the
  `vacuum` storage version) with source pointers.
- [examples](./examples.md) — the runnable examples under `examples/`.
- [release](./release.md) — the actual release convention: `vX.Y.Z` git tags served
  by the Go module proxy; no OCI artifact, no release workflow.
- [log](./log.md) — curated history, v0.1.0 (2024) to v0.6.0 (the module-identity
  migration to `krateo-platformops`).
- [llms.txt](./llms.txt) — the version-pinned agent index of this bundle.

## Repo shape

```
crdgen.go                  # the whole public API: Options, Result, Generate
internal/
├── transpiler/            # JSON Schema -> Go struct/field model
│   └── jsonschema/        # draft-07 subset parser + $ref resolver
├── coder/                 # jennifer codegen: types.go, markers, scheme wiring
└── assets/                # go.mod template of the generated throwaway module
testdata/                  # schemas exercised by the tests
examples/generate-crd/     # the runnable example (own Go module)
```
