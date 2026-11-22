#!/bin/sh
# Atalaia package: before removal (.deb: "remove" or "upgrade", .rpm: 0 = removal, 1 = upgrade).
# On a real removal, stops and disables the service; on an upgrade, leaves it alone.
set -e

case "$1" in
  remove | 0)
    if command -v systemctl >/dev/null 2>&1 && [ -d /run/systemd/system ]; then
      systemctl disable --now atalaia.service || true
    fi
    ;;
esac
exit 0
