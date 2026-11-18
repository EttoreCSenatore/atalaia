---
id: ferramentas
title: Ferramentas de rede
sidebar_position: 5
---

# Ferramentas de rede

Em **Ferramentas** (e na página de cada dispositivo) ficam as ações que **mexem na rede**. Por isso elas têm regras mais rígidas que o resto do app.

![Ferramentas de rede](/img/screenshots/app-ferramentas.png)

## O que tem

| Ferramenta | O que faz | Limites |
|---|---|---|
| **Ping** | Envia até 10 pedidos de eco ICMP e mostra o tempo de resposta. | Só IPv4 unicast; recusa multicast e broadcast. |
| **Consulta DNS** | Registros A, AAAA, CNAME, MX, NS, TXT e PTR, pelo resolvedor do sistema. | Até 20 respostas, texto limpo de caracteres de controle. |
| **Wake-on-LAN** | Envia o "pacote mágico" para acordar um dispositivo. | Só dispositivos do inventário, cujo endereço está numa rede local do host. |
| **Varredura de portas** | Tenta conectar em cerca de 65 portas TCP comuns de **um dispositivo** e lista as abertas. | Veja abaixo. |

Cada varredura guarda o resultado e, na seguinte, mostra **o que abriu e o que fechou**: uma porta nova aberta num dispositivo é um bom motivo para olhar com atenção.

## Regras de segurança das ações ativas

- Só o administrador autenticado.
- **Uma por vez**: se outra está rodando, a nova é recusada (429).
- A varredura de portas e o Wake-on-LAN só agem em **dispositivos que já estão no inventário e cujo endereço está no escopo** (as redes privadas do próprio host). Não há campo para digitar um alvo qualquer.
- A varredura de portas exige a **confirmação** "esta é a minha própria rede e o meu próprio dispositivo" e tem **intervalo mínimo de 10 s** entre varreduras e limite de 60 s por execução.
- A varredura só completa o *handshake* TCP e fecha: **não envia dados** nem tenta identificar versões ou explorar nada.
- **Toda ação fica no registro de auditoria** (data, ação, alvo e resultado), visível na própria página, e é apagado junto com o histórico, conforme a retenção.

## Ping no Docker e no Linux

O ping usa ICMP **sem privilégios** (um *ping socket*). Ele funciona quando o usuário que roda o Atalaia está no intervalo `net.ipv4.ping_group_range` do sistema; se não estiver, o app avisa. Para liberar para todos os usuários:

```bash
sudo sysctl -w net.ipv4.ping_group_range="0 2147483647"
```

Com a imagem Docker (rede do host, sem capabilities) vale a configuração do host.

## O que ainda não tem

Traceroute, teste de velocidade (exige falar com um servidor externo, o que merece um aviso de privacidade próprio), verificação de CVEs, de TLS e de painéis expostos. Veja o [roadmap](../roadmap.md).
