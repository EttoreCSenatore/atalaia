---
id: glossario
title: Glossário
sidebar_position: 4
---

# Glossário

**ARP**: o protocolo que liga um endereço IP ao endereço físico (MAC) na rede local. O Atalaia lê a tabela ARP do sistema para saber quem está na rede.

**ARP spoofing**: um aparelho que responde por um IP que não é dele (por exemplo, o do roteador) para interceptar o tráfego. O Atalaia procura dois sinais disso.

**Ação ativa**: algo que o Atalaia faz **contra** um aparelho (ping, varredura de portas, verificação de exposição, Wake-on-LAN, traceroute, teste de velocidade, verificação de DHCP). Pede confirmação, fica no **registro de auditoria** e pode ser desligada (**modo somente observação**).

**Código de configuração**: o código de uso único que o servidor escreve no log na primeira vez, para criar a conta.

**Conhecido**: um dispositivo que você marcou como seu. Não gera alerta de "novo".

**DHCP**: o serviço da rede que entrega endereços IP aos aparelhos. Um **DHCP falso** é um segundo servidor que pode desviar o tráfego.

**Escopo**: as redes que o Atalaia pode sondar: só as redes privadas IPv4 da própria máquina (`atalaia scope`).

**Evento**: algo que aconteceu com um dispositivo: entrou, ficou online ou offline, mudou de endereço, ou um sinal de segurança.

**Exposição**: o que um aparelho oferece na rede que não deveria (serviços arriscados, TLS fraco, páginas de administração).

**Fabricante (OUI)**: os 3 primeiros bytes do MAC dizem quem fabricou a placa de rede, pela lista pública do IEEE.

**Gateway**: o roteador por onde a rede sai para a internet.

**MAC**: o endereço físico de uma placa de rede (`aa:bb:cc:dd:ee:ff`). **MAC aleatório**: um MAC inventado pelo aparelho por privacidade; o Atalaia sinaliza quando percebe.

**mDNS, SSDP, NetBIOS**: protocolos com que os aparelhos se anunciam na rede local; o Atalaia os usa para descobrir nomes, modelos e serviços.

**Migração**: a mudança do formato do banco numa versão nova. O Atalaia faz uma cópia antes.

**Modo somente observação**: o servidor iniciado com `--no-active-tools`: só descobre e avisa, nunca age.

**Registro de auditoria**: a lista das ações ativas feitas, com hora e alvo (em **Ferramentas**, **Registro de auditoria**).

**Sinal de vida (heartbeat)**: uma chamada que o Atalaia faz a um serviço de monitoramento depois de cada varredura ou backup; se parar de chegar, você é avisado.

**Teste de restauração**: restaurar um backup numa cópia de rascunho e conferir que ele abre e está íntegro.

**Varredura**: uma passada pela rede para ver quem está lá (a cada 5 minutos, por padrão). A **varredura de portas** é outra coisa: uma ação ativa contra um aparelho.
