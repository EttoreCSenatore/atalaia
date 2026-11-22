---
id: tarefas
title: How to
sidebar_position: 2
---

# How to

Everyday tasks, each in a few steps.

## Find out what an unknown device is

1. On the dashboard, open the device.
2. Look at the **vendor** (from the MAC), the **name** (DNS, mDNS, NetBIOS), the **model** and the announced **services**: often that already tells you ("Brother", `_ipp._tcp` = printer).
3. If the MAC is **random**, it is almost always a modern phone, tablet or laptop; check when it joined in the **history** and who was home.
4. Still no idea? **Scan common TCP ports** shows what it offers on the network (a camera usually has 554, RTSP).
5. If it is nobody's, change the Wi-Fi password: Atalaia shows, it does not block.

## Name it, tag it and mark it known

On the device page: **Name**, **Tags** (comma separated, e.g. `iot, living-room`) and **Known device**. Known devices do not raise "new device" alerts and can raise one when they go offline (an optional rule).

## Be told when something changes

**Settings > Alerts**: pick the rules (new device, known device offline, ARP spoofing signs, rogue DHCP, new ports) and at least one channel. Use **Send test** after setting it up. Details in [Alerts](../funcionalidades/alertas.md).

## See whether a device is exposed

1. On the device page, **Scan common TCP ports** (confirm; it goes to the audit log).
2. Then **Check exposure**: it flags risky services (Telnet, SMB, databases…), reads what the web pages answer and checks the TLS certificates.

Only for devices in the inventory, on your local network. See [Network tools](../funcionalidades/ferramentas.md).

## Open Atalaia from another device on the network

Listen on every address with HTTPS on: `ATALAIA_LISTEN=0.0.0.0:8443` and `ATALAIA_TLS_SELF_SIGNED=1` (Docker: `environment:`; package: `/etc/default/atalaia`). The first time the browser warns about the certificate: compare the fingerprint in Atalaia's log before accepting it. See [Deployment](../tecnico/implantacao.md#built-in-https).

## Open it over a VPN (Tailscale and similar)

Tell it which VPN network to trust: `ATALAIA_ALLOW_CIDR=100.64.0.0/10` (Tailscale). Without that, VPN addresses get 403. Never expose Atalaia straight to the internet.

## Check for an attack on the network

The **history** shows **ARP spoofing** signs in red (the gateway changed its MAC, or one MAC answers for several addresses). The **rogue DHCP** check is optional: turn it on in **Settings > DHCP servers** (it needs UDP port 68). See [Passive security](../funcionalidades/seguranca-passiva.md).

## Take the list to a spreadsheet

On the dashboard, **Download the list** (CSV or JSON), or `atalaia export --format csv`. The CSV neutralizes formulas: names coming from the network never become spreadsheet commands.

## Clean up old devices

On the device page, **Forget this device** (its history goes too). To do it automatically, in **Settings**, **Forget devices offline for more than N days**: it only affects the ones you never touched.

## Make Atalaia observe only

Start it with `--no-active-tools` (or `ATALAIA_NO_ACTIVE_TOOLS=1`): no ping, traceroute, speed test, Wake-on-LAN, port scan, exposure or DHCP check. Discovery and alerts go on.

## Change the password

**Settings > Change password**. The other sessions are signed out. Forgot it? See the [FAQ](faq.md#i-forgot-the-password).
