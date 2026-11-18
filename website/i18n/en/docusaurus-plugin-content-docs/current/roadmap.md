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
| **M4.2** | Passive security: rogue DHCP server and changed services (optional, needs extra capabilities) | Planned |
| **M5.1** | Tools: ping, DNS, Wake-on-LAN and port scan with confirmation, limits and an audit log | Done (see [Tools](funcionalidades/ferramentas.md)) |
| **M5.2** | Active security: risky services, exposed web panels and TLS audit | Done (see [Tools](funcionalidades/ferramentas.md)) |
| **M5.3** | CVEs by service version, traceroute and speed test | Planned |
| **M6.1** | Built-in HTTPS (own certificate, or self-signed with renewal) | Done (see [Deployment](tecnico/implantacao.md)) |
| **M6.2** | Security self-review of the code and gosec in CI | Done (see [Security and privacy](tecnico/seguranca-e-privacidade.md)) |
| **M6.3** | systemd unit, .deb/.rpm packages and an observe-only mode | Done (see [Deployment](tecnico/implantacao.md)) |

**v0.1** is the end of M3: discovery, history and alerts. v0.1.0 has been published. The next steps are **M5.3** and **M4.2**.

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
