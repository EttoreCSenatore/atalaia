---
id: modelo-de-ameacas
title: Modelo de ameaças
sidebar_position: 2
---

# Modelo de ameaças

O Atalaia é uma ferramenta de segurança que também é um alvo: ele conhece todos os dispositivos da rede e pode sondá-los. Este documento diz o que protegemos, de quem, e o que o projeto **não** faz.

## O que protegemos

- O inventário da rede (MACs, nomes, serviços), que é dado pessoal em muitos contextos.
- As credenciais do painel e os segredos de notificação (tokens).
- A própria rede: o Atalaia não pode virar uma ferramenta de ataque.

## Quem pode atacar

| Ator | Exemplo |
|---|---|
| Outro dispositivo da rede | Tenta abrir o painel, adivinhar a senha ou enviar pacotes falsos ao agente |
| Pessoa na internet | Alcança o painel se o usuário o expuser por engano |
| Usuário mal-intencionado | Usa o Atalaia para varrer a rede de outra pessoa |
| Cadeia de suprimentos | Dependência ou imagem comprometida |

## Ameaças e mitigações

| Ameaça | Mitigação |
|---|---|
| Painel acessível sem login | Login obrigatório no primeiro uso, sem senha padrão |
| Senha e sessão lidas na rede local (HTTP puro) | HTTPS embutido opcional (TLS 1.2+, cookie `Secure`, HSTS com certificado próprio); sem ele, o aviso de HTTP puro no console |
| Painel exposto à internet | Escuta em `127.0.0.1` por padrão; mesmo com `--listen`, recusa origens que não são da rede local; documentação recomenda VPN ou proxy com TLS |
| DNS rebinding e CSRF contra o painel | Verificação do `Host`, cabeçalho `X-Atalaia` e conferência do `Origin`, cookie `SameSite=Strict` |
| Adivinhar a senha | argon2id, bloqueio progressivo por endereço |
| Uso para varrer redes de terceiros | Escopo fixo nas redes privadas do host (`internal/scope`), sem alvo arbitrário |
| Varredura ativa sem consentimento | Opcional, exige confirmação "esta é a minha rede", limitada em taxa, registrada em log de auditoria |
| Agente comprometido virar porta de entrada | Roda sem root, só com `CAP_NET_RAW`; imagem mínima sem privilégios extras |
| Administrador comprometido disparando ações ativas | Modo somente observação (`--no-active-tools`) as desliga por completo; a unidade systemd roda sem privilégios e com filtro de chamadas de sistema |
| Atalaia usado como ferramenta de ataque (varrer redes ou máquinas de terceiros) | Ações ativas só sobre dispositivos do inventário no escopo local, sem alvo livre; confirmação, uma por vez, intervalo mínimo e registro de auditoria; a varredura só conecta, não envia dados |
| Outro dispositivo se passar pelo roteador (ARP spoofing) | Detecção passiva de dois sinais (MAC do gateway mudou; um MAC com vários IPs) com alerta; heurística, com limites documentados |
| Segredo de um canal de alerta vazar (token do Telegram, senha de e-mail) | Guardado só no banco local `0600`, nunca devolvido pela API nem repetido em erros; e-mail recusa enviar senha sem TLS |
| Nome de dispositivo malicioso usado para atacar o alerta | Enviado como texto simples, sem marcação; quebras de linha removidas dos títulos |
| Dados vazarem | Armazenamento local, sem telemetria, retenção configurável |
| Entrada maliciosa vinda da rede (nomes, banners) | Tratada como dado não confiável: escapada na interface, nunca executada |
| Dependência ou imagem comprometida | CI com auditoria de vulnerabilidades e de licenças, SBOM e Dependabot |

## Fora do escopo, de propósito

Bloquear dispositivos com ARP spoofing, desautenticar Wi-Fi, explorar vulnerabilidades e fazer ataque de força bruta. A única verificação de credenciais prevista usa uma lista pública e curta de senhas padrão de fábrica, é opcional e fica registrada.

## Uso responsável

Use o Atalaia só em redes que você administra ou em que tem autorização por escrito. Varrer redes alheias pode ser crime. Veja também [Segurança e privacidade](seguranca-e-privacidade.md).
