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
| `int` to `uint16`/`uint32`/`byte` conversions flagged (G115) | Fixed: a port outside 1-65535 becomes port 0 (refused); the argon2 key size is capped at 128 bytes; explicit masks when reading the routing table |
| Logout cookie without `Secure` (G124) | Fixed; new `--secure-cookies` (`ATALAIA_SECURE_COOKIES=1`) marks the cookie `Secure` behind a TLS-terminating proxy |
| `Secure` cookie only when there is HTTPS (G124) | **Accepted**: a `Secure` cookie over plain HTTP would be dropped by the browser |
| File reads by a variable path (G304) | **Accepted**: paths come from administrator flags or system files (`/proc`), never from network input |
| `password` field serialized to JSON (G117) | **Accepted**: it is the SMTP password deliberately stored in the local database (`0600`) and never returned by the API |
| ntfy/webhook addresses may point to the local network | **Accepted**: that is the use case (Home Assistant); only the signed-in administrator configures channels |
| Channel secrets (Telegram token, SMTP password) are stored in clear in the database | **Accepted and documented**: the file is `0600`; protect the database backup as you would a password |
| VPN addresses in CGNAT (for example Tailscale, `100.64.0.0/10`) do not count as "local network" | **Known limit**: the server answers 403; use the local-network IP or a proxy on the same machine |

This review was done by the project itself and does **not** replace an independent penetration test.

## How to report a vulnerability

Use GitHub's [private vulnerability reporting](https://github.com/EttoreCSenatore/atalaia/security/advisories/new). Do not open a public issue with details and do not send real personal data. The repository's `SECURITY.md` has the same text.

## Supply chain

CI audits dependency vulnerabilities and licenses, and Dependabot keeps dependencies and GitHub Actions up to date.
