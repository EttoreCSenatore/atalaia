---
id: faq
title: Perguntas frequentes
sidebar_position: 3
---

# Perguntas frequentes

## Esqueci a senha

Quem tem acesso ao arquivo do banco já controla o Atalaia, então a recuperação é pela linha de comando, na máquina dele:

1. Pare o servidor (`sudo systemctl stop atalaia` ou `docker compose stop`).
2. Rode `atalaia reset-password --yes` com o mesmo banco: pacote `sudo atalaia reset-password --db /var/lib/private/atalaia/atalaia.db --yes`; Docker `docker compose run --rm atalaia reset-password --yes`.
3. Inicie o servidor e pegue o novo **código de configuração** no log para criar a conta de novo.

Dispositivos, histórico, alertas e configurações continuam.

## Um celular aparece como novo toda hora

Celulares e notebooks modernos usam **MAC aleatório** por rede e às vezes trocam. Para o Atalaia, um MAC novo é um aparelho novo. Opções: desligue o MAC aleatório para a sua rede no aparelho (configuração de Wi-Fi > "endereço privado"), ou deixe o Atalaia **esquecer** os offline há N dias que você nunca mexeu (Configurações).

## Por que ele não vê todos os meus dispositivos?

- Ele só enxerga a **mesma rede** (mesma sub-rede IPv4) da máquina em que roda; redes de convidados e VLANs separadas ficam de fora.
- Aparelhos dormindo podem não responder a uma varredura; eles aparecem na próxima.
- No Docker, o contêiner precisa da **rede do host** (`network_mode: host`), que o compose já usa.
- **IPv6** não é coberto (veja o [registro de decisões](../tecnico/decisoes.md)).

## Ele é seguro de deixar rodando?

Ele roda sem root, só responde à rede local, exige login, não manda nada para a nuvem e as ações ativas (varredura de portas etc.) só acontecem quando você pede, com confirmação e registro. Ainda **não passou por um teste de invasão independente**; veja [Segurança e privacidade](../tecnico/seguranca-e-privacidade.md) e o [modelo de ameaças](../tecnico/modelo-de-ameacas.md).

## Ele manda algum dado para fora?

Não, a não ser o que você ligar: os alertas vão direto da máquina para o serviço que você escolher, o teste de velocidade fala com speed.cloudflare.com e os sinais de vida vão para a URL que você configurar. Não há telemetria.

## Posso usar numa rede que não é minha?

Só com autorização por escrito de quem a administra. Mesmo sendo defensivo, varrer redes alheias pode ser ilegal.

## Por que não procura vulnerabilidades (CVEs)?

Identificar a versão exata do software de cada aparelho pela rede é pouco confiável, e um alerta errado de CVE assusta mais do que ajuda. O Atalaia aponta o que é verificável: portas e serviços arriscados expostos, TLS fraco, interfaces web abertas. Veja [Ferramentas](../funcionalidades/ferramentas.md#verificação-de-cves-por-que-não-tem).

## O alerta de "ARP spoofing" é um ataque?

Pode ser, mas também acontece quando você troca o roteador, ou com repetidores e pontes Wi-Fi que respondem por vários aparelhos. Se você não mudou nada na rede, investigue; marcar os aparelhos como conhecidos reduz os falsos positivos.

## Quanto espaço ele usa?

Pouco: o banco de uma casa comum fica em poucos megabytes, e os 7 backups automáticos somam algo parecido. O app avisa quando o disco do banco está quase cheio.

## Como atualizo?

Veja [Atualização](../operacao/atualizacao.md). O banco é copiado antes de mudar de esquema, e dá para voltar atrás.

## Funciona no Windows ou no macOS?

A descoberta usa recursos do Linux (tabela ARP do kernel, sockets sem privilégio). Use um Linux: um Raspberry Pi já basta.

## Dá para instalar no celular?

Dá: veja [Instalar como app](../funcionalidades/app-web.md#instalar-como-app). A opção de instalar só aparece com HTTPS e um certificado em que o aparelho confia; sem isso, use o navegador normalmente.
