---
id: seguranca-passiva
title: Segurança passiva
sidebar_position: 4
---

# Segurança passiva

## Sinais de ARP spoofing

No *ARP spoofing* um dispositivo da rede finge ser outro (quase sempre o roteador) para passar o tráfego dos demais por ele e poder espiá-lo ou alterá-lo. O Atalaia procura os dois sinais dessa técnica que dá para ver **sem privilégios e sem enviar nada além da varredura normal**: ele só confere o que o kernel já aprendeu sobre a rede.

## O que é detectado

| Evento | Quando acontece | Por que importa |
|---|---|---|
| **MAC do roteador mudou** (`gateway_changed`) | O endereço do gateway padrão (lido da tabela de rotas) passa a responder por outro MAC. | É o sinal clássico de interceptação. Também ocorre ao trocar o roteador: aí, é só conferir. |
| **Responde por vários endereços** (`multi_ip`) | Um mesmo MAC responde por 3 ou mais endereços IP na mesma varredura. | Quem envenena a rede responde por vários IPs. Roteadores, Docker e máquinas com endereços virtuais também fazem isso. |

Cada evento é registrado **uma vez, quando começa**, no histórico (em destaque vermelho), e dispara um alerta se a regra **Sinais de ARP spoofing** estiver ligada (padrão: ligada; veja [Alertas](alertas.md)). Se o dispositivo parar e recomeçar, avisa de novo. Diferente dos alertas de dispositivo novo, esses **valem também na primeira varredura**.

## Dispositivos conhecidos

Um dispositivo marcado como **conhecido** nunca é apontado como "responde por vários endereços": é a forma de silenciar um roteador ou um servidor com endereços virtuais que você reconhece. O alerta de troca de MAC do roteador não tem essa exceção.

## Limites (leia antes de confiar)

- É **heurística**: pode haver falso positivo (explicados acima) e falso negativo. Um atacante cuidadoso que responda por um endereço só, sem trocar o do roteador, **não será visto**.
- A conferência acontece **a cada varredura** (padrão: 5 minutos), não em tempo real: o Atalaia não escuta pacotes ARP.
- Só IPv4, só Linux, só a rede local.

## Servidor DHCP falso

Quem distribui os endereços da rede (o DHCP) também diz qual é o gateway e o DNS. Um **segundo servidor DHCP** que você não configurou (um roteador sobrando em modo roteador, um aparelho malicioso) pode mandar os dispositivos para um gateway ou DNS errado e desviar o tráfego.

Em **Configurações → Servidores DHCP** o Atalaia envia **uma descoberta DHCP em broadcast**, escuta as ofertas e lembra quais servidores responderam, com o gateway e o DNS que cada um ofereceria. Ele **nunca envia o pedido** que fecha o empréstimo, então não toma nenhum endereço. A **primeira verificação** só registra quais servidores são seus; depois disso, um servidor novo gera o evento **Novo servidor DHCP respondeu** (em vermelho no histórico, se o endereço dele é um dispositivo do inventário) e um alerta (regra ligada por padrão). Trocou de roteador? Use **Esquecer servidores conhecidos** e a próxima verificação recomeça.

- É **opcional e começa desligado**: você pode só **Verificar agora** (com confirmação) ou ligar a verificação **depois de cada varredura**.
- Precisa **escutar a porta UDP 68, privilegiada**. Rode como root, ou dê a capacidade `CAP_NET_BIND_SERVICE`: no systemd, crie `sudo systemctl edit atalaia` com `[Service]` / `AmbientCapabilities=CAP_NET_BIND_SERVICE` / `CapabilityBoundingSet=CAP_NET_BIND_SERVICE`; no Docker, `cap_add: [NET_BIND_SERVICE]` no compose. Sem isso o app explica o erro.
- Coexiste com o cliente DHCP da própria máquina (a porta é compartilhada) e fica desligado no [modo somente observação](../tecnico/implantacao.md).
- Limites: um servidor falso que **não responder** à nossa descoberta (ou que só responda a clientes conhecidos) não aparece, e uma resposta só vem dos servidores da rede onde o Atalaia está.

## Serviços alterados

Cada varredura de portas (veja [Ferramentas](ferramentas.md)) é comparada com a anterior do mesmo dispositivo. Se **uma porta nova abriu**, a página do dispositivo destaca e sai um alerta (regra **Uma varredura de portas encontra portas recém-abertas**, ligada por padrão). A varredura de portas é uma ação ativa, então **não roda sozinha**: a comparação acontece quando você varre.

## O que ainda vem

Varredura de portas agendada (opcional) e a detecção de serviços por mudança de versão ficam para o futuro. Veja o [roadmap](../roadmap.md).
