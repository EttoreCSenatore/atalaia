---
id: alertas
title: Alerts and known devices
sidebar_position: 3
---

# Alerts and known devices

Atalaia can tell you when something happens on your network. The message goes **straight from this machine** to the service you choose: there is no server of ours in between.

## Names, tags and "known"

On each device's page you can give it a **name** (it wins over the name discovered on the network and scans never overwrite it), up to 8 **tags** (`iot`, `guest`, `phone`…) and mark the device as **known**. The unknown ones are the ones worth a second look. Names and tags are searchable on the dashboard.

## When to alert

| Rule | Default | What it does |
|---|---|---|
| New device | on | A device that was never seen before joins the network. |
| Known device went offline | off | A device marked as known disappeared (after the grace period). |
| Signs of ARP spoofing | on | The router's MAC changed, or a device answers for many addresses. See [Passive security](seguranca-passiva.md). |

The **first scan** of an empty database only builds the inventory and **never** triggers new-device alerts; otherwise you would get one message for every gadget in the house. When one scan finds several devices at once, **a single message** is sent with the list (at most 10 lines).

## Channels

| Channel | What to fill in |
|---|---|
| **ntfy** | Topic address (`https://ntfy.sh/your-secret-topic` or your own server) and, optionally, an access token. |
| **Telegram** | Bot token (created with @BotFather) and the chat ID. |
| **Webhook** | An `http://` or `https://` address; it receives the JSON `{"title","message","source":"atalaia"}` (works for Home Assistant, n8n, etc.). |
| **E-mail** | Server `host:port`, user and password (optional), sender and recipient. |

Every channel has a **Send test** button. Channels live in **Settings → Alerts**, up to 8 of them.

![Alerts settings](/img/screenshots/app-alertas-en.png)

## Security

- Tokens and passwords stay **only in this machine's SQLite database** (a `0600` file) and are **never returned** by the API: the screen only shows that a secret is saved, and leaving the field empty keeps the existing one.
- E-mail uses STARTTLS when the server offers it and **refuses to send the password without TLS**.
- The text comes from the network (device names), so it is sent as **plain text**, never as markup, and line breaks are removed from titles (header injection).
- Error messages never repeat the channel's address, which may contain a token.
- ntfy and webhook addresses may point to the local network (that is the Home Assistant case). Only the signed-in administrator configures channels.
- A failure in one channel is logged and does not affect the others or the scan.

See the endpoints in [API and server](../tecnico/api.md).
