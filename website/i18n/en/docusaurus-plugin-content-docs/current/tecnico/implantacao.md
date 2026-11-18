---
id: implantacao
title: Implantação
sidebar_position: 3
---

# Deployment

## Docker (recommended)

The image is about 26 MB, runs without root, with no extra privileges and a read-only file system.

```bash
git clone https://github.com/EttoreCSenatore/atalaia && cd atalaia/deploy
docker compose up -d --build
docker compose logs atalaia          # look for the "Setup code"
```

The container uses the **host network** (`network_mode: host`): without it, it would only see Docker's virtual network and would not discover the real devices. Data lives in the `atalaia-data` volume.

## Listen on the local network

By default the server listens on `127.0.0.1:8080` only. To open it from a phone or another computer on the network:

```bash
ATALAIA_LISTEN=0.0.0.0:8080 docker compose up -d
```

The server still refuses requests that do not come from the local network, but the traffic is plain HTTP. For use beyond a trusted network, put a TLS reverse proxy in front (for example, Caddy) or use a VPN, and pass the name with `--allow-host`.

## Binary

```bash
cd web && npm ci && npm run build && cd ..
go build -o atalaia ./cmd/atalaia
./atalaia serve --listen 0.0.0.0:8080
```

## Backup and updates

All state is one SQLite file (`/data/atalaia.db` in the container). To back it up, copy the volume or the file while the server is stopped. To update, run `docker compose up -d --build` again: the database migrates by itself, and an older version refuses to open a newer database.

## What CI checks

CI builds the image, validates the compose file, scans for vulnerabilities with Trivy (fails on critical ones with a fix) and runs a smoke test: the server starts, the app and API answer, the CSP is present, `/api/devices` requires login and a forged `Host` header is refused.
