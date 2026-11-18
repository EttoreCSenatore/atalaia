# Changelog / Histórico de mudanças

## Unreleased

**EN** — M4.1 passive security: rootless detection of two ARP-spoofing signs on every scan (the default gateway's address answering from a new MAC; one MAC answering for 3+ addresses), shown in red in the history, with a default-on alert rule that also fires on the first scan; known devices are exempt from the multi-address check; one observation per MAC in scans (no more false "address changed" events for multi-address devices); bilingual docs with documented limits.

**PT-BR** — Segurança passiva (M4.1): detecção sem privilégios de dois sinais de ARP spoofing a cada varredura (o endereço do gateway respondendo com outro MAC; um MAC respondendo por 3+ endereços), destacados em vermelho no histórico, com regra de alerta ligada por padrão que vale também na primeira varredura; dispositivos conhecidos ficam isentos da checagem de vários endereços; uma observação por MAC nas varreduras (sem falsos eventos de "mudou de endereço"); documentação bilíngue com os limites.

**EN** — Release automation: pushing a `vX.Y.Z` tag builds Linux binaries (amd64, arm64, armv7) with checksums, publishes a multi-arch image to GHCR (SBOM + provenance) and creates the GitHub release.

**PT-BR** — Automação de versões: ao enviar uma tag `vX.Y.Z`, o CI gera binários Linux (amd64, arm64, armv7) com checksums, publica uma imagem multi-arquitetura no GHCR (SBOM + proveniência) e cria a release no GitHub.

**EN** — M3 alerts: new-device and known-device-offline alerts through ntfy, Telegram, webhook and e-mail (STARTTLS, never sends a password without TLS); per-device name, tags and "known" flag; alert settings and a send-test button in the web app; the first scan never alerts; secrets stay server-side and are never returned by the API; bilingual docs with screenshots.

**PT-BR** — Alertas (M3): avisos de dispositivo novo e de dispositivo conhecido offline por ntfy, Telegram, webhook e e-mail (STARTTLS, nunca envia senha sem TLS); nome, etiquetas e marca "conhecido" por dispositivo; configuração de alertas e botão de teste no app web; a primeira varredura nunca alerta; segredos ficam no servidor e a API nunca os devolve; documentação bilíngue com capturas de tela.

**EN** — M2.2 web app: Svelte 5 UI embedded in the binary (setup, login, dashboard with search and live updates, device page, history, settings; pt-BR and English; light/dark), hardened ~26 MB Docker image + compose (host network, read-only, no capabilities), CI jobs for the app and the image (Trivy, smoke test), bilingual docs with screenshots.

**PT-BR** — App web (M2.2): interface Svelte 5 embutida no binário (configuração inicial, login, painel com busca e atualização ao vivo, página do dispositivo, histórico, configurações; pt-BR e inglês; claro/escuro), imagem Docker de ~26 MB e compose endurecidos (rede do host, somente leitura, sem capabilities), jobs de CI do app e da imagem (Trivy, teste de fumaça), documentação bilíngue com capturas de tela.

**EN** — M2.1 server: `atalaia serve` runs scheduled background scans and an HTTP API with one-time-code first-run setup, argon2id login, hashed server-side sessions, lockout, local-network-only + Host + CSRF protections, live events (SSE), event retention.

**PT-BR** — Servidor (M2.1): `atalaia serve` roda varreduras agendadas em segundo plano e uma API HTTP com configuração inicial por código de uso único, login argon2id, sessões com hash no servidor, bloqueio de tentativas, proteções (só rede local, Host, CSRF), eventos ao vivo (SSE) e retenção de eventos.

**EN** — M1 discovery: `atalaia scan|devices|history` (Linux, IPv4): UDP sweep + system ARP table, names by mDNS and reverse DNS, vendor from the embedded IEEE registry (random MACs flagged), SQLite store with joined/online/offline/ip_changed events and a grace period, sanitizing of text from the network, bilingual docs page.

**PT-BR** — Descoberta (M1): `atalaia scan|devices|history` (Linux, IPv4): varredura UDP + tabela ARP do sistema, nomes por mDNS e DNS reverso, fabricante pela lista do IEEE embutida (MAC aleatório sinalizado), banco SQLite com eventos joined/online/offline/ip_changed e prazo de tolerância, limpeza de textos vindos da rede, página de documentação bilíngue.

**EN** — M0 foundation: repository, AGPL-3.0 license, bilingual Docusaurus documentation on GitHub Pages, CI, Go skeleton with the scan-scope guard (`atalaia scope`).

**PT-BR** — Fundação (M0): repositório, licença AGPL-3.0, documentação bilíngue em Docusaurus no GitHub Pages, CI, esqueleto em Go com o guarda de escopo de varredura (`atalaia scope`).
