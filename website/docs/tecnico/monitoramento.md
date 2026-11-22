---
id: monitoramento
title: Monitoramento
sidebar_position: 4
---

# Monitoramento

Como saber se o próprio Atalaia está funcionando. Tudo aqui é opcional, exceto o endereço de saúde, e nada disso informa dispositivos, endereços ou nomes da sua rede: só estados, contagens, horários e tamanhos.

## Endereço de saúde: `GET /api/health`

Responde **sem login** (mas só para a rede local, como o resto do servidor), para Uptime Kuma, Gatus e similares:

```json
{"status": "ok", "version": "v0.7.0", "database": "ok", "last_scan": "2026-11-22T12:00:00Z"}
```

| `status` | HTTP | Quando |
|---|---|---|
| `ok` | 200 | Tudo certo |
| `degraded` | 200 | Só avisos em `warnings`: `disk_low` (disco do banco quase cheio), `backup_late` (último backup automático com mais de 26 h), `restore_test_failed` (o último teste de restauração falhou) |
| `error` | 503 | Problemas em `problems`: `database` (o banco não responde), `last_scan_failed` (a última varredura falhou), `scans_stale` (nenhuma varredura há mais de 3 intervalos) |

O texto dos erros nunca aparece aqui (ele pode citar a interface de rede); veja o log do servidor ou o app. A versão aparece sem login: quem está na sua rede local consegue saber qual versão você roda.

No Uptime Kuma: monitor do tipo **HTTP(s)** apontando para `http://ENDERECO:8080/api/health` (com HTTPS autoassinado, marque "ignorar erros de TLS").

### `atalaia health` e o Docker

`atalaia health` consulta esse endereço e sai com 0 quando está tudo certo, 1 quando não. Ele lê as mesmas variáveis do `serve` (`ATALAIA_LISTEN`, `ATALAIA_TLS_*`) ou recebe `--url`. A imagem Docker já usa esse comando no `HEALTHCHECK`, então `docker ps` mostra `healthy` ou `unhealthy`.

## Sinais de vida (Healthchecks.io e similares)

Um serviço como o [Healthchecks.io](https://healthchecks.io/) avisa quando um sinal **para de chegar**, o que pega até o caso de a máquina inteira cair.

| Sinal | Como ligar | O que envia |
|---|---|---|
| Varredura | `--heartbeat-scan URL` ou `ATALAIA_HEARTBEAT_SCAN` | `GET URL` depois de cada varredura; `GET URL/fail` quando ela falha |
| Backup | `--heartbeat-backup URL` (backups automáticos), `atalaia backup --heartbeat URL` (à mão) ou `ATALAIA_HEARTBEAT_BACKUP` | `GET URL` quando o backup (e o teste de restauração, quando roda) dá certo; `GET URL/fail` quando falha |

A URL é um segredo (quem a tiver pode mandar sinais falsos): ela nunca aparece nos logs. Espera até 10 segundos e não segue redirecionamentos. No Healthchecks.io, configure o período igual ao intervalo de varredura (5 minutos por padrão) com uma tolerância de alguns minutos.

## Prometheus: `/metrics`

Desligado por padrão. Liga com um token aleatório de 32 caracteres ou mais (ao menos 10 diferentes; `openssl rand -hex 32` serve), de um destes jeitos:

- `--metrics-token-file ARQUIVO` ou `ATALAIA_METRICS_TOKEN_FILE` (melhor: o token não aparece na lista de processos);
- uma credencial do systemd chamada `metrics-token` (`LoadCredential=metrics-token:/etc/atalaia/metrics-token`, veja `/etc/default/atalaia`);
- `ATALAIA_METRICS_TOKEN` (prático no Docker).

```yaml
scrape_configs:
  - job_name: atalaia
    authorization:
      credentials_file: /etc/prometheus/atalaia-token
    static_configs:
      - targets: ["192.168.1.10:8080"]
```

| Métrica | O quê |
|---|---|
| `atalaia_build_info{version}` | Versão em uso (sempre 1) |
| `atalaia_database_up` | 1 quando o banco responde |
| `atalaia_devices`, `atalaia_devices_online`, `atalaia_devices_known` | Contagens do inventário |
| `atalaia_scans_total`, `atalaia_scan_failures_total` | Varreduras e falhas desde que o servidor subiu |
| `atalaia_last_scan_timestamp_seconds`, `atalaia_last_scan_duration_seconds`, `atalaia_last_scan_success` | A última varredura |
| `atalaia_last_backup_timestamp_seconds` | Último backup (automático ou `atalaia backup`) que deu certo |
| `atalaia_last_restore_test_timestamp_seconds`, `atalaia_last_restore_test_success` | O último teste de restauração |
| `atalaia_database_size_bytes` | Banco e o seu log de escrita |
| `atalaia_disk_free_bytes`, `atalaia_disk_size_bytes` | Disco onde fica o banco |

Sem o token, `/metrics` responde 401; desligado, 404.

## Disco quase cheio

Quando o disco do banco tem menos de 512 MB ou menos de 5% livres, o app mostra um aviso em todas as páginas e o endereço de saúde passa a `degraded`.

![Aviso de disco quase cheio (dados fictícios)](/img/screenshots/app-disco-cheio.png)
