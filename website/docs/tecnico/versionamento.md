---
id: versionamento
title: Versões e novidades
sidebar_position: 7
---

# Versões e novidades

## Números de versão

O Atalaia segue o [versionamento semântico](https://semver.org/lang/pt-BR/): `MAIOR.MENOR.CORREÇÃO`. Enquanto a versão começar com `0.`, uma versão menor pode mudar comportamento; as mudanças que exigem atenção aparecem nas novidades. Só a **versão mais recente** recebe correções de segurança (veja o `SECURITY.md`).

## Uma fonte só: `releases.yaml`

O arquivo `releases.yaml`, na raiz do repositório, diz o que mudou em cada versão, em português e inglês, para quem usa. Ele vai embutido no binário e alimenta:

| Onde | O quê |
|---|---|
| App web, página **Novidades** (`/whats-new`) | Todas as versões, da mais nova para a mais antiga, com a marca "Em uso" na versão rodando |
| App web, rodapé e Configurações | A versão em uso e o link para as Novidades |
| App web, aviso no topo | "O Atalaia foi atualizado para a vX.Y.Z", quando roda uma versão mais nova que a última vista neste navegador |
| Release no GitHub | O texto da release (inglês e português) |
| API | `GET /api/releases` (com login) |

O `CHANGELOG.md` continua com os detalhes técnicos de cada versão.

Formato de uma entrada (a mais nova fica no topo):

```yaml
releases:
  - version: 0.7.0          # sem o "v"
    date: 2026-12-01        # AAAA-MM-DD
    en:
      title: Health checks and backups
      items:
        - A health endpoint for monitoring.
    pt-BR:
      title: Verificação de saúde e backups
      items:
        - Um endereço de saúde para monitoramento.
```

## O que a CI confere

`go run ./cmd/releasenotes check` roda em todo PR e falha se:

- o arquivo tiver um campo desconhecido, uma versão fora do formato `X.Y.Z` ou uma data fora do formato `AAAA-MM-DD`;
- as versões não estiverem da mais nova para a mais antiga, sem repetição, ou as datas estiverem fora de ordem;
- faltar título ou itens em uma das línguas, ou o número de itens for diferente entre elas;
- uma versão do `releases.yaml` não tiver o título `## vX.Y.Z — data` no `CHANGELOG.md` com a mesma data, ou o contrário.

Na release, o mesmo comando com `-tag` confere que a tag enviada é a primeira versão do arquivo; se não for, nada é publicado (nem binários, nem imagem).

## Aviso de novidades

O app guarda no navegador (`localStorage`) a última versão que mostrou. Na primeira vez que um navegador abre o app, nada aparece como novo. Depois de uma atualização, o aviso aparece no topo e um ponto aparece ao lado de **Novidades** no rodapé; abrir a página ou dispensar o aviso marca a versão como vista. Nada disso sai do navegador.

![Aviso de atualização no topo do painel (dados fictícios)](/img/screenshots/app-aviso-novidades.png)

## Como publicar uma versão

1. No `CHANGELOG.md`, troque "Unreleased" por `## vX.Y.Z — AAAA-MM-DD` (e abra um "Unreleased" vazio).
2. No topo do `releases.yaml`, acrescente a versão com a mesma data e os destaques nas duas línguas.
3. Rode `go run ./cmd/releasenotes check -changelog CHANGELOG.md -tag vX.Y.Z` e `go run ./cmd/releasenotes notes vX.Y.Z` para ver o texto.
4. Integre o PR e envie a tag: `git tag vX.Y.Z && git push origin vX.Y.Z`.
