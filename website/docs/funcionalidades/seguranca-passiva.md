---
id: seguranca-passiva
title: Segurança passiva (ARP)
sidebar_position: 4
---

# Segurança passiva: sinais de ARP spoofing

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

## O que ainda vem

Detecção de servidor **DHCP falso** e de **serviços alterados** exigem enviar pacotes especiais (porta 68, sockets brutos) e capacidades extras no contêiner; ficam para um próximo passo, como recurso opcional. Veja o [roadmap](../roadmap.md).
