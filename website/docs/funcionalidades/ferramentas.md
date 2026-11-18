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

| **Verificação de exposição** | Parte da última varredura de portas: aponta serviços arriscados, lê o que as portas web respondem e confere o TLS. | Mesmas regras da varredura de portas; exige uma varredura antes. |

Cada varredura guarda o resultado e, na seguinte, mostra **o que abriu e o que fechou**: uma porta nova aberta num dispositivo é um bom motivo para olhar com atenção.

## Verificação de exposição

Depois de uma varredura de portas, **Verificar exposição** responde "o que há de arriscado neste dispositivo?":

- **Serviços arriscados abertos**, em três graus (crítico, alto, médio): telnet, Docker API, Redis, MongoDB, SMB, RDP, VNC e outros que raramente deveriam estar acessíveis na rede. É uma **tabela de regras**, não um banco de vulnerabilidades.
- **Interfaces web**: uma única requisição `GET /` por porta web (sem seguir redirecionamentos, lendo no máximo 64 KiB) para mostrar status, cabeçalho `Server`, título da página e se pede login. Serve para achar painéis de administração que você nem lembrava.
- **TLS** nas portas 443, 8443, 465, 993, 995, 636 e 8883: versão negociada, certificado (assunto, validade, autoassinado), se **ainda aceita TLS 1.0/1.1** e chave RSA menor que 2048 bits.

O certificado não é verificado (equipamentos domésticos usam autoassinados); o relatório apenas informa. Portas que começam em texto puro e depois sobem para TLS (STARTTLS) não são cobertas.

## Regras de segurança das ações ativas

- Só o administrador autenticado.
- **Uma por vez**: se outra está rodando, a nova é recusada (429).
- A varredura de portas e o Wake-on-LAN só agem em **dispositivos que já estão no inventário e cujo endereço está no escopo** (as redes privadas do próprio host). Não há campo para digitar um alvo qualquer.
- A varredura de portas exige a **confirmação** "esta é a minha própria rede e o meu próprio dispositivo" e tem **intervalo mínimo de 10 s** entre varreduras e limite de 60 s por execução.
- A varredura de portas só completa o *handshake* TCP e fecha: **não envia dados**. A verificação de exposição envia **uma requisição HTTP e um handshake TLS** por porta aberta aplicável, e nada que tente explorar ou adivinhar senhas.
- **Toda ação fica no registro de auditoria** (data, ação, alvo e resultado), visível na própria página, e é apagado junto com o histórico, conforme a retenção.

## Modo somente observação

Para desligar todas as ações ativas (ping, Wake-on-LAN, varredura de portas, verificação de exposição) use `--no-active-tools`. Veja [Implantação](../tecnico/implantacao.md).

## Ping no Docker e no Linux

O ping usa ICMP **sem privilégios** (um *ping socket*). Ele funciona quando o usuário que roda o Atalaia está no intervalo `net.ipv4.ping_group_range` do sistema; se não estiver, o app avisa. Para liberar para todos os usuários:

```bash
sudo sysctl -w net.ipv4.ping_group_range="0 2147483647"
```

Com a imagem Docker (rede do host, sem capabilities) vale a configuração do host.

## O que ainda não tem

**Verificação de CVEs** (exige um banco de vulnerabilidades e identificar versões com confiabilidade, o que ainda não fazemos), traceroute e teste de velocidade (exige falar com um servidor externo, o que merece um aviso de privacidade próprio). Veja o [roadmap](../roadmap.md).
