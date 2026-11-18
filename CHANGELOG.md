# Changelog / Histórico de mudanças

## Unreleased

**EN** — M6.3 packaging: `.deb` and `.rpm` packages (amd64, arm64, armv7) in every release, with a hardened systemd unit (DynamicUser, no capabilities, read-only file system, syscall filter, IP sockets only; sandbox tested with the real binary) and `/etc/default/atalaia` options; observe-only mode (`--no-active-tools`) that disables ping, Wake-on-LAN, port scan and exposure check; `ATALAIA_LISTEN` is now read by the binary.

**PT-BR** — Pacotes (M6.3): `.deb` e `.rpm` (amd64, arm64, armv7) em toda versão, com unidade systemd endurecida (DynamicUser, sem capabilities, sistema de arquivos somente leitura, filtro de chamadas de sistema, só sockets IP; sandbox testado com o binário real) e opções em `/etc/default/atalaia`; modo somente observação (`--no-active-tools`) que desliga ping, Wake-on-LAN, varredura de portas e verificação de exposição; `ATALAIA_LISTEN` agora é lido pelo binário.

**EN** — M6.2 security self-review: gosec added to CI (G304/G117 accepted, documented); fixes for integer conversions (ports, argon2 key size, route parsing), the logout cookie's `Secure` flag, and a new `--secure-cookies` for TLS-terminating proxies; the review's findings and accepted risks are documented in the security page.

**PT-BR** — Autorrevisão de segurança (M6.2): gosec na CI (G304/G117 aceitos e documentados); correções em conversões de inteiros (portas, tamanho da chave argon2, leitura da tabela de rotas), na flag `Secure` do cookie de logout e novo `--secure-cookies` para proxies que terminam TLS; achados e riscos aceitos documentados na página de segurança.

**EN** — M6.1 built-in HTTPS: `--tls-self-signed` (ECDSA P-256 certificate made, reused and renewed by Atalaia, fingerprint printed) or `--tls-cert`/`--tls-key` (picked up on change, HSTS enabled); TLS 1.2+, `Secure` session cookie; HSTS never with the self-signed certificate; env vars for Docker; bilingual docs.

**PT-BR** — HTTPS embutido (M6.1): `--tls-self-signed` (certificado ECDSA P-256 criado, reaproveitado e renovado pelo Atalaia, com impressão digital no console) ou `--tls-cert`/`--tls-key` (recarregado quando muda, HSTS ligado); TLS 1.2+, cookie de sessão `Secure`; nunca HSTS com o certificado autoassinado; variáveis de ambiente para Docker; documentação bilíngue.

**EN** — Fix found while testing on a real network: the machine running Atalaia was missing from its own inventory (the ARP table never lists the host itself); it is now added, using only the interface on the gateway's network.

**PT-BR** — Correção achada testando numa rede real: a máquina que roda o Atalaia não aparecia no próprio inventário (a tabela ARP nunca lista o próprio host); agora ela é incluída, usando só a interface da rede do roteador.

## v0.3.0 — 2026-11-18

**EN** — M5.2 exposure check: from a device's last port scan, flags risky open services (rule table, 3 levels), reads what web ports answer (status, Server, title, login) with one non-redirecting GET, and audits TLS (version, certificate validity, self-signed, legacy TLS 1.0/1.1, weak RSA). Same fences as the port scan (confirmation, local scope, one at a time, interval, audit log). Bilingual docs.

**PT-BR** — Verificação de exposição (M5.2): a partir da última varredura de portas do dispositivo, aponta serviços arriscados abertos (tabela de regras, 3 níveis), lê o que as portas web respondem (status, Server, título, login) com um GET sem seguir redirecionamentos, e audita o TLS (versão, validade do certificado, autoassinado, TLS 1.0/1.1, RSA fraco). Mesmas cercas da varredura de portas (confirmação, escopo local, uma por vez, intervalo, auditoria). Documentação bilíngue.

## v0.2.0 — 2026-11-18

**EN** — M5.1 tools: ping (unprivileged ICMP), DNS lookup, Wake-on-LAN and a TCP port scan (~65 ports, handshake only) with a diff against the previous scan. Active actions are fenced: administrator only, one at a time, inventory devices in the local scope only (no free target), explicit confirmation and a 10 s minimum interval for scans, and an audit log purged with the retention. New Tools page, device actions, bilingual docs.

**PT-BR** — Ferramentas (M5.1): ping (ICMP sem privilégios), consulta DNS, Wake-on-LAN e varredura de portas TCP (~65 portas, só o handshake) com comparação com a anterior. As ações ativas têm cercas: só o administrador, uma por vez, só dispositivos do inventário no escopo local (sem alvo livre), confirmação explícita e intervalo mínimo de 10 s nas varreduras, e registro de auditoria apagado junto com a retenção. Nova página Ferramentas, ações no dispositivo, documentação bilíngue.

**EN** — M4.1 passive security: rootless detection of two ARP-spoofing signs on every scan (the default gateway's address answering from a new MAC; one MAC answering for 3+ addresses), shown in red in the history, with a default-on alert rule that also fires on the first scan; known devices are exempt from the multi-address check; one observation per MAC in scans (no more false "address changed" events for multi-address devices); bilingual docs with documented limits.

**PT-BR** — Segurança passiva (M4.1): detecção sem privilégios de dois sinais de ARP spoofing a cada varredura (o endereço do gateway respondendo com outro MAC; um MAC respondendo por 3+ endereços), destacados em vermelho no histórico, com regra de alerta ligada por padrão que vale também na primeira varredura; dispositivos conhecidos ficam isentos da checagem de vários endereços; uma observação por MAC nas varreduras (sem falsos eventos de "mudou de endereço"); documentação bilíngue com os limites.

## v0.1.0 — 2026-11-18

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
