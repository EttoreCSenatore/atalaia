#!/usr/bin/env python3
"""Confere que tudo o que o CI, a release e a imagem baixam está fixado (M7.1, cadeia de suprimentos).

Regras:
- `uses:` nos workflows aponta para um commit (SHA de 40 caracteres) com a versão em comentário
  (`# v7.0.1`), nunca para uma etiqueta móvel como `@v7`. Ações locais (`./...`) passam.
- Imagens de registro usadas nos workflows (`docker run org/imagem:tag`) e todo `FROM` dos Dockerfiles
  levam o digest (`@sha256:...`). `FROM` que aponta para um estágio anterior do mesmo Dockerfile passa.
- `go run modulo@versao` usa uma versão exata (`vX.Y.Z`), nunca `@latest` nem um ramo.

Falha (código 1) listando cada linha fora da regra. Só biblioteca padrão.
"""

import re
import sys
from pathlib import Path

RAIZ = Path(__file__).resolve().parent.parent

USES = re.compile(r"^\s*-?\s*uses:\s*(\S+)(.*)$")
FIXADO = re.compile(r"^[\w.-]+/[\w./-]+@[0-9a-f]{40}$")
COMENTARIO = re.compile(r"#\s*v\d+(\.\d+)*")
IMAGEM = re.compile(r"(?<![\w./$-])([a-z0-9.-]+(?:/[a-z0-9._-]+)+:[A-Za-z0-9._-]+)(@sha256:[0-9a-f]{64})?")
GO_RUN = re.compile(r"\bgo run\s+(\S+@\S+)")
VERSAO_GO = re.compile(r"@v\d+\.\d+\.\d+$")
FROM = re.compile(r"^\s*FROM\s+(?:--\S+\s+)*(\S+)(?:\s+AS\s+(\S+))?", re.I)


def workflows(caminho: Path) -> list[str]:
    erros = []
    for n, linha in enumerate(caminho.read_text().splitlines(), 1):
        onde = f"{caminho.relative_to(RAIZ)}:{n}"
        m = USES.match(linha)
        if m:
            alvo, resto = m.group(1), m.group(2)
            if alvo.startswith("./"):
                continue
            if not FIXADO.match(alvo):
                erros.append(f"{onde}: ação sem SHA de commit: {alvo}")
            elif not COMENTARIO.search(resto):
                erros.append(f"{onde}: falta a versão em comentário (# vX.Y.Z): {alvo}")
            continue
        if "docker run" in linha:
            for img in IMAGEM.finditer(linha):
                if not img.group(2):
                    erros.append(f"{onde}: imagem sem digest: {img.group(1)}")
        for mod in GO_RUN.findall(linha):
            if not VERSAO_GO.search(mod):
                erros.append(f"{onde}: go run sem versão exata: {mod}")
    return erros


def dockerfile(caminho: Path) -> list[str]:
    erros, estagios = [], set()
    for n, linha in enumerate(caminho.read_text().splitlines(), 1):
        m = FROM.match(linha)
        if not m:
            continue
        imagem, nome = m.group(1), m.group(2)
        if imagem.lower() not in estagios and "@sha256:" not in imagem:
            erros.append(f"{caminho.relative_to(RAIZ)}:{n}: FROM sem digest: {imagem}")
        if nome:
            estagios.add(nome.lower())
    return erros


def main() -> int:
    erros = []
    for wf in sorted((RAIZ / ".github" / "workflows").glob("*.y*ml")):
        erros += workflows(wf)
    for df in sorted(RAIZ.glob("**/Dockerfile*")):
        if "node_modules" not in df.parts:
            erros += dockerfile(df)
    for e in erros:
        print(e)
    if erros:
        print(f"\n{len(erros)} referência(s) sem fixação. Veja docs/tecnico/cadeia-de-suprimentos.")
        return 1
    print("Tudo fixado: ações por SHA, imagens por digest, go run por versão.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
