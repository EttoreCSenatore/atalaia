---
id: seguranca-passiva
title: Passive security (ARP)
sidebar_position: 4
---

# Passive security: signs of ARP spoofing

In *ARP spoofing* a device on the network pretends to be another one (almost always the router) so the others' traffic passes through it and can be spied on or altered. Atalaia looks for the two signs of this technique that can be seen **without privileges and without sending anything beyond the normal scan**: it only checks what the kernel already learned about the network.

## What is detected

| Event | When it happens | Why it matters |
|---|---|---|
| **Gateway MAC changed** (`gateway_changed`) | The default gateway's address (read from the routing table) starts answering with another MAC. | It is the classic sign of interception. It also happens when you replace the router: then just confirm it. |
| **Answers for many addresses** (`multi_ip`) | One MAC answers for 3 or more IP addresses in the same scan. | Whoever poisons the network answers for several IPs. Routers, Docker and machines with virtual addresses do it too. |

Each event is recorded **once, when it starts**, in the history (highlighted in red), and triggers an alert if the **Signs of ARP spoofing** rule is on (default: on; see [Alerts](alertas.md)). If the device stops and starts again, you are told again. Unlike new-device alerts, these **also fire on the first scan**.

## Known devices

A device marked as **known** is never flagged as "answers for many addresses": it is how you silence a router or a server with virtual addresses that you recognise. The router MAC-change alert has no such exception.

## Limits (read before trusting)

- It is a **heuristic**: there can be false positives (explained above) and false negatives. A careful attacker who answers for a single address, without taking over the router's, **will not be seen**.
- The check runs **at every scan** (default: 5 minutes), not in real time: Atalaia does not listen to ARP packets.
- IPv4 only, Linux only, local network only.

## What is still to come

Detecting a **rogue DHCP server** and **changed services** requires sending special packets (port 68, raw sockets) and extra capabilities in the container; they are left for a next step, as an optional feature. See the [roadmap](../roadmap.md).
