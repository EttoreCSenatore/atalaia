---
id: desenvolvimento
title: Desenvolvimento
sidebar_position: 4
---

# Development

## Requirements

- Go (version in `go.mod`) for the agent.
- Node 24+ for the documentation site.
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

## Workflow

One step per pull request: branch, commits, green CI, diff review and squash merge. Documentation (pt-BR and English) and project memory are updated in the same step.
