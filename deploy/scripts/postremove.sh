#!/bin/sh
# Atalaia package: after removal. Never deletes the data (the database, its backups and the
# TLS certificate stay in /var/lib/private/atalaia, reached through /var/lib/atalaia), not even on purge:
# it is the user's network history, and deleting it is their call.
set -e

if command -v systemctl >/dev/null 2>&1 && [ -d /run/systemd/system ]; then
  systemctl daemon-reload || true
fi

case "$1" in
  remove | purge | 0)
    if [ -d /var/lib/private/atalaia ] || [ -d /var/lib/atalaia ]; then
      echo "Atalaia's data was kept in /var/lib/private/atalaia (database, backups, certificate). Delete it yourself if you no longer need it."
    fi
    ;;
esac
exit 0
