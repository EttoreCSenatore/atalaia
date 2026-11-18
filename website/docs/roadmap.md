---
id: roadmap
title: Roadmap
sidebar_position: 3
---

# Roadmap

Cada etapa vira um pull request, revisado e integrado antes da próxima. Documentação (pt-BR e inglês) e memória do projeto são atualizadas a cada etapa.

| Etapa | Entrega | Estado |
|---|---|---|
| **M0** | Fundação: repositório, licença, documentação bilíngue no GitHub Pages, CI, esqueleto em Go e guarda de escopo de varredura | Concluído |
| **M1** | Descoberta: ARP e mDNS, lista de dispositivos com fabricante e nome, SQLite, CLI | Concluído (ver [Descoberta](funcionalidades/descoberta.md)) |
| **M2.1** | Servidor: API HTTP, login seguro, varredura agendada, eventos ao vivo, retenção | Concluído (ver [API](tecnico/api.md)) |
| **M2.2** | App web embutido: painel, detalhe do dispositivo, linha do tempo e imagem Docker | Concluído (ver [App web](funcionalidades/app-web.md) e [Implantação](tecnico/implantacao.md)) |
| **M3** | Alertas: dispositivo novo, adaptadores de notificação (ntfy, Telegram, webhook, e-mail), nomes, etiquetas e dispositivos conhecidos | Concluído (ver [Alertas](funcionalidades/alertas.md)) |
| **M4.1** | Segurança passiva: sinais de ARP spoofing (MAC do roteador mudou, um MAC com vários IPs) | Concluído (ver [Segurança passiva](funcionalidades/seguranca-passiva.md)) |
| **M4.2** | Segurança passiva: servidor DHCP falso e serviços alterados (opcional, exige capacidades extras) | Planejado |
| **M5.1** | Ferramentas: ping, DNS, Wake-on-LAN e varredura de portas com confirmação, limites e registro de auditoria | Concluído (ver [Ferramentas](funcionalidades/ferramentas.md)) |
| **M5.2** | Segurança ativa: serviços arriscados, painéis web expostos e auditoria de TLS | Concluído (ver [Ferramentas](funcionalidades/ferramentas.md)) |
| **M5.3** | CVEs por versão de serviço, traceroute e teste de velocidade | Planejado |
| **M6.1** | HTTPS embutido (certificado próprio ou autoassinado com renovação) | Concluído (ver [Implantação](tecnico/implantacao.md)) |
| **M6.2** | Autorrevisão de segurança do código | Planejado |
| **M6.3** | Unidade systemd, pacotes .deb/.rpm e modo sem ferramentas ativas | Planejado |

A **v0.1** é o fim do M3: descoberta, histórico e alertas. A v0.1.0 foi publicada. Os próximos passos são o **M5.3** e o **M4.2**.

## Depois

- Integração opcional com nmap e nuclei para varreduras mais profundas.
- Verificação de exposição pela internet (precisa de um ponto externo).
- Segurança de Wi-Fi (ponto de acesso falso), somente em Linux com adaptador compatível.
- Vários locais num só painel.
- Integrações com roteadores (OpenWrt, UniFi, MikroTik) e app instalável (PWA).

## Fora do escopo

- Bloquear dispositivos por ARP spoofing, desautenticação de Wi-Fi, exploração de falhas e ataque de força bruta: são técnicas ofensivas.
- Varrer redes que não são suas.
- Guardar dados na nuvem do projeto.
