---
type: Usage
title: Using crdgen
description: go get + the minimal Generate call, the runtime prerequisites (Go toolchain, module proxy) and how to debug a failed generation.
resource: github.com/krateo-platformops/crdgen
tags: [crdgen, usage, go-get]
timestamp: 2026-08-07T00:00:00Z
---

# Usage

## Install

```sh
go get github.com/krateo-platformops/crdgen@v0.6.0
```

`v0.6.0` is the first tag under the `krateo-platformops` module identity (earlier tags
declare the pre-migration module path and are not importable under this one).

## Minimal code

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/krateo-platformops/crdgen"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type fileGetter struct{ path string }

func (g *fileGetter) Get() ([]byte, error) { return os.ReadFile(g.path) }

func main() {
	res := crdgen.Generate(context.Background(), crdgen.Options{
		WorkDir: "myapp", // names the temp module dir
		GVK: schema.GroupVersionKind{
			Group:   "composition.krateo.io",
			Version: "v1-2-3", // chart version, CRD-legal form (dots -> dashes)
			Kind:    "MyApp",  // chart name, PascalCase
		},
		Categories:           []string{"compositions", "comps"},
		Managed:              true, // ConditionedStatus + READY column
		SpecJsonSchemaGetter: &fileGetter{"values.schema.json"},
	})
	if res.Err != nil {
		panic(res.Err)
	}
	fmt.Println(string(res.Manifest)) // the CRD YAML
}
```

The runnable version of this — including a realistic `values.schema.json` — is
[examples/generate-crd](../examples/generate-crd/README.md).

## Runtime prerequisites (they are real)

`Generate` is not pure-Go at runtime — it materializes a Go module in
`os.TempDir()` and execs the toolchain (see [overview](./overview.md)):

- **A Go toolchain on `PATH`** of the *calling process* (container images embedding
  this library must ship `go`).
- **Module-proxy network access** on first run: the generated module `go mod tidy`s
  controller-tools, controller-runtime, apimachinery and provider-runtime. Subsequent
  runs hit the local module cache. Air-gapped consumers need a reachable
  `GOPROXY`/`GOMODCACHE` ([configuration](./configuration.md)).
- **Seconds per call**: budget for `go mod tidy` + `go run controller-gen`. Callers
  that generate repeatedly should cache on `Result.Digest` (sha256 of the input
  schemas) — that is exactly what it exists for.

## Feeding it a chart

The intended input is a chart's `values.schema.json`: the schema's root object becomes
the CRD's `.spec`, so a CR instance of the generated kind carries the chart's values
as its spec. Version and kind are derived from the chart by convention:
`Kind = PascalCase(chart name)`, `Version = "v" + chart version with dots →
dashes` (`1.2.3` → `v1-2-3`) — renaming a chart therefore changes the CRD kind.

An optional `StatusJsonSchemaGetter` contributes `.status` fields; with
`Managed: true` they are merged alongside the provider-runtime condition machinery.

## When it fails

- Schema constructs outside the supported subset fail fast with a transpiler error
  (type unions, definitions containing only `additionalProperties`) — the complete
  mapping table and limits are in [api.md](./api.md).
- controller-gen failures (e.g. a `type: number` field) surface as
  `Result.Err` with the full tool output, including the workdir path.
- To inspect the generated module, set `CRDGEN_CLEAN_WORKDIR=1` (any value) and rerun:
  the workdir under `$TMPDIR` is kept, with `apis/…/types.go` and `crds/` inside.
