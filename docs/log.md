---
type: Log
title: crdgen — curated history
description: Notable changes and decisions from v0.1.0 (2024) to v0.6.0 (module-identity migration), plus where the lineage's active development moved.
resource: github.com/krateo-platformops/crdgen
tags: [crdgen, history]
timestamp: 2026-08-07T00:00:00Z
---

# Log

Curated history — notable changes and decisions, newest first. Release notes belong in
GitHub Releases; this is the narrative.

- **2026-08-03 — v0.6.0: module identity migrated to `krateo-platformops`** (full
  independence from the pre-migration org). Code-only rename of the module path and
  internal imports; the *generated* throwaway module still points at the
  pre-migration org for its own module path and its provider-runtime dependency
  (see [overview](./overview.md), Stage 3).
- **2026 — active development of the lineage moved to plumbing/core-provider.** The
  composition engine's CRD generation is the direct (exec-free) transpiler in the
  plumbing repo's `crdgen` package, with version lifecycle (the `vacuum` storage
  version, stale-version pruning) in core-provider — see
  [api.md](./api.md). This repo remains the standalone library; it is not currently
  imported by the engine.
- **2025-09-05 — v0.5.0**: enums inside arrays hoisted onto the array field (#39);
  array/nested-array test-case fixes (#41, #43).
- **2025-05 — v0.4.1…v0.4.3**: `title` support end-to-end (`+kubebuilder:title`,
  controller-tools bumped to a version that emits it) (#35, #37).
- **2025-03-19 — v0.4.0**: KRA-133 robustness pass (#34): duplicate struct-name
  disambiguation (parent prefix + random suffix), collision-safe generation; Apache-2
  LICENSE added.
- **2024-12-30 — v0.3.9**: `additionalProperties` aligned to Kubernetes CRD rules
  (#31) — typed map values, `true` → preserve-unknown-fields.
- **2024-10..12 — v0.3.7/v0.3.8**: descriptions propagated into the generated CRD
  (`openAPIV3Schema.description`, then per-attribute fixes); `optional` keyword
  dropped in favor of the JSON-Schema `required` array.
- **2024-05-14 — v0.3.5**: type unions made a **hard error** ("multiple types in
  schema", #17/#19) instead of silently mis-generating — the boundary the successor
  transpiler later refined (see [api.md](./api.md)).
- **2024-04 — v0.3.0…v0.3.4**: `default` handling (quoted string defaults, bool
  defaults, string-typed bools), `go mod tidy` pinned against auto-upgrading the
  generated module's deps (#15).
- **2024-02 — v0.1.0/v0.2.0**: initial release: JSON-Schema→Go-types→controller-gen
  pipeline, GVK returned with the result, spec/status schemas split (#2, #4).
