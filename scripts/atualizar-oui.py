#!/usr/bin/env python3
"""Gera `internal/identity/oui.tsv.gz` a partir das listas públicas de OUI do IEEE (MA-L, MA-M e MA-S).

Uso:
    python3 -I scripts/atualizar-oui.py                    # baixa de https://standards-oui.ieee.org/
    python3 -I scripts/atualizar-oui.py --de-pasta PASTA   # usa oui.csv, mam.csv e oui36.csv já baixados

O arquivo gerado tem uma linha por prefixo: `PREFIXO<TAB>FABRICANTE` (prefixo em hexadecimal maiúsculo, 6, 7 ou 9 dígitos).
Os nomes vêm de terceiros e são tratados como dado não confiável: caracteres de controle são removidos, espaços
colapsados e o tamanho limitado. Só biblioteca padrão.
"""

import argparse
import csv
import gzip
import re
import sys
import urllib.request
from pathlib import Path

BASE = "https://standards-oui.ieee.org/"
ARQUIVOS = {"oui.csv": "oui/oui.csv", "mam.csv": "oui28/mam.csv", "oui36.csv": "oui36/oui36.csv"}
DESTINO = Path(__file__).resolve().parent.parent / "internal" / "identity" / "oui.tsv.gz"
TAMANHO_MAXIMO = 80
HEX = re.compile(r"^[0-9A-F]{6}$|^[0-9A-F]{7}$|^[0-9A-F]{9}$")


def limpar(nome: str) -> str:
    nome = re.sub(r"[\x00-\x1f\x7f-\x9f  ]", " ", nome)
    return re.sub(r"\s+", " ", nome).strip()[:TAMANHO_MAXIMO]


def ler(texto: str) -> dict[str, str]:
    saida: dict[str, str] = {}
    for linha in csv.DictReader(texto.splitlines()):
        prefixo = (linha.get("Assignment") or "").strip().upper()
        nome = limpar(linha.get("Organization Name") or "")
        if HEX.match(prefixo) and nome:
            saida[prefixo] = nome
    return saida


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--de-pasta", type=Path)
    args = ap.parse_args()
    todos: dict[str, str] = {}
    for arquivo, caminho in ARQUIVOS.items():
        if args.de_pasta:
            texto = (args.de_pasta / arquivo).read_text(encoding="utf-8", errors="replace")
        else:
            req = urllib.request.Request(BASE + caminho, headers={"User-Agent": "atalaia-oui-update"})
            with urllib.request.urlopen(req, timeout=60) as r:  # noqa: S310 (URL fixa, https)
                texto = r.read().decode("utf-8", errors="replace")
        todos.update(ler(texto))
    if len(todos) < 30000:
        print(f"erro: só {len(todos)} prefixos; a lista parece incompleta", file=sys.stderr)
        return 1
    corpo = "".join(f"{p}\t{n}\n" for p, n in sorted(todos.items())).encode("utf-8")
    with gzip.GzipFile(DESTINO, "wb", mtime=0) as f:  # mtime 0: saída reprodutível
        f.write(corpo)
    print(f"ok: {len(todos)} prefixos em {DESTINO} ({DESTINO.stat().st_size // 1024} KiB)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
