---
id: implantacao
title: Implantação
sidebar_position: 3
---

# Implantação

## Docker (recomendado)

A imagem tem cerca de 26 MB, roda sem root, sem privilégios extras e com o sistema de arquivos somente leitura.

```bash
git clone https://github.com/EttoreCSenatore/atalaia && cd atalaia/deploy
docker compose up -d --build
docker compose logs atalaia          # procure o "Setup code"
```

O contêiner usa a **rede do host** (`network_mode: host`): sem isso ele só enxergaria a rede virtual do Docker e não descobriria os dispositivos reais. Os dados ficam no volume `atalaia-data`.

## Escutar na rede local

Por padrão o servidor escuta só em `127.0.0.1:8080`. Para abrir pelo celular ou por outro computador da rede:

```bash
ATALAIA_LISTEN=0.0.0.0:8080 docker compose up -d
```

O servidor continua recusando pedidos que não vêm da rede local, mas o tráfego é HTTP puro. Para uso além de uma rede confiável, coloque na frente um proxy reverso com TLS (por exemplo, Caddy) ou use uma VPN, e informe o nome com `--allow-host`.

## Binário

```bash
cd web && npm ci && npm run build && cd ..
go build -o atalaia ./cmd/atalaia
./atalaia serve --listen 0.0.0.0:8080
```

## Backup e atualização

Todo o estado é um arquivo SQLite (`/data/atalaia.db` no contêiner). Para fazer backup, copie o volume ou o arquivo com o servidor parado. Para atualizar, rode `docker compose up -d --build` de novo: o banco migra sozinho e uma versão mais antiga se recusa a abrir um banco mais novo.

## O que a CI verifica

A CI constrói a imagem, valida o compose, varre vulnerabilidades com o Trivy (falha em críticas com correção) e faz um teste de fumaça: o servidor sobe, o app e a API respondem, a CSP está presente, `/api/devices` exige login e um cabeçalho `Host` falso é recusado.
