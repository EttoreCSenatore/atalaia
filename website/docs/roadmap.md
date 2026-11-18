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
| **M4** | Segurança passiva: ARP spoofing, DHCP falso, serviços alterados | Planejado |
| **M5** | Segurança ativa e ferramentas: portas, CVEs, painéis expostos, TLS, ping, traceroute, DNS, Wake-on-LAN, velocidade | Planejado |
| **M6** | Endurecimento, autorrevisão, versão v0.2 e pacotes (Linux, macOS, Windows, ARM) | Planejado |

A **v0.1** é o fim do M3: descoberta, histórico e alertas. O próximo passo é publicar a **v0.1.0** e seguir para o **M4**.

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
