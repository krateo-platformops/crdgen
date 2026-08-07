---
type: Example
title: generate-crd — a composition CRD from a chart values.schema.json
description: The canonical crdgen use — feed a sample chart values.schema.json to crdgen.Generate and print the resulting CustomResourceDefinition.
resource: github.com/krateo-platformops/crdgen/examples/generate-crd
tags: [crdgen, example, crd]
timestamp: 2026-08-07T00:00:00Z
---

# generate-crd

Generates a `SampleApp` composition CRD (`sampleapps.composition.krateo.io`, version
`v1-2-3`, categories `compositions/comps`, `Managed` status) from the
[`values.schema.json`](./values.schema.json) in this directory, and prints the CRD
YAML on stdout. The schema exercises the mapping rules documented in
[docs/api.md](../../docs/api.md): scalars with defaults, enum, min/max, a required
object, a typed `additionalProperties` map, and a string array.

## Preconditions

- Go ≥ 1.25 on `PATH` — `crdgen.Generate` shells out to `go mod tidy` and
  `go run … controller-gen` in a temp workdir.
- Network access to a Go module proxy on first run (the generated workdir module
  fetches controller-tools and friends); later runs hit the local module cache.

This example is its own Go module pinned to the released library
(`github.com/krateo-platformops/crdgen v0.6.0`) — it runs as-is from a repo checkout
or copied anywhere.

## The one command

```sh
go run .
```

The CRD YAML lands on stdout (redirect to a file to `kubectl apply` it); the GVK and
the spec digest are logged on stderr. To inspect the intermediate generated Go module,
rerun with `CRDGEN_CLEAN_WORKDIR=1 go run .` and look under the workdir path embedded
in any error (or `$TMPDIR`).
