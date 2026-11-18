---
id: intro
title: Visão geral
slug: /intro
sidebar_position: 1
---

# Atalaia

**Atalaia** (Portuguese for "watchtower") is a **free, open-source, self-hosted** network watchtower. You install it on a machine in your network (a Raspberry Pi, a mini-PC, a NAS or your own computer) and open the web app from any browser to see **what is connected**, **what changed** and **what is exposed**.

It is an independent project, inspired by the category of device-discovery tools such as Fing. It is not affiliated with Fing Limited and does not use its trademarks, images or databases.

![Atalaia site home page](/img/screenshots/site-inicio-en.png)

## What it does

| Feature | What for |
|---|---|
| **Device discovery** | Lists everything on the network (ARP, mDNS, SSDP, NetBIOS, DHCP) with vendor, name and type. |
| **History and alerts** | Online/offline timeline and a notification when a new device joins (ntfy, Telegram, e-mail, webhook). |
| **Passive security** | Detects ARP spoofing, duplicate IPs, MAC changes, rogue DHCP servers and new services. |
| **Active security (opt-in)** | Ports and services, known CVEs, exposed panels, TLS and HTTP headers, UPnP, open DNS. |
| **Network tools** | Ping, traceroute, DNS, Wake-on-LAN, speed and internet quality. |

## Why it cannot run in the browser alone

A browser cannot send ARP or ICMP, nor listen for mDNS: it only speaks HTTP. So Atalaia has two parts: a **Go agent** that runs on your network and does the discovery, and a **web app** served by the agent itself. See the [architecture](tecnico/arquitetura.md).

## Principles

- **Defensive**: it is for auditing **your own** network, not attacking other people's. See the [threat model](tecnico/modelo-de-ameacas.md).
- **Secure by default**: login required, LAN-only, scans limited to your own networks, active checks opt-in and logged. See [Security and privacy](tecnico/seguranca-e-privacidade.md).
- **Privacy**: data stays on your machine. Nothing goes to the cloud.
- **Free software**: code under AGPL-3.0-or-later, documentation under CC BY-SA 4.0. See [License](licenca.md).
- **Bilingual**: documentation in Portuguese and English.

## Current state

The foundation (M0) and **device discovery** (M1) are done: `atalaia scan` finds the devices on the network, identifies vendor and name and keeps the history. `atalaia serve` runs scheduled scans, the API with secure login and the **web app** (dashboard, devices, history), also as a Docker image. Alerts come in M3. See [Discovery](funcionalidades/descoberta.md) and the [roadmap](roadmap.md).
