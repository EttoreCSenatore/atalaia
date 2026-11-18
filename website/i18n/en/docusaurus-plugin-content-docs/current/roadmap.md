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
| **M3** | Alerts: new device, notification adapters, names, tags and known devices | Planned |
| **M4** | Passive security: ARP spoofing, rogue DHCP, changed services | Planned |
| **M5** | Active security and tools: ports, CVEs, exposed panels, TLS, ping, traceroute, DNS, Wake-on-LAN, speed | Planned |
| **M6** | Hardening, self-review, v0.2 release and packages (Linux, macOS, Windows, ARM) | Planned |

**v0.1** is the end of M3: discovery, history and alerts. The next step is **M3 (alerts)**.

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
