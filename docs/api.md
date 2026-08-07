---
type: API
title: crdgen API and schema→CRD mapping rules
description: The exported Go surface (Options, JsonSchemaGetter, Result, Generate), the complete JSON-Schema→CRD mapping rules of this library, and the production composition pipeline's rules — int-or-string handling and the vacuum storage-version design — with source pointers.
resource: https://pkg.go.dev/github.com/krateo-platformops/crdgen
tags: [crdgen, api, crd, json-schema, int-or-string, vacuum]
timestamp: 2026-08-07T00:00:00Z
---

# API

## Exported Go surface (`crdgen.go` — the whole package)

```go
type JsonSchemaGetter interface {
	Get() ([]byte, error)
}

type Options struct {
	WorkDir                string
	GVK                    schema.GroupVersionKind
	Categories             []string
	SpecJsonSchemaGetter   JsonSchemaGetter
	StatusJsonSchemaGetter JsonSchemaGetter
	Managed                bool
	Verbose                bool
}

type Result struct {
	WorkDir  string // the temp module dir (useful with CRDGEN_CLEAN_WORKDIR)
	Manifest []byte // the CRD YAML
	Digest   string // sha256(spec schema [+ status schema]) — cache key
	GVK      schema.GroupVersionKind
	Err      error
}

func Generate(ctx context.Context, opts Options) Result
```

Semantics worth knowing:

- Errors are returned **inside `Result.Err`**, not as a second return value; always
  check it before touching `Manifest`.
- The `context.Context` parameter is accepted but not currently consulted by the
  implementation — a `Generate` in flight cannot be cancelled through it.
- `Digest` hashes only the *input schemas*, not GVK/options: it detects schema
  drift, which is the change that matters for regeneration.
- Field-by-field `Options` reference: [configuration.md](./configuration.md).

## Schema→CRD mapping rules (this library)

The input is JSON Schema draft-07 (subset); the root object becomes the CRD's
`.spec`. Rules as implemented in `internal/transpiler` (schema → Go model) and
`internal/coder` (Go model → kubebuilder markers), verified against
`crdgen.Generate` output:

| Schema construct | CRD result |
|---|---|
| `type: object` + `properties` | nested `object` with typed `properties` (a generated Go struct; colliding struct names are parent-prefixed, then random-letter suffixed) |
| `type: string` / `integer` / `boolean` | same scalar type (`string`/`int`/`bool` fields) |
| `type: array` + `items` | `array` with typed `items`; `items` absent → `[]any` |
| `type: number` | **generation fails**: maps to `float64` and controller-gen refuses floats without `allowDangerousTypes` (not passed) — "not all generators ran successfully" |
| `type: ["a","b"]` (type union, incl. `["integer","string"]`) | **hard error** — `multiple types in schema '<path>'` (`internal/transpiler/transpiler.go`); a Go field has exactly one type. See the pipeline section below for how the successor handles this |
| missing `type` | inferred: `properties` present → `object`, `items` present → `array`; otherwise error (unless `$ref`) |
| `$ref` + `definitions` | resolved and inlined as the referenced struct; recursive refs are broken via the cached generated type name |
| `required` | listed fields are non-pointer + `required` in the CRD; all others get `+optional`, a pointer type and `omitempty` |
| `default` | `+kubebuilder:default` (strings quoted). Array/object defaults render through `%v` and are NOT valid marker syntax — they break generation; scalars only |
| `enum` | `+kubebuilder:validation:Enum` (string members quoted); an array whose `items` carry `enum` hoists it onto the array field |
| `minimum` / `maximum` / `multipleOf` | corresponding validation markers — **truncated to int** in the marker rendering; fractional bounds lose precision |
| `pattern` | `+kubebuilder:validation:Pattern` |
| `title` / `description` | carried onto the schema node (`title` via the `+kubebuilder:title` marker, requires the pinned controller-tools ≥ v0.18) |
| `additionalProperties: <schema>` | `map[string]T`; an object with ONLY typed additionalProperties collapses to the map itself (error when that object is a `definitions` entry) |
| `additionalProperties: true` | `x-kubernetes-preserve-unknown-fields: true` (via `+kubebuilder:pruning:PreserveUnknownFields`) |
| `additionalProperties: false` | no marker — the struct is closed anyway (unknown fields are pruned by the API server) |
| `oneOf` / `anyOf` / `allOf` | parsed into the model but **silently ignored** by the transpiler — no constraint reaches the CRD |

Envelope (from `Options`): `scope: Namespaced` always; `categories` from
`Options.Categories`; printer columns AGE + (when `Managed`) READY on
`.status.conditions[?(@.type=='Ready')].status`; `Managed` also inlines the
provider-runtime `ConditionedStatus` and `failedObjectRef` into `.status` and makes
the kind satisfy the provider-runtime managed-resource condition interface. The CRD
version name is `Options.GVK.Version` verbatim — callers normalize the chart version
(`1.2.3` → `v1-2-3`) before calling.

## The composition pipeline today (successor implementation + version lifecycle)

Production composition CRDs — including the `Installer` CRD generated from the
installer chart's `values.schema.json` — are produced by this library's successor:
the `crdgen` package in [plumbing](https://github.com/krateo-platformops/plumbing)
(`crdgen/transpile.go`, a direct JSON-Schema→structural-OpenAPI transpiler, no
toolchain exec), driven by
[core-provider](https://github.com/krateo-platformops/core-provider)
(`internal/tools/crd/`). Their rules complete the contract and are documented here
because this repo is the lineage's anchor; each rule below is traced to that source.

### Int-or-string (Kubernetes quantities), proven by the installer CRD

A chart schema meets Kubernetes' `IntOrString`/`Quantity` shape three ways, with three
different fates:

1. **Explicit `x-kubernetes-int-or-string: true` in `values.schema.json`** — carried
   verbatim onto the CRD as a typeless int-or-string node (plumbing
   `crdgen/transpile.go`, `XIntOrString` passthrough). This is the supported way to
   express a quantity, and the installer chart uses it throughout
   (`chart/values.schema.json`, e.g. every `resources.requests/limits.cpu`); the
   generated `Installer` CRD carries the extension on those fields.
2. **A `oneOf`/`anyOf` union of exactly `{type: integer|number}` and
   `{type: string}`** — normalized by core-provider *before* generation
   (`internal/tools/crd/generation/generation.go`, `normalizeIntOrStringUnions`): the
   union is collapsed to `type: string` (a quantity is always writable as a string),
   because a bare union is not structural and previously failed the whole generation.
3. **A bare type union `"type": ["integer","string"]`** — the direct transpiler keeps
   the FIRST non-null type (`primaryType` in plumbing `crdgen/transpile.go`), so the
   field lands as `type: integer` and the string form is dropped (verified against
   plumbing main and v1.13.2). Prefer form 1: state the extension explicitly in the
   chart schema. (Contrast: this library errors on the same input, rule table above.)

`type: ["T","null"]` is the exception union both implementations honor: it maps to
`type: T` + `nullable: true` in the successor (this library rejects it).

### The `vacuum` storage-version design (core-provider)

A composition chart version bump appends a NEW served version to the existing CRD
(`v1-2-3`, `v1-2-4`, … — one per chart version). Kubernetes requires exactly one
`storage: true` version, and converting stored instances between arbitrary chart
schemas is not possible — so core-provider stores everything in a synthetic,
schema-permissive version named `vacuum`:

- **`AppendVersion`** (`internal/tools/crd/generation/generation.go`): when a version
  not yet on the CRD is appended and no `vacuum` version exists — i.e. **on the first
  version bump** of that CRD — it injects `vacuum` with `served: false`,
  `storage: true` and a wide-open schema (`spec`/`status` both
  `x-kubernetes-preserve-unknown-fields: true`). Every non-vacuum version is then set
  `served: true, storage: false`. Conversion strategy is `None` — all real versions
  are schema-passthrough, and losslessness across heterogeneous version schemas comes
  from the permissive storage version, not from conversion
  (`internal/tools/crd/crd.go`, `ApplyOrUpdateCRD`).
- **`RemoveStaleVersions`** (same file): prunes named stale served versions but
  **never removes `vacuum`** — the API server forbids dropping the storage version,
  and keeping it preserves every stored instance regardless of which served version
  wrote it.
- The VERSION printer column (`AddCompositionVersionColumn`) is stamped on every
  served version and skips `vacuum` (never listed by kubectl).

Net contract: a freshly created composition CRD has one real version
(`storage: true`); from the first bump onward it has N served real versions +
`vacuum` as the permanent, non-served storage version.
