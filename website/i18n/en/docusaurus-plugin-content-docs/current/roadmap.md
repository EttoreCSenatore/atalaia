---
id: roadmap
title: Roadmap
sidebar_position: 3
---

# Roadmap

Each step becomes a pull request, reviewed and merged before the next one. Documentation (pt-BR and English) and project memory are updated at every step.

| Step | Deliverable | State |
|---|---|---|
| **M0** | Foundation: repository, license, bilingual documentation on GitHub Pages, CI, Go skeleton and the scan-scope guard | Done |
| **M1** | Discovery: ARP and mDNS, device list with vendor and name, SQLite, CLI | Done (see [Discovery](funcionalidades/descoberta.md)) |
| **M2.1** | Server: HTTP API, secure login, scheduled scans, live events, retention | Done (see [API](tecnico/api.md)) |
| **M2.2** | Embedded web app: dashboard, device detail, timeline and Docker image | Done (see [Web app](funcionalidades/app-web.md) and [Deployment](tecnico/implantacao.md)) |
| **M3** | Alerts: new device, notification adapters (ntfy, Telegram, webhook, e-mail), names, tags and known devices | Done (see [Alerts](funcionalidades/alertas.md)) |
| **M4.1** | Passive security: signs of ARP spoofing (gateway MAC changed, one MAC with many IPs) | Done (see [Passive security](funcionalidades/seguranca-passiva.md)) |
| **M4.2** | Rogue DHCP server (optional, needs port 68) and a newly-opened-ports alert | Done (see [Passive security](funcionalidades/seguranca-passiva.md)) |
| **M5.1** | Tools: ping, DNS, Wake-on-LAN and port scan with confirmation, limits and an audit log | Done (see [Tools](funcionalidades/ferramentas.md)) |
| **M5.2** | Active security: risky services, exposed web panels and TLS audit | Done (see [Tools](funcionalidades/ferramentas.md)) |
| **M5.3** | Traceroute and speed test (with a privacy notice). CVEs: decided **not** to do for now, with the reason in the docs | Done (see [Tools](funcionalidades/ferramentas.md)) |
| **M6.1** | Built-in HTTPS (own certificate, or self-signed with renewal) | Done (see [Deployment](tecnico/implantacao.md)) |
| **M6.2** | Security self-review of the code and gosec in CI | Done (see [Security and privacy](tecnico/seguranca-e-privacidade.md)) |
| **M6.3** | systemd unit, .deb/.rpm packages and observe-only mode | Done (see [Deployment](tecnico/implantacao.md)) |
| **v0.5** | Identification: mDNS services, SSDP/UPnP, NetBIOS, type guess with icons, names from DHCP (opt-in), VPN via `--allow-cidr`, scheduled scans | Done (see [Discovery](funcionalidades/descoberta.md)) |
| **v0.6** | Security review (login lockout under parallel attempts, global limit) and inventory housekeeping (forget, auto-prune, filter, export, backup) | Done (see [Security and privacy](tecnico/seguranca-e-privacidade.md)) |
| **v0.7** | Operations and maturity (M7.1–M7.9): pinned supply chain, what's new in the app, health and metrics, automatic backups with restore tests, safe upgrades, tested packages, manual and demo | Done (see [Getting started](manual/primeiros-passos.md)) |

The latest published version is **v0.7.1** (fixes from the v0.7 security review), with the **M7** block (M7.10, Spanish, won't be done).

## M7: Operations and maturity

Inspired by Kaladrius (see the [comparison](tecnico/comparacao-kaladrius.md)). One pull request per step, in this order; each has an issue in the GitHub **M7** milestone.

| Step | Delivery | State |
|---|---|---|
| **M7.1** | Supply chain: Dependabot with a 7-day wait and grouped PRs; GitHub actions pinned by SHA and images by digest, checked in CI | Done (see [Supply chain](tecnico/cadeia-de-suprimentos.md)) |
| **M7.2** | Repository files: CONTRIBUTING, CODE_OF_CONDUCT, CODEOWNERS, NOTICE (name, brand, no link to Fing) and issue and PR templates | Done |
| **M7.3** | Versions and what's new: `releases.yaml` as the single source of the **What's new** page in the app, the version in the footer and the release text; CI checks the version | Done (see [Versions and what's new](tecnico/versionamento.md)) |
| **M7.4** | Health and monitoring: `GET /api/health` without login (version and database), optional Healthchecks.io signals (scan, backup), optional `/metrics` for Prometheus and a nearly-full-disk warning | Done (see [Monitoring](tecnico/monitoramento.md)) |
| **M7.5** | Scheduled backup: a daily backup inside `atalaia serve` (Docker, packages and binary) with rotation, a recorded weekly restore test and a warning when the backup is late | Done (see [Deployment](tecnico/implantacao.md)) |
| **M7.6** | Safe upgrades: automatic backup before migrating and a CI test that upgrades databases created by every published release | Done (see [Versions and what's new](tecnico/versionamento.md)) |
| **M7.7** | Package lifecycle: install and remove scripts, an install test of the `.deb` and the `.rpm` in CI | Done (see [Deployment](tecnico/implantacao.md)) |
| **M7.8** | Documentation: task-based manual (first steps, FAQ, glossary), an operations section (backup, monitoring, upgrade), a decision log and a single configuration reference | Done (see [Getting started](manual/primeiros-passos.md)) |
| **M7.9** | Public demo: the web app in demo mode, with fictitious data, on GitHub Pages | Done ([open the demo](https://ettorecsenatore.github.io/atalaia/demo/)) |
| **M7.10** | (Optional) Spanish translation | Won't do: the project stays in Portuguese and English |

## Later

- Optional nmap and nuclei integration for deeper scans.
- Internet exposure check (needs an outside vantage point).
- Wi-Fi security (rogue access point), Linux with a compatible adapter only.
- Several sites in one dashboard.
- Router integrations (OpenWrt, UniFi, MikroTik) and an installable app (PWA).

## Out of scope

- Blocking devices through ARP spoofing, Wi-Fi deauthentication, exploitation and brute-forcing: these are offensive techniques.
- Scanning networks that are not yours.
- Storing data in a project-run cloud.
