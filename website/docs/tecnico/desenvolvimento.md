---
id: desenvolvimento
title: Desenvolvimento
sidebar_position: 4
---

# Desenvolvimento

## Requisitos

- Go (versão em `go.mod`) para o agente.
- Node 24+ para o site de documentação (a CI e a imagem usam o Node 26, LTS).
- Python 3 para os scripts de verificação da CI.

Sem Go instalado, dá para usar o Docker:

```bash
docker run --rm -v "$PWD":/src -w /src golang:1-alpine go test ./...
```

## Agente

```bash
go run ./cmd/atalaia version   # versão
go run ./cmd/atalaia scope     # redes que o Atalaia pode sondar
go run ./cmd/atalaia scan      # descobre e guarda os dispositivos
go run ./cmd/atalaia serve     # API em http://127.0.0.1:8080
go vet ./... && go test -race ./...
```

## App web

```bash
cd web && npm ci
npm run dev          # http://localhost:5173, usa a API de `atalaia serve` na porta 8080
npm run check && npm test
npm run build        # gera internal/web/dist, que o Go embute
```

Para refazer as imagens da documentação (dados fictícios): `PLAYWRIGHT=/caminho/playwright/index.mjs npm run screenshots -- ../website/static/img/screenshots`.

## Lista de fabricantes (IEEE)

O arquivo `internal/identity/oui.tsv.gz` é gerado a partir das listas públicas do IEEE. Para atualizar:

```bash
python3 -I scripts/atualizar-oui.py
```

## Documentação

```bash
cd website && npm ci && npm start    # servidor local, em português
npm run start -- --locale en         # em inglês
```

Cada página em `website/docs/` precisa da versão em inglês em `website/i18n/en/docusaurus-plugin-content-docs/current/`, com a mesma estrutura. A CI confere isso com `scripts/checar-traducao.py`.

## Verificações de qualidade e segurança

```bash
gofmt -l .                                  # deve ficar vazio
go vet ./... && go test -race ./...
go run honnef.co/go/tools/cmd/staticcheck@v0.8.2 ./...
go run github.com/securego/gosec/v2/cmd/gosec@v2.30.0 -quiet -exclude=G304,G117 ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.10.0 ./...
```

A CI roda as mesmas verificações, mais a das licenças das dependências, o Trivy na imagem e um teste de fumaça do contêiner.

## Pacotes e versões

Os pacotes `.deb`/`.rpm` são montados pelo [nfpm](https://nfpm.goreleaser.com/) a partir de `deploy/nfpm.yaml`. Uma versão sai ao enviar uma tag `vX.Y.Z`: o fluxo `release.yml` compila os binários (amd64, arm64, armv7), monta os pacotes, publica a imagem no GHCR e cria a release com `SHA256SUMS`. Antes, mova as entradas "Unreleased" do `CHANGELOG.md` para a nova versão e acrescente a versão no topo do `releases.yaml`; a release confere que a tag é a primeira versão de lá e usa o texto dela. Passo a passo em [Versões e novidades](versionamento.md).

## Fluxo de trabalho

Uma etapa por pull request: branch, commits, CI verde, revisão do diff e integração com squash. Documentação (pt-BR e inglês) e memória do projeto são atualizadas na mesma etapa. Para contribuições de fora, o [CONTRIBUTING](https://github.com/EttoreCSenatore/atalaia/blob/main/CONTRIBUTING.md) resume as regras, e os modelos de issue (bug, ideia) e de PR trazem a lista de verificação; o `CODEOWNERS` pede a revisão da pessoa mantenedora em todo PR.
