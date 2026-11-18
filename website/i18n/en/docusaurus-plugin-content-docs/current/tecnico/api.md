---
id: api
title: API e servidor
sidebar_position: 2
---

# API and server

`atalaia serve` runs the background scanner and serves the HTTP API (and the web app) in the same process.

```bash
atalaia serve                                  # http://127.0.0.1:8080, scans every 5 minutes
atalaia serve --listen 0.0.0.0:8080            # the whole local network
atalaia serve --interval 2m --allow-host atalaia.lan
```

## First run

There is no default password. On the first run the server prints a one-time **setup code** in its console. Whoever creates the administrator account must read that code, so only someone with access to the server's console can "take over" the installation, even if other people on the network reach the port first.

## Protections

| Protection | What it does |
|---|---|
| Local network only | Requests from public addresses (for example, a port opened on the router) get 403. Loopback, private networks and link-local are accepted. |
| `Host` check | Stops DNS rebinding: only IP addresses, `localhost`, the machine's name and `--allow-host` names are accepted. |
| Session | `HttpOnly` and `SameSite=Strict` cookie (`Secure` when there is TLS). The server stores only the SHA-256 hash of the token. Valid for 7 days. |
| CSRF | Every request that changes something needs the `X-Atalaia` header, and `Origin`, when present, must be the server itself. |
| Password | argon2id (19 MiB, 2 passes), 10 to 128 characters. Changing the password ends the other sessions. |
| Attempts | 5 wrong passwords block the address for 1 minute, doubling up to 15 minutes. |
| Headers | Strict CSP, `nosniff`, `X-Frame-Options: DENY`, `no-referrer`, `no-store` on API responses. |
| Body | JSON up to 4 KiB, no unknown fields. |

:::warning Plain HTTP
The server speaks HTTP. On a trusted local network that is acceptable, but the password and cookie are not encrypted on the wire. For anything beyond that, put a TLS reverse proxy or a VPN in front.
:::

## Endpoints

| Method and path | Login | Description |
|---|---|---|
| `GET /api/status` | no | Version, whether setup is needed, whether there is a session. With a session, includes the scan state. |
| `POST /api/setup` | no | `{code, password}`: creates the administrator and opens the session. |
| `POST /api/login` | no | `{password}`. |
| `POST /api/logout` | yes | Ends the session. |
| `POST /api/password` | yes | `{current, new}`. |
| `GET /api/devices` | yes | Known devices. |
| `GET /api/events?limit=N` | yes | History (1 to 1000), newest first. |
| `POST /api/scan` | yes | Asks for a scan now (answers 202). |
| `GET /api/stream` | yes | Live events (SSE): `hello`, `scan_started`, `scan_finished`, `scan_failed`. |
| `GET/PUT /api/settings` | yes | `retention_days` (0 to 3650; 0 keeps forever; default 90). |

Example, with the session in `cookies.txt`:

```bash
curl -b cookies.txt -H 'X-Atalaia: 1' -X POST http://127.0.0.1:8080/api/scan
```

## Retention

After every scan the server deletes events older than `retention_days` and expired sessions. The devices themselves are not deleted.
