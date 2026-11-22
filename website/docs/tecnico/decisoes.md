---
id: decisoes
title: Registro de decisões
sidebar_position: 10
---

# Registro de decisões

As escolhas de projeto que moldam o Atalaia, com o motivo de cada uma. Uma decisão nova (ou revista) entra aqui no mesmo PR que a implementa.

| # | Decisão | Por quê | Desde |
|---|---|---|---|
| 1 | Licença **AGPL-3.0-or-later** (código) e CC BY-SA 4.0 (documentação) | Quem oferecer uma versão modificada como serviço precisa abrir o código; o app mostra o link do código-fonte | M0 |
| 2 | **Go, um binário só, SQLite embutido** (sem cgo) | Instala em qualquer Linux, até num Raspberry Pi, sem banco externo nem dependências | M0, M1 |
| 3 | **Só Linux** | A descoberta usa a tabela ARP do kernel e sockets sem privilégio do Linux | M1 |
| 4 | **Descoberta sem root**: varredura UDP + tabela ARP, sem sockets brutos nem `CAP_NET_RAW` | Menos privilégio, menos risco; a tabela ARP já tem o que precisamos | M1 |
| 5 | **Escopo fixo**: só as redes privadas IPv4 da própria máquina; ferramentas só contra dispositivos do inventário | Impede usar o Atalaia como scanner contra terceiros | M1, M5.1 |
| 6 | **Sem IPv6** por enquanto | A descoberta por vizinhança IPv6 é outra técnica, com endereços temporários que mudam; o ganho numa rede doméstica não compensa agora | v0.5 |
| 7 | **Sem verificação de CVEs** | Adivinhar a versão do software pela rede é pouco confiável; alerta falso de CVE assusta mais do que ajuda. Mostramos o que é verificável | v0.4 |
| 8 | **Servidor só para a rede local**, checagem de `Host` e proteção CSRF | O app vigia a rede local; abrir para a internet não é um caso de uso. VPN entra por `--allow-cidr` | M2.1, v0.5 |
| 9 | **Uma conta de administrador**, argon2id, código de configuração de uso único | Casa e pequeno escritório têm um responsável; o código impede que o primeiro a chegar tome o servidor | M2.1 |
| 10 | **Ações ativas cercadas**: confirmação, uma por vez, intervalo mínimo, registro de auditoria; **modo somente observação** | Separar observar de agir deixa claro quando o Atalaia toca outro aparelho | M5.1, M6.3 |
| 11 | **HSTS só com o seu próprio certificado** | Com o autoassinado, HSTS impediria aceitar o aviso do navegador e trancaria você fora | M6.1 |
| 12 | **Verificação de DHCP e aprendizado de nomes desligados por padrão** | Precisam de porta privilegiada (UDP 67/68); quem quiser liga e dá a permissão | M4.2, v0.5 |
| 13 | **Teste de velocidade opcional, com confirmação a cada uso** | Ele revela o seu IP público à Cloudflare | M5.3 |
| 14 | **Sem telemetria**; alertas saem direto da máquina para o serviço escolhido | Os dados da sua rede são seus | M3 |
| 15 | **Segredos dos alertas nunca voltam pela API** | Uma sessão roubada não vaza tokens e senhas | M3 |
| 16 | Docker com **rede do host** | Sem ela o contêiner só vê a rede virtual do Docker | M2.2 |
| 17 | Pacotes com **usuário descartável** (`DynamicUser`) e unidade endurecida; **nunca ligam o serviço sozinhos nem apagam dados** | Menos privilégio; instalar não deve mudar o que roda, e o histórico é do usuário | M6.3, M7.7 |
| 18 | **Backups dentro do `atalaia serve`**, não num timer do systemd | Funciona igual no Docker, nos pacotes e no binário | M7.5 |
| 19 | **Sem atualização automática** | Pacotes são atualizados pelo apt/dnf; rever se houver um repositório de pacotes | M7 |
| 20 | **Cadeia de suprimentos fixada**: ações por SHA, imagens por digest, Dependabot com 7 dias de espera | Uma etiqueta pode passar a apontar para código malicioso; pacotes comprometidos costumam ser retirados em dias | M7.1 |
| 21 | **`releases.yaml` como fonte única** das novidades | Versão, novidades no app e texto da release nunca divergem | M7.3 |
| 22 | **Saúde sem login, sem dados da rede; `/metrics` desligado e com token** | Monitores precisam de um endereço simples; contagens bastam e nada identifica aparelhos | M7.4 |
| 23 | **Migrações só para frente e só crescem**, com backup antes | Descer migrações exige código que quase nunca roda; uma cópia antes é mais confiável | M7.6 |
| 24 | **Documentação bilíngue**, português primeiro | O projeto nasceu no Brasil; inglês alcança o resto | M0 |
| 25 | **Nada de dados de rede reais no repositório**: capturas e testes usam dados fictícios | Privacidade de quem testa e de quem aparece na rede | M2.2 |
