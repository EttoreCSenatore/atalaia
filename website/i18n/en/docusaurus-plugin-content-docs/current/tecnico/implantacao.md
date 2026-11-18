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

The server still refuses requests that do not come from the local network, but the traffic is plain HTTP. Turn on the [built-in HTTPS](#built-in-https) (`--tls-self-signed`), put a TLS reverse proxy in front (for example, Caddy) or use a VPN, and pass the name with `--allow-host`.

## Ready-made releases

Every release (from v0.1.0) publishes, on the [Releases](https://github.com/EttoreCSenatore/atalaia/releases) page:

- Linux binaries (`amd64`, `arm64`, `armv7`) with the web app inside, and a `SHA256SUMS` file (check it with `sha256sum -c SHA256SUMS`);
- the image `ghcr.io/ettorecsenatore/atalaia:0.1.0` (and `latest`), for `amd64` and `arm64`, with SBOM and provenance.

```bash
docker run -d --name atalaia --network host --read-only --cap-drop ALL \
  -v atalaia-data:/data ghcr.io/ettorecsenatore/atalaia:latest
docker logs atalaia                  # look for the "Setup code"
```

## Built-in HTTPS

Atalaia can serve HTTPS by itself, without a proxy (TLS 1.2 or higher; the session cookie becomes `Secure`):

| Option | When to use it |
|---|---|
| `--tls-self-signed` (or `ATALAIA_TLS_SELF_SIGNED=1`) | Home network. Atalaia creates a self-signed certificate in `<database folder>/tls/` (ECDSA P-256, 1 year, `0600` key) covering `localhost`, the machine name, the local addresses and any `--allow-host`, and **renews it by itself** 30 days before it expires. The browser will warn once: compare the **SHA-256 fingerprint** printed on the console before accepting. |
| `--tls-cert file.pem --tls-key key.pem` (or `ATALAIA_TLS_CERT` / `ATALAIA_TLS_KEY`) | You already have a certificate (for example from an internal CA, or Let's Encrypt via DNS). A replaced file is picked up without a restart. With this option the server also sends `Strict-Transport-Security`. |

```bash
atalaia serve --listen 0.0.0.0:8443 --tls-self-signed
```

HSTS is **not** sent with the self-signed certificate: it would stop the browser from letting you accept the warning. Behind a proxy that terminates TLS, Atalaia sees plain HTTP; in that case prefer the built-in HTTPS, or make sure the proxy talks to it only over the local machine.

## .deb and .rpm packages (systemd)

Every release publishes `.deb` (Debian, Ubuntu, Raspberry Pi OS) and `.rpm` (Fedora, RHEL and derivatives) packages for `amd64`, `arm64` and `armv7`, listed in `SHA256SUMS`.

```bash
sudo apt install ./atalaia_0.4.0_amd64.deb        # or: sudo dnf install ./atalaia-0.4.0-1.x86_64.rpm
sudo systemctl enable --now atalaia
journalctl -u atalaia | grep -A1 "Setup code"    # first-run setup code
```

The package installs the binary at `/usr/bin/atalaia`, the `atalaia.service` unit and the options file `/etc/default/atalaia` (which an upgrade never overwrites). **Nothing is turned on unless you ask.** The unit runs as a **throwaway unprivileged user** (`DynamicUser`), with no capabilities, a read-only file system (only `/var/lib/atalaia` is writable), no access to `/home`, a system-call filter and IP sockets only. Options (`ATALAIA_LISTEN`, HTTPS, observe-only mode…) go in `/etc/default/atalaia`.

## Observe-only mode

With `--no-active-tools` (or `ATALAIA_NO_ACTIVE_TOOLS=1`) the server **only observes**: ping, Wake-on-LAN, port scan and exposure check are off (the server answers 403 and the app hides the buttons). Passive discovery, history, alerts and DNS lookup keep working. It is the right mode when Atalaia runs on a network where you do not want anyone, not even a compromised administrator, to be able to trigger active actions.

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
