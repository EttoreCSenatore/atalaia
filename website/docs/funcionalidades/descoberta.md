---
id: descoberta
title: Descoberta de dispositivos
sidebar_position: 1
---

# Descoberta de dispositivos

A partir do M1, o Atalaia descobre os dispositivos da rede local, identifica o fabricante e o nome de cada um e guarda o histórico num arquivo SQLite local. Por enquanto é pela linha de comando; o app web vem no M2.

## Como funciona

1. O Atalaia calcula as **redes que pode sondar**: só as redes IPv4 privadas das interfaces da própria máquina (veja `atalaia scope`). Nada fora delas é sondado.
2. Envia **um byte UDP** para cada endereço da rede (só a /24 onde a máquina está, mesmo que a rede seja maior). Isso faz o sistema resolver cada endereço por ARP.
3. Lê a **tabela de vizinhos do sistema** (`/proc/net/arp`) e confere de novo cada endereço contra o escopo.
4. Procura o **nome** de cada dispositivo (mDNS e DNS reverso) e o **fabricante** pelo MAC (lista pública do IEEE, MA-L, MA-M e MA-S, embutida no programa).

Não precisa de root nem de sockets brutos. Por enquanto roda só em **Linux** e só com **IPv4**.

## Escanear

```bash
atalaia scan
```

![Saída do atalaia scan com dados fictícios](/img/screenshots/cli-descoberta-scan.png)

Os dados da imagem são fictícios. `(random MAC)` indica um endereço MAC aleatório: celulares e notebooks modernos usam um endereço diferente em cada rede, por privacidade. Esses endereços não têm fabricante.

Opções: `--db ARQUIVO` (padrão `~/.local/share/atalaia/atalaia.db` ou `$ATALAIA_DB`), `--no-names` (não procura nomes) e `--offline-after 10m` (tempo sem ver um dispositivo até marcá-lo como offline; evita que um celular em repouso "pisque").

## Listar e ver o histórico

```bash
atalaia devices            # tabela (ou --json)
atalaia history --limit 5
```

![Saída do atalaia history com dados fictícios](/img/screenshots/cli-descoberta-historico.png)

Eventos: `joined` (visto pela primeira vez), `online` (voltou), `offline` (sumiu além do prazo) e `ip_changed` (mesmo dispositivo, endereço novo).

## Segurança e privacidade

- O banco guarda o inventário da rede, que é dado pessoal. O arquivo é criado com permissão `0600` e a pasta com `0700`, e nada sai da máquina.
- Nomes e fabricantes vêm da rede e são **dados não confiáveis**: caracteres de controle e de inversão de texto são removidos antes de guardar ou mostrar.
- Dentro do Docker, use `--network host`; sem isso o contêiner só enxerga a própria rede virtual.

## Limites atuais

- Só Linux e IPv4. macOS e Windows ficam para depois.
- Não há descoberta por SSDP/UPnP, NetBIOS nem pelos aluguéis do DHCP: aparelhos sem nome mDNS ou DNS reverso aparecem só com fabricante (ou, com MAC aleatório, sem nenhum).
- Endereços de VPN em CGNAT (por exemplo Tailscale, `100.64.0.0/10`) não são aceitos como "rede local" pelo servidor.
- Dispositivos que não respondem a ARP (raro) não aparecem; os que estão dormindo podem faltar em uma varredura, daí o prazo de offline.
- O próprio computador onde o Atalaia roda aparece na lista (a tabela ARP nunca lista a própria máquina, então ele é incluído à parte, só pela interface da rede principal, a que contém o roteador).
