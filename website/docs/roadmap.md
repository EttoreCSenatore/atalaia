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
| **M4.2** | Servidor DHCP falso (opcional, exige a porta 68) e alerta de portas recém-abertas | Concluído (ver [Segurança passiva](funcionalidades/seguranca-passiva.md)) |
| **M5.1** | Ferramentas: ping, DNS, Wake-on-LAN e varredura de portas com confirmação, limites e registro de auditoria | Concluído (ver [Ferramentas](funcionalidades/ferramentas.md)) |
| **M5.2** | Segurança ativa: serviços arriscados, painéis web expostos e auditoria de TLS | Concluído (ver [Ferramentas](funcionalidades/ferramentas.md)) |
| **M5.3** | Traceroute e teste de velocidade (com aviso de privacidade). CVEs: decidido **não** fazer por ora, com o motivo na documentação | Concluído (ver [Ferramentas](funcionalidades/ferramentas.md)) |
| **M6.1** | HTTPS embutido (certificado próprio ou autoassinado com renovação) | Concluído (ver [Implantação](tecnico/implantacao.md)) |
| **M6.2** | Autorrevisão de segurança do código e gosec na CI | Concluído (ver [Segurança e privacidade](tecnico/seguranca-e-privacidade.md)) |
| **M6.3** | Unidade systemd, pacotes .deb/.rpm e modo somente observação | Concluído (ver [Implantação](tecnico/implantacao.md)) |
| **v0.5** | Identificação: serviços mDNS, SSDP/UPnP, NetBIOS, palpite de tipo com ícones, nomes por DHCP (opcional), VPN por `--allow-cidr`, varreduras agendadas | Concluído (ver [Descoberta](funcionalidades/descoberta.md)) |
| **v0.6** | Revisão de segurança (bloqueio de login sob tentativas paralelas, limite global) e organização do inventário (esquecer, limpeza automática, filtro, exportação, backup) | Concluído (ver [Segurança e privacidade](tecnico/seguranca-e-privacidade.md)) |
| **v0.7** | Operação e maturidade (M7.1–M7.9): cadeia de suprimentos fixada, novidades no app, saúde e métricas, backups automáticos com teste de restauração, atualização segura, pacotes testados, manual e demonstração | Concluído (ver [Primeiros passos](manual/primeiros-passos.md)) |

A versão publicada mais recente é a **v0.7.1** (correções da revisão de segurança da v0.7), com o bloco **M7** (o M7.10, espanhol, não será feito).

## M7: Operação e maturidade

Inspirado no Kaladrius (veja a [comparação](tecnico/comparacao-kaladrius.md)). Um pull request por passo, nesta ordem; cada um tem uma issue no marco **M7** do GitHub.

| Passo | Entrega | Estado |
|---|---|---|
| **M7.1** | Cadeia de suprimentos: Dependabot com espera de 7 dias e PRs agrupados; ações do GitHub fixadas por SHA e imagens por digest, com verificação na CI | Concluído (ver [Cadeia de suprimentos](tecnico/cadeia-de-suprimentos.md)) |
| **M7.2** | Arquivos do repositório: CONTRIBUTING, CODE_OF_CONDUCT, CODEOWNERS, NOTICE (nome, marca, sem ligação com a Fing) e modelos de issue e PR | Concluído |
| **M7.3** | Versões e novidades: `releases.yaml` como fonte única da página **Novidades** no app, da versão no rodapé e do texto da release; a CI confere a versão | Concluído (ver [Versões e novidades](tecnico/versionamento.md)) |
| **M7.4** | Saúde e monitoramento: `GET /api/health` sem login (versão e banco), sinais opcionais para Healthchecks.io (varredura, backup), `/metrics` opcional para Prometheus e aviso de disco quase cheio | Concluído (ver [Monitoramento](tecnico/monitoramento.md)) |
| **M7.5** | Backup agendado: backup diário dentro do `atalaia serve` (Docker, pacotes e binário) com rotação, teste de restauração semanal registrado e aviso quando o backup atrasa | Concluído (ver [Implantação](tecnico/implantacao.md)) |
| **M7.6** | Atualização segura: backup automático antes de migrar e teste na CI que atualiza bancos criados por cada versão publicada | Concluído (ver [Versões e novidades](tecnico/versionamento.md)) |
| **M7.7** | Ciclo de vida do pacote: scripts de instalação e remoção, teste de instalação do `.deb` e do `.rpm` na CI | Concluído (ver [Implantação](tecnico/implantacao.md)) |
| **M7.8** | Documentação: manual por tarefas (primeiros passos, FAQ, glossário), seção de operação (backup, monitoramento, atualização), registro de decisões e referência única de configuração | Concluído (ver [Primeiros passos](manual/primeiros-passos.md)) |
| **M7.9** | Demonstração pública: o app web em modo demonstração, com dados fictícios, no GitHub Pages | Concluído ([abrir a demonstração](https://ettorecsenatore.github.io/atalaia/demo/)) |
| **M7.10** | (Opcional) Tradução para o espanhol | Não será feito: o projeto fica em português e inglês |

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
