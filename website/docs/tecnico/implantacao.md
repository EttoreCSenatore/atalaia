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

O servidor continua recusando pedidos que não vêm da rede local, mas o tráfego é HTTP puro. Ligue o [HTTPS embutido](#https-embutido) (`--tls-self-signed`), coloque na frente um proxy reverso com TLS (por exemplo, Caddy) ou use uma VPN, e informe o nome com `--allow-host`.

## Versões prontas

Cada versão (a partir da v0.1.0) publica, na página de [Releases](https://github.com/EttoreCSenatore/atalaia/releases):

- binários Linux (`amd64`, `arm64`, `armv7`) com o app web dentro, e um arquivo `SHA256SUMS` (confira com `sha256sum -c SHA256SUMS`);
- a imagem `ghcr.io/ettorecsenatore/atalaia:0.1.0` (e `latest`), para `amd64` e `arm64`, com SBOM e proveniência.

```bash
docker run -d --name atalaia --network host --read-only --cap-drop ALL \
  -v atalaia-data:/data ghcr.io/ettorecsenatore/atalaia:latest
docker logs atalaia                  # procure o "Setup code"
```

## HTTPS embutido

O Atalaia pode servir HTTPS sozinho, sem proxy (TLS 1.2 ou superior; o cookie de sessão passa a ser `Secure`):

| Opção | Quando usar |
|---|---|
| `--tls-self-signed` (ou `ATALAIA_TLS_SELF_SIGNED=1`) | Rede doméstica. O Atalaia cria em `<pasta do banco>/tls/` um certificado autoassinado (ECDSA P-256, 1 ano, chave `0600`) que cobre `localhost`, o nome da máquina, os endereços locais e os `--allow-host`, e o **renova sozinho** 30 dias antes de vencer. O navegador avisará uma vez: confira a **impressão digital SHA-256** impressa no console antes de aceitar. |
| `--tls-cert arquivo.pem --tls-key chave.pem` (ou `ATALAIA_TLS_CERT` / `ATALAIA_TLS_KEY`) | Você já tem um certificado (por exemplo, de uma CA interna ou do Let's Encrypt via DNS). Troca de arquivo é percebida sem reiniciar. Com esta opção o servidor também envia `Strict-Transport-Security`. |

```bash
atalaia serve --listen 0.0.0.0:8443 --tls-self-signed
```

O HSTS **não** é enviado com o certificado autoassinado: ele impediria o navegador de deixar você aceitar o aviso. Atrás de um proxy que termina o TLS, o Atalaia enxerga HTTP puro; nesse caso prefira o HTTPS embutido, ou garanta que o proxy fale com ele só pela máquina local.

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
