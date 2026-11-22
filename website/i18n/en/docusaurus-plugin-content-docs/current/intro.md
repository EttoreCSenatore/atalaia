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
| **Device discovery** | Lists what is on the network (UDP sweep + ARP table, names by mDNS and reverse DNS) with vendor; flags random MACs. |
| **History and alerts** | Online/offline timeline, names and tags, and alerts for new devices and other events (ntfy, Telegram, e-mail, webhook). |
| **Passive security** | Signs of ARP spoofing (router MAC changed, one MAC with many IPs), a rogue DHCP server (opt-in) and newly opened ports. |
| **Active security (opt-in)** | Port scan, risky services, exposed web interfaces and a TLS audit, with confirmation, limits and an audit log. No CVE matching, on purpose. |
| **Network tools** | Ping, traceroute, DNS lookup, Wake-on-LAN and a download speed test. |

## Why it cannot run in the browser alone

A browser cannot send ARP or ICMP, nor listen for mDNS: it only speaks HTTP. So Atalaia has two parts: a **Go agent** that runs on your network and does the discovery, and a **web app** served by the agent itself. See the [architecture](tecnico/arquitetura.md).

## Principles

- **Defensive**: it is for auditing **your own** network, not attacking other people's. See the [threat model](tecnico/modelo-de-ameacas.md).
- **Secure by default**: login required, LAN-only, scans limited to your own networks, active checks opt-in and logged. See [Security and privacy](tecnico/seguranca-e-privacidade.md).
- **Privacy**: data stays on your machine. Nothing goes to the cloud.
- **Free software**: code under AGPL-3.0-or-later, documentation under CC BY-SA 4.0. See [License](licenca.md).
- **Bilingual**: documentation in Portuguese and English.

## Current state

Atalaia is at **v0.6** and already usable: `atalaia serve` discovers and identifies the devices, keeps the history, serves the **web app** with secure login (HTTP or built-in HTTPS), runs defensive checks, makes daily backups and alerts you through ntfy, Telegram, webhook or e-mail. To begin, see [Getting started](manual/primeiros-passos.md), or try the [demo](pathname:///atalaia/demo/) with fictitious data. There are `.deb` and `.rpm` packages, Linux binaries (amd64, arm64, armv7) and a Docker image. It is a new project, maintained by one person and **not yet independently pen-tested**; see the [known limits](funcionalidades/descoberta.md#current-limits) and the [roadmap](roadmap.md).

