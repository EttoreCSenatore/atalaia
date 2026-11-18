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
| `GET /api/devices` | yes | Known devices, with `model`, `kind` (the guessed type) and `services` (advertised mDNS services). |
| `PUT /api/devices/{mac}` | yes | `{label, tags, known}`: name (up to 64 characters), up to 8 tags (24 characters, no commas) and whether it is known. |
| `POST /api/devices/{mac}/wake` | yes | Sends the magic packet (Wake-on-LAN). Inventory devices in scope only. |
| `POST /api/devices/{mac}/portscan` | yes | `{confirm: true}` required. Scans ~65 TCP ports; at most one per 10 s; returns `ports` and, if there was a previous result, `opened`/`closed`. |
| `POST /api/devices/{mac}/exposure` | yes | `{confirm: true}`; needs a port scan first (409). Returns `risks`, `web` and `tls`. Same limits as the scan. |
| `GET /api/devices/{mac}/exposure` | yes | The last exposure report. |
| `GET /api/devices/{mac}/ports` | yes | The device's last port scan. |
| `POST /api/tools/ping` | yes | `{host, count}` (1 to 10). Unprivileged ICMP; 501 if the system does not allow it. |
| `POST /api/tools/dns` | yes | `{name, type}` (A, AAAA, CNAME, MX, NS, TXT, PTR). |
| `POST /api/tools/traceroute` | yes | `{host}`: hops to the destination (IPv4, up to 20, no privileges). |
| `POST /api/tools/speedtest` | yes | `{confirm: true}`: downloads a test file from a third-party server and returns Mbit/s and latency. |
| `GET /api/dhcp` | yes | Known DHCP servers, whether the scheduled check is on, and the last result. |
| `PUT /api/dhcp` | yes | `{enabled}`: turns the check after every scan on or off. Disabled in observe-only mode. |
| `PUT /api/dhcp/names` | yes | `{enabled}`: turns learning names from DHCP requests on or off (listen-only; port 67). |
| `POST /api/dhcp/check` | yes | `{confirm: true}`: one DHCP discovery now (needs port 68). |
| `POST /api/dhcp/reset` | yes | `{confirm: true}`: forgets the known servers. |
| `GET /api/audit` | yes | Log of active actions, newest first. |
| `GET /api/events?limit=N` | yes | History (1 to 1000), newest first. Kinds: `joined`, `online`, `offline`, `ip_changed`, `gateway_changed`, `multi_ip`, `rogue_dhcp`. |
| `POST /api/scan` | yes | Asks for a scan now (answers 202). |
| `GET /api/stream` | yes | Live events (SSE): `hello`, `scan_started`, `scan_finished`, `scan_failed`. |
| `GET/PUT /api/settings` | yes | `retention_days` (0 to 3650; 0 keeps forever; default 90) and `port_scan_hours` (0 = off, up to 720; only with active tools). You can send either. |
| `GET/PUT /api/alerts` | yes | `{rules: {new_device, known_offline, arp_anomaly, rogue_dhcp, new_ports}, channels: [...]}`. Secrets (`token`, `password`) are never returned by GET; in PUT, a channel sent without a secret keeps the one already saved. Up to 8 channels, body up to 32 KiB. |
| `POST /api/alerts/test` | yes | `{id}`: sends a test message through a saved channel. |

Example, with the session in `cookies.txt`:

```bash
curl -b cookies.txt -H 'X-Atalaia: 1' -X POST http://127.0.0.1:8080/api/scan
```

## Retention

After every scan the server deletes events older than `retention_days` and expired sessions. The devices themselves are not deleted.
