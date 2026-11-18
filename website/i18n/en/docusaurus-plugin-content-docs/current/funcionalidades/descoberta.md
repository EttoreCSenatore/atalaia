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
4. It looks up each device's **name** (mDNS, reverse DNS, NetBIOS) and **vendor** from the MAC (the public IEEE list, MA-L, MA-M and MA-S, embedded in the program).
5. It asks each one what it **says it is** (see below) and guesses the type.

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

## What Atalaia learns about each device

Besides the vendor (from the MAC), a device can **say what it is**. Atalaia asks, on your local network and only the addresses it already found:

| Source | What it gives | How |
|---|---|---|
| **mDNS (service types)** | The services the device advertises (`_airplay`, `_googlecast`, `_ipp` for printers, `_hap` for HomeKit, Apple's `_companion-link`…) and the **model** (`Chromecast Ultra`, `iPhone14,5`, the printer model) | One direct question to each device's port 5353 |
| **SSDP/UPnP** | Manufacturer, model, friendly name and type (`MediaRenderer`, router…) of TVs, consoles, NAS boxes and routers | One multicast search; the description is only downloaded **from the very address that answered** |
| **NetBIOS** | The computer name and workgroup of Windows PCs, Samba and many NAS boxes | One direct query to UDP port 137 |

With that Atalaia **guesses the type** of the device (router, printer, TV, speaker, camera, NAS, computer, phone, smart-home…) and shows an icon on the dashboard. It is a **heuristic**: strong signals (the router, a printing service, a UPnP type) count more than weak ones (a name, a vendor), and with no signal at all it **does not guess**. A random MAC with no announcement at all shows as "phone, tablet or laptop", which is what it almost always is.

What a device announces **never erases** what you wrote (the name and tags are yours), and a scan where it stayed silent does not erase what was already known.

**Honest limit**: it only finds out about devices that talk. In a test on a real home network, most devices were phones with a random MAC that **announce nothing** (no mDNS, no UPnP, no NetBIOS): for those the best option is to give a name and tags by hand.

Everything that comes from the network is treated as **untrusted**: text is stripped of control characters and truncated, UPnP XML is read without expanding entities, and a device cannot make Atalaia download anything from another address.

### Names from DHCP (optional)

Phones and laptops **announce their name** (and, on Android, a "vendor class" such as `android-dhcp-14`) in the DHCP request they broadcast when they join the network or renew their address, even with a random MAC. In **Settings → DHCP servers** you can turn on **Learn device names from their DHCP requests**: Atalaia only **listens** (it sends and answers nothing) and uses what it hears to give a name and type to devices that have none.

- **Off by default**: it needs to listen on UDP port 67, a privileged port (root or `CAP_NET_BIND_SERVICE`, like the [rogue DHCP check](seguranca-passiva.md)). Without the capability the app explains the error.
- It only learns **when the device renews or reconnects**, which can take hours; names show up over time.
- It never replaces a name that you or another source already set, and a learned name does not erase yours.
- It is not an active tool: it stays available in [observe-only mode](../tecnico/implantacao.md).

## Security and privacy

- The database holds the network inventory, which is personal data. The file is created with `0600` permissions and its folder with `0700`, and nothing leaves the machine.
- Names and vendors come from the network and are **untrusted data**: control and text-direction characters are stripped before they are stored or shown.
- Inside Docker, use `--network host`; without it the container only sees its own virtual network.

## Current limits

- Linux and IPv4 only. macOS and Windows come later. **IPv6**: we decided not to cover it for now (it would mean reading the neighbor table over netlink and probing the ND multicast group; almost every device on a home network also has an IPv4 address, which is already discovered).
- Devices that announce nothing (many phones with a random MAC) show no name or model.
- VPN addresses in CGNAT (for example Tailscale, `100.64.0.0/10`) are not accepted as "local network" by the server.
- Devices that do not answer ARP (rare) do not show up; sleeping ones may be missing from one sweep, hence the offline grace period.
- The computer running Atalaia is in the list (the ARP table never lists the machine itself, so it is added separately, using only the main network interface, the one that contains the router).
