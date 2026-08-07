---
type: Runbook
title: How a crdgen release ships
description: The actual release convention — a vX.Y.Z git tag served by the Go module proxy; no OCI artifact, no release workflow; what CI runs and what a consumer pins.
resource: github.com/krateo-platformops/crdgen
tags: [crdgen, release, go-module, tags]
timestamp: 2026-08-07T00:00:00Z
---

# Release runbook

This is a **tag-only Go library**. There is no release workflow, no container image,
no OCI/Helm artifact, no changelog automation — a release IS a semver git tag on
`main`, and the Go module proxy does the distribution. (This differs from the
platform's chart/component repos and their canonical `release-oci.yaml`; stating
reality per the documentation standard.)

## Versioning convention (derived from the actual tags)

- Tags are `vX.Y.Z` **with the `v` prefix** (Go modules require it) — `v0.1.0` …
  `v0.6.0`. Note the contrast with the org's chart repos, which tag plain semver
  without the prefix.
- Still `v0.x`: minor bumps carry features/breaking changes, patch bumps fixes.
- `v0.6.0` (2026-08-03) is the **module-identity migration** tag — the first tag whose
  `go.mod` declares `github.com/krateo-platformops/crdgen`. Tags `v0.5.0` and earlier
  declare the pre-migration module path and are not importable under the new identity;
  treat `v0.6.0` as the floor for new consumers.

## Shipping a release

1. Land the changes on `main` (PR; CI must be green — see below).
2. Tag and push:

   ```sh
   git tag v0.7.0
   git push origin v0.7.0
   ```

3. Nothing else is required: the module proxy serves the tag on first
   `go get github.com/krateo-platformops/crdgen@v0.7.0`. Optionally draft a GitHub
   Release on the tag for human-readable notes (release notes live there, not in the
   repo — [log.md](./log.md) is curated history, not a changelog).
4. Update [llms.txt](./llms.txt)'s pin and the version referenced in
   [usage.md](./usage.md)/README when the release changes what they state.

## What CI runs (`.github/workflows/test.yaml`)

On every push and pull request: `go test -race -coverprofile=… ./…` + Codecov upload.
The two root end-to-end tests (`cdrgen_test.go`, `crdgen_example_test.go`) are behind
the `integration` build tag — they exec the real toolchain — and are therefore NOT
exercised by CI; run them locally with `go test -tags integration .` before tagging.
The PR pipeline also runs the org's documentation-conformance check (`lint-docs`, the
shared workflow from the org's `.github` repo) over this bundle.

## What a consumer pins

A Go `require github.com/krateo-platformops/crdgen vX.Y.Z` — nothing more. The
runtime toolchain contract (the generated module's pinned deps: controller-tools
v0.18.0, controller-runtime v0.20.0, apimachinery v0.33.0, provider-runtime v0.9.1)
ships inside the library (`internal/assets/files/go.mod.tpl`) and changes only with a
crdgen release.
