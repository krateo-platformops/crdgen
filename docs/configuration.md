---
type: Configuration
title: crdgen configuration
description: The whole config surface the library reads — Options fields, the CRDGEN_CLEAN_WORKDIR env var, and the ambient Go-toolchain settings its subprocesses inherit.
resource: github.com/krateo-platformops/crdgen
tags: [crdgen, configuration, env]
timestamp: 2026-08-07T00:00:00Z
---

# Configuration

This is a library: there are no config files, no flags, no build tags. The entire
surface is the `Options` struct passed to `Generate`, one environment variable, and
the Go-toolchain environment inherited by the subprocesses it spawns.

## `Options` (per call)

| Field | Type | Effect |
|---|---|---|
| `WorkDir` | `string` | Names the throwaway module directory (`$TMPDIR/github.com/<legacy-org>/<WorkDir>`) and its module path. Concurrent `Generate` calls MUST use distinct `WorkDir`s — the dir is created, written and (by default) removed per call. |
| `GVK` | `schema.GroupVersionKind` | Used verbatim: `Group` → CRD group, `Version` → the served version name (must already be CRD-legal, e.g. `v1-2-3`), `Kind` → the kind (PascalCase). |
| `Categories` | `[]string` | `+kubebuilder:resource:categories={…}` — e.g. `compositions, comps`. Scope is always `Namespaced`. |
| `SpecJsonSchemaGetter` | `JsonSchemaGetter` | REQUIRED. Source of the spec schema (the chart's `values.schema.json`). |
| `StatusJsonSchemaGetter` | `JsonSchemaGetter` | Optional. Source of a status schema. |
| `Managed` | `bool` | Inlines the provider-runtime `ConditionedStatus` + `failedObjectRef` into `.status`, generates `GetCondition`/`SetConditions`, adds the READY printer column. |
| `Verbose` | `bool` | Routes the library's debug log to stderr; when false the log is discarded. NOTE: this toggles the process-global `log` default output. |

## Environment variables (read by the library)

| Variable | Default | Effect |
|---|---|---|
| `CRDGEN_CLEAN_WORKDIR` | unset | Unset/empty: the temp workdir is removed after generation. Set to ANY non-empty value (the value itself is not interpreted — `FALSE` also keeps it): the workdir is kept for inspection. Debug aid only. |

## Ambient toolchain configuration (inherited, not read)

`Generate` execs `go mod tidy` and `go run … controller-gen` in the workdir; those
subprocesses inherit the caller's environment, so the standard Go variables govern
them even though the library never reads them itself:

- `GOPROXY` / `GONOSUMDB` / `GOFLAGS` — where the generated module's dependencies
  (controller-tools v0.18.0, controller-runtime v0.20.0, apimachinery v0.33.0,
  provider-runtime v0.9.1, pinned in `internal/assets/files/go.mod.tpl`) are fetched
  from; point `GOPROXY` at a mirror for air-gapped use.
- `GOMODCACHE` / `GOPATH` — the cache that makes every call after the first fast.
- `TMPDIR` — where workdirs are created (`os.TempDir()`).

There is nothing else: no config file is searched, no other env var is consulted.
