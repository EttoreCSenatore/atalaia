---
id: app-web
title: App web
sidebar_position: 2
---

# Web app

:::tip Try it without installing
The [demo](https://ettorecsenatore.github.io/atalaia/demo/) runs the app in your browser with fictitious data: nothing touches a real network or sends anything, and changes last until you reload the page.
:::

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

How long the history is kept (default 90 days; 0 keeps everything), password change (ends the other sessions), backups (last backup, last restore test and **Back up now**; see [Deployment](../tecnico/implantacao.md)), language and the version.

![Backups section (fictitious data)](/img/screenshots/app-backups-en.png)

![Settings screen](/img/screenshots/app-configuracoes-en.png)

## Install as an app

Atalaia can be installed as an app (PWA) on a phone or computer: it gets its own icon, opens in its own window and, when the server is down or you are away from home, it still opens and says "Could not reach the server" instead of a browser error page.

- **Chrome, Edge (Android, Windows, Linux, macOS):** browser menu › **Install app** (or the install icon in the address bar).
- **Safari (iPhone, iPad):** Share › **Add to Home Screen**.

Browsers only allow installing over **HTTPS with a certificate the device trusts** (or on `localhost`). With the self-signed certificate of the [built-in HTTPS](../tecnico/implantacao.md#built-in-https) or over plain HTTP, the app works the same, the install option just does not appear. To get it, use your own certificate (`--tls-cert`/`--tls-key`, for example one from `tailscale cert`) or a reverse proxy with TLS.

Only the app "shell" is kept on the device (HTML, JavaScript, CSS and icons). **No API response is kept** by the installed app: the device list, history and settings still come from the server every time. Each new Atalaia version replaces the kept shell the next time the app opens with the server up.

## What's new

The **What's new** page (linked in the footer and in Settings) shows what changed in each version, with a "Running" mark on the running version. After an update, a notice at the top says which version Atalaia was updated to. See [Versions and what's new](../tecnico/versionamento.md).

![What's new page (fictitious data)](/img/screenshots/app-novidades-en.png)
