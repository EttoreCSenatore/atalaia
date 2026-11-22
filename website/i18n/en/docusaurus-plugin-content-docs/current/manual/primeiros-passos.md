---
id: primeiros-passos
title: Getting started
sidebar_position: 1
---

# Getting started

From nothing to your first alert in about 15 minutes. Use Atalaia only on networks you administer or have written authorization for.

## 1. Pick where to run it

Atalaia has to sit **inside the network** it watches and stay on: a Raspberry Pi, a mini PC, a NAS or a home server running Linux. It does not run in the browser or in the cloud (see [why](../intro.md#why-it-cannot-run-in-the-browser-alone)).

## 2. Install it

The simplest way is Docker:

```bash
git clone https://github.com/EttoreCSenatore/atalaia && cd atalaia/deploy
ATALAIA_LISTEN=0.0.0.0:8080 docker compose up -d --build
```

Prefer a `.deb`/`.rpm` package or the binary? See [Deployment](../tecnico/implantacao.md). `ATALAIA_LISTEN=0.0.0.0:8080` opens the app to the local network (the default is this machine only); on a network that is not only yours, turn on HTTPS too (`ATALAIA_TLS_SELF_SIGNED=1`).

## 3. Create the account

The first time, the server writes a one-time **setup code** to its log:

```bash
docker compose logs atalaia | grep -A1 "Setup code"      # package: journalctl -u atalaia | grep -A1 "Setup code"
```

Open `http://MACHINE-ADDRESS:8080`, type the code and choose a password. There is one account, the administrator.

## 4. Look at the first scan

The dashboard shows the devices found, with vendor, name and a type guess. The first scan **never** raises alerts: whatever is on the network now is your baseline. A new scan runs every 5 minutes, or when you click **Scan now**.

## 5. Say what you know

Open each device you recognise, give it a **Name** (say "Living room TV") and tick **Known device**. Whatever stays unticked deserves a second look. Phones with a **random MAC** change it now and then and may show up as new again; see the [FAQ](faq.md#a-phone-keeps-showing-up-as-new).

## 6. Get alerts on your phone

In **Settings > Alerts**, click **Add channel** and pick **ntfy** (the simplest: install the ntfy app and subscribe to a topic with a hard-to-guess name). Click **Send test**. From then on, an unknown device joining the network raises an alert. Telegram, webhook and e-mail work too; see [Alerts](../funcionalidades/alertas.md).

## 7. Done. What next?

- Atalaia already **backs up** its database every day and test-restores it every week; copy the backups folder somewhere else now and then ([Backup and restore](../operacao/backup-e-restauracao.md)).
- Want to know if it stops? See [Monitoring](../tecnico/monitoramento.md).
- More everyday tasks: [How to](tarefas.md).
