# AGENTS.md

This file is for autonomous coding agents working in `map-ascii`.
It captures repo-specific commands and coding conventions.
Follow this document unless the user gives explicit overrides.

## Project Summary

- Language: Go (`go 1.22` in `go.mod`).
- Module path: `github.com/Kivayan/map-ascii`.
- App type: CLI that renders an ASCII world map from a PNG land mask.
- Entrypoint: `cmd/map-ascii/main.go`.
- Core logic: `internal/mask.go` and `internal/render.go`.
- Default mask path: `data/landmask_3600x1800.png`.
- Output artifacts are typically written under `out/`.

## Repository Layout

- `cmd/map-ascii/main.go`: flag parsing, argument validation, I/O orchestration.
- `internal/mask.go`: mask loading, validation, lon/lat sampling.
- `internal/render.go`: ASCII generation, marker drawing, char mapping.
- `data/`: static assets (mask PNG files).
- `README.md`: CLI usage and flag docs.

## Environment and Setup

- Check toolchain: `go version`.
- Download modules: `go mod download`.
- Verify dependencies: `go mod verify`.
- After dependency edits: `go mod tidy`.
- No Makefile, no Taskfile, no golangci config in this repo.

## Build and Run Commands

- Run CLI quickly: `go run ./cmd/map-ascii --size 60 --supersample 3`
- Run with marker: `go run ./cmd/map-ascii --size 120 --marker-lon -73.9857 --marker-lat 40.7484`
- Run with explicit mask: `go run ./cmd/map-ascii --mask data/landmask_3600x1800.png --size 120`
- Write output file: `go run ./cmd/map-ascii --size 120 --output out/world_120.txt`
- Build command binary: `go build ./cmd/map-ascii`
- Build all packages: `go build ./...`

## Lint / Static Checks

- Format code: `gofmt -w ./cmd ./internal`
- Optional import normalization: `goimports -w ./cmd ./internal`
- Static analysis: `go vet ./...`
- If both formatters are available, run `goimports` after `gofmt`.

## Code Style Rules

Use idiomatic Go and preserve existing conventions.

- Always run `gofmt` on edited Go files.
- Keep imports in default Go order (stdlib, then module-local).
- Do not alias imports unless there is a collision or clear readability gain.
- Keep functions small and focused.
- Prefer early returns for validation and error paths.
- Prefer simple helpers over deep nested conditionals.
- Keep ASCII behavior intact for marker characters.

## Types and API Design

- Use concrete types in fields and function signatures.
- Use pointers when nil is meaningful or mutation/shared ownership is required.
- Use values for small immutable structs.
- Keep exported surface minimal; prefer unexported identifiers by default.
- Validate boundary invariants at I/O edges and constructors.
- Validate numeric ranges explicitly; fail fast with descriptive errors.

## Naming Conventions

- Exported identifiers: `CamelCase` with domain meaning.
- Unexported identifiers: `camelCase`.
- Use standard Go acronym style (`ASCII`, `PNG`, `URL`) when natural.
- Boolean helpers should read naturally (`isFinite`, `validateMask`).
- Avoid vague abbreviations except common ones (`err`, `ctx`, `lon`, `lat`).

## Error Handling

- Do not panic for expected user/runtime failures.
- Return `error` values and handle them at call sites.
- Wrap low-level failures with `%w` and operation context.
- Keep error strings lowercase and without trailing punctuation.
- Include actionable context (`open mask file`, `decode mask PNG`, etc.).

## CLI and I/O Practices

- Keep flag parsing/wiring in `cmd/map-ascii/main.go`.
- Keep rendering and sampling logic in `internal/`.
- Create parent directories before writes (`os.MkdirAll`).
- Use explicit file permissions (`0o644` for outputs).
- Preserve current defaults unless behavior changes are requested.

## Agent Change Management

- Keep edits minimal and scoped to user intent.
- Do not reformat unrelated files.
- Do not change public behavior unless requested.
- Update `README.md` when CLI flags or behavior change.
- If dependencies change, run `go mod tidy` and include `go.mod`/`go.sum` diffs.

## Quick Pre-Submit Checklist

- Build passes: `go build ./...`
- Formatting applied: `gofmt -w ./cmd ./internal`
- Vet passes: `go vet ./...`
- Tests pass: `go test ./...`
- Docs updated when user-facing behavior changed.

This repository is intentionally small and direct.
Favor readability and correctness over abstraction.
