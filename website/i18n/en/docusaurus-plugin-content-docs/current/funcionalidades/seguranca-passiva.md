---
id: seguranca-passiva
title: Passive security
sidebar_position: 4
---

# Passive security

## Signs of ARP spoofing

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

## Rogue DHCP server

Whoever hands out the network's addresses (DHCP) also says what the gateway and DNS are. A **second DHCP server** you did not set up (a spare router in router mode, a malicious device) can send devices to a wrong gateway or DNS and redirect their traffic.

In **Settings → DHCP servers** Atalaia sends **one DHCP discovery broadcast**, listens for the offers and remembers which servers answered, with the gateway and DNS each one would hand out. It **never sends the request** that completes a lease, so it takes no address. The **first check** only records which servers are yours; after that, a new server produces the **New DHCP server answered** event (red in the history, if its address is a device in the inventory) and an alert (rule on by default). Replaced your router? Use **Forget known servers** and the next check starts again.

- It is **optional and starts off**: you can just **Check now** (with confirmation) or turn the check on **after every scan**.
- It needs to **listen on UDP port 68, a privileged port**. Run as root, or give it the `CAP_NET_BIND_SERVICE` capability: with systemd, `sudo systemctl edit atalaia` with `[Service]` / `AmbientCapabilities=CAP_NET_BIND_SERVICE` / `CapabilityBoundingSet=CAP_NET_BIND_SERVICE`; with Docker, `cap_add: [NET_BIND_SERVICE]` in the compose file. Without it the app explains the error.
- It coexists with the machine's own DHCP client (the port is shared) and is off in [observe-only mode](../tecnico/implantacao.md).
- Limits: a rogue server that **does not answer** our discovery (or only answers known clients) is not seen, and replies only come from servers on the network where Atalaia runs.

## Changed services

Each port scan (see [Tools](ferramentas.md)) is compared with the previous one of the same device. If **a new port opened**, the device page highlights it and an alert is sent (the rule **A port scan finds newly opened ports**, on by default). A port scan is an active action, so it **does not run by itself**: the comparison happens when you scan.

## What is still to come

Scheduled (optional) port scans and service detection by version change are left for the future. See the [roadmap](../roadmap.md).
