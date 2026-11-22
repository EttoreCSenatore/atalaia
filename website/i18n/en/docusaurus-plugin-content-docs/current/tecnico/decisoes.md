---
id: decisoes
title: Decision log
sidebar_position: 10
---

# Decision log

The design choices that shape Atalaia, with the reason for each. A new (or revised) decision goes here in the same PR that implements it.

| # | Decision | Why | Since |
|---|---|---|---|
| 1 | **AGPL-3.0-or-later** licence (code) and CC BY-SA 4.0 (docs) | Whoever offers a modified version as a service must publish the code; the app shows the source link | M0 |
| 2 | **Go, a single binary, embedded SQLite** (no cgo) | Installs on any Linux, even a Raspberry Pi, with no external database or dependencies | M0, M1 |
| 3 | **Linux only** | Discovery uses the kernel's ARP table and Linux unprivileged sockets | M1 |
| 4 | **Rootless discovery**: UDP sweep + ARP table, no raw sockets or `CAP_NET_RAW` | Less privilege, less risk; the ARP table already has what we need | M1 |
| 5 | **Fixed scope**: only the machine's own private IPv4 networks; tools only against inventory devices | Stops Atalaia being used as a scanner against others | M1, M5.1 |
| 6 | **No IPv6** for now | IPv6 neighbour discovery is a different technique, with temporary addresses that change; the gain on a home network does not pay off yet | v0.5 |
| 7 | **No CVE matching** | Guessing software versions over the network is unreliable; a false CVE alert scares more than it helps. We show what can be verified | v0.4 |
| 8 | **Local-network-only server**, `Host` check and CSRF protection | The app watches the local network; opening it to the internet is not a use case. A VPN comes in through `--allow-cidr` | M2.1, v0.5 |
| 9 | **One administrator account**, argon2id, one-time setup code | A home or small office has one person in charge; the code stops whoever comes first from taking the server | M2.1 |
| 10 | **Fenced active actions**: confirmation, one at a time, minimum interval, audit log; **observe-only mode** | Separating watching from acting makes it clear when Atalaia touches another device | M5.1, M6.3 |
| 11 | **HSTS only with your own certificate** | With the self-signed one, HSTS would block accepting the browser's warning and lock you out | M6.1 |
| 12 | **DHCP check and name learning off by default** | They need a privileged port (UDP 67/68); whoever wants them turns them on and grants the permission | M4.2, v0.5 |
| 13 | **Opt-in speed test, confirmed on every use** | It reveals your public IP to Cloudflare | M5.3 |
| 14 | **No telemetry**; alerts go straight from the machine to the chosen service | Your network's data is yours | M3 |
| 15 | **Alert secrets never come back through the API** | A stolen session does not leak tokens and passwords | M3 |
| 16 | Docker with the **host network** | Without it the container only sees Docker's virtual network | M2.2 |
| 17 | Packages with a **throwaway user** (`DynamicUser`) and a hardened unit; they **never turn the service on or delete data** | Less privilege; installing should not change what runs, and the history belongs to the user | M6.3, M7.7 |
| 18 | **Backups inside `atalaia serve`**, not a systemd timer | Works the same with Docker, the packages and the binary | M7.5 |
| 19 | **No automatic updates** | Packages are updated by apt/dnf; revisit if there is a package repository | M7 |
| 20 | **Pinned supply chain**: actions by SHA, images by digest, Dependabot with a 7-day wait | A tag can start pointing at malicious code; compromised packages are usually pulled within days | M7.1 |
| 21 | **`releases.yaml` as the single source** of what's new | Version, the app's what's new and the release text never drift apart | M7.3 |
| 22 | **Health without login and without network data; `/metrics` off and behind a token** | Monitors need a simple endpoint; counts are enough and nothing identifies devices | M7.4 |
| 23 | **Migrations forward only and append-only**, with a backup first | Down migrations need code that almost never runs; a copy first is more reliable | M7.6 |
| 24 | **Bilingual documentation**, Portuguese first | The project was born in Brazil; English reaches everyone else | M0 |
| 25 | **No real network data in the repository**: screenshots and tests use fictitious data | Privacy of whoever tests and of whoever is on the network | M2.2 |
| 26 | **Static demo** on GitHub Pages: API simulated in the browser, no server | Nobody has to install to look; nothing runs or is stored on the publishing side, and the real app does not carry this code | M7.9 |
