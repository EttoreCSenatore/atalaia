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
| **Descoberta de dispositivos** | Lista tudo o que está na rede (ARP, mDNS, SSDP, NetBIOS, DHCP) com fabricante, nome e tipo. |
| **Histórico e alertas** | Linha do tempo online/offline e aviso quando um dispositivo novo entra (ntfy, Telegram, e-mail, webhook). |
| **Segurança passiva** | Detecta ARP spoofing, IP duplicado, troca de MAC, servidor DHCP falso e serviços novos. |
| **Segurança ativa (opcional)** | Portas e serviços, CVEs conhecidas, painéis expostos, TLS e cabeçalhos HTTP, UPnP, DNS aberto. |
| **Ferramentas de rede** | Ping, traceroute, DNS, Wake-on-LAN, velocidade e qualidade da internet. |

## Por que não roda só no navegador

Um navegador não consegue enviar ARP, ICMP nem ouvir mDNS: ele só fala HTTP. Por isso o Atalaia tem duas partes: um **agente em Go**, que roda na sua rede e faz a descoberta, e um **app web** servido pelo próprio agente. Veja a [arquitetura](tecnico/arquitetura.md).

## Princípios

- **Defensivo**: serve para auditar a **sua** rede, não para atacar a dos outros. Veja o [modelo de ameaças](tecnico/modelo-de-ameacas.md).
- **Seguro por padrão**: login obrigatório, só na rede local, varreduras limitadas às suas redes, verificações ativas opcionais e registradas. Veja [Segurança e privacidade](tecnico/seguranca-e-privacidade.md).
- **Privacidade**: os dados ficam na sua máquina. Nada vai para a nuvem.
- **Software livre**: código sob AGPL-3.0-or-later, documentação sob CC BY-SA 4.0. Veja [Licença](licenca.md).
- **Bilíngue**: documentação em português e inglês.

## Estado atual

A fundação (M0) e a **descoberta de dispositivos** (M1) estão prontas: `atalaia scan` encontra os dispositivos da rede, identifica fabricante e nome e guarda o histórico. `atalaia serve` roda a varredura agendada, a API com login seguro e o **app web** (painel, dispositivos, histórico), também em imagem Docker. Os alertas vêm no M3. Veja [Descoberta](funcionalidades/descoberta.md) e o [roadmap](roadmap.md).
