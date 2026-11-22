---
id: primeiros-passos
title: Primeiros passos
sidebar_position: 1
---

# Primeiros passos

Do zero até receber o primeiro alerta, em uns 15 minutos. Use o Atalaia só em redes que você administra ou em que tem autorização por escrito.

## 1. Escolha onde rodar

O Atalaia precisa ficar **dentro da rede** que vai vigiar, ligado o tempo todo: um Raspberry Pi, um mini PC, um NAS ou um servidor caseiro com Linux. Ele não roda no navegador nem na nuvem (veja [por quê](../intro.md#por-que-não-roda-só-no-navegador)).

## 2. Instale

O jeito mais simples é o Docker:

```bash
git clone https://github.com/EttoreCSenatore/atalaia && cd atalaia/deploy
ATALAIA_LISTEN=0.0.0.0:8080 docker compose up -d --build
```

Prefere um pacote `.deb`/`.rpm` ou o binário? Veja [Implantação](../tecnico/implantacao.md). O `ATALAIA_LISTEN=0.0.0.0:8080` abre o app para a rede local (o padrão é só a própria máquina); numa rede que não é só sua, ligue também o HTTPS (`ATALAIA_TLS_SELF_SIGNED=1`).

## 3. Crie a conta

Na primeira vez, o servidor escreve no log um **código de configuração** de uso único:

```bash
docker compose logs atalaia | grep -A1 "Setup code"      # pacote: journalctl -u atalaia | grep -A1 "Setup code"
```

Abra `http://ENDEREÇO-DA-MÁQUINA:8080`, digite o código e escolha uma senha. Só existe uma conta, a de administrador.

## 4. Veja a primeira varredura

O painel mostra os dispositivos encontrados, com fabricante, nome e um palpite de tipo. A primeira varredura **nunca** gera alertas: tudo o que está na rede agora é a sua base. Uma varredura nova sai a cada 5 minutos, ou quando você clica **Escanear agora**.

## 5. Diga o que você conhece

Abra cada dispositivo que você reconhece, dê um **Nome** (por exemplo "TV da sala") e marque **Dispositivo conhecido**. O que sobrar desmarcado é o que merece uma segunda olhada. Celulares com **MAC aleatório** trocam de endereço de vez em quando e podem aparecer de novo como novos; veja o [FAQ](faq.md#um-celular-aparece-como-novo-toda-hora).

## 6. Receba alertas no celular

Em **Configurações > Alertas**, clique **Adicionar canal** e escolha **ntfy** (o mais simples: instale o app ntfy no celular e assine um tópico com um nome difícil de adivinhar). Clique **Enviar teste**. A partir daí, um dispositivo desconhecido que entrar na rede gera um aviso. Telegram, webhook e e-mail também funcionam; veja [Alertas](../funcionalidades/alertas.md).

## 7. Pronto. E depois?

- O Atalaia já faz **backup** do banco todo dia e testa a restauração toda semana; copie a pasta de backups para outro lugar de vez em quando ([Backup e restauração](../operacao/backup-e-restauracao.md)).
- Quer saber se ele parou? Veja [Monitoramento](../tecnico/monitoramento.md).
- Mais tarefas do dia a dia: [Como fazer](tarefas.md).
