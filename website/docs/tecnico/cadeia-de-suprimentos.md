---
id: cadeia-de-suprimentos
title: Cadeia de suprimentos
sidebar_position: 8
---

# Cadeia de suprimentos

O Atalaia roda na sua rede com acesso a ela, então o que entra no binário e na imagem importa tanto quanto o código. Estas regras valem para a CI, a release e a imagem Docker.

## O que está fixado

| O quê | Como | Exemplo |
|---|---|---|
| Ações do GitHub (`uses:`) | SHA do commit, com a versão em comentário | `actions/checkout@3d3c42e… # v7.0.1` |
| Imagens base do Dockerfile (`FROM`) | Digest `sha256` além da tag | `golang:1-alpine@sha256:70bd7dd…` |
| Imagens usadas na CI (Trivy) | Digest `sha256` além da tag | `aquasec/trivy:0.76.2@sha256:8f6bb94…` |
| Ferramentas Go (`go run`) | Versão exata; o Go confere o hash no banco de checksums | `govulncheck@v1.10.0` |
| Dependências Go e npm | `go.sum` e `package-lock.json` versionados | — |

Uma etiqueta como `@v7` ou `:1-alpine` pode passar a apontar para outro código a qualquer momento, inclusive se a conta do autor for invadida. Um SHA ou digest não muda.

## Verificação na CI

`scripts/checar-fixacao.py` roda no job do Go e falha se encontrar uma ação sem SHA (ou sem o comentário de versão), um `FROM` ou `docker run` sem digest, ou um `go run` sem versão exata. Para rodar localmente:

```bash
python3 -I scripts/checar-fixacao.py
```

## Atualizações (Dependabot)

- **Espera de 7 dias** (`cooldown`): o Dependabot só propõe versões publicadas há pelo menos 7 dias. Pacotes comprometidos costumam ser descobertos e retirados nesse prazo. Alertas de segurança do Dependabot não esperam.
- **PRs agrupados**: atualizações menores e de correção vêm num PR só por ecossistema (Go, web, docs, ações); versões maiores vêm separadas, para revisar com calma.
- **Limite de PRs abertos** por ecossistema, para a fila não crescer sem revisão.
- O Dependabot atualiza o SHA ou o digest **e** o comentário de versão juntos.

A imagem do Trivy na CI fica fora do alcance do Dependabot (está num `docker run`); ela é atualizada à mão, com o digest novo.

## Ao atualizar à mão

Para uma ação, pegue o SHA do commit da versão (`gh api repos/DONO/ACAO/commits/vX.Y.Z --jq .sha`) e escreva `DONO/ACAO@SHA # vX.Y.Z`. Para uma imagem, pegue o digest do índice multiplataforma no registro e escreva `imagem:tag@sha256:…`. Depois rode o script acima.

## Na release

A imagem publicada no GHCR sai com proveniência e SBOM (`provenance: true`, `sbom: true`), e os binários e pacotes com `SHA256SUMS`.
