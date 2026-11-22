---
id: desenvolvimento
title: Desenvolvimento
sidebar_position: 4
---

# Development

## Requirements

- Go (version in `go.mod`) for the agent.
- Node 24+ for the documentation site (CI and the image use Node 26, LTS).
- Python 3 for the CI check scripts.

Without Go installed, Docker works:

```bash
docker run --rm -v "$PWD":/src -w /src golang:1-alpine go test ./...
```

## Agent

```bash
go run ./cmd/atalaia version   # version
go run ./cmd/atalaia scope     # networks Atalaia may probe
go run ./cmd/atalaia scan      # discover and store the devices
go run ./cmd/atalaia serve     # API at http://127.0.0.1:8080
go vet ./... && go test -race ./...
```

## Web app

```bash
cd web && npm ci
npm run dev          # http://localhost:5173, uses the API of `atalaia serve` on port 8080
npm run check && npm test
npm run build        # writes internal/web/dist, which Go embeds
```

To redo the documentation images (fictitious data): `PLAYWRIGHT=/path/to/playwright/index.mjs npm run screenshots -- ../website/static/img/screenshots`.

## Public demo

`npm run build:demo` (in `web/`) builds the app in demo mode into `web/dist-demo/`, for `/atalaia/demo/`: the API is answered in the browser by `src/lib/demo.ts` with the fictitious data from `tests/fixtures.mjs` (the same as the screenshots), routes live in the `#` (GitHub Pages only serves files) and actions that would touch a network are refused with a message. CI publishes it with the docs and checks that the normal build does **not** carry the demo code or data (`VITE_DEMO` becomes a constant at build time).

## Vendor list (IEEE)

The file `internal/identity/oui.tsv.gz` is built from the IEEE public lists. To update it:

```bash
python3 -I scripts/atualizar-oui.py
```

## Documentation

```bash
cd website && npm ci && npm start    # local server, in Portuguese
npm run start -- --locale en         # in English
```

Every page in `website/docs/` needs an English version in `website/i18n/en/docusaurus-plugin-content-docs/current/` with the same structure. CI checks this with `scripts/checar-traducao.py`.

## Quality and security checks

```bash
gofmt -l .                                  # must be empty
go vet ./... && go test -race ./...
go run honnef.co/go/tools/cmd/staticcheck@v0.8.2 ./...
go run github.com/securego/gosec/v2/cmd/gosec@v2.30.0 -quiet -exclude=G304,G117 ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.10.0 ./...
```

CI runs the same checks, plus the dependency license check, Trivy on the image and a smoke test of the container.

## Packages and releases

The `.deb`/`.rpm` packages are built by [nfpm](https://nfpm.goreleaser.com/) from `deploy/nfpm.yaml`. A release goes out when a `vX.Y.Z` tag is pushed: the `release.yml` workflow builds the binaries (amd64, arm64, armv7), builds the packages, publishes the image to GHCR and creates the release with `SHA256SUMS`. Before that, move the "Unreleased" entries of `CHANGELOG.md` under the new version and add the version at the top of `releases.yaml`; the release checks that the tag is the first version there and uses its text. Step by step in [Versions and what's new](versionamento.md).

## Workflow

One step per pull request: branch, commits, green CI, diff review and squash merge. Documentation (pt-BR and English) and project memory are updated in the same step. For outside contributions, [CONTRIBUTING](https://github.com/EttoreCSenatore/atalaia/blob/main/CONTRIBUTING.md) sums up the rules, and the issue (bug, idea) and PR templates carry the checklist; `CODEOWNERS` requests the maintainer's review on every PR.
