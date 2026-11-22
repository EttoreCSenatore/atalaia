---
id: backup-e-restauracao
title: Backup and restore
sidebar_position: 2
---

# Backup and restore

All of Atalaia's state is **one SQLite file**: devices, history, settings, the account and the **alert channels' secrets**. Treat the copies as you would a password.

## What already happens by itself

- **Every day**, `atalaia serve` copies the database into `backups/` next to it and keeps the newest 7.
- **Every week**, it test-restores the newest one (opens a scratch copy as Atalaia would and runs the integrity check) and records the result.
- **Before upgrading the schema** of the database in a new version, it makes an `atalaia-pre-migration-*.db` copy.
- **Settings > Backups** shows the last backup and the last test; a notice appears when the backup is more than 26 hours late.

Options and paths: [Configuration reference](configuracao.md).

## What you need to do

The automatic copies are **on the same disk** as the database. They protect against mistakes and corruption, not against losing the disk or the machine. **Copy the backups folder somewhere else** now and then (another disk, a NAS, a USB stick in a drawer):

```bash
# package
sudo rsync -a /var/lib/private/atalaia/backups/ /mnt/nas/atalaia-backups/
# Docker (the volume is deploy_atalaia-data with the repository's compose file)
docker run --rm -v deploy_atalaia-data:/data:ro -v /mnt/nas/atalaia-backups:/out alpine cp -a /data/backups/. /out/
```

To be alerted if backups stop, use `--heartbeat-backup` ([Monitoring](../tecnico/monitoramento.md)).

## Check a copy

```bash
atalaia verify-backup /path/atalaia-20261122-030000.db
atalaia verify-backup --dir /path/to/folder      # the newest
```

The test restores into a scratch copy, runs the migrations and the integrity check and counts the devices, without touching the tested file.

## Restore

1. **Stop** the server: `sudo systemctl stop atalaia` or `docker compose stop`.
2. **Keep** the current database under another name (it may be useful): `atalaia.db` → `atalaia.db.before-restore`. Also delete `atalaia.db-wal` and `atalaia.db-shm` if they exist.
3. **Copy** the chosen backup in place of the database, named `atalaia.db`:
   - package: `sudo cp FILE /var/lib/private/atalaia/atalaia.db`;
   - Docker: `docker run --rm -v deploy_atalaia-data:/data -v "$PWD":/in alpine cp /in/FILE /data/atalaia.db && docker run --rm -v deploy_atalaia-data:/data alpine chown 65532:65532 /data/atalaia.db`.
4. **Start** the server. If the backup is from an earlier version, it is migrated (with a copy first).

A backup from a **newer** version than the one installed does not open: upgrade Atalaia first.

## A one-off copy

```bash
atalaia backup /path/new-file.db                # never overwrites
atalaia backup --dir /path/folder --keep 30     # dated, rotated (for your own scheduler)
```

It works while the server runs and writes with mode `0600`.
