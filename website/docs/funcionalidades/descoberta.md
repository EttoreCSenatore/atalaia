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
4. Procura o **nome** de cada dispositivo (mDNS, DNS reverso, NetBIOS) e o **fabricante** pelo MAC (lista pública do IEEE, MA-L, MA-M e MA-S, embutida no programa).
5. Pergunta a cada um o que ele **diz que é** (veja abaixo) e adivinha o tipo.

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

## O que o Atalaia descobre sobre cada dispositivo

Além do fabricante (pelo MAC), cada dispositivo pode **dizer o que é**. O Atalaia pergunta, na própria rede local e só aos endereços que já encontrou:

| Fonte | O que traz | Como |
|---|---|---|
| **mDNS (tipos de serviço)** | Os serviços que o aparelho anuncia (`_airplay`, `_googlecast`, `_ipp` de impressoras, `_hap` do HomeKit, `_companion-link` da Apple…) e o **modelo** (`Chromecast Ultra`, `iPhone14,5`, o modelo da impressora) | Uma pergunta direta à porta 5353 de cada dispositivo |
| **SSDP/UPnP** | Fabricante, modelo, nome amigável e tipo (`MediaRenderer`, roteador…) de TVs, consoles, NAS e roteadores | Uma busca em multicast; a descrição só é baixada **do próprio endereço que respondeu** |
| **NetBIOS** | O nome do computador e o grupo de trabalho de PCs Windows, Samba e muitos NAS | Uma consulta direta à porta UDP 137 |

Com isso o Atalaia **adivinha o tipo** do dispositivo (roteador, impressora, TV, caixa de som, câmera, NAS, computador, celular, casa inteligente…) e mostra um ícone no painel. É uma **heurística**: sinais fortes (o roteador, um serviço de impressão, um tipo UPnP) valem mais que sinais fracos (um nome, um fabricante), e sem sinal nenhum ele **não chuta**. Um MAC aleatório sem nenhum anúncio aparece como "celular, tablet ou notebook", que é o que ele quase sempre é.

O que o aparelho anuncia **nunca apaga** o que você escreveu (o nome e as etiquetas são seus), e uma varredura em que ele ficou calado não apaga o que já se sabia.

**Limite honesto**: só descobre quem fala. Num teste em uma rede doméstica real, a maioria dos aparelhos eram celulares com MAC aleatório que **não anunciam nada** (nem mDNS, nem UPnP, nem NetBIOS): nesses casos o melhor é dar um nome e etiquetas à mão.

Tudo o que vem da rede é tratado como **não confiável**: texto é limpo de caracteres de controle e cortado, o XML do UPnP é lido sem expandir entidades, e um aparelho não consegue fazer o Atalaia baixar nada de outro endereço.

### Nomes pelo DHCP (opcional)

Celulares e notebooks **anunciam o nome** (e, no Android, uma "classe de fabricante" como `android-dhcp-14`) no pedido DHCP que enviam em broadcast quando entram na rede ou renovam o endereço, mesmo com MAC aleatório. Em **Configurações → Servidores DHCP** dá para ligar **Aprender o nome dos dispositivos pelos pedidos DHCP deles**: o Atalaia só **escuta** (não envia nem responde nada) e usa o que ouvir para dar nome e tipo a dispositivos que estão sem nome.

- **Desligado por padrão**: precisa escutar a porta UDP 67, que é privilegiada (root ou `CAP_NET_BIND_SERVICE`, como a [verificação de DHCP falso](seguranca-passiva.md)). Sem a capacidade, o app explica o erro.
- Só aprende **quando o aparelho renova ou reconecta**, o que pode levar horas; por isso os nomes vão aparecendo ao longo do tempo.
- Nunca troca um nome que você ou outra fonte já definiu, e um nome aprendido não apaga o seu.
- Não é uma ferramenta ativa: continua disponível no [modo somente observação](../tecnico/implantacao.md).

## Limpar a lista

Na página de um dispositivo, **Esquecer este dispositivo** remove o dispositivo, o histórico e os resultados de varredura (fica no registro de auditoria). Em **Configurações**, **Esquecer dispositivos offline há mais de N dias** limpa sozinho os que você nunca mexeu (não conhecidos, sem nome, sem etiquetas); o padrão é nunca. Isso evita o acúmulo de celulares que trocam o MAC aleatório. No painel dá para **filtrar por tipo**.

## Segurança e privacidade

- O banco guarda o inventário da rede, que é dado pessoal. O arquivo é criado com permissão `0600` e a pasta com `0700`, e nada sai da máquina.
- Nomes e fabricantes vêm da rede e são **dados não confiáveis**: caracteres de controle e de inversão de texto são removidos antes de guardar ou mostrar.
- Dentro do Docker, use `--network host`; sem isso o contêiner só enxerga a própria rede virtual.

## Limites atuais

- Só Linux e IPv4. macOS e Windows ficam para depois. **IPv6**: decidimos não cobrir por ora (exigiria ler a tabela de vizinhos por netlink e sondar o grupo multicast ND; quase todo dispositivo de uma rede doméstica também tem um endereço IPv4, que já é descoberto).
- Aparelhos que não anunciam nada (muitos celulares com MAC aleatório) aparecem sem nome nem modelo.
- Endereços de VPN em CGNAT (Tailscale, `100.64.0.0/10`) só são aceitos com `--allow-cidr` (veja [Implantação](../tecnico/implantacao.md#acesso-por-vpn-tailscale-e-similares)).
- Dispositivos que não respondem a ARP (raro) não aparecem; os que estão dormindo podem faltar em uma varredura, daí o prazo de offline.
- O próprio computador onde o Atalaia roda aparece na lista (a tabela ARP nunca lista a própria máquina, então ele é incluído à parte, só pela interface da rede principal, a que contém o roteador).
