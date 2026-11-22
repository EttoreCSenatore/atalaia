---
id: atualizacao
title: Atualização
sidebar_position: 3
---

# Atualização

Veja o que mudou na página **Novidades** do app (depois de atualizar) ou nas [releases](https://github.com/EttoreCSenatore/atalaia/releases). Só a versão mais recente recebe correções de segurança.

## Como atualizar

| Instalação | Comando |
|---|---|
| Docker (compose do repositório) | `git pull && docker compose up -d --build` (na pasta `deploy`) |
| Imagem do GHCR | troque a tag (`ghcr.io/ettorecsenatore/atalaia:X.Y.Z`) e recrie o contêiner |
| Pacote `.deb` | baixe o novo e `sudo apt install ./atalaia_X.Y.Z_amd64.deb` |
| Pacote `.rpm` | `sudo dnf upgrade ./atalaia-X.Y.Z-1.x86_64.rpm` |
| Binário | pare, troque o arquivo, inicie |

Confira os arquivos baixados com o `SHA256SUMS` da release (`sha256sum -c SHA256SUMS --ignore-missing`).

No Debian e derivados, se você editou o `/etc/default/atalaia` e a versão nova também mudou esse arquivo, o `apt` pergunta qual manter: responda **N** (manter a sua). A versão nova fica ao lado, em `/etc/default/atalaia.dpkg-dist`, para você ver as opções novas (todas estão na [referência de configuração](configuracao.md)). Numa atualização automática, use `-o Dpkg::Options::=--force-confold` para manter a sua sem perguntar.

## O que acontece com o banco

- Se a versão nova mudou o formato do banco, o Atalaia **copia o banco antes** (`atalaia-pre-migration-schemaN-DATA.db`, ao lado dele) e só então migra. Se a cópia falhar, ele não migra e não muda nada.
- A CI testa a atualização a partir dos bancos criados por **cada versão publicada**.
- Com o pacote, seu `/etc/default/atalaia` é mantido, e o serviço só é reiniciado se já estava rodando.

## Voltar para a versão anterior

Uma versão antiga **se recusa** a abrir um banco migrado por uma mais nova. Para voltar:

1. Pare o servidor.
2. Instale a versão anterior.
3. Restaure a cópia `atalaia-pre-migration-*.db` (ou um backup de antes da atualização) como em [Backup e restauração](backup-e-restauracao.md#restaurar).

O que aconteceu na rede entre a atualização e a volta fica de fora (está no banco novo, que você guardou com outro nome).
