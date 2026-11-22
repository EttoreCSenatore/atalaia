# Atalaia

**English** · [Português](README.md)

[![CI](https://github.com/EttoreCSenatore/atalaia/actions/workflows/ci.yml/badge.svg)](https://github.com/EttoreCSenatore/atalaia/actions/workflows/ci.yml)
[![License: AGPL-3.0-or-later](https://img.shields.io/badge/license-AGPL--3.0--or--later-blue.svg)](LICENSE)
[![Docs](https://img.shields.io/badge/docs-pt--BR%20%7C%20EN-1d4ed8.svg)](https://ettorecsenatore.github.io/atalaia/en/)

**Watch your network.** A free, open-source, self-hosted network watchtower: discover the devices on your network, track history, get alerts and run **defensive** security checks. A Go agent with the web app embedded.

> **State: v0.6.** `atalaia serve` discovers and identifies the devices, keeps the history, serves a web app with login (Linux, Docker, .deb/.rpm packages), runs defensive checks and alerts you through ntfy, Telegram, webhook or e-mail when a new device appears. See the [roadmap](https://ettorecsenatore.github.io/atalaia/en/docs/roadmap).

## Try it (Linux)

```bash
go run ./cmd/atalaia scan      # discover the devices on your network
go run ./cmd/atalaia devices   # list them
go run ./cmd/atalaia history   # what changed
go run ./cmd/atalaia serve     # web app with login at http://127.0.0.1:8080
```

With Docker: `cd deploy && docker compose up -d --build` (the setup code shows in `docker compose logs atalaia`). Step by step in [Getting started](https://ettorecsenatore.github.io/atalaia/en/docs/manual/primeiros-passos); questions in the [FAQ](https://ettorecsenatore.github.io/atalaia/en/docs/manual/faq).

![Atalaia dashboard (fictitious data)](website/static/img/screenshots/app-painel-en.png)

## What it does

- **Device discovery** (ARP table, mDNS, reverse DNS) with vendor and name.
- **History and alerts**: online/offline timeline and new-device notifications (ntfy, Telegram, e-mail, webhook).
- **Passive and active security**: signs of ARP spoofing, rogue DHCP (opt-in), ports and risky services, TLS, exposed web interfaces. No CVE matching, on purpose.
- **Network tools**: ping, traceroute, DNS, Wake-on-LAN, download speed.

## Principles

- **Defensive**: for auditing **your own** network. Scope fixed to the host's own private networks; active checks are opt-in and logged.
- **Secure by default**: login required, LAN-only, no root, no telemetry.
- **Free software**: code AGPL-3.0-or-later ([LICENSE](LICENSE)); documentation CC BY-SA 4.0 ([LICENSE-DOCS.md](LICENSE-DOCS.md)).
- Independent project: not affiliated with Fing Limited and does not use its trademarks or data (see [NOTICE](NOTICE)).

## Development

```bash
go run ./cmd/atalaia scope          # networks Atalaia may probe
go vet ./... && go test -race ./...
cd website && npm ci && npm start   # documentation
```

Without Go installed: `docker run --rm -v "$PWD":/src -w /src golang:1-alpine go test ./...`. See [Development](https://ettorecsenatore.github.io/atalaia/en/docs/tecnico/desenvolvimento). Vulnerabilities: [SECURITY.md](SECURITY.md). To contribute: [CONTRIBUTING.md](CONTRIBUTING.md) and the [Code of Conduct](CODE_OF_CONDUCT.md).

Use only on networks you administer or have written authorization for.
