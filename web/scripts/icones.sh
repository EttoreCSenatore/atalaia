#!/bin/sh
# Gera os ícones PNG do app instalável (PWA) a partir dos SVGs. Precisa de rsvg-convert (librsvg).
# Generates the installable app (PWA) PNG icons from the SVGs. Needs rsvg-convert (librsvg).
set -eu
cd "$(dirname "$0")/.."
rsvg-convert -w 192 -h 192 public/favicon.svg -o public/icons/icon-192.png
rsvg-convert -w 512 -h 512 public/favicon.svg -o public/icons/icon-512.png
rsvg-convert -w 512 -h 512 pwa/icon-maskable.svg -o public/icons/maskable-512.png
rsvg-convert -w 180 -h 180 pwa/icon-maskable.svg -o public/icons/apple-touch-icon.png
