# Changelog / Histórico de mudanças

## Unreleased

**EN** — Installable app (PWA): web app manifest, icons (with a maskable one) and a hand-written service worker that keeps only the app shell, so the installed app opens without the server and says it cannot reach it; it never answers or caches `/api/` or `/metrics`. Registered only in a secure context and never in the demo. Docs: "Install as an app", FAQ, decision log row 28.

**PT-BR** — App instalável (PWA): manifesto do app web, ícones (com um "maskable") e um service worker escrito à mão que guarda só o casco do app, para o app instalado abrir sem o servidor e dizer que não o alcança; ele nunca responde nem guarda `/api/` ou `/metrics`. Só é registrado em contexto seguro e nunca na demonstração. Documentação: "Instalar como app", perguntas frequentes, linha 28 do registro de decisões.

**EN** — Upgrade-test database for v0.7.1 added.

**PT-BR** — Banco de teste de atualização da v0.7.1 incluído.

## v0.7.1 — 2026-11-22

**EN** — Security review of v0.7: automatic backups refuse a folder other users can write to, and a backup is never written through a symlink planted under its name; databases open with SQLite's `trusted_schema` off (`verify-backup` may open a foreign file); `/metrics` tokens must also have at least 10 different characters; old scratch copies are cleaned only in the automatic backups' own folder. Scanners (govulncheck, gosec, staticcheck, npm audit, Trivy) found nothing. Findings and accepted risks are in the security page.

**PT-BR** — Revisão de segurança da v0.7: os backups automáticos recusam uma pasta gravável por outros usuários, e um backup nunca é gravado através de um link simbólico plantado com o nome dele; os bancos abrem com o `trusted_schema` do SQLite desligado (o `verify-backup` pode abrir um arquivo de fora); o token de `/metrics` também precisa de ao menos 10 caracteres diferentes; cópias de rascunho velhas só são limpas na própria pasta dos backups automáticos. As ferramentas (govulncheck, gosec, staticcheck, npm audit, Trivy) não acharam nada. Achados e riscos aceitos estão na página de segurança.

**EN** — Decision: no Spanish translation (M7.10 dropped); the project stays in Portuguese and English (decision log row 27).

**PT-BR** — Decisão: sem tradução para o espanhol (M7.10 descartado); o projeto fica em português e inglês (registro de decisões, linha 27).

**EN** — Packages: a Debian upgrade no longer stops unattended when `/etc/default/atalaia` was edited and the new version also changed it (the test now keeps your file, as unattended upgrades do; the docs explain the `apt` prompt). The release workflow now tests the new `.deb`/`.rpm` as an upgrade from the previous release in Debian 13 and Fedora 44 before publishing, and CI's test packages change the options file like real releases. The v0.6.0 → v0.7.0 upgrade of the published packages was checked this way. Upgrade-test database for v0.7.0 added.

**PT-BR** — Pacotes: uma atualização no Debian não para mais sem resposta quando o `/etc/default/atalaia` foi editado e a versão nova também o mudou (o teste agora mantém o seu arquivo, como as atualizações automáticas; a documentação explica a pergunta do `apt`). O fluxo de release agora testa o `.deb`/`.rpm` novo como atualização a partir da release anterior no Debian 13 e no Fedora 44 antes de publicar, e os pacotes de teste da CI mudam o arquivo de opções como as releases de verdade. A atualização v0.6.0 → v0.7.0 dos pacotes publicados foi conferida assim. Banco de teste de atualização da v0.7.0 incluído.

## v0.7.0 — 2026-11-22

**EN** — Docs: comparison with Kaladrius and the M7 "Operations and maturity" roadmap; roadmap brought up to date (v0.5, v0.6).

**PT-BR** — Documentação: comparação com o Kaladrius e o roteiro M7 "Operação e maturidade"; roadmap atualizado (v0.5, v0.6).

**EN** — Supply chain (M7.1): GitHub actions pinned by commit SHA and Docker images by digest, checked in CI by `scripts/checar-fixacao.py`; Dependabot waits 7 days before proposing a version, groups minor/patch updates and limits open PRs. New docs page "Supply chain".

**PT-BR** — Cadeia de suprimentos (M7.1): ações do GitHub fixadas por SHA de commit e imagens Docker por digest, conferidas na CI por `scripts/checar-fixacao.py`; o Dependabot espera 7 dias antes de propor uma versão, agrupa as atualizações menores e de correção e limita os PRs abertos. Nova página "Cadeia de suprimentos".

**EN** — Repository files (M7.2): CONTRIBUTING and CODE_OF_CONDUCT (bilingual), NOTICE (name use, no affiliation with Fing), CODEOWNERS, issue forms (bug, idea) and a PR checklist; README status brought up to v0.6.

**PT-BR** — Arquivos do repositório (M7.2): CONTRIBUTING e CODE_OF_CONDUCT (bilíngues), NOTICE (uso do nome, sem ligação com a Fing), CODEOWNERS, formulários de issue (bug, ideia) e lista de verificação no PR; estado do README atualizado para v0.6.

**EN** — Build: Node 26 (LTS) for the image, CI and release; Docker release actions moved to their new majors (login v4, setup-qemu v4, setup-buildx v4, build-push v7), tested with a non-publishing multi-arch build.

**PT-BR** — Build: Node 26 (LTS) na imagem, na CI e na release; ações Docker da release nas novas versões maiores (login v4, setup-qemu v4, setup-buildx v4, build-push v7), testadas com um build multiarquitetura sem publicação.

**EN** — What's new (M7.3): `releases.yaml` is the single source for a **What's new** page in the web app (with a notice after an update and the running version in the footer and Settings), `GET /api/releases` and the GitHub release text. CI checks the file and that it matches CHANGELOG.md; a release only publishes when its tag is the newest entry. New dependency: go.yaml.in/yaml/v3 (MIT). New docs page "Versions and what's new".

**PT-BR** — Novidades (M7.3): `releases.yaml` é a fonte única da página **Novidades** no app web (com aviso depois de uma atualização e a versão em uso no rodapé e em Configurações), do `GET /api/releases` e do texto da release no GitHub. A CI confere o arquivo e que ele bate com o CHANGELOG.md; uma release só é publicada se a tag for a versão mais nova de lá. Nova dependência: go.yaml.in/yaml/v3 (MIT). Nova página "Versões e novidades".

**EN** — Health and monitoring (M7.4): `GET /api/health` without login for uptime monitors (status, version, database, last scan; HTTP 503 when the database fails, the last scan failed or scans stopped; never network data); `atalaia health` and a Docker `HEALTHCHECK`; Healthchecks.io-style heartbeats after every scan (`--heartbeat-scan`) and backup (`atalaia backup --heartbeat`), with `/fail` on failure and the secret URL never logged; optional Prometheus `/metrics` behind a bearer token (file, systemd credential or env; 32+ characters) with counts, scan and backup times and disk space; a notice in the app when the database's disk is almost full (under 512 MB or 5%). New docs page "Monitoring".

**PT-BR** — Saúde e monitoramento (M7.4): `GET /api/health` sem login para monitores (estado, versão, banco, última varredura; HTTP 503 quando o banco falha, a última varredura falhou ou as varreduras pararam; nunca dados da rede); `atalaia health` e `HEALTHCHECK` no Docker; sinais no estilo Healthchecks.io depois de cada varredura (`--heartbeat-scan`) e backup (`atalaia backup --heartbeat`), com `/fail` na falha e a URL secreta nunca no log; `/metrics` opcional para o Prometheus com token (arquivo, credencial do systemd ou variável; 32+ caracteres) com contagens, horários de varredura e backup e espaço em disco; aviso no app quando o disco do banco está quase cheio (menos de 512 MB ou 5%). Nova página "Monitoramento".

**EN** — Automatic backups (M7.5): `atalaia serve` copies the database once a day into `backups/` next to it (`--backup-dir`, `--backup-keep` 7, `--backup-interval` 24h; on by default, `--backup-keep 0` turns it off), catches up after downtime, and test-restores the newest copy every week (scratch copy, opened with migrations, SQLite integrity check), recording the result. Settings shows the last backup and restore test with a **Back up now** button (`POST /api/backup`, audited); a notice appears when the last backup is more than 26 hours old; health warns `backup_late` / `restore_test_failed`; new restore-test metrics; `--heartbeat-backup`. New `atalaia backup --dir FOLDER --keep N` and `atalaia verify-backup FILE | --dir FOLDER`. Built into the server instead of a systemd timer, so it works the same with Docker, the packages and the binary.

**PT-BR** — Backups automáticos (M7.5): o `atalaia serve` copia o banco uma vez por dia para `backups/` ao lado dele (`--backup-dir`, `--backup-keep` 7, `--backup-interval` 24h; ligado por padrão, `--backup-keep 0` desliga), recupera o atraso depois de ficar parado e testa a restauração da cópia mais nova toda semana (cópia de rascunho, aberta com as migrações, verificação de integridade do SQLite), registrando o resultado. Configurações mostra o último backup e o último teste, com o botão **Fazer backup agora** (`POST /api/backup`, auditado); um aviso aparece quando o último backup tem mais de 26 horas; a saúde avisa `backup_late` / `restore_test_failed`; novas métricas do teste de restauração; `--heartbeat-backup`. Novos `atalaia backup --dir PASTA --keep N` e `atalaia verify-backup ARQUIVO | --dir PASTA`. Feito dentro do servidor em vez de um timer do systemd, para funcionar igual no Docker, nos pacotes e no binário.

**EN** — Safe upgrades (M7.6): before migrating a database from an earlier version, Atalaia copies it to `atalaia-pre-migration-schemaN-DATE.db` next to it (0600, newest 3 kept) and refuses to migrate if the copy fails. Released migrations are now append-only (hash list checked by a test). New upgrade test: a database made by the binary of every published version (v0.1.0–v0.6.0, generated by `scripts/gerar-bancos-de-versoes.py` on a throwaway Docker network with a fictitious device, binaries checked against SHA256SUMS) is opened by the current code and must reach the latest schema, pass `integrity_check`, keep its devices and events and take a new scan.

**PT-BR** — Atualização segura (M7.6): antes de migrar um banco de uma versão anterior, o Atalaia o copia para `atalaia-pre-migration-schemaN-DATA.db` ao lado dele (0600, as 3 mais novas ficam) e se recusa a migrar se a cópia falhar. As migrações publicadas agora só crescem (lista de hashes conferida por teste). Novo teste de atualização: um banco criado pelo binário de cada versão publicada (v0.1.0–v0.6.0, gerados por `scripts/gerar-bancos-de-versoes.py` numa rede Docker descartável com um dispositivo fictício, binários conferidos pelo SHA256SUMS) é aberto pelo código atual e precisa chegar ao esquema mais novo, passar no `integrity_check`, manter dispositivos e eventos e aceitar uma varredura nova.

**EN** — Packages (M7.7): `.deb`/`.rpm` install scripts that reload systemd, never enable or start the service by themselves, restart it on upgrade only if it was running, stop and disable it on removal, and never delete the data (not even on purge). New CI job installs the packages in clean Debian 13 and Fedora 44 containers: install, file modes, `systemd-analyze verify`, upgrade keeping a changed `/etc/default/atalaia`, removal and purge keeping the data (`scripts/testar-pacotes.sh`).

**PT-BR** — Pacotes (M7.7): scripts de instalação do `.deb`/`.rpm` que recarregam o systemd, nunca ligam nem iniciam o serviço sozinhos, reiniciam na atualização só se ele já rodava, param e desligam na remoção e nunca apagam os dados (nem no purge). Novo job da CI instala os pacotes em contêineres limpos de Debian 13 e Fedora 44: instalação, permissões, `systemd-analyze verify`, atualização mantendo um `/etc/default/atalaia` alterado, remoção e purge mantendo os dados (`scripts/testar-pacotes.sh`).

**EN** — Documentation (M7.8): a task-based **Manual** (getting started, how-to, FAQ, glossary), an **Operations** section (one configuration reference for every command, option, variable, file and port; backup and restore with a step-by-step restore; upgrading and going back) and a **decision log** of 25 design choices. CI checks that the configuration reference (pt-BR and English) names every command, option and `ATALAIA_` variable in the code. New `atalaia reset-password --yes` for a forgotten password: removes the account (devices, history and settings stay), so the next start prints a new setup code. The "current state" on the home page now says v0.6.

**PT-BR** — Documentação (M7.8): **Manual** por tarefas (primeiros passos, como fazer, perguntas frequentes, glossário), seção de **Operação** (referência única de configuração com todo comando, opção, variável, arquivo e porta; backup e restauração com o passo a passo da restauração; atualização e como voltar atrás) e **registro de decisões** com 25 escolhas de projeto. A CI confere que a referência de configuração (pt-BR e inglês) cita todo comando, opção e variável `ATALAIA_` do código. Novo `atalaia reset-password --yes` para senha esquecida: remove a conta (dispositivos, histórico e configurações ficam), e o próximo início mostra um código de configuração novo. O "estado atual" da página inicial agora diz v0.6.

**EN** — Public demo (M7.9): the web app runs at https://ettorecsenatore.github.io/atalaia/demo/ with fictitious data (the screenshot fixtures), with a "demo" banner; the API is answered in the browser, routes use the hash, and actions that would touch a network or send something are refused. Nothing is sent anywhere. CI builds it into the Pages site and checks that the real app's build carries none of the demo code or data. "Demo" link in the docs menu.

**PT-BR** — Demonstração pública (M7.9): o app web roda em https://ettorecsenatore.github.io/atalaia/demo/ com dados fictícios (os das capturas de tela) e uma faixa de "demonstração"; a API é respondida no navegador, as rotas usam o `#` e as ações que tocariam uma rede ou enviariam algo são recusadas. Nada é enviado a lugar nenhum. A CI monta a demonstração no site do Pages e confere que o build do app real não leva nada do código nem dos dados dela. Link "Demonstração" no menu da documentação.

## v0.6.0 — 2026-11-19

**EN** — Inventory housekeeping: forget a device (with its history), opt-in forgetting of untouched devices offline for N days, a type filter on the dashboard, CSV/JSON export (spreadsheet-safe CSV) from the dashboard and `atalaia export`, and `atalaia backup` for a consistent 0600 copy while the server runs.

**PT-BR** — Organização do inventário: esquecer um dispositivo (com o histórico), esquecer automaticamente (opcional) os não mexidos offline há N dias, filtro por tipo no painel, exportação CSV/JSON (CSV seguro para planilhas) pelo painel e `atalaia export`, e `atalaia backup` para uma cópia consistente 0600 com o servidor rodando.

**EN** — Security review fixes: parallel password guesses could all pass the lockout check (attempts are now counted before the password is checked); a global limit across all addresses (about 30 a minute, at most a 1-minute lockout); at most 2 concurrent argon2id checks (memory exhaustion); a data race on the DHCP-names counter; SSDP answers capped at 256 per round; staticcheck added to CI.

**PT-BR** — Correções da revisão de segurança: tentativas de senha em paralelo podiam todas passar pela checagem do bloqueio (agora a tentativa é contada antes de conferir a senha); limite global para todos os endereços (cerca de 30 por minuto, bloqueio de no máximo 1 minuto); no máximo 2 verificações argon2id simultâneas (esgotamento de memória); condição de corrida no contador de nomes por DHCP; respostas SSDP limitadas a 256 por rodada; staticcheck na CI.

## v0.5.0 — 2026-11-18

**EN** — Usability fixes: `--allow-cidr` (env `ATALAIA_ALLOW_CIDR`) to accept a VPN such as Tailscale (over-wide networks refused); a clear "port already in use" error that suggests `--listen`; opt-in scheduled port scans of known, online, in-scope devices (off by default, audited, one at a time, off in observe-only mode); the redirect target in the exposure report; a "This machine" marker.

**PT-BR** — Correções de usabilidade: `--allow-cidr` (env `ATALAIA_ALLOW_CIDR`) para aceitar uma VPN como o Tailscale (redes largas demais são recusadas); erro claro de "porta já em uso" que sugere `--listen`; varreduras de portas agendadas opcionais de dispositivos conhecidos, online e no escopo (desligadas por padrão, auditadas, uma de cada vez, desativadas no modo somente observação); destino do redirecionamento no relatório de exposição; marca "Esta máquina".

**EN** — Opt-in learning of device names from DHCP requests (listen-only on UDP 67; hostname and Android vendor class; never replaces a name that is already set; announcements from devices not yet in the inventory wait for the next scan); a decision on IPv6 (not covered, reasons documented). Verified against the real network environment: the socket binds next to the system's DHCP client and parses a real UDP packet on the LAN address.

**PT-BR** — Aprendizado opcional de nomes pelos pedidos DHCP (só escuta a porta UDP 67; nome e classe de fabricante do Android; nunca troca um nome já definido; anúncios de dispositivos ainda fora do inventário esperam a próxima varredura); decisão sobre IPv6 (não coberto, motivos documentados). Verificado no ambiente de rede real: o socket abre ao lado do cliente DHCP do sistema e interpreta um pacote UDP real no endereço da LAN.

**EN** — Richer device identification: mDNS service types and model (direct query to each device), SSDP/UPnP (multicast search, description fetched only from the answering address, XML without entity expansion), NetBIOS names; a type guess (router, printer, TV, speaker, camera, NAS, computer, phone, smart-home…) with icons on the dashboard; model/type/services in the API and device page; what a device announces never erases user names and silence never erases what was known. Tested on a real LAN, where most devices were random-MAC phones that announce nothing (documented as the honest limit).

**PT-BR** — Identificação mais rica dos dispositivos: tipos de serviço e modelo mDNS (consulta direta a cada aparelho), SSDP/UPnP (busca em multicast, descrição baixada só do endereço que respondeu, XML sem expansão de entidades), nomes NetBIOS; palpite de tipo (roteador, impressora, TV, caixa de som, câmera, NAS, computador, celular, casa inteligente…) com ícones no painel; modelo/tipo/serviços na API e na página do dispositivo; o que o aparelho anuncia nunca apaga nomes do usuário e o silêncio nunca apaga o que já se sabia. Testado numa rede real, onde a maioria eram celulares com MAC aleatório que não anunciam nada (documentado como limite honesto).

**EN** — Documentation and licensing audit: the docs and READMEs no longer claim features that do not exist (SSDP/NetBIOS discovery, CVE matching, UPnP and open-DNS checks, Rapid7/NVD data sources, `CAP_NET_RAW`); architecture page rewritten from the real packages; developer page covers gosec, packages and releases; known limits added; SECURITY.md states supported versions and scope; the web app now shows a **Source code** link (AGPL section 13) on every page and in Settings.

**PT-BR** — Auditoria de documentação e licença: a documentação e os READMEs não prometem mais recursos que não existem (descoberta por SSDP/NetBIOS, CVEs, UPnP e DNS aberto, fontes Rapid7/NVD, `CAP_NET_RAW`); página de arquitetura reescrita a partir dos pacotes reais; página de desenvolvimento cobre gosec, pacotes e versões; limites conhecidos acrescentados; o SECURITY.md informa versões suportadas e escopo; o app web agora mostra o link **Código-fonte** (AGPL, seção 13) em todas as páginas e em Configurações.

## v0.4.0 — 2026-11-18

**EN** — M5.3 tools: unprivileged traceroute (UDP with a small TTL, ICMP answers read from the socket error queue; verified on loopback and against the internet) and a download speed test (opt-in with a per-use confirmation, because it tells speed.cloudflare.com your public IP; redirects refused, size and time capped); both off in observe-only mode and in the audit log. CVE matching was deliberately not built: the docs explain why (no reliable versions, a database to maintain, false positives and false negatives).

**PT-BR** — Ferramentas (M5.3): traceroute sem privilégios (UDP com TTL baixo, respostas ICMP lidas da fila de erros do socket; testado no loopback e na internet) e teste de velocidade de download (opt-in com confirmação a cada uso, porque informa seu IP público a speed.cloudflare.com; redirecionamentos recusados, tamanho e tempo limitados); ambos desligados no modo somente observação e no registro de auditoria. A verificação de CVEs foi deliberadamente não implementada: a documentação explica o motivo (sem versões confiáveis, banco para manter, falsos positivos e negativos).

**EN** — M4.2 passive security: optional rogue-DHCP check (one DISCOVER broadcast, offers listened to, never a REQUEST; first check is the baseline; a new server raises an event and an alert; check now or after every scan; needs UDP port 68, so root or `CAP_NET_BIND_SERVICE`; off in observe-only mode; verified against a real DHCP server); alert for ports that opened since the previous port scan; alert rules `rogue_dhcp` and `new_ports`; bilingual docs.

**PT-BR** — Segurança passiva (M4.2): verificação opcional de servidor DHCP falso (um DISCOVER em broadcast, escuta as ofertas, nunca envia REQUEST; a primeira verificação é a base; um servidor novo gera evento e alerta; verificar agora ou depois de cada varredura; exige a porta UDP 68, ou seja root ou `CAP_NET_BIND_SERVICE`; desligada no modo somente observação; testada contra um servidor DHCP real); alerta de portas que abriram desde a varredura anterior; regras de alerta `rogue_dhcp` e `new_ports`; documentação bilíngue.

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
