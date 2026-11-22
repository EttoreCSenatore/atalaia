#!/bin/sh
# Atalaia package: after install or upgrade (.deb: "configure [old-version]", .rpm: 1 = install, 2 = upgrade).
# Never enables or starts the service by itself; on an upgrade, restarts it only if it was already running.
set -e

has_systemd() { command -v systemctl >/dev/null 2>&1 && [ -d /run/systemd/system ]; }

upgrade=no
case "$1" in
  configure) [ -n "$2" ] && upgrade=yes ;;
  2) upgrade=yes ;;
esac

if has_systemd; then
  systemctl daemon-reload || true
  if [ "$upgrade" = yes ]; then
    systemctl try-restart atalaia.service || true
  fi
fi

if [ "$upgrade" = no ]; then
  cat <<'MSG'
Atalaia is installed. It does not start by itself:
  sudo systemctl enable --now atalaia
  journalctl -u atalaia | grep -A1 "Setup code"     # the code for the first login
Options (listen address, HTTPS, backups, monitoring) are in /etc/default/atalaia.
Use Atalaia only on networks you administer or have written authorization for.
MSG
fi
exit 0
