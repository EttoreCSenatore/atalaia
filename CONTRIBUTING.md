# Contribuindo / Contributing

[Português](#português) · [English](#english)

## Português

Obrigado pelo interesse! O Atalaia é mantido por uma pessoa, então contribuições pequenas e focadas são as mais fáceis de revisar.

### Antes de começar

- **Vulnerabilidade?** Não abra issue: use o [relato privado](https://github.com/EttoreCSenatore/atalaia/security/advisories/new) (veja [SECURITY.md](SECURITY.md)).
- **Bug ou ideia?** Abra uma issue com o modelo correspondente. Para mudanças grandes, combine na issue antes de escrever o código.
- **Nunca** coloque dados reais da sua rede (MACs, IPs, nomes de dispositivos, capturas de tela reais) em issues, PRs, testes ou documentação. Use dados fictícios, como os de `web/tests/fixtures.mjs`.
- Use o Atalaia só em redes que você administra ou em que tem autorização por escrito. Contribuições que o tornem ofensivo (varredura fora do escopo local, exploração, evasão) não serão aceitas.

### Ambiente

Requisitos, comandos e verificações estão em [Desenvolvimento](https://ettorecsenatore.github.io/atalaia/docs/tecnico/desenvolvimento). O mínimo antes de abrir um PR:

```bash
gofmt -l .                       # deve ficar vazio
go vet ./... && go test -race ./...
python3 -I scripts/checar-fixacao.py
cd web && npm ci && npm run check && npm test
```

### O pull request

- Uma mudança por PR, a partir de um branch, com uma descrição do quê e do porquê. O PR é integrado com squash depois da CI verde e da revisão.
- **Documentação nas duas línguas**: toda página em `website/docs/` tem a versão em inglês em `website/i18n/en/...` com a mesma estrutura (a CI confere). Textos do app ficam em `web/src/messages/` (pt-BR e en, a CI confere a paridade).
- Entrada em `CHANGELOG.md`, em "Unreleased", em inglês e português, para mudanças visíveis.
- Testes para o comportamento novo; nenhum teste toca a rede real.
- Dependências novas só com licença compatível com a AGPL (a CI confere) e com motivo no PR.
- Ações do GitHub fixadas por SHA e imagens por digest ([Cadeia de suprimentos](https://ettorecsenatore.github.io/atalaia/docs/tecnico/cadeia-de-suprimentos)).

### Licença

Ao contribuir, você concorda que o código é licenciado sob a [AGPL-3.0-or-later](LICENSE) e a documentação sob a [CC BY-SA 4.0](LICENSE-DOCS.md), as mesmas licenças do projeto. Não há CLA. Siga o [Código de Conduta](CODE_OF_CONDUCT.md).

## English

Thanks for your interest! Atalaia is maintained by one person, so small, focused contributions are the easiest to review.

### Before you start

- **A vulnerability?** Do not open an issue: use [private reporting](https://github.com/EttoreCSenatore/atalaia/security/advisories/new) (see [SECURITY.md](SECURITY.md)).
- **A bug or an idea?** Open an issue with the matching template. For large changes, agree on the approach in the issue before writing code.
- **Never** put real data from your network (MACs, IPs, device names, real screenshots) in issues, PRs, tests or docs. Use fictitious data, like `web/tests/fixtures.mjs`.
- Use Atalaia only on networks you administer or have written authorization for. Contributions that make it offensive (scanning outside the local scope, exploitation, evasion) will not be accepted.

### Environment

Requirements, commands and checks are in [Development](https://ettorecsenatore.github.io/atalaia/en/docs/tecnico/desenvolvimento). The minimum before opening a PR:

```bash
gofmt -l .                       # must be empty
go vet ./... && go test -race ./...
python3 -I scripts/checar-fixacao.py
cd web && npm ci && npm run check && npm test
```

### The pull request

- One change per PR, from a branch, describing what and why. The PR is squash-merged after green CI and review.
- **Docs in both languages**: every page in `website/docs/` has its English version in `website/i18n/en/...` with the same structure (CI checks it). App texts live in `web/src/messages/` (pt-BR and en, CI checks parity).
- A `CHANGELOG.md` entry under "Unreleased", in English and Portuguese, for visible changes.
- Tests for new behaviour; no test touches the real network.
- New dependencies only with an AGPL-compatible licence (CI checks it) and a reason in the PR.
- GitHub actions pinned by SHA and images by digest ([Supply chain](https://ettorecsenatore.github.io/atalaia/en/docs/tecnico/cadeia-de-suprimentos)).

### Licence

By contributing you agree that your code is licensed under [AGPL-3.0-or-later](LICENSE) and your docs under [CC BY-SA 4.0](LICENSE-DOCS.md), the project's own licences. There is no CLA. Follow the [Code of Conduct](CODE_OF_CONDUCT.md).
