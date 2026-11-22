---
id: comparacao-kaladrius
title: Comparison with Kaladrius
sidebar_position: 9
---

# What to learn from Kaladrius

[Kaladrius](https://github.com/facchini-miguel/kaladrius) (Apache 2.0) is a management system for clinics, in Django. The **product** has nothing to do with Atalaia, but the way it is **operated and maintained** is more mature than ours. This page compares the two (as of November 2026) and lists what is worth bringing over. We bring **ideas**, not code.

## Comparison

| Topic | Kaladrius | Atalaia today | Bring over? |
|---|---|---|---|
| Dependencies (Dependabot) | Weekly, a **7-day wait** before proposing a new version, grouped PRs | Weekly, no wait, no groups | **Done** (M7.1) |
| GitHub actions and images | Pinned by **SHA** and by **digest** | By tag (`@v7`, `node:24-alpine`) | **Yes** (M7.1) |
| Repository files | SECURITY, CONTRIBUTING, CODE_OF_CONDUCT, CODEOWNERS, NOTICE | SECURITY only | **Done** (M7.2) |
| Versions and what's new | A single `releases.yaml` feeds the "What's new" page in the app, the version in the footer and the GitHub release | A text CHANGELOG; version only on the About card | **Done** (M7.3) |
| Health and monitoring | A health page without login, signals to Healthchecks.io (system, backup, restore, update, disk), numbers for Prometheus | Nothing without login; no warning if Atalaia stops | **Done** (M7.4) |
| Backup | Daily via a systemd timer, a **weekly restore test**, a warning when it is late | Manual `atalaia backup` | **Yes** (M7.5) |
| Migrations | CI tests going up and down each migration | Forward-only migrations; no test from older releases' databases | **Adapt** (M7.6): test upgrading from the database of every published release, and back up automatically before migrating |
| Packages | `.deb`/`.rpm` with install and remove scripts, an install test in CI, a **package repository** for updates | `.deb`/`.rpm` without scripts or install test | **Yes** (M7.7); the package repository comes later |
| Automatic update | A timer every 5 min with **automatic rollback** when health fails | None | **Not for now**: with a package, the system (apt/dnf) updates. To be discussed again with the package repository |
| Documentation | User manual, operations (backup, monitoring, upgrade), decision log, data model, FAQ | Mostly technical; no task-based manual, no decision log | **Yes** (M7.8) |
| Public demo | A demo server reset every night | Screenshots only | **Adapt** (M7.9): the web app in demo mode, with fictitious data, on GitHub Pages itself, with no server |
| Languages | pt-BR, English and Spanish | pt-BR and English | **Maybe** (M7.10) |
| Stack, Google Cloud, funding | Django/HTMX, a Google Cloud guide, FUNDING | Go + Svelte, self-hosted | **Does not** apply |

## What Atalaia already does better (and keeps)

- A single static binary, no external database; a ~26 MB unprivileged image.
- Security static analysis in CI (gosec, staticcheck, govulncheck, Trivy) and a documented self-review.
- Fenced active tools (scope, confirmation, audit log, observe-only mode).

The steps are in the [roadmap](../roadmap.md#m7-operations-and-maturity).
