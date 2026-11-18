---
id: api
title: API e servidor
sidebar_position: 2
---

# API e servidor

`atalaia serve` roda a varredura em segundo plano e serve a API HTTP (e o app web) no mesmo processo.

```bash
atalaia serve                                  # http://127.0.0.1:8080, varre a cada 5 minutos
atalaia serve --listen 0.0.0.0:8080            # toda a rede local
atalaia serve --interval 2m --allow-host atalaia.lan
```

## Primeiro uso

Não existe senha padrão. Na primeira execução o servidor imprime no console um **código de configuração** de uso único. Quem cria a conta de administrador precisa ler esse código, então só quem tem acesso ao console do servidor consegue "tomar" a instalação, mesmo que outras pessoas da rede alcancem a porta antes.

## Proteções

| Proteção | O que faz |
|---|---|
| Só rede local | Pedidos de endereços públicos (por exemplo, uma porta aberta no roteador) recebem 403. Aceita loopback, redes privadas e link-local. |
| Verificação do `Host` | Barra DNS rebinding: só aceita endereços IP, `localhost`, o nome da máquina e os `--allow-host`. |
| Sessão | Cookie `HttpOnly` e `SameSite=Strict` (`Secure` quando há TLS). O servidor guarda só o hash SHA-256 do token. Validade de 7 dias. |
| CSRF | Todo pedido que altera algo exige o cabeçalho `X-Atalaia`, e o `Origin`, quando existe, precisa ser o próprio servidor. |
| Senha | argon2id (19 MiB, 2 passes), 10 a 128 caracteres. Trocar a senha encerra as outras sessões. |
| Tentativas | 5 senhas erradas bloqueiam o endereço por 1 minuto, dobrando até 15 minutos. |
| Cabeçalhos | CSP restritiva, `nosniff`, `X-Frame-Options: DENY`, `no-referrer`, `no-store` nas respostas da API. |
| Corpo | JSON de até 4 KiB, sem campos desconhecidos. |

:::warning HTTP puro
O servidor fala HTTP. Numa rede local confiável isso é aceitável, mas a senha e o cookie não são criptografados no trajeto. Para qualquer coisa além disso, coloque um proxy reverso com TLS ou uma VPN na frente.
:::

## Endpoints

| Método e caminho | Login | Descrição |
|---|---|---|
| `GET /api/status` | não | Versão, se falta configurar, se há sessão. Com sessão, inclui o estado da varredura. |
| `POST /api/setup` | não | `{code, password}`: cria o administrador e já abre a sessão. |
| `POST /api/login` | não | `{password}`. |
| `POST /api/logout` | sim | Encerra a sessão. |
| `POST /api/password` | sim | `{current, new}`. |
| `GET /api/devices` | sim | Dispositivos conhecidos. |
| `PUT /api/devices/{mac}` | sim | `{label, tags, known}`: nome (até 64 caracteres), até 8 etiquetas (24 caracteres, sem vírgula) e se é conhecido. |
| `POST /api/devices/{mac}/wake` | sim | Envia o pacote mágico (Wake-on-LAN). Só dispositivos do inventário, no escopo. |
| `POST /api/devices/{mac}/portscan` | sim | `{confirm: true}` obrigatório. Varre ~65 portas TCP; no máx. uma a cada 10 s; devolve `ports` e, se havia resultado anterior, `opened`/`closed`. |
| `POST /api/devices/{mac}/exposure` | sim | `{confirm: true}`; exige uma varredura de portas antes (409). Devolve `risks`, `web` e `tls`. Mesmos limites da varredura. |
| `GET /api/devices/{mac}/exposure` | sim | Último relatório de exposição. |
| `GET /api/devices/{mac}/ports` | sim | Último resultado da varredura de portas. |
| `POST /api/tools/ping` | sim | `{host, count}` (1 a 10). ICMP sem privilégios; 501 se o sistema não permitir. |
| `POST /api/tools/dns` | sim | `{name, type}` (A, AAAA, CNAME, MX, NS, TXT, PTR). |
| `GET /api/audit` | sim | Registro das ações ativas, do mais novo para o mais antigo. |
| `GET /api/events?limit=N` | sim | Histórico (1 a 1000), do mais novo para o mais antigo. Tipos: `joined`, `online`, `offline`, `ip_changed`, `gateway_changed`, `multi_ip`. |
| `POST /api/scan` | sim | Pede uma varredura agora (responde 202). |
| `GET /api/stream` | sim | Eventos ao vivo (SSE): `hello`, `scan_started`, `scan_finished`, `scan_failed`. |
| `GET/PUT /api/settings` | sim | `retention_days` (0 a 3650; 0 guarda para sempre; padrão 90). |
| `GET/PUT /api/alerts` | sim | `{rules: {new_device, known_offline, arp_anomaly}, channels: [...]}`. Os segredos (`token`, `password`) nunca voltam no GET; no PUT, um canal enviado sem segredo mantém o que já está salvo. Até 8 canais, corpo de até 32 KiB. |
| `POST /api/alerts/test` | sim | `{id}`: envia uma mensagem de teste por um canal salvo. |

Exemplo, com a sessão em `cookies.txt`:

```bash
curl -b cookies.txt -H 'X-Atalaia: 1' -X POST http://127.0.0.1:8080/api/scan
```

## Retenção

Depois de cada varredura o servidor apaga os eventos mais antigos que `retention_days` e as sessões vencidas. Os dispositivos em si não são apagados.
