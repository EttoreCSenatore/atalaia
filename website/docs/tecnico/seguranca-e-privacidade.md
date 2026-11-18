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
- Sem telemetria e sem contas em nuvem. **Exceção que você liga:** se configurar um canal de alerta (ntfy, Telegram, e-mail, webhook), o texto do alerta (nome, fabricante, IP e MAC do dispositivo) é enviado a esse serviço. Com ntfy.sh ou Telegram isso significa um terceiro; para manter tudo em casa, use um servidor ntfy ou um webhook seu. Veja [Alertas](../funcionalidades/alertas.md).
- O banco SQLite é criado com permissão `0600` (pasta `0700`).
- Tudo o que vem da rede (nomes, fabricantes) é tratado como dado não confiável e limpo antes de ser guardado ou mostrado.

## Privacidade e LGPD

Endereços MAC, nomes de dispositivos e horários de conexão podem identificar pessoas (por exemplo, o celular de um morador). Por isso:

- Os dados ficam num arquivo local do usuário. O projeto não recebe nada.
- A retenção do histórico será configurável (M2).
- Quem opera o Atalaia numa rede de terceiros (uma empresa, um condomínio) é o controlador desses dados e deve informar as pessoas.

A documentação de conformidade do projeto é feita com **revisão assistida por IA** e **não é aconselhamento jurídico**.

## Como relatar uma vulnerabilidade

Use o [relato privado de vulnerabilidades do GitHub](https://github.com/EttoreCSenatore/atalaia/security/advisories/new). Não abra uma issue pública com detalhes e não envie dados pessoais reais. O arquivo `SECURITY.md` do repositório tem o mesmo texto.

## Cadeia de suprimentos

A CI audita vulnerabilidades e licenças das dependências, e o Dependabot mantém as dependências e as ações do GitHub atualizadas.
