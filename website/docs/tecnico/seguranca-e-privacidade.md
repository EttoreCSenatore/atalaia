---
id: seguranca-e-privacidade
title: Segurança e privacidade
sidebar_position: 3
---

# Segurança e privacidade

## Padrões seguros

- Login obrigatório, sem senha padrão: a conta é criada com um código de uso único impresso no console.
- Escuta em `127.0.0.1` por padrão; com `--listen` o servidor ainda recusa pedidos que não vêm da rede local, e confere o cabeçalho `Host`. Veja [API e servidor](api.md).
- Varreduras ativas só nas redes privadas do próprio host, limitadas a no máximo uma /16.
- Verificações ativas são opcionais, limitadas em taxa e registradas em log de auditoria.
- Execução sem root, com `CAP_NET_RAW`.
- Sem telemetria e sem contas em nuvem. **Exceção que você liga:** se configurar um canal de alerta (ntfy, Telegram, e-mail, webhook), o texto do alerta (nome, fabricante, IP e MAC do dispositivo) é enviado a esse serviço. Com ntfy.sh ou Telegram isso significa um terceiro; para manter tudo em casa, use um servidor ntfy ou um webhook seu. Veja [Alertas](../funcionalidades/alertas.md). **Outra exceção que você liga:** o teste de velocidade baixa um arquivo de `speed.cloudflare.com`, que passa a ver o IP público da sua rede; o app pede confirmação a cada uso.
- O banco SQLite é criado com permissão `0600` (pasta `0700`).
- Tudo o que vem da rede (nomes, fabricantes) é tratado como dado não confiável e limpo antes de ser guardado ou mostrado.

## Privacidade e LGPD

Endereços MAC, nomes de dispositivos e horários de conexão podem identificar pessoas (por exemplo, o celular de um morador). Por isso:

- Os dados ficam num arquivo local do usuário. O projeto não recebe nada.
- A retenção do histórico será configurável (M2).
- Quem opera o Atalaia numa rede de terceiros (uma empresa, um condomínio) é o controlador desses dados e deve informar as pessoas.

A documentação de conformidade do projeto é feita com **revisão assistida por IA** e **não é aconselhamento jurídico**.

## Autorrevisão de segurança (M6.2)

O código passou por uma revisão manual das partes críticas (login, sessões, proteção da rede local, ferramentas ativas, segredos) e por análise estática com o **gosec**, agora também **na CI**. O que a revisão encontrou e o que foi decidido:

| Achado | Decisão |
|---|---|
| Tentativas de senha em paralelo passavam todas pela checagem do bloqueio antes de qualquer falha ser contada (revisão de 2026-11-19) | Corrigido: a tentativa é contada **antes** de conferir a senha |
| Trocar de endereço IP na rede multiplicava as tentativas | Corrigido: além do limite por endereço, um limite para **todos juntos** (cerca de 30 por minuto, bloqueio de no máximo 1 minuto, para que um atacante não consiga trancar o dono do lado de fora) |
| Muitos logins ao mesmo tempo podiam esgotar a memória (cada argon2id usa ~19 MiB) | Corrigido: no máximo 2 verificações de senha simultâneas |
| Respostas SSDP falsificadas em massa | Corrigido: no máximo 256 dispositivos por rodada de descoberta |
| Conversões `int`→`uint16`/`uint32`/`byte` apontadas (G115) | Corrigidas: porta fora de 1–65535 vira porta 0 (recusada); tamanho da chave do argon2 limitado a 128 bytes; máscaras explícitas na leitura da tabela de rotas |
| Cookie de logout sem `Secure` (G124) | Corrigido; novo `--secure-cookies` (`ATALAIA_SECURE_COOKIES=1`) marca o cookie `Secure` atrás de um proxy que termina TLS |
| Cookie `Secure` só quando há HTTPS (G124) | **Aceito**: um cookie `Secure` em HTTP puro seria descartado pelo navegador |
| Leitura de arquivos por caminho variável (G304) | **Aceito**: os caminhos vêm de flags do administrador ou de arquivos do sistema (`/proc`), nunca de entrada da rede |
| Campo `password` serializado em JSON (G117) | **Aceito**: é a senha SMTP guardada de propósito no banco local (`0600`) e nunca devolvida pela API |
| Endereços de ntfy/webhook podem apontar para a rede local | **Aceito**: é o caso de uso (Home Assistant); só o administrador autenticado configura canais |
| Segredos dos canais (token do Telegram, senha SMTP) ficam em texto no banco | **Aceito e documentado**: o arquivo é `0600`; proteja o backup do banco como protegeria uma senha |
| Endereços de VPN em CGNAT (por exemplo Tailscale, `100.64.0.0/10`) não contam como "rede local" | **Resolvido com opção**: o servidor responde 403 por padrão; `--allow-cidr 100.64.0.0/10` aceita a VPN (veja [Implantação](implantacao.md#acesso-por-vpn-tailscale-e-similares)) |

Esta revisão foi feita pelo próprio projeto, **não** substitui um teste de invasão independente.

## Como relatar uma vulnerabilidade

Use o [relato privado de vulnerabilidades do GitHub](https://github.com/EttoreCSenatore/atalaia/security/advisories/new). Não abra uma issue pública com detalhes e não envie dados pessoais reais. O arquivo `SECURITY.md` do repositório tem o mesmo texto.

## Cadeia de suprimentos

A CI audita vulnerabilidades e licenças das dependências, e o Dependabot mantém as dependências e as ações do GitHub atualizadas.
