---
type: ExampleIndex
title: crdgen examples
description: Index of the runnable examples under examples/ — one line each, with preconditions.
resource: github.com/krateo-platformops/crdgen
tags: [crdgen, examples]
timestamp: 2026-08-07T00:00:00Z
---

# Examples

One directory per example under [`examples/`](../examples/); each is its own Go
module with a `README.md` stating preconditions and the one command.

- [generate-crd](../examples/generate-crd/README.md) — the canonical library use:
  feed a sample chart `values.schema.json` to `crdgen.Generate` (`Managed`, categories
  `compositions/comps`, chart-derived GVK) and print the resulting CRD. `go run .` —
  requires a Go toolchain and module-proxy access.
