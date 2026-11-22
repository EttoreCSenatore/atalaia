---
id: atualizacao
title: Upgrading
sidebar_position: 3
---

# Upgrading

See what changed on the app's **What's new** page (after upgrading) or in the [releases](https://github.com/EttoreCSenatore/atalaia/releases). Only the latest version receives security fixes.

## How to upgrade

| Install | Command |
|---|---|
| Docker (the repository's compose file) | `git pull && docker compose up -d --build` (in the `deploy` folder) |
| GHCR image | change the tag (`ghcr.io/ettorecsenatore/atalaia:X.Y.Z`) and recreate the container |
| `.deb` package | download the new one and `sudo apt install ./atalaia_X.Y.Z_amd64.deb` |
| `.rpm` package | `sudo dnf upgrade ./atalaia-X.Y.Z-1.x86_64.rpm` |
| Binary | stop, replace the file, start |

Check the downloaded files against the release's `SHA256SUMS` (`sha256sum -c SHA256SUMS --ignore-missing`).

On Debian and derivatives, if you edited `/etc/default/atalaia` and the new version also changed that file, `apt` asks which one to keep: answer **N** (keep yours). The new one is left next to it as `/etc/default/atalaia.dpkg-dist`, so you can see the new options (all of them are in the [configuration reference](configuracao.md)). For an unattended upgrade, use `-o Dpkg::Options::=--force-confold` to keep yours without being asked.

## What happens to the database

- If the new version changed the database format, Atalaia **copies the database first** (`atalaia-pre-migration-schemaN-DATE.db`, next to it) and only then migrates. If the copy fails, it does not migrate and changes nothing.
- CI tests the upgrade from databases made by **every published version**.
- With the package, your `/etc/default/atalaia` is kept, and the service is restarted only if it was already running.

## Going back to the previous version

An older version **refuses** to open a database migrated by a newer one. To go back:

1. Stop the server.
2. Install the previous version.
3. Restore the `atalaia-pre-migration-*.db` copy (or a backup from before the upgrade) as in [Backup and restore](backup-e-restauracao.md#restore).

What happened on the network between the upgrade and going back is left out (it is in the new database, which you kept under another name).
