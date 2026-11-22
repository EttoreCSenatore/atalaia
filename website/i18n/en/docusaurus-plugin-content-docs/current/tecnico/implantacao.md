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

## Access over a VPN (Tailscale and similar)

By default the server only answers "local" addresses (loopback, private and link-local). The addresses of VPNs such as **Tailscale** (`100.64.0.0/10`, the CGNAT range) are **not** in that list and get a 403. To accept them, tell Atalaia which network you trust:

```bash
atalaia serve --listen 0.0.0.0:8443 --tls-self-signed --allow-cidr 100.64.0.0/10
# or: ATALAIA_ALLOW_CIDR=100.64.0.0/10   (several, comma separated)
```

Networks that are too wide are refused (anything shorter than /8 for IPv4). Use it together with [HTTPS](#built-in-https) and pass the machine's name with `--allow-host`.

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

What the package scripts do:

| When | What |
|---|---|
| Install | Reloads systemd and shows how to turn the service on; does **not** enable or start Atalaia |
| Upgrade | Reloads systemd and restarts the service **only if it was already running**; keeps your `/etc/default/atalaia` |
| Removal | Stops and disables the service; **keeps the data** (database, backups and certificate in `/var/lib/private/atalaia`), even with `apt purge`: delete it by hand if you no longer need it |

On every release, the published packages are tested as an **upgrade from the previous release** before the release goes out. CI installs the packages in clean Debian 13 and Fedora 44 containers on every change: it installs them, runs `systemd-analyze verify` on the unit, upgrades to a newer version (with a changed `/etc/default/atalaia`), removes them and, on Debian, purges them, checking files, modes and that the data stays.

## Observe-only mode

With `--no-active-tools` (or `ATALAIA_NO_ACTIVE_TOOLS=1`) the server **only observes**: ping, Wake-on-LAN, port scan and exposure check are off (the server answers 403 and the app hides the buttons). Passive discovery, history, alerts and DNS lookup keep working. It is the right mode when Atalaia runs on a network where you do not want anyone, not even a compromised administrator, to be able to trigger active actions.

## Binary

```bash
cd web && npm ci && npm run build && cd ..
go build -o atalaia ./cmd/atalaia
./atalaia serve --listen 0.0.0.0:8080
```

## Backup and updates

All state is one SQLite file (`/data/atalaia.db` in the container, `/var/lib/atalaia/atalaia.db` with the package).

**Automatic backups (on by default).** While it runs, `atalaia serve` copies the database once a day into the `backups/` folder next to it (`/data/backups` in the container, `/var/lib/atalaia/backups` with the package), keeps the newest 7 and, once a week, **test-restores** the newest one: it restores it into a scratch copy, opens it as Atalaia would and runs SQLite's integrity check. It works the same with Docker, the packages and the binary. If the machine was off, the late backup runs as soon as the server is back. The app shows the last backup and the last test in **Settings > Backups** (with a **Back up now** button) and a notice at the top when the last backup is more than 26 hours old.

| Option | Variable | Default |
|---|---|---|
| `--backup-dir` | `ATALAIA_BACKUP_DIR` | `backups/` next to the database |
| `--backup-keep` | `ATALAIA_BACKUP_KEEP` | 7 (0 turns scheduled backups off) |
| `--backup-interval` | — | 24h (at least 1h) |
| `--heartbeat-backup` | `ATALAIA_HEARTBEAT_BACKUP` | off (see [Monitoring](monitoramento.md)) |

The copies are **on the same disk** as the database: they protect against mistakes and corruption, not against losing the disk. Copy the folder somewhere else now and then (another disk, a NAS). To check a copy by hand: `atalaia verify-backup FILE` or `atalaia verify-backup --dir FOLDER` (the newest).

For a one-off consistent copy **without stopping the server** (or from your own scheduler, with `--dir FOLDER --keep N`):

```bash
atalaia backup --db /var/lib/atalaia/atalaia.db ~/atalaia-$(date +%F).db
```

The copy is created with mode `0600` and never overwrites an existing file: it holds the device list **and the alert channels' secrets**, so keep it as you would a password. To restore, stop the server and put the copy in place of the database. To update, install the new version (or `docker compose up -d --build`): the database migrates by itself, and an older version refuses to open a newer database. **Before migrating, Atalaia copies the database** to `atalaia-pre-migration-schemaN-DATE.db` next to it (mode `0600`, the newest 3 are kept); if the copy fails (a full disk, for example), it does not migrate and changes nothing. To go back, stop the server, install the previous version and put that copy in place of the database. CI tests the upgrade from databases made by **every published version** (see [Versions and what's new](versionamento.md)).

To take the list to a spreadsheet: `atalaia export --format csv` (or `json`), or the **Download the list** link on the dashboard. The CSV neutralizes formulas: names coming from the network never become spreadsheet commands.

To know whether Atalaia is working (a health endpoint without login, Healthchecks.io pings after every scan and backup, Prometheus metrics and the Docker image's `HEALTHCHECK`), see [Monitoring](monitoramento.md).

## What CI checks

CI builds the image, validates the compose file, scans for vulnerabilities with Trivy (fails on critical ones with a fix) and runs a smoke test: the server starts, the app and API answer, the CSP is present, `/api/devices` requires login and a forged `Host` header is refused.
