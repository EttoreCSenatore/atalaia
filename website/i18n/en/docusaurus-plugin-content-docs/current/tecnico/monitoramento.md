---
id: monitoramento
title: Monitoring
sidebar_position: 4
---

# Monitoring

How to know that Atalaia itself is working. Everything here is optional except the health endpoint, and none of it reports devices, addresses or names from your network: only states, counts, times and sizes.

## Health endpoint: `GET /api/health`

Answers **without login** (but only to the local network, like the rest of the server), for Uptime Kuma, Gatus and the like:

```json
{"status": "ok", "version": "v0.7.0", "database": "ok", "last_scan": "2026-11-22T12:00:00Z"}
```

| `status` | HTTP | When |
|---|---|---|
| `ok` | 200 | All good |
| `degraded` | 200 | Only warnings in `warnings`: `disk_low` (the database's disk is almost full) |
| `error` | 503 | Problems in `problems`: `database` (the database does not answer), `last_scan_failed` (the last scan failed), `scans_stale` (no scan for more than 3 intervals) |

Error texts never show here (they may name the network interface); see the server log or the app. The version shows without login: anyone on your local network can tell which version you run.

In Uptime Kuma: an **HTTP(s)** monitor pointing at `http://ADDRESS:8080/api/health` (with the self-signed HTTPS certificate, tick "ignore TLS errors").

### `atalaia health` and Docker

`atalaia health` queries that endpoint and exits 0 when all is good, 1 when not. It reads the same variables as `serve` (`ATALAIA_LISTEN`, `ATALAIA_TLS_*`) or takes `--url`. The Docker image already uses it as its `HEALTHCHECK`, so `docker ps` shows `healthy` or `unhealthy`.

## Heartbeats (Healthchecks.io and similar)

A service such as [Healthchecks.io](https://healthchecks.io/) alerts you when a ping **stops arriving**, which catches even the whole machine going down.

| Signal | How to turn it on | What it sends |
|---|---|---|
| Scan | `--heartbeat-scan URL` or `ATALAIA_HEARTBEAT_SCAN` | `GET URL` after every scan; `GET URL/fail` when it fails |
| Backup | `atalaia backup --heartbeat URL` or `ATALAIA_HEARTBEAT_BACKUP` | `GET URL` when the backup works; `GET URL/fail` when it fails |

The URL is a secret (whoever has it can send fake pings): it never shows in the logs. It waits at most 10 seconds and follows no redirects. In Healthchecks.io, set the period to the scan interval (5 minutes by default) with a grace time of a few minutes.

## Prometheus: `/metrics`

Off by default. Turn it on with a token of 32 characters or more, in one of these ways:

- `--metrics-token-file FILE` or `ATALAIA_METRICS_TOKEN_FILE` (best: the token does not show in the process list);
- a systemd credential named `metrics-token` (`LoadCredential=metrics-token:/etc/atalaia/metrics-token`, see `/etc/default/atalaia`);
- `ATALAIA_METRICS_TOKEN` (handy with Docker).

```yaml
scrape_configs:
  - job_name: atalaia
    authorization:
      credentials_file: /etc/prometheus/atalaia-token
    static_configs:
      - targets: ["192.168.1.10:8080"]
```

| Metric | What |
|---|---|
| `atalaia_build_info{version}` | Running version (always 1) |
| `atalaia_database_up` | 1 when the database answers |
| `atalaia_devices`, `atalaia_devices_online`, `atalaia_devices_known` | Inventory counts |
| `atalaia_scans_total`, `atalaia_scan_failures_total` | Scans and failures since the server started |
| `atalaia_last_scan_timestamp_seconds`, `atalaia_last_scan_duration_seconds`, `atalaia_last_scan_success` | The last scan |
| `atalaia_last_backup_timestamp_seconds` | Last `atalaia backup` that worked |
| `atalaia_database_size_bytes` | The database and its write-ahead log |
| `atalaia_disk_free_bytes`, `atalaia_disk_size_bytes` | The disk that holds the database |

Without the token, `/metrics` answers 401; when off, 404.

## Disk almost full

When the database's disk has less than 512 MB or less than 5% free, the app shows a notice on every page and the health endpoint turns `degraded`.

![Disk almost full notice (fictitious data)](/img/screenshots/app-disco-cheio-en.png)
