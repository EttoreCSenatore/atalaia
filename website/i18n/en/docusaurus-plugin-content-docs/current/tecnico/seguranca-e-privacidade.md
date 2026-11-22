---
id: seguranca-e-privacidade
title: Segurança e privacidade
sidebar_position: 3
---

# Security and privacy

## Secure defaults

- Login required, no default password: the account is created with a one-time code printed in the console.
- Listens on `127.0.0.1` by default; with `--listen` the server still refuses requests that do not come from the local network, and checks the `Host` header. See [API and server](api.md).
- Active scans only on the host's own private networks, capped at a /16.
- Active checks are opt-in, rate-limited and written to an audit log.
- Runs without root, with `CAP_NET_RAW`.
- No telemetry and no cloud accounts. **The exception you turn on:** if you configure an alert channel (ntfy, Telegram, e-mail, webhook), the alert text (the device's name, vendor, IP and MAC) is sent to that service. With ntfy.sh or Telegram that means a third party; to keep everything at home, use your own ntfy server or webhook. See [Alerts](../funcionalidades/alertas.md). **Another exception you turn on:** the speed test downloads a file from `speed.cloudflare.com`, which then sees your network's public IP; the app asks for confirmation every time.
- The SQLite database is created with `0600` permissions (folder `0700`).
- Everything that comes from the network (names, vendors) is treated as untrusted and sanitized before it is stored or shown.

## Privacy and data protection

MAC addresses, device names and connection times can identify people (for example, a resident's phone). So:

- Data stays in a local file owned by the user. The project receives nothing.
- History retention will be configurable (M2).
- Whoever runs Atalaia on a third party's network (a company, a building) is the controller of that data and must inform the people affected.

The project's compliance documentation is done with **AI-assisted review** and is **not legal advice**.

## Security self-review (M6.2)

The code went through a manual review of the critical parts (login, sessions, local-network protection, active tools, secrets) and through static analysis with **gosec**, now **in CI** too. What the review found and what was decided:

| Finding | Decision |
|---|---|
| Parallel password attempts all passed the lockout check before any failure was counted (review of 2026-11-19) | Fixed: the attempt is counted **before** the password is checked |
| Changing IP address on the LAN multiplied the attempts | Fixed: besides the per-address limit, a limit for **all of them together** (about 30 a minute, a lockout of at most 1 minute, so an attacker cannot lock the owner out) |
| Many logins at once could exhaust memory (each argon2id uses ~19 MiB) | Fixed: at most 2 password checks at the same time |
| Mass spoofed SSDP answers | Fixed: at most 256 devices per discovery round |
| `int` to `uint16`/`uint32`/`byte` conversions flagged (G115) | Fixed: a port outside 1-65535 becomes port 0 (refused); the argon2 key size is capped at 128 bytes; explicit masks when reading the routing table |
| Logout cookie without `Secure` (G124) | Fixed; new `--secure-cookies` (`ATALAIA_SECURE_COOKIES=1`) marks the cookie `Secure` behind a TLS-terminating proxy |
| `Secure` cookie only when there is HTTPS (G124) | **Accepted**: a `Secure` cookie over plain HTTP would be dropped by the browser |
| File reads by a variable path (G304) | **Accepted**: paths come from administrator flags or system files (`/proc`), never from network input |
| `password` field serialized to JSON (G117) | **Accepted**: it is the SMTP password deliberately stored in the local database (`0600`) and never returned by the API |
| ntfy/webhook addresses may point to the local network | **Accepted**: that is the use case (Home Assistant); only the signed-in administrator configures channels |
| Channel secrets (Telegram token, SMTP password) are stored in clear in the database | **Accepted and documented**: the file is `0600`; protect the database backup as you would a password |
| VPN addresses in CGNAT (for example Tailscale, `100.64.0.0/10`) do not count as "local network" | **Solved with an option**: the server answers 403 by default; `--allow-cidr 100.64.0.0/10` accepts the VPN (see [Deployment](implantacao.md#access-over-a-vpn-tailscale-and-similar)) |

This review was done by the project itself and does **not** replace an independent penetration test.

## v0.7 review (2026-11-22)

A manual review of everything v0.7 added (the health endpoint without login, `/metrics`, automatic backups and "Back up now", the restore test, the copy before migrating, `reset-password`, the package scripts, the public demo), with govulncheck, gosec, staticcheck, `npm audit` and Trivy (vulnerabilities and secrets in the image) finding nothing.

| Finding | Decision |
|---|---|
| A symlink planted under the next backup's name made the copy be written wherever it pointed (only possible in a backup folder other users can write to) | Fixed: the name is checked without following links, the result must be a regular file, and automatic backups **refuse a folder other users can write to** |
| `verify-backup` opens a database file that may come from elsewhere | Hardened: the database opens with `trusted_schema` off (views and triggers of a foreign file cannot run functions with side effects) |
| A long but repetitive `/metrics` token (`aaaa…`) was accepted | Fixed: besides 32+ characters, the token needs at least 10 different characters |
| The cleanup of old scratch copies also ran in the folder of a file given to `verify-backup` | Fixed: only automatic backups clean up, and only in their own folder |
| `/api/health` answers without login and shows the version | **Accepted**: local network only, with the `Host` check; nothing about devices and no error texts |
| `POST /api/backup` can be called repeatedly | **Accepted**: administrator only, CSRF-protected, one backup at a time, and rotation caps how many copies stay |
| `reset-password` asks for no password | **Accepted**: it needs access to the database file, which already gives full control of Atalaia |
| The heartbeat URL is a secret | It never goes to the log or to error messages; it waits at most 10 s and follows no redirects |
| The demo simulates the API with fictitious data | The real build does not carry that code (CI checks it) |

## How to report a vulnerability

Use GitHub's [private vulnerability reporting](https://github.com/EttoreCSenatore/atalaia/security/advisories/new). Do not open a public issue with details and do not send real personal data. The repository's `SECURITY.md` has the same text.

## Supply chain

CI audits dependency vulnerabilities and licenses, and Dependabot keeps dependencies and GitHub Actions up to date, with a 7-day wait; actions and images are pinned by SHA and digest (see [Supply chain](cadeia-de-suprimentos.md)).
