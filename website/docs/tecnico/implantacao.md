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

## Acesso por VPN (Tailscale e similares)

Por padrão o servidor só responde a endereços "locais" (loopback, privados e link-local). Os endereços de VPNs como o **Tailscale** (`100.64.0.0/10`, a faixa CGNAT) **não** entram nessa lista e recebem 403. Para aceitá-los, diga ao Atalaia qual rede é de confiança:

```bash
atalaia serve --listen 0.0.0.0:8443 --tls-self-signed --allow-cidr 100.64.0.0/10
# ou: ATALAIA_ALLOW_CIDR=100.64.0.0/10   (várias, separadas por vírgula)
```

Redes largas demais são recusadas (qualquer coisa menor que /8 em IPv4). Use junto com [HTTPS](#https-embutido) e informe o nome da máquina com `--allow-host`.

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

## Pacotes .deb e .rpm (systemd)

Cada versão publica pacotes `.deb` (Debian, Ubuntu, Raspberry Pi OS) e `.rpm` (Fedora, RHEL e derivados) para `amd64`, `arm64` e `armv7`, listados no `SHA256SUMS`.

```bash
sudo apt install ./atalaia_0.4.0_amd64.deb        # ou: sudo dnf install ./atalaia-0.4.0-1.x86_64.rpm
sudo systemctl enable --now atalaia
journalctl -u atalaia | grep -A1 "Setup code"    # código da primeira configuração
```

O pacote instala o binário em `/usr/bin/atalaia`, a unidade `atalaia.service` e o arquivo de opções `/etc/default/atalaia` (que um upgrade nunca sobrescreve). **Nada é ligado sem você pedir.** A unidade roda como um **usuário descartável sem privilégios** (`DynamicUser`), sem capabilities, com o sistema de arquivos em modo somente leitura (só `/var/lib/atalaia` é gravável), sem acesso a `/home`, com filtro de chamadas de sistema e só sockets IP. As opções (`ATALAIA_LISTEN`, HTTPS, modo somente observação…) ficam em `/etc/default/atalaia`.

## Modo somente observação

Com `--no-active-tools` (ou `ATALAIA_NO_ACTIVE_TOOLS=1`) o servidor **só observa**: ping, Wake-on-LAN, varredura de portas e verificação de exposição ficam desligados (o servidor responde 403 e o app esconde os botões). A descoberta passiva, o histórico, os alertas e a consulta DNS continuam. É o modo indicado quando o Atalaia roda numa rede onde você não quer que ninguém, nem um administrador comprometido, possa disparar ações ativas.

## Binário

```bash
cd web && npm ci && npm run build && cd ..
go build -o atalaia ./cmd/atalaia
./atalaia serve --listen 0.0.0.0:8080
```

## Backup e atualização

Todo o estado é um arquivo SQLite (`/data/atalaia.db` no contêiner, `/var/lib/atalaia/atalaia.db` com o pacote).

**Backups automáticos (ligados por padrão).** Enquanto roda, o `atalaia serve` faz uma cópia do banco por dia na pasta `backups/` ao lado do banco (`/data/backups` no contêiner, `/var/lib/atalaia/backups` com o pacote), guarda as 7 mais novas e, uma vez por semana, **testa a restauração** da mais nova: restaura numa cópia de rascunho, abre como o Atalaia abriria e roda a verificação de integridade do SQLite. Funciona igual no Docker, nos pacotes e no binário. Se a máquina ficou desligada, o backup atrasado sai logo que o servidor volta. O app mostra o último backup e o último teste em **Configurações > Backups** (com o botão **Fazer backup agora**) e um aviso no topo quando o último backup tem mais de 26 horas.

| Opção | Variável | Padrão |
|---|---|---|
| `--backup-dir` | `ATALAIA_BACKUP_DIR` | `backups/` ao lado do banco |
| `--backup-keep` | `ATALAIA_BACKUP_KEEP` | 7 (0 desliga os backups agendados) |
| `--backup-interval` | — | 24h (mínimo 1h) |
| `--heartbeat-backup` | `ATALAIA_HEARTBEAT_BACKUP` | desligado (veja [Monitoramento](monitoramento.md)) |

As cópias ficam **no mesmo disco** do banco: protegem contra erro e corrupção, não contra a perda do disco. Copie a pasta para outro lugar de vez em quando (outro disco, um NAS). Para conferir uma cópia à mão: `atalaia verify-backup ARQUIVO` ou `atalaia verify-backup --dir PASTA` (a mais nova).

Para uma cópia avulsa e consistente **sem parar o servidor** (ou num agendador seu, com `--dir PASTA --keep N`):

```bash
atalaia backup --db /var/lib/atalaia/atalaia.db ~/atalaia-$(date +%F).db
```

A cópia é criada com permissão `0600` e nunca sobrescreve um arquivo existente: ela contém a lista de dispositivos **e os segredos dos canais de alerta**, então guarde-a como guardaria uma senha. Para restaurar, pare o servidor e coloque a cópia no lugar do banco. Para atualizar, instale a versão nova (ou `docker compose up -d --build`): o banco migra sozinho e uma versão mais antiga se recusa a abrir um banco mais novo. **Antes de migrar, o Atalaia copia o banco** para `atalaia-pre-migration-schemaN-DATA.db` ao lado dele (permissão `0600`, as 3 mais novas ficam); se a cópia falhar (disco cheio, por exemplo), ele não migra e não muda nada. Para voltar atrás, pare o servidor, instale a versão anterior e coloque essa cópia no lugar do banco. A CI testa a atualização a partir dos bancos criados por **cada versão publicada** (veja [Versões e novidades](versionamento.md)).

Para levar a lista para uma planilha: `atalaia export --format csv` (ou `json`), ou o link **Baixar a lista** no painel. O CSV neutraliza fórmulas: nomes vindos da rede nunca viram comandos na planilha.

Para saber se o Atalaia está funcionando (endereço de saúde sem login, sinais para o Healthchecks.io depois de cada varredura e backup, métricas para o Prometheus e o `HEALTHCHECK` da imagem Docker), veja [Monitoramento](monitoramento.md).

## O que a CI verifica

A CI constrói a imagem, valida o compose, varre vulnerabilidades com o Trivy (falha em críticas com correção) e faz um teste de fumaça: o servidor sobe, o app e a API respondem, a CSP está presente, `/api/devices` exige login e um cabeçalho `Host` falso é recusado.
