---
id: desenvolvimento
title: Desenvolvimento
sidebar_position: 4
---

# Desenvolvimento

## Requisitos

- Go (versão em `go.mod`) para o agente.
- Node 24+ para o site de documentação.
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
go vet ./... && go test -race ./...
```

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

## Fluxo de trabalho

Uma etapa por pull request: branch, commits, CI verde, revisão do diff e integração com squash. Documentação (pt-BR e inglês) e memória do projeto são atualizadas na mesma etapa.
