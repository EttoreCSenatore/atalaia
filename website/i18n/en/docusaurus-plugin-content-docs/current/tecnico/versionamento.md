---
id: versionamento
title: Versions and what's new
sidebar_position: 7
---

# Versions and what's new

## Version numbers

Atalaia follows [semantic versioning](https://semver.org/): `MAJOR.MINOR.PATCH`. While the version starts with `0.`, a minor version may change behaviour; changes that need attention are listed in what's new. Only the **latest version** receives security fixes (see `SECURITY.md`).

## One source: `releases.yaml`

The `releases.yaml` file, at the repository root, says what changed in each version, in Portuguese and English, for users. It is embedded in the binary and feeds:

| Where | What |
|---|---|
| Web app, **What's new** page (`/whats-new`) | Every version, newest first, with a "Running" mark on the running version |
| Web app, footer and Settings | The running version and the link to What's new |
| Web app, notice at the top | "Atalaia was updated to vX.Y.Z", when a version newer than the last one seen in this browser is running |
| GitHub release | The release text (English and Portuguese) |
| API | `GET /api/releases` (login required) |

`CHANGELOG.md` keeps the technical details of each version.

Format of an entry (the newest goes on top):

```yaml
releases:
  - version: 0.7.0          # no "v"
    date: 2026-12-01        # YYYY-MM-DD
    en:
      title: Health checks and backups
      items:
        - A health endpoint for monitoring.
    pt-BR:
      title: Verificação de saúde e backups
      items:
        - Um endereço de saúde para monitoramento.
```

## What CI checks

`go run ./cmd/releasenotes check` runs on every PR and fails if:

- the file has an unknown field, a version not in `X.Y.Z` form or a date not in `YYYY-MM-DD` form;
- versions are not newest first, without repeats, or dates are out of order;
- a language lacks a title or items, or the two languages have a different number of items;
- a version in `releases.yaml` has no `## vX.Y.Z — date` heading in `CHANGELOG.md` with the same date, or the other way round.

On release, the same command with `-tag` checks that the pushed tag is the first version in the file; if not, nothing is published (no binaries, no image).

## What's new notice

The app keeps the last version it showed in the browser (`localStorage`). The first time a browser opens the app, nothing shows as new. After an update, the notice appears at the top and a dot appears next to **What's new** in the footer; opening the page or dismissing the notice marks the version as seen. None of this leaves the browser.

![Update notice at the top of the dashboard (fictitious data)](/img/screenshots/app-aviso-novidades-en.png)

## How to publish a version

1. In `CHANGELOG.md`, replace "Unreleased" with `## vX.Y.Z — YYYY-MM-DD` (and open an empty "Unreleased").
2. At the top of `releases.yaml`, add the version with the same date and the highlights in both languages.
3. Run `go run ./cmd/releasenotes check -changelog CHANGELOG.md -tag vX.Y.Z` and `go run ./cmd/releasenotes notes vX.Y.Z` to see the text.
4. Merge the PR and push the tag: `git tag vX.Y.Z && git push origin vX.Y.Z`.
