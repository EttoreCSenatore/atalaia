---
id: configuracao
title: Referência de configuração
sidebar_position: 1
---

# Referência de configuração

Tudo o que se configura no Atalaia, num lugar só. Cada opção de linha de comando existe como `--nome`; as que têm variável de ambiente podem ir em `/etc/default/atalaia` (pacotes) ou em `environment:` (Docker). A CI confere que esta página cita todo comando, opção e variável do código.

## Comandos

| Comando | O quê |
|---|---|
| `atalaia serve` | Roda o app web e as varreduras em segundo plano (o modo normal) |
| `atalaia scan` | Uma varredura agora, pela linha de comando |
| `atalaia devices` | Lista os dispositivos (`--json` para JSON) |
| `atalaia history` | O que aconteceu (entradas, online, offline, endereço novo) |
| `atalaia export` | A lista em CSV (seguro para planilhas) ou JSON |
| `atalaia backup` | Cópia consistente do banco, com o servidor rodando |
| `atalaia verify-backup` | Teste de restauração de uma cópia |
| `atalaia reset-password` | Esqueceu a senha: remove a conta (dispositivos ficam) |
| `atalaia health` | Consulta a saúde de um servidor rodando (sai 0 ou 1) |
| `atalaia scope` | As redes que o Atalaia pode sondar |
| `atalaia version` | A versão |

## `atalaia serve`

| Opção | Variável | Padrão | O quê |
|---|---|---|---|
| `--db` | `ATALAIA_DB` | veja [Arquivos](#arquivos-e-pastas) | Arquivo do banco (vale para todos os comandos) |
| `--listen` | `ATALAIA_LISTEN` | `127.0.0.1:8080` | Endereço e porta; `0.0.0.0:8443` abre para a rede local |
| `--interval` | — | `5m` | Intervalo entre varreduras (0 = só quando pedir) |
| `--allow-host` | — | — | Nome extra pelo qual o app pode ser aberto (repetível) |
| `--allow-cidr` | `ATALAIA_ALLOW_CIDR` | — | Rede extra tratada como local, ex.: `100.64.0.0/10` (Tailscale); repetível, vírgulas na variável |
| `--tls-self-signed` | `ATALAIA_TLS_SELF_SIGNED=1` | desligado | HTTPS com certificado autoassinado feito e renovado pelo Atalaia |
| `--tls-cert`, `--tls-key` | `ATALAIA_TLS_CERT`, `ATALAIA_TLS_KEY` | — | HTTPS com o seu certificado (recarregado quando muda; liga HSTS) |
| `--secure-cookies` | `ATALAIA_SECURE_COOKIES=1` | desligado | Cookie `Secure` mesmo em HTTP (atrás de um proxy com TLS) |
| `--no-active-tools` | `ATALAIA_NO_ACTIVE_TOOLS=1` | desligado | Modo somente observação: sem ping, traceroute, teste de velocidade, Wake-on-LAN, varredura de portas, verificação de exposição e de DHCP |
| `--heartbeat-scan` | `ATALAIA_HEARTBEAT_SCAN` | — | URL avisada depois de cada varredura ([Monitoramento](../tecnico/monitoramento.md)) |
| `--metrics-token-file` | `ATALAIA_METRICS_TOKEN_FILE`, `ATALAIA_METRICS_TOKEN` | desligado | Liga `/metrics` com esse token (32+ caracteres, ao menos 10 diferentes, ex.: `openssl rand -hex 32`); também lê a credencial `metrics-token` do systemd |
| `--backup-dir` | `ATALAIA_BACKUP_DIR` | `backups/` ao lado do banco | Pasta dos backups automáticos; precisa ser privada (não gravável por outros usuários) |
| `--backup-keep` | `ATALAIA_BACKUP_KEEP` | `7` | Quantos backups automáticos guardar (0 desliga) |
| `--backup-interval` | — | `24h` | Intervalo entre backups automáticos (mínimo 1h) |
| `--heartbeat-backup` | `ATALAIA_HEARTBEAT_BACKUP` | — | URL avisada depois de cada backup automático |

Um dispositivo vira offline depois de `--interval` × 3 sem aparecer (no mínimo 10 minutos).

## Outros comandos

| Comando | Opção | Padrão | O quê |
|---|---|---|---|
| `scan` | `--no-names` | desligado | Não procura nomes (DNS reverso, mDNS) |
| `scan` | `--offline-after` | `10m` | Tempo sem ver um dispositivo até marcá-lo offline |
| `devices` | `--json` | desligado | Saída em JSON |
| `history` | `--limit` | `30` | Quantos eventos |
| `export` | `--format` | `csv` | `csv` ou `json` |
| `backup` | `--dir`, `--keep` | —, `7` | Cópia datada numa pasta, guardando as N mais novas (em vez de um arquivo) |
| `backup` | `--heartbeat` | `ATALAIA_HEARTBEAT_BACKUP` | URL avisada depois do backup |
| `verify-backup` | `--dir` | — | Testa a cópia mais nova da pasta (ou passe um arquivo) |
| `reset-password` | `--yes` | — | Sem `--yes` só explica; pare o servidor antes |
| `health` | `--url` | a partir de `ATALAIA_LISTEN` e `ATALAIA_TLS_*` | Servidor a consultar |

## Variáveis de ambiente

| Variável | Equivale a |
|---|---|
| `ATALAIA_DB` | `--db` |
| `ATALAIA_LISTEN` | `--listen` (também lida por `atalaia health`) |
| `ATALAIA_ALLOW_CIDR` | `--allow-cidr` (separadas por vírgula) |
| `ATALAIA_TLS_SELF_SIGNED` | `--tls-self-signed` (valor `1`) |
| `ATALAIA_TLS_CERT`, `ATALAIA_TLS_KEY` | `--tls-cert`, `--tls-key` |
| `ATALAIA_SECURE_COOKIES` | `--secure-cookies` (valor `1`) |
| `ATALAIA_NO_ACTIVE_TOOLS` | `--no-active-tools` (valor `1`) |
| `ATALAIA_HEARTBEAT_SCAN` | `--heartbeat-scan` |
| `ATALAIA_HEARTBEAT_BACKUP` | `--heartbeat-backup` (`serve`) e `--heartbeat` (`backup`) |
| `ATALAIA_METRICS_TOKEN_FILE` | `--metrics-token-file` |
| `ATALAIA_METRICS_TOKEN` | O próprio token (prefira o arquivo) |
| `ATALAIA_BACKUP_DIR`, `ATALAIA_BACKUP_KEEP` | `--backup-dir`, `--backup-keep` |

## Configurações no app

Guardadas no banco e mudadas em **Configurações** (ou pela [API](../tecnico/api.md)):

| Configuração | Padrão |
|---|---|
| Retenção do histórico | 90 dias (0 guarda tudo) |
| Esquecer dispositivos offline há mais de N dias | nunca |
| Varreduras de portas agendadas dos dispositivos conhecidos | desligado |
| Alertas: regras e canais (ntfy, Telegram, webhook, e-mail) | dispositivo novo, ARP, DHCP falso e portas novas ligados; sem canais |
| Verificação de DHCP falso e aprendizado de nomes por DHCP | desligados |
| Idioma | o do navegador (pt-BR ou inglês) |

## Arquivos e pastas

| O quê | Docker | Pacote (.deb/.rpm) | Binário |
|---|---|---|---|
| Banco | `/data/atalaia.db` | `/var/lib/atalaia/atalaia.db` | `$ATALAIA_DB` ou `~/.local/share/atalaia/atalaia.db` |
| Backups automáticos | `/data/backups/` | `/var/lib/atalaia/backups/` | `backups/` ao lado do banco |
| Cópias antes de migrar | ao lado do banco (`atalaia-pre-migration-*.db`) | idem | idem |
| Certificado autoassinado | `/data/tls/` | `/var/lib/atalaia/tls/` | `tls/` ao lado do banco |
| Opções | `environment:` do compose | `/etc/default/atalaia` | linha de comando |

Com o pacote, `/var/lib/atalaia` aponta para `/var/lib/private/atalaia` (é assim que o systemd guarda os dados de um `DynamicUser`).

## Rede

| Porta | Direção | Quando |
|---|---|---|
| TCP 8080 (ou a de `--listen`) | entrada | O app web e a API; só para a rede local |
| UDP (varredura) | saída, só a rede local | A cada varredura |
| UDP 5353, 1900, 137 | saída, só a rede local | mDNS, SSDP e NetBIOS para identificar dispositivos |
| UDP 68 / 67 | entrada e saída | Só com a verificação de DHCP falso / o aprendizado de nomes ligados (precisa de root ou `CAP_NET_BIND_SERVICE`) |
| Internet | saída | Só o que você ligar: alertas, teste de velocidade (speed.cloudflare.com), sinais de vida |
