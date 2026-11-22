---
id: faq
title: FAQ
sidebar_position: 3
---

# Frequently asked questions

## I forgot the password

Whoever can reach the database file already controls Atalaia, so recovery is from the command line, on its machine:

1. Stop the server (`sudo systemctl stop atalaia` or `docker compose stop`).
2. Run `atalaia reset-password --yes` on the same database: package `sudo atalaia reset-password --db /var/lib/private/atalaia/atalaia.db --yes`; Docker `docker compose run --rm atalaia reset-password --yes`.
3. Start the server and take the new **setup code** from its log to create the account again.

Devices, history, alerts and settings stay.

## A phone keeps showing up as new

Modern phones and laptops use a **random MAC** per network and sometimes change it. To Atalaia, a new MAC is a new device. Options: turn off the random MAC for your network on the device (Wi-Fi settings > "private address"), or let Atalaia **forget** untouched devices offline for N days (Settings).

## Why does it not see all my devices?

- It only sees the **same network** (same IPv4 subnet) as the machine it runs on; guest networks and separate VLANs are left out.
- Sleeping devices may not answer a scan; they show up in the next one.
- With Docker, the container needs the **host network** (`network_mode: host`), which the compose file already uses.
- **IPv6** is not covered (see the [decision log](../tecnico/decisoes.md)).

## Is it safe to leave running?

It runs without root, answers only the local network, needs a login, sends nothing to the cloud, and active actions (port scans and so on) only happen when you ask, with confirmation and a log. It has **not had an independent penetration test** yet; see [Security and privacy](../tecnico/seguranca-e-privacidade.md) and the [threat model](../tecnico/modelo-de-ameacas.md).

## Does it send any data out?

No, except what you turn on: alerts go straight from the machine to the service you chose, the speed test talks to speed.cloudflare.com and heartbeats go to the URL you set. There is no telemetry.

## Can I use it on a network that is not mine?

Only with written authorization from whoever runs it. Even a defensive scan of someone else's network can be illegal.

## Why does it not look for vulnerabilities (CVEs)?

Telling the exact software version of each device over the network is unreliable, and a wrong CVE alert scares more than it helps. Atalaia shows what can be verified: risky exposed ports and services, weak TLS, open web interfaces. See [Tools](../funcionalidades/ferramentas.md#cve-checks-why-there-are-none).

## Is an "ARP spoofing" alert an attack?

It can be, but it also happens when you replace the router, or with Wi-Fi repeaters and bridges that answer for several devices. If you changed nothing on the network, look into it; marking devices as known cuts false positives.

## How much space does it use?

Little: a typical home's database is a few megabytes, and the 7 automatic backups add about the same. The app warns when the database's disk is almost full.

## How do I upgrade?

See [Upgrading](../operacao/atualizacao.md). The database is copied before its format changes, and you can go back.

## Does it work on Windows or macOS?

Discovery uses Linux features (the kernel's ARP table, unprivileged sockets). Use Linux: a Raspberry Pi is enough.
