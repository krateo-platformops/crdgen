---
type: Architecture
title: crdgen architecture
description: How Generate turns a values.schema.json into a CRD — the parse/transpile/codegen/controller-gen pipeline, the temp-workdir lifecycle, and the design trade-offs.
resource: github.com/krateo-platformops/crdgen
tags: [crdgen, architecture, controller-gen, jennifer]
timestamp: 2026-08-07T00:00:00Z
---

# Architecture

`crdgen.Generate` (the only entry point, `crdgen.go`) is a four-stage pipeline. The
defining design choice: instead of building CRD OpenAPI structs directly, it
**generates the Go code a controller author would write** and lets the canonical
Kubernetes tooling (`controller-gen`) derive the CRD from it. Everything downstream of
"valid Go types" is therefore exactly as correct as controller-gen itself — and
everything the Go type system or controller-gen cannot express is a hard limit of this
library (see the mapping rules in [api.md](./api.md)).

## Stage 1 — fetch + parse the schemas

The caller supplies the spec schema (and optionally a status schema) through the
`JsonSchemaGetter` interface — one `Get() ([]byte, error)` method, so schemas can come
from a file, an OCI layer, or HTTP without the library caring.
`internal/transpiler/jsonschema` parses a draft-07 subset into a `Schema` tree
(`properties`, `required`, `items`, `definitions`, `$ref`, `additionalProperties`,
`enum`, `default`, `minimum`/`maximum`/`multipleOf`, `pattern`, `title`,
`description`) and links parents for name derivation. A `$ref` resolver
(`refsresolver.go`) indexes every schema and definition by URI fragment. A node with
no `type` is repaired when guessable: `properties` present → `object`, `items`
present → `array` (`FixMissingTypeValue`).

## Stage 2 — transpile to a Go struct model

`internal/transpiler` walks the tree into a flat `map[string]Struct` (fields carry
name, JSON name, Go type, required, default, validation keywords). Objects become
pointer-to-struct types; name collisions are disambiguated by prefixing the parent
name, then (still colliding) a random letter suffix. Arrays become `[]T`; `$ref`s
resolve to the referenced struct (recursion is broken by caching the generated type
name on the schema node). A schema whose `type` is a list of more than one type is
**rejected with an error** — the Go-codegen approach has no way to express a union
field (this is where the successor transpiler in plumbing behaves differently; see
[api.md](./api.md)).

## Stage 3 — codegen the throwaway module

`internal/coder` renders real Go source with [jennifer](https://github.com/dave/jennifer)
under `$TMPDIR/github.com/<legacy-org>/<WorkDir>/`:

- `apis/<kind>/<version>/types.go` — the spec/status structs, every field annotated
  with kubebuilder markers (`+kubebuilder:default`, `:validation:Minimum/Maximum/
  MultipleOf/Pattern/Enum`, `+kubebuilder:title`, `+optional`), the root type with
  `+kubebuilder:object:root=true`, `:subresource:status`,
  `:resource:scope=Namespaced,categories={…}` and the AGE (+READY when `Managed`)
  printer columns.
- `apis/<kind>/<version>/groupversion_info.go` — `+groupName`/`+versionName` and the
  scheme builder; the package name is the CRD version with `-` → `_` (`v1-2-3` →
  `v1_2_3`), the CRD version string stays exactly as passed in `Options.GVK.Version`.
- `managed.go` / `managed_list.go` (only when `Options.Managed`) — `GetCondition`/
  `SetConditions` plumbing over the provider-runtime `ConditionedStatus` that is
  inlined into `.status`, plus `failedObjectRef`.
- `apis/generate.go` + `hack/boilerplate.go.txt` — the `go:generate` scaffolding.
- `go.mod` — rendered from `internal/assets/files/go.mod.tpl`, pinning the generated
  module's dependencies: provider-runtime v0.9.1, apimachinery v0.33.0,
  controller-runtime v0.20.0, controller-tools v0.18.0.

Note: the generated module's path and its provider-runtime dependency still live under
the pre-migration GitHub org (the old org name, spelled without the hyphen — see
`crdgen.go` `defaultCodeGeneratorOptions` and `internal/coder/types.go`). That is the
generated code's identity, not a doc link; it keeps resolving as long as GitHub
redirects the old org.

## Stage 4 — exec the toolchain, harvest the CRD

`Generate` runs, in the workdir: `go mod tidy`, then
`go run --tags generate sigs.k8s.io/controller-tools/cmd/controller-gen
object:headerFile=./hack/boilerplate.go.txt paths=./... crd:crdVersions=v1
output:artifacts:config=./crds`. The first file in `crds/` is the result
(`Result.Manifest`). `Result.Digest` is the sha256 of the input spec (+ status)
schema — a change-detection key for callers that cache generated CRDs. The workdir is
removed afterwards unless `CRDGEN_CLEAN_WORKDIR` is set
([configuration.md](./configuration.md)).

## Consequences of the design

- **Correctness by proxy**: markers, deepcopy, structural-schema pruning are all
  controller-gen's, battle-tested.
- **Runtime cost**: every `Generate` call is a `go mod tidy` + `go run` — seconds,
  network on first run (module proxy), a Go toolchain required on PATH. This is the
  reason the production engine moved to the exec-free successor in
  [plumbing](https://github.com/krateo-platformops/plumbing) (its `crdgen` package
  transpiles JSON Schema straight to structural OpenAPI v3 in-process).
- **Expressiveness ceiling**: what Go + controller-gen cannot express fails loudly —
  type unions error in Stage 2, `type: number` fields make controller-gen itself
  refuse (float64 needs `allowDangerousTypes`, which is not passed). The full list:
  [api.md](./api.md).
