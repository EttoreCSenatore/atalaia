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
| **M1** | Descoberta: ARP e mDNS, lista de dispositivos com fabricante e nome, SQLite, CLI | Planejado |
| **M2** | App web embutido: painel, detalhe do dispositivo, linha do tempo, login e imagem Docker | Planejado |
| **M3** | Alertas: dispositivo novo, adaptadores de notificação, nomes, etiquetas e dispositivos conhecidos | Planejado |
| **M4** | Segurança passiva: ARP spoofing, DHCP falso, serviços alterados | Planejado |
| **M5** | Segurança ativa e ferramentas: portas, CVEs, painéis expostos, TLS, ping, traceroute, DNS, Wake-on-LAN, velocidade | Planejado |
| **M6** | Endurecimento, autorrevisão, versão v0.2 e pacotes (Linux, macOS, Windows, ARM) | Planejado |

A **v0.1** é o fim do M3: descoberta, histórico e alertas.

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
