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
- O app instalável (PWA) guarda no aparelho só o casco do app (HTML, JavaScript, CSS, ícones). O service worker nunca responde nem guarda `/api/` ou `/metrics`, então os dados da rede não ficam no cache por causa dele; um teste confere isso. Veja [Instalar como app](../funcionalidades/app-web.md#instalar-como-app).

## Privacidade e LGPD

Endereços MAC, nomes de dispositivos e horários de conexão podem identificar pessoas (por exemplo, o celular de um morador). Por isso:

- Os dados ficam num arquivo local do usuário. O projeto não recebe nada.
- O histórico de eventos é apagado depois de 90 dias por padrão; o prazo muda em Configurações (0 guarda para sempre).
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

## Revisão da v0.7 (2026-11-22)

Revisão manual de tudo o que a v0.7 acrescentou (endereço de saúde sem login, `/metrics`, backups automáticos e "Fazer backup agora", teste de restauração, cópia antes de migrar, `reset-password`, scripts dos pacotes, demonstração pública), com govulncheck, gosec, staticcheck, `npm audit` e Trivy (vulnerabilidades e segredos na imagem) sem achados.

| Achado | Decisão |
|---|---|
| Um link simbólico plantado com o nome do próximo backup fazia a cópia ser gravada onde ele apontasse (só possível numa pasta de backups gravável por outros usuários) | Corrigido: o nome é conferido sem seguir links, o resultado precisa ser um arquivo comum, e o backup automático **recusa uma pasta gravável por outros usuários** |
| `verify-backup` abre um arquivo de banco que pode vir de fora | Endurecido: o banco abre com `trusted_schema` desligado (views e triggers de um arquivo estranho não rodam funções com efeitos) |
| Um token de `/metrics` longo mas repetitivo (`aaaa…`) era aceito | Corrigido: além de 32+ caracteres, o token precisa ter pelo menos 10 caracteres diferentes |
| A limpeza de cópias de rascunho velhas rodava também na pasta de um arquivo passado ao `verify-backup` | Corrigido: só o backup automático limpa, e só na própria pasta |
| `/api/health` responde sem login e mostra a versão | **Aceito**: só para a rede local e com a checagem de `Host`; nada sobre dispositivos nem textos de erro |
| `POST /api/backup` pode ser chamado várias vezes | **Aceito**: só o administrador, com proteção CSRF, um backup por vez, e a rotação limita quantas cópias ficam |
| `reset-password` não pede senha | **Aceito**: exige acesso ao arquivo do banco, que já dá controle total do Atalaia |
| A URL dos sinais de vida é um segredo | Nunca vai para o log nem para mensagens de erro; espera no máximo 10 s e não segue redirecionamentos |
| A demonstração simula a API com dados fictícios | O build real não leva esse código (a CI confere) |

## Como relatar uma vulnerabilidade

Use o [relato privado de vulnerabilidades do GitHub](https://github.com/EttoreCSenatore/atalaia/security/advisories/new). Não abra uma issue pública com detalhes e não envie dados pessoais reais. O arquivo `SECURITY.md` do repositório tem o mesmo texto.

## Cadeia de suprimentos

A CI audita vulnerabilidades e licenças das dependências, e o Dependabot mantém as dependências e as ações do GitHub atualizadas, com espera de 7 dias; ações e imagens são fixadas por SHA e digest (ver [Cadeia de suprimentos](cadeia-de-suprimentos.md)).
