---
id: cadeia-de-suprimentos
title: Supply chain
sidebar_position: 8
---

# Supply chain

Atalaia runs on your network with access to it, so what goes into the binary and the image matters as much as the code. These rules apply to CI, the release and the Docker image.

## What is pinned

| What | How | Example |
|---|---|---|
| GitHub actions (`uses:`) | Commit SHA, with the version in a comment | `actions/checkout@3d3c42e… # v7.0.1` |
| Dockerfile base images (`FROM`) | `sha256` digest besides the tag | `golang:1-alpine@sha256:70bd7dd…` |
| Images used in CI (Trivy) | `sha256` digest besides the tag | `aquasec/trivy:0.76.2@sha256:8f6bb94…` |
| Go tools (`go run`) | Exact version; Go checks the hash against the checksum database | `govulncheck@v1.10.0` |
| Go and npm dependencies | `go.sum` and `package-lock.json` committed | — |

A tag like `@v7` or `:1-alpine` can start pointing at different code at any time, even if the author's account is taken over. A SHA or digest does not change.

## CI check

`scripts/checar-fixacao.py` runs in the Go job and fails on an action without a SHA (or without the version comment), a `FROM` or `docker run` without a digest, or a `go run` without an exact version. To run it locally:

```bash
python3 -I scripts/checar-fixacao.py
```

## Updates (Dependabot)

- **7-day wait** (`cooldown`): Dependabot only proposes versions published at least 7 days ago. Compromised packages are usually found and pulled within that window. Dependabot security alerts do not wait.
- **Grouped PRs**: minor and patch updates come in one PR per ecosystem (Go, web, docs, actions); major versions come separately, to review with care.
- **Open-PR limit** per ecosystem, so the queue does not grow unreviewed.
- Dependabot updates the SHA or digest **and** the version comment together.

The Trivy image in CI is out of Dependabot's reach (it is in a `docker run`); it is updated by hand, with the new digest.

## Updating by hand

For an action, get the commit SHA of the version (`gh api repos/OWNER/ACTION/commits/vX.Y.Z --jq .sha`) and write `OWNER/ACTION@SHA # vX.Y.Z`. For an image, get the multi-platform index digest from the registry and write `image:tag@sha256:…`. Then run the script above.

## In the release

The image published on GHCR ships with provenance and an SBOM (`provenance: true`, `sbom: true`), and the binaries and packages with `SHA256SUMS`.
