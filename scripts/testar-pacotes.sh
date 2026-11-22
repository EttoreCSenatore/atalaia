#!/bin/sh
# Install test for the .deb and .rpm packages (M7.7), run as root inside a clean Debian or Fedora
# container. Usage: testar-pacotes.sh OLD-PACKAGE NEW-PACKAGE
# Installs OLD, checks files and the systemd unit, upgrades to NEW (a changed /etc/default/atalaia
# must survive), removes it (the data must survive) and, on Debian, purges it (the data still survives).
# The container has no running systemd: the scripts must cope and never start the service.
set -eu
old=$1 new=$2
fail() { echo "FAIL: $*" >&2; exit 1; }
step() { echo "== $*"; }

if command -v apt-get >/dev/null 2>&1; then
  kind=deb
  export DEBIAN_FRONTEND=noninteractive
  apt-get update -qq
  install() { apt-get install -y -qq "$@" >/dev/null; }
  upgrade() { apt-get install -y -qq "$1" >/dev/null; }
  remove() { apt-get remove -y -qq atalaia >/dev/null; }
  installed() { dpkg-query -W -f='${Status}' atalaia 2>/dev/null | grep -q 'install ok installed'; }
elif command -v dnf >/dev/null 2>&1; then
  kind=rpm
  install() { dnf install -y -q "$@" >/dev/null; }
  upgrade() { dnf upgrade -y -q "$1" >/dev/null; }
  remove() { dnf remove -y -q atalaia >/dev/null; }
  installed() { rpm -q atalaia >/dev/null 2>&1; }
else
  fail "neither apt-get nor dnf"
fi

step "$kind: install $(basename "$old") (with systemd, for systemd-analyze)"
install systemd "$old"
installed || fail "not installed"
v1=$(atalaia version) || fail "atalaia version failed"
echo "   $v1"
for f in /usr/bin/atalaia:755 /usr/lib/systemd/system/atalaia.service:644 /etc/default/atalaia:644 /usr/share/doc/atalaia/copyright:644; do
  p=${f%%:*} m=${f##*:}
  [ "$(stat -c %a "$p")" = "$m" ] || fail "$p mode $(stat -c %a "$p"), want $m"
done
[ ! -e /etc/systemd/system/multi-user.target.wants/atalaia.service ] || fail "the package enabled the service"

step "systemd-analyze verify"
out=$(systemd-analyze verify /usr/lib/systemd/system/atalaia.service 2>&1) || fail "systemd-analyze: $out"
# warnings about the unit itself (unknown keys, bad values) fail the test; ones about the container do not
if echo "$out" | grep 'atalaia.service' | grep -vi 'network-online\|/run/systemd'; then fail "unit warnings"; fi

step "upgrade to $(basename "$new"); a changed /etc/default/atalaia is kept"
echo "ATALAIA_LISTEN=0.0.0.0:8443  # test edit" >> /etc/default/atalaia
upgrade "$new"
v2=$(atalaia version)
[ "$v2" != "$v1" ] || fail "not upgraded: still $v2"
echo "   $v2"
grep -q 'test edit' /etc/default/atalaia || fail "/etc/default/atalaia was replaced"
[ ! -e /etc/systemd/system/multi-user.target.wants/atalaia.service ] || fail "the upgrade enabled the service"

step "remove: the binary goes, the data stays"
mkdir -p /var/lib/private/atalaia
echo data > /var/lib/private/atalaia/atalaia.db
remove
installed && fail "still installed"
[ ! -e /usr/bin/atalaia ] || fail "binary left behind"
[ -f /var/lib/private/atalaia/atalaia.db ] || fail "data deleted on remove"

if [ "$kind" = deb ]; then
  step "purge: the data still stays"
  dpkg --purge atalaia >/dev/null
  [ -f /var/lib/private/atalaia/atalaia.db ] || fail "data deleted on purge"
fi
echo "OK: $kind package installs, upgrades and removes cleanly"
