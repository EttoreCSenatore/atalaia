---
id: intro
title: Visão geral
slug: /intro
sidebar_position: 1
---

# Atalaia

**Atalaia** (em português, a torre de vigia) é um vigia de rede **livre, de código aberto e auto-hospedado**. Você o instala numa máquina da sua rede (um Raspberry Pi, um mini-PC, um NAS ou o seu computador) e abre o app web de qualquer navegador para ver **o que está conectado**, **o que mudou** e **o que está exposto**.

É um projeto independente, inspirado na categoria de ferramentas de descoberta de dispositivos como o Fing. Não tem ligação com a Fing Limited e não usa suas marcas, imagens nem bases de dados.

![Página inicial do site do Atalaia](/img/screenshots/site-inicio.png)

## O que ele faz

| Recurso | Para quê |
|---|---|
| **Descoberta de dispositivos** | Lista o que está na rede (varredura UDP + tabela ARP, nomes por mDNS e DNS reverso) com fabricante; marca MACs aleatórios. |
| **Histórico e alertas** | Linha do tempo online/offline, nomes e etiquetas, e aviso de dispositivo novo e outros eventos (ntfy, Telegram, e-mail, webhook). |
| **Segurança passiva** | Sinais de ARP spoofing (MAC do roteador mudou, um MAC com vários IPs), servidor DHCP falso (opcional) e portas recém-abertas. |
| **Segurança ativa (opcional)** | Varredura de portas, serviços arriscados, interfaces web expostas e auditoria de TLS, com confirmação, limites e registro de auditoria. Sem verificação de CVEs, de propósito. |
| **Ferramentas de rede** | Ping, traceroute, consulta DNS, Wake-on-LAN e teste de velocidade de download. |

## Por que não roda só no navegador

Um navegador não consegue enviar ARP, ICMP nem ouvir mDNS: ele só fala HTTP. Por isso o Atalaia tem duas partes: um **agente em Go**, que roda na sua rede e faz a descoberta, e um **app web** servido pelo próprio agente. Veja a [arquitetura](tecnico/arquitetura.md).

## Princípios

- **Defensivo**: serve para auditar a **sua** rede, não para atacar a dos outros. Veja o [modelo de ameaças](tecnico/modelo-de-ameacas.md).
- **Seguro por padrão**: login obrigatório, só na rede local, varreduras limitadas às suas redes, verificações ativas opcionais e registradas. Veja [Segurança e privacidade](tecnico/seguranca-e-privacidade.md).
- **Privacidade**: os dados ficam na sua máquina. Nada vai para a nuvem.
- **Software livre**: código sob AGPL-3.0-or-later, documentação sob CC BY-SA 4.0. Veja [Licença](licenca.md).
- **Bilíngue**: documentação em português e inglês.

## Estado atual

O Atalaia está na **v0.6** e já é usável: `atalaia serve` descobre e identifica os dispositivos, guarda o histórico, serve o **app web** com login seguro (HTTP ou HTTPS embutido), faz verificações defensivas, backups diários e avisa por ntfy, Telegram, webhook ou e-mail. Para começar, veja [Primeiros passos](manual/primeiros-passos.md). Há pacotes `.deb` e `.rpm`, binários Linux (amd64, arm64, armv7) e uma imagem Docker. É um projeto novo, mantido por uma pessoa e **ainda sem teste de invasão independente**; veja os [limites conhecidos](funcionalidades/descoberta.md#limites-atuais) e o [roadmap](roadmap.md).

