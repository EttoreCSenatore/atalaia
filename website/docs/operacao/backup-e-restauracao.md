---
id: backup-e-restauracao
title: Backup e restauração
sidebar_position: 2
---

# Backup e restauração

Todo o estado do Atalaia é **um arquivo SQLite**: dispositivos, histórico, configurações, a conta e os **segredos dos canais de alerta**. Trate as cópias como trataria uma senha.

## O que já acontece sozinho

- **Todo dia**, o `atalaia serve` copia o banco para `backups/` ao lado dele e guarda as 7 mais novas.
- **Toda semana**, ele testa a restauração da mais nova (abre uma cópia de rascunho como o Atalaia abriria e roda a verificação de integridade) e registra o resultado.
- **Antes de atualizar o esquema** do banco numa versão nova, ele faz uma cópia `atalaia-pre-migration-*.db`.
- **Configurações > Backups** mostra o último backup e o último teste; um aviso aparece quando o backup atrasa mais de 26 horas.

Opções e caminhos: [Referência de configuração](configuracao.md).

## O que você precisa fazer

As cópias automáticas ficam **no mesmo disco** do banco. Elas protegem contra erro e corrupção, não contra perder o disco ou a máquina. **Copie a pasta de backups para outro lugar** de vez em quando (outro disco, um NAS, um pendrive guardado):

```bash
# pacote
sudo rsync -a /var/lib/private/atalaia/backups/ /mnt/nas/atalaia-backups/
# Docker (o volume se chama deploy_atalaia-data com o compose do repositório)
docker run --rm -v deploy_atalaia-data:/data:ro -v /mnt/nas/atalaia-backups:/out alpine cp -a /data/backups/. /out/
```

Para ser avisado se o backup parar, use `--heartbeat-backup` ([Monitoramento](../tecnico/monitoramento.md)).

## Conferir uma cópia

```bash
atalaia verify-backup /caminho/atalaia-20261122-030000.db
atalaia verify-backup --dir /caminho/da/pasta      # a mais nova
```

O teste restaura numa cópia de rascunho, roda as migrações e a verificação de integridade e conta os dispositivos, sem mexer no arquivo testado.

## Restaurar

1. **Pare** o servidor: `sudo systemctl stop atalaia` ou `docker compose stop`.
2. **Guarde** o banco atual com outro nome (pode ser útil): `atalaia.db` → `atalaia.db.antes-de-restaurar`. Apague também `atalaia.db-wal` e `atalaia.db-shm`, se existirem.
3. **Copie** o backup escolhido para o lugar do banco, com o nome `atalaia.db`:
   - pacote: `sudo cp ARQUIVO /var/lib/private/atalaia/atalaia.db`;
   - Docker: `docker run --rm -v deploy_atalaia-data:/data -v "$PWD":/in alpine cp /in/ARQUIVO /data/atalaia.db && docker run --rm -v deploy_atalaia-data:/data alpine chown 65532:65532 /data/atalaia.db`.
4. **Inicie** o servidor. Se o backup for de uma versão anterior, ele é migrado (com uma cópia antes).

Um backup de uma versão **mais nova** do que a instalada não abre: atualize o Atalaia primeiro.

## Uma cópia avulsa

```bash
atalaia backup /caminho/novo-arquivo.db          # nunca sobrescreve
atalaia backup --dir /caminho/pasta --keep 30   # datada, com rotação (para o seu próprio agendador)
```

Funciona com o servidor rodando e grava com permissão `0600`.
