---
id: descoberta
title: Descoberta de dispositivos
sidebar_position: 1
---

# Device discovery

Since M1, Atalaia discovers the devices on the local network, identifies the vendor and name of each one and keeps the history in a local SQLite file. For now it is on the command line; the web app comes in M2.

## How it works

1. Atalaia works out the **networks it may probe**: only the private IPv4 networks of the machine's own interfaces (see `atalaia scope`). Nothing outside them is probed.
2. It sends **one UDP byte** to every address of the network (only the /24 the machine is in, even if the network is larger). That makes the system resolve each address with ARP.
3. It reads the **system neighbor table** (`/proc/net/arp`) and checks each address against the scope again.
4. It looks up each device's **name** (mDNS and reverse DNS) and **vendor** from the MAC (the public IEEE list, MA-L, MA-M and MA-S, embedded in the program).

It needs no root and no raw sockets. For now it runs on **Linux** only and with **IPv4** only.

## Scan

```bash
atalaia scan
```

![Output of atalaia scan with fictitious data](/img/screenshots/cli-descoberta-scan.png)

The data in the image is fictitious. `(random MAC)` marks a randomized MAC address: modern phones and laptops use a different address on each network, for privacy. Such addresses have no vendor.

Options: `--db FILE` (default `~/.local/share/atalaia/atalaia.db` or `$ATALAIA_DB`), `--no-names` (skip name lookups) and `--offline-after 10m` (how long without seeing a device before it is marked offline; this stops a sleeping phone from "blinking").

## List and history

```bash
atalaia devices            # table (or --json)
atalaia history --limit 5
```

![Output of atalaia history with fictitious data](/img/screenshots/cli-descoberta-historico.png)

Events: `joined` (seen for the first time), `online` (came back), `offline` (gone past the grace period) and `ip_changed` (same device, new address).

## Security and privacy

- The database holds the network inventory, which is personal data. The file is created with `0600` permissions and its folder with `0700`, and nothing leaves the machine.
- Names and vendors come from the network and are **untrusted data**: control and text-direction characters are stripped before they are stored or shown.
- Inside Docker, use `--network host`; without it the container only sees its own virtual network.

## Current limits

- Linux and IPv4 only. macOS and Windows come later.
- There is no discovery by SSDP/UPnP, NetBIOS or DHCP leases: devices with no mDNS name or reverse DNS show only a vendor (or, with a random MAC, nothing).
- VPN addresses in CGNAT (for example Tailscale, `100.64.0.0/10`) are not accepted as "local network" by the server.
- Devices that do not answer ARP (rare) do not show up; sleeping ones may be missing from one sweep, hence the offline grace period.
- The computer running Atalaia is in the list (the ARP table never lists the machine itself, so it is added separately, using only the main network interface, the one that contains the router).
