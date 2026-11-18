---
id: ferramentas
title: Network tools
sidebar_position: 5
---

# Network tools

**Tools** (and each device's page) hold the actions that **touch the network**. That is why they follow stricter rules than the rest of the app.

![Network tools](/img/screenshots/app-ferramentas-en.png)

## What is there

| Tool | What it does | Limits |
|---|---|---|
| **Ping** | Sends up to 10 ICMP echo requests and shows the response time. | IPv4 unicast only; refuses multicast and broadcast. |
| **DNS lookup** | A, AAAA, CNAME, MX, NS, TXT and PTR records, through the system resolver. | Up to 20 answers, text stripped of control characters. |
| **Wake-on-LAN** | Sends the "magic packet" to wake a device. | Only devices in the inventory whose address is on a local network of the host. |
| **Port scan** | Tries to connect to about 65 common TCP ports of **one device** and lists the open ones. | See below. |

Each scan stores its result and the next one shows **what opened and what closed**: a new open port on a device is a good reason to take a closer look.

## Safety rules for active actions

- Only the signed-in administrator.
- **One at a time**: if another is running, the new one is refused (429).
- Port scan and Wake-on-LAN only act on **devices already in the inventory whose address is in scope** (the host's own private networks). There is no field to type an arbitrary target.
- The port scan requires the **confirmation** "this is my own network and my own device", has a **10 s minimum interval** between scans and a 60 s limit per run.
- The scan only completes the TCP handshake and closes: it **sends no data** and does not try to identify versions or exploit anything.
- **Every action goes to the audit log** (time, action, target and result), visible on the page itself, and it is deleted along with the history, following the retention setting.

## Ping on Docker and Linux

Ping uses **unprivileged** ICMP (a *ping socket*). It works when the user running Atalaia is inside the system's `net.ipv4.ping_group_range`; if not, the app says so. To allow every user:

```bash
sudo sysctl -w net.ipv4.ping_group_range="0 2147483647"
```

With the Docker image (host network, no capabilities) the host's setting applies.

## What is not there yet

Traceroute, speed test (it needs to talk to an outside server, which deserves its own privacy notice), CVE checks, TLS checks and exposed-panel checks. See the [roadmap](../roadmap.md).
