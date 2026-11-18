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

| **Exposure check** | Starts from the last port scan: flags risky services, reads what the web ports answer and checks TLS. | Same rules as the port scan; needs a scan first. |

Each scan stores its result and the next one shows **what opened and what closed**: a new open port on a device is a good reason to take a closer look.

## Exposure check

After a port scan, **Check exposure** answers "what is risky on this device?":

- **Risky open services**, in three levels (critical, high, medium): telnet, Docker API, Redis, MongoDB, SMB, RDP, VNC and others that should rarely be reachable on a network. It is a **rule table**, not a vulnerability database.
- **Web interfaces**: a single `GET /` request per web port (redirects are not followed, at most 64 KiB read) to show the status, the `Server` header, the page title and whether it asks for a login. It helps find admin panels you forgot about.
- **TLS** on ports 443, 8443, 465, 993, 995, 636 and 8883: negotiated version, certificate (subject, validity, self-signed), whether it **still accepts TLS 1.0/1.1**, and RSA keys under 2048 bits.

The certificate is not verified (home equipment uses self-signed ones); the report only informs. Ports that start in clear text and then upgrade to TLS (STARTTLS) are not covered.

## Safety rules for active actions

- Only the signed-in administrator.
- **One at a time**: if another is running, the new one is refused (429).
- Port scan and Wake-on-LAN only act on **devices already in the inventory whose address is in scope** (the host's own private networks). There is no field to type an arbitrary target.
- The port scan requires the **confirmation** "this is my own network and my own device", has a **10 s minimum interval** between scans and a 60 s limit per run.
- The port scan only completes the TCP handshake and closes: it **sends no data**. The exposure check sends **one HTTP request and one TLS handshake** per applicable open port, and nothing that tries to exploit or guess passwords.
- **Every action goes to the audit log** (time, action, target and result), visible on the page itself, and it is deleted along with the history, following the retention setting.

## Observe-only mode

To turn off every active action (ping, Wake-on-LAN, port scan, exposure check) use `--no-active-tools`. See [Deployment](../tecnico/implantacao.md).

## Ping on Docker and Linux

Ping uses **unprivileged** ICMP (a *ping socket*). It works when the user running Atalaia is inside the system's `net.ipv4.ping_group_range`; if not, the app says so. To allow every user:

```bash
sudo sysctl -w net.ipv4.ping_group_range="0 2147483647"
```

With the Docker image (host network, no capabilities) the host's setting applies.

## What is not there yet

**CVE checks** (they need a vulnerability database and reliable version detection, which we do not do yet), traceroute and speed test (it needs to talk to an outside server, which deserves its own privacy notice). See the [roadmap](../roadmap.md).
