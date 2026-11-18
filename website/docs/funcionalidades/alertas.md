---
id: alertas
title: Alertas e dispositivos conhecidos
sidebar_position: 3
---

# Alertas e dispositivos conhecidos

O Atalaia pode avisar você quando algo acontece na rede. A mensagem sai **direto desta máquina** para o serviço que você escolher: não há servidor nosso no meio.

## Nomes, etiquetas e "conhecido"

Na página de cada dispositivo você pode dar um **nome** (que vale mais que o nome descoberto na rede e nunca é sobrescrito pelas varreduras), até 8 **etiquetas** (`iot`, `visita`, `celular`…) e marcar o dispositivo como **conhecido**. Os desconhecidos são os que merecem uma segunda olhada. Nome e etiquetas entram na busca do painel.

## Quando avisar

| Regra | Padrão | O que faz |
|---|---|---|
| Dispositivo novo | ligada | Um dispositivo que nunca tinha sido visto entra na rede. |
| Conhecido ficou offline | desligada | Um dispositivo marcado como conhecido sumiu (depois do prazo de tolerância). |

A **primeira varredura** de um banco vazio só monta o inventário e **nunca** dispara alertas; senão você receberia um aviso para cada aparelho da casa. Quando uma varredura encontra vários dispositivos de uma vez, sai **uma só mensagem** com a lista (no máximo 10 linhas).

## Canais

| Canal | O que informar |
|---|---|
| **ntfy** | Endereço do tópico (`https://ntfy.sh/seu-topico-secreto` ou o seu servidor) e, se quiser, um token de acesso. |
| **Telegram** | Token do bot (criado com o @BotFather) e o ID do chat. |
| **Webhook** | Um endereço `http://` ou `https://`; recebe um JSON `{"title","message","source":"atalaia"}` (serve para Home Assistant, n8n etc.). |
| **E-mail** | Servidor `host:porta`, usuário e senha (opcionais), remetente e destinatário. |

Cada canal tem o botão **Enviar teste**. Os canais vão em **Configurações → Alertas** e dá para ter até 8.

![Configuração de alertas](/img/screenshots/app-alertas.png)

## Segurança

- Tokens e senhas ficam **só no banco SQLite desta máquina** (arquivo `0600`) e **nunca são devolvidos** pela API: a tela mostra apenas que há um segredo salvo, e deixar o campo vazio mantém o que já existe.
- O e-mail usa STARTTLS quando o servidor oferece e **se recusa a enviar a senha sem TLS**.
- Os textos vêm da rede (nomes de dispositivos), então são enviados como **texto simples**, nunca como marcação, e quebras de linha são removidas dos títulos (injeção de cabeçalhos).
- Mensagens de erro nunca repetem o endereço do canal, que pode conter um token.
- Os endereços de ntfy e webhook podem apontar para a rede local (é o caso do Home Assistant). Só o administrador autenticado configura canais.
- Falha em um canal é registrada no log e não atrapalha os outros nem a varredura.

Veja os endpoints em [API e servidor](../tecnico/api.md).
