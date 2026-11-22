---
id: tarefas
title: Como fazer
sidebar_position: 2
---

# Como fazer

Tarefas do dia a dia, cada uma em poucos passos.

## Descobrir o que é um dispositivo desconhecido

1. No painel, abra o dispositivo.
2. Olhe o **fabricante** (pelo MAC), o **nome** (DNS, mDNS, NetBIOS), o **modelo** e os **serviços** anunciados: muitas vezes isso já diz o que é ("Brother", `_ipp._tcp` = impressora).
3. Se o MAC for **aleatório**, é quase sempre um celular, tablet ou notebook moderno; veja quando ele entrou no **histórico** e quem estava em casa.
4. Ainda sem ideia? **Varrer portas TCP comuns** mostra o que ele oferece na rede (uma câmera costuma ter a 554, RTSP).
5. Se não for de ninguém, troque a senha do Wi-Fi: o Atalaia mostra, mas não bloqueia.

## Dar nome, etiquetas e marcar como conhecido

Na página do dispositivo: **Nome**, **Etiquetas** (separadas por vírgula, ex.: `iot, sala`) e **Dispositivo conhecido**. Conhecidos não geram alerta de "dispositivo novo" e podem gerar alerta quando ficam offline (regra opcional).

## Ser avisado quando algo muda

**Configurações > Alertas**: escolha as regras (dispositivo novo, conhecido offline, sinais de ARP spoofing, DHCP falso, portas novas) e pelo menos um canal. Use **Enviar teste** depois de configurar. Detalhes em [Alertas](../funcionalidades/alertas.md).

## Ver se um dispositivo está exposto

1. Na página do dispositivo, **Varrer portas TCP comuns** (confirme; fica no registro de auditoria).
2. Depois, **Verificar exposição**: aponta serviços arriscados (Telnet, SMB, bancos de dados…), lê o que as páginas web respondem e confere os certificados TLS.

Só funciona para dispositivos do inventário, na sua rede local. Veja [Ferramentas de rede](../funcionalidades/ferramentas.md).

## Abrir o Atalaia de outro aparelho da rede

Escute em todos os endereços com HTTPS ligado: `ATALAIA_LISTEN=0.0.0.0:8443` e `ATALAIA_TLS_SELF_SIGNED=1` (Docker: `environment:`; pacote: `/etc/default/atalaia`). Na primeira vez o navegador avisa sobre o certificado: compare a impressão digital que aparece no log do Atalaia antes de aceitar. Veja [Implantação](../tecnico/implantacao.md#https-embutido).

## Abrir pela VPN (Tailscale e similares)

Diga qual rede da VPN é de confiança: `ATALAIA_ALLOW_CIDR=100.64.0.0/10` (Tailscale). Sem isso, endereços de VPN recebem 403. Nunca exponha o Atalaia direto para a internet.

## Conferir se há ataque na rede

O **histórico** destaca em vermelho os sinais de **ARP spoofing** (o gateway mudou de MAC, ou um MAC responde por vários endereços). A verificação de **DHCP falso** é opcional: ligue em **Configurações > Servidores DHCP** (precisa da porta UDP 68). Veja [Segurança passiva](../funcionalidades/seguranca-passiva.md).

## Levar a lista para uma planilha

No painel, **Baixar a lista** (CSV ou JSON), ou `atalaia export --format csv`. O CSV neutraliza fórmulas: nomes vindos da rede nunca viram comandos na planilha.

## Limpar dispositivos antigos

Na página do dispositivo, **Esquecer este dispositivo** (apaga também o histórico dele). Para limpar sozinho, em **Configurações**, **Esquecer dispositivos offline há mais de N dias**: só afeta os que você nunca mexeu.

## Deixar o Atalaia só observando

Inicie com `--no-active-tools` (ou `ATALAIA_NO_ACTIVE_TOOLS=1`): sem ping, traceroute, teste de velocidade, Wake-on-LAN, varredura de portas, exposição nem DHCP. A descoberta e os alertas continuam.

## Trocar a senha

**Configurações > Trocar a senha**. As outras sessões são encerradas. Esqueceu a senha? Veja o [FAQ](faq.md#esqueci-a-senha).
