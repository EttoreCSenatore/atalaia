---
id: comparacao-kaladrius
title: Comparação com o Kaladrius
sidebar_position: 9
---

# O que aprender com o Kaladrius

O [Kaladrius](https://github.com/facchini-miguel/kaladrius) (Apache 2.0) é um sistema de gestão para clínicas, em Django. O **produto** não tem nada a ver com o Atalaia, mas a forma como ele é **operado e mantido** está mais madura que a nossa. Esta página compara os dois (em novembro de 2026) e lista o que vale trazer. Trazemos **ideias**, não código.

## Comparação

| Tema | Kaladrius | Atalaia hoje | Trazer? |
|---|---|---|---|
| Dependências (Dependabot) | Semanal, **espera de 7 dias** antes de propor uma versão nova, PRs agrupados | Semanal, sem espera, sem grupos | **Feito** (M7.1) |
| Ações do GitHub e imagens | Fixadas por **SHA** e por **digest** | Por etiqueta (`@v7`, `node:24-alpine`) | **Sim** (M7.1) |
| Arquivos do repositório | SECURITY, CONTRIBUTING, CODE_OF_CONDUCT, CODEOWNERS, NOTICE | Só SECURITY | **Feito** (M7.2) |
| Versões e novidades | `releases.yaml` único alimenta a página "Novidades" no app, a versão no rodapé e a release do GitHub | CHANGELOG em texto; versão só na tela Sobre | **Feito** (M7.3) |
| Saúde e monitoramento | Página de saúde sem login, sinais para Healthchecks.io (sistema, backup, restauração, atualização, disco), números para Prometheus | Nada sem login; nenhum aviso se o Atalaia parar | **Feito** (M7.4) |
| Backup | Diário por timer do systemd, **teste de restauração semanal**, aviso quando atrasa | `atalaia backup` manual | **Feito** (M7.5) |
| Migrações | CI testa subir e descer cada migração | Migrações só para frente; sem teste a partir de bancos de versões antigas | **Adaptado** (M7.6): testar a atualização a partir do banco de cada versão publicada e fazer backup automático antes de migrar |
| Pacotes | `.deb`/`.rpm` com scripts de instalação e remoção, teste de instalação na CI, **repositório de pacotes** para atualizações | `.deb`/`.rpm` sem scripts nem teste de instalação | **Sim** (M7.7); o repositório de pacotes fica para depois |
| Atualização automática | Timer a cada 5 min com **rollback automático** se a saúde falhar | Nenhuma | **Não por ora**: num pacote, quem atualiza é o sistema (apt/dnf). Volta a ser discutido com o repositório de pacotes |
| Documentação | Manual do usuário, operação (backup, monitoramento, atualização), registro de decisões, modelo de dados, FAQ | Boa parte técnica; sem manual por tarefas, sem registro de decisões | **Sim** (M7.8) |
| Demonstração pública | Servidor de demonstração reiniciado toda noite | Só capturas de tela | **Adaptar** (M7.9): o app web em modo demonstração, com dados fictícios, no próprio GitHub Pages, sem servidor |
| Idiomas | pt-BR, inglês e espanhol | pt-BR e inglês | **Talvez** (M7.10) |
| Pilha, nuvem do Google, financiamento | Django/HTMX, guia do Google Cloud, FUNDING | Go + Svelte, auto-hospedado | **Não** se aplica |

## O que o Atalaia já faz melhor (e mantém)

- Um binário estático, sem banco externo; imagem de ~26 MB sem privilégios.
- Análise estática de segurança na CI (gosec, staticcheck, govulncheck, Trivy) e uma autorrevisão documentada.
- Ferramentas ativas cercadas (escopo, confirmação, auditoria, modo somente observação).

O roteiro com os passos está no [roadmap](../roadmap.md#m7-operação-e-maturidade).
