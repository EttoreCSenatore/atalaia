---
id: app-web
title: App web
sidebar_position: 2
---

# Web app

`atalaia serve` serves a web app that shows the network inventory, the history and the settings. It comes **inside the binary** (and the Docker image): there is nothing to install in the browser and no cloud service. It works in Portuguese and English, with light and dark themes (it follows the system's) and on phones.

```bash
atalaia serve          # open http://127.0.0.1:8080
```

The images below use fictitious data.

## First visit

There is no default password. The server prints a **setup code** in its console; type it and choose a password to create the administrator account.

![Initial setup screen](/img/screenshots/app-configuracao-inicial-en.png)

After that, access is by password. Five wrong passwords block the address for a while.

![Sign-in screen](/img/screenshots/app-login-en.png)

## Dashboard

Shows how many devices there are, how many are online, the new ones from the last 24 hours, and the list with search and filter. **Scan now** asks for a scan right away; the dashboard updates itself when it finishes (live updates).

![Dashboard with the device list](/img/screenshots/app-painel-en.png)

## Device

Each device has a page with its address, vendor, first and last time seen and its own history. Devices with a random MAC are explained on the page.

![A device page](/img/screenshots/app-dispositivo-en.png)

## History

The timeline of everything that changed: new devices, devices that came back, that disappeared and that changed address.

![History timeline](/img/screenshots/app-historico-en.png)

## Settings

How long the history is kept (default 90 days; 0 keeps everything), password change (ends the other sessions), language and the version.

![Settings screen](/img/screenshots/app-configuracoes-en.png)

## What's new

The **What's new** page (linked in the footer and in Settings) shows what changed in each version, with a "Running" mark on the running version. After an update, a notice at the top says which version Atalaia was updated to. See [Versions and what's new](../tecnico/versionamento.md).

![What's new page (fictitious data)](/img/screenshots/app-novidades-en.png)
