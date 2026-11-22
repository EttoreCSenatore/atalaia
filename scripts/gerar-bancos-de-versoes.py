#!/usr/bin/env python3
"""Gera os bancos de teste de atualização (M7.6): um banco criado pelo binário de cada versão publicada.

Para cada versão pedida (ex.: v0.1.0 v0.2.0), o script:
1. baixa `atalaia_<versão>_linux_amd64.tar.gz` e o `SHA256SUMS` da release no GitHub e confere o hash;
2. cria uma rede Docker isolada e descartável (10.123.45.0/28) com um contêiner "dispositivo" fictício;
3. roda o binário daquela versão num contêiner nessa rede: `atalaia scan` duas vezes (o dispositivo some entre
   as duas para gerar eventos de offline);
4. grava o banco em internal/store/testdata/released/<versão>.db e apaga a rede e os contêineres.

Nada toca a rede de verdade: a varredura só vê a rede Docker descartável, e os MACs e IPs são os que o
Docker inventa. O teste `TestUpgradeFromReleasedDatabases` abre cada banco com o código atual.

Uso (precisa de Docker e de acesso à internet para baixar a release):
    python3 -I scripts/gerar-bancos-de-versoes.py v0.6.0
"""

import hashlib
import os
import shutil
import subprocess
import sys
import tarfile
import tempfile
import urllib.request
from pathlib import Path

REPO = "EttoreCSenatore/atalaia"
RAIZ = Path(__file__).resolve().parent.parent
DESTINO = RAIZ / "internal" / "store" / "testdata" / "released"
REDE = "atalaia-fixtures"
SUBREDE = "10.123.45.0/28"
# busybox estático, só para manter o "dispositivo" fictício de pé e rodar o binário
IMAGEM = "busybox:1.37.0-musl@sha256:f6a95d611e0a8780fabdb2aec43e31275b2720dfefa3e41b16835bfc42454478"


def docker(*args: str, check: bool = True) -> subprocess.CompletedProcess:
    return subprocess.run(["docker", *args], check=check, capture_output=True, text=True)


def baixar(url: str, destino: Path) -> None:
    with urllib.request.urlopen(url, timeout=60) as r, open(destino, "wb") as f:  # noqa: S310 (URL fixa do GitHub)
        shutil.copyfileobj(r, f)


def binario(versao: str, pasta: Path) -> Path:
    nome = f"atalaia_{versao}_linux_amd64.tar.gz"
    base = f"https://github.com/{REPO}/releases/download/{versao}/"
    baixar(base + nome, pasta / nome)
    baixar(base + "SHA256SUMS", pasta / "SHA256SUMS")
    esperado = next(l.split()[0] for l in (pasta / "SHA256SUMS").read_text().splitlines() if l.endswith(nome))
    obtido = hashlib.sha256((pasta / nome).read_bytes()).hexdigest()
    if obtido != esperado:
        sys.exit(f"{nome}: SHA-256 não confere ({obtido} != {esperado})")
    with tarfile.open(pasta / nome) as t:
        membro = next(m for m in t.getmembers() if m.isfile() and Path(m.name).name == "atalaia")
        membro.name = "atalaia"
        t.extract(membro, pasta, filter="data")
    return pasta / "atalaia"


def gerar(versao: str) -> None:
    with tempfile.TemporaryDirectory() as tmp:
        pasta = Path(tmp)
        binario(versao, pasta)
        dispositivo = f"impressora-teste-{versao.replace('.', '-')}"
        docker("network", "rm", REDE, check=False)
        docker("network", "create", "--internal", "--subnet", SUBREDE, REDE)
        try:
            docker("run", "-d", "--rm", "--name", dispositivo, "--network", REDE, "--ip", "10.123.45.5", IMAGEM, "sleep", "300")
            scan = ["run", "--rm", "-u", f"{os.getuid()}:{os.getgid()}", "--network", REDE, "--ip", "10.123.45.2",
                    "-v", f"{pasta}:/w", "-w", "/w",
                    "-e", "ATALAIA_DB=/w/atalaia.db", IMAGEM, "/w/atalaia", "scan"]
            r = docker(*scan, check=False)
            if r.returncode != 0:
                sys.exit(f"{versao}: scan falhou:\n{r.stdout}{r.stderr}")
            docker("rm", "-f", dispositivo, check=False)
            # segunda varredura sem o dispositivo: --offline-after curto gera o evento de offline
            r = docker(*scan, "--offline-after", "1s", check=False)
            if r.returncode != 0:
                r = docker(*scan, check=False)  # versões sem --offline-after
            if r.returncode != 0:
                sys.exit(f"{versao}: segunda varredura falhou:\n{r.stdout}{r.stderr}")
        finally:
            docker("rm", "-f", dispositivo, check=False)
            docker("network", "rm", REDE, check=False)
        DESTINO.mkdir(parents=True, exist_ok=True)
        for sufixo in ("-wal", "-shm"):
            if (pasta / f"atalaia.db{sufixo}").exists():
                sys.exit(f"{versao}: o banco ficou com {sufixo}; a varredura não fechou o banco")
        shutil.copyfile(pasta / "atalaia.db", DESTINO / f"{versao}.db")
        print(f"{versao}: {DESTINO / f'{versao}.db'}")


def main() -> int:
    versoes = sys.argv[1:]
    if not versoes or any(not v.startswith("v") for v in versoes):
        print(__doc__)
        return 2
    for v in versoes:
        gerar(v)
    return 0


if __name__ == "__main__":
    sys.exit(main())
