---
id: configuracao
title: Configuration reference
sidebar_position: 1
---

# Configuration reference

Everything you can configure in Atalaia, in one place. Every command-line option is `--name`; the ones with an environment variable can go in `/etc/default/atalaia` (packages) or under `environment:` (Docker). CI checks that this page names every command, option and variable in the code.

## Commands

| Command | What |
|---|---|
| `atalaia serve` | Runs the web app and the background scans (the normal mode) |
| `atalaia scan` | One scan now, from the command line |
| `atalaia devices` | Lists the devices (`--json` for JSON) |
| `atalaia history` | What happened (joined, online, offline, new address) |
| `atalaia export` | The list as CSV (spreadsheet-safe) or JSON |
| `atalaia backup` | A consistent copy of the database, while the server runs |
| `atalaia verify-backup` | Restore test of a copy |
| `atalaia reset-password` | Forgot the password: removes the account (devices stay) |
| `atalaia health` | Asks a running server for its health (exits 0 or 1) |
| `atalaia scope` | The networks Atalaia may probe |
| `atalaia version` | The version |

## `atalaia serve`

| Option | Variable | Default | What |
|---|---|---|---|
| `--db` | `ATALAIA_DB` | see [Files](#files-and-folders) | Database file (all commands) |
| `--listen` | `ATALAIA_LISTEN` | `127.0.0.1:8080` | Address and port; `0.0.0.0:8443` opens it to the local network |
| `--interval` | — | `5m` | Time between scans (0 = only when asked) |
| `--allow-host` | — | — | Extra name the app may be opened with (repeatable) |
| `--allow-cidr` | `ATALAIA_ALLOW_CIDR` | — | Extra network treated as local, e.g. `100.64.0.0/10` (Tailscale); repeatable, commas in the variable |
| `--tls-self-signed` | `ATALAIA_TLS_SELF_SIGNED=1` | off | HTTPS with a self-signed certificate Atalaia makes and renews |
| `--tls-cert`, `--tls-key` | `ATALAIA_TLS_CERT`, `ATALAIA_TLS_KEY` | — | HTTPS with your certificate (reloaded when it changes; turns on HSTS) |
| `--secure-cookies` | `ATALAIA_SECURE_COOKIES=1` | off | `Secure` cookie even over HTTP (behind a TLS proxy) |
| `--no-active-tools` | `ATALAIA_NO_ACTIVE_TOOLS=1` | off | Observe-only mode: no ping, traceroute, speed test, Wake-on-LAN, port scan, exposure or DHCP check |
| `--heartbeat-scan` | `ATALAIA_HEARTBEAT_SCAN` | — | URL pinged after every scan ([Monitoring](../tecnico/monitoramento.md)) |
| `--metrics-token-file` | `ATALAIA_METRICS_TOKEN_FILE`, `ATALAIA_METRICS_TOKEN` | off | Turns on `/metrics` with this token (32+ characters); also reads the systemd credential `metrics-token` |
| `--backup-dir` | `ATALAIA_BACKUP_DIR` | `backups/` next to the database | Folder for automatic backups |
| `--backup-keep` | `ATALAIA_BACKUP_KEEP` | `7` | Automatic backups to keep (0 turns them off) |
| `--backup-interval` | — | `24h` | Time between automatic backups (at least 1h) |
| `--heartbeat-backup` | `ATALAIA_HEARTBEAT_BACKUP` | — | URL pinged after every automatic backup |

A device goes offline after `--interval` × 3 without being seen (at least 10 minutes).

## Other commands

| Command | Option | Default | What |
|---|---|---|---|
| `scan` | `--no-names` | off | Do not look up names (reverse DNS, mDNS) |
| `scan` | `--offline-after` | `10m` | Time without seeing a device before marking it offline |
| `devices` | `--json` | off | JSON output |
| `history` | `--limit` | `30` | How many events |
| `export` | `--format` | `csv` | `csv` or `json` |
| `backup` | `--dir`, `--keep` | —, `7` | A dated copy in a folder, keeping the newest N (instead of one file) |
| `backup` | `--heartbeat` | `ATALAIA_HEARTBEAT_BACKUP` | URL pinged after the backup |
| `verify-backup` | `--dir` | — | Tests the newest copy in the folder (or pass a file) |
| `reset-password` | `--yes` | — | Without `--yes` it only explains; stop the server first |
| `health` | `--url` | from `ATALAIA_LISTEN` and `ATALAIA_TLS_*` | Server to ask |

## Environment variables

| Variable | Same as |
|---|---|
| `ATALAIA_DB` | `--db` |
| `ATALAIA_LISTEN` | `--listen` (also read by `atalaia health`) |
| `ATALAIA_ALLOW_CIDR` | `--allow-cidr` (comma separated) |
| `ATALAIA_TLS_SELF_SIGNED` | `--tls-self-signed` (value `1`) |
| `ATALAIA_TLS_CERT`, `ATALAIA_TLS_KEY` | `--tls-cert`, `--tls-key` |
| `ATALAIA_SECURE_COOKIES` | `--secure-cookies` (value `1`) |
| `ATALAIA_NO_ACTIVE_TOOLS` | `--no-active-tools` (value `1`) |
| `ATALAIA_HEARTBEAT_SCAN` | `--heartbeat-scan` |
| `ATALAIA_HEARTBEAT_BACKUP` | `--heartbeat-backup` (`serve`) and `--heartbeat` (`backup`) |
| `ATALAIA_METRICS_TOKEN_FILE` | `--metrics-token-file` |
| `ATALAIA_METRICS_TOKEN` | The token itself (prefer the file) |
| `ATALAIA_BACKUP_DIR`, `ATALAIA_BACKUP_KEEP` | `--backup-dir`, `--backup-keep` |

## Settings in the app

Kept in the database and changed in **Settings** (or through the [API](../tecnico/api.md)):

| Setting | Default |
|---|---|
| History retention | 90 days (0 keeps everything) |
| Forget devices offline for more than N days | never |
| Scheduled port scans of known devices | off |
| Alerts: rules and channels (ntfy, Telegram, webhook, e-mail) | new device, ARP, rogue DHCP and new ports on; no channels |
| Rogue-DHCP check and learning names from DHCP | off |
| Language | the browser's (pt-BR or English) |

## Files and folders

| What | Docker | Package (.deb/.rpm) | Binary |
|---|---|---|---|
| Database | `/data/atalaia.db` | `/var/lib/atalaia/atalaia.db` | `$ATALAIA_DB` or `~/.local/share/atalaia/atalaia.db` |
| Automatic backups | `/data/backups/` | `/var/lib/atalaia/backups/` | `backups/` next to the database |
| Copies before migrating | next to the database (`atalaia-pre-migration-*.db`) | same | same |
| Self-signed certificate | `/data/tls/` | `/var/lib/atalaia/tls/` | `tls/` next to the database |
| Options | compose `environment:` | `/etc/default/atalaia` | command line |

With the package, `/var/lib/atalaia` points to `/var/lib/private/atalaia` (that is how systemd keeps a `DynamicUser`'s data).

## Network

| Port | Direction | When |
|---|---|---|
| TCP 8080 (or the `--listen` one) | in | The web app and the API; local network only |
| UDP (sweep) | out, local network only | Every scan |
| UDP 5353, 1900, 137 | out, local network only | mDNS, SSDP and NetBIOS to identify devices |
| UDP 68 / 67 | in and out | Only with the rogue-DHCP check / DHCP name learning on (needs root or `CAP_NET_BIND_SERVICE`) |
| Internet | out | Only what you turn on: alerts, speed test (speed.cloudflare.com), heartbeats |
