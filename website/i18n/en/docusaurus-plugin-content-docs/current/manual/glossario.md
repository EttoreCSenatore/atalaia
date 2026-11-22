---
id: glossario
title: Glossary
sidebar_position: 4
---

# Glossary

**ARP**: the protocol that maps an IP address to the physical address (MAC) on the local network. Atalaia reads the system's ARP table to know who is on the network.

**ARP spoofing**: a device answering for an IP that is not its own (the router's, for example) to intercept traffic. Atalaia looks for two signs of it.

**Active action**: something Atalaia does **to** a device (ping, port scan, exposure check, Wake-on-LAN, traceroute, speed test, DHCP check). It asks for confirmation, goes to the **audit log** and can be turned off (**observe-only mode**).

**Setup code**: the one-time code the server writes to its log the first time, to create the account.

**Known**: a device you marked as yours. It raises no "new device" alert.

**DHCP**: the network service that hands out IP addresses. A **rogue DHCP** is a second server that can divert traffic.

**Scope**: the networks Atalaia may probe: only the machine's own private IPv4 networks (`atalaia scope`).

**Event**: something that happened to a device: it joined, went online or offline, changed address, or a security sign.

**Exposure**: what a device offers on the network that it should not (risky services, weak TLS, admin pages).

**Vendor (OUI)**: the first 3 bytes of the MAC say who made the network card, from the public IEEE list.

**Gateway**: the router through which the network reaches the internet.

**MAC**: the physical address of a network card (`aa:bb:cc:dd:ee:ff`). **Random MAC**: a MAC the device makes up for privacy; Atalaia flags it when it can tell.

**mDNS, SSDP, NetBIOS**: protocols devices use to announce themselves on the local network; Atalaia uses them to find names, models and services.

**Migration**: the change of the database format in a new version. Atalaia makes a copy first.

**Observe-only mode**: the server started with `--no-active-tools`: it only discovers and alerts, never acts.

**Audit log**: the list of active actions taken, with time and target (in **Tools**, **Audit log**).

**Heartbeat**: a call Atalaia makes to a monitoring service after every scan or backup; if it stops arriving, you are alerted.

**Restore test**: restoring a backup into a scratch copy and checking that it opens and is intact.

**Scan**: a pass over the network to see who is there (every 5 minutes by default). A **port scan** is something else: an active action against one device.
