---
id: app-web
title: App web
sidebar_position: 2
---

# App web

O `atalaia serve` serve um app web que mostra o inventário da rede, o histórico e as configurações. Ele vem **dentro do binário** (e da imagem Docker): não há nada para instalar no navegador nem serviço na nuvem. Funciona em português e inglês, com tema claro e escuro (segue o do sistema) e no celular.

```bash
atalaia serve          # abra http://127.0.0.1:8080
```

As imagens abaixo usam dados fictícios.

## Primeiro acesso

Na primeira vez não há senha padrão. O servidor imprime um **código de configuração** no console; digite-o e escolha uma senha para criar a conta de administrador.

![Tela de configuração inicial](/img/screenshots/app-configuracao-inicial.png)

Depois disso, o acesso é por senha. Cinco senhas erradas bloqueiam o endereço por um tempo.

![Tela de entrada](/img/screenshots/app-login.png)

## Painel

Mostra quantos dispositivos existem, quantos estão online, os novos das últimas 24 horas, e a lista com busca e filtro. **Escanear agora** pede uma varredura na hora; o painel se atualiza sozinho quando ela termina (atualização ao vivo).

![Painel com a lista de dispositivos](/img/screenshots/app-painel.png)

## Dispositivo

Cada dispositivo tem uma página com endereço, fabricante, primeira e última vez visto e o histórico dele. Dispositivos com MAC aleatório são explicados na própria página.

![Página de um dispositivo](/img/screenshots/app-dispositivo.png)

## Histórico

A linha do tempo de tudo o que mudou: dispositivos novos, que voltaram, que sumiram e que trocaram de endereço.

![Linha do tempo do histórico](/img/screenshots/app-historico.png)

## Configurações

Quanto tempo o histórico é guardado (padrão 90 dias; 0 guarda tudo), troca de senha (encerra as outras sessões), backups (último backup, último teste de restauração e **Fazer backup agora**; veja [Implantação](../tecnico/implantacao.md)), idioma e a versão.

![Seção Backups (dados fictícios)](/img/screenshots/app-backups.png)

![Tela de configurações](/img/screenshots/app-configuracoes.png)

## Novidades

A página **Novidades** (link no rodapé e em Configurações) mostra o que mudou em cada versão, com a marca "Em uso" na versão rodando. Depois de uma atualização, um aviso no topo diz para qual versão o Atalaia foi atualizado. Veja [Versões e novidades](../tecnico/versionamento.md).

![Página Novidades (dados fictícios)](/img/screenshots/app-novidades.png)
