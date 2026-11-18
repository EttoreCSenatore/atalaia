---
id: seguranca-e-privacidade
title: Segurança e privacidade
sidebar_position: 3
---

# Security and privacy

## Secure defaults

- Login required on first run, no default password.
- Listens on the local network only, never on the internet.
- Active scans only on the host's own private networks, capped at a /16.
- Active checks are opt-in, rate-limited and written to an audit log.
- Runs without root, with `CAP_NET_RAW`.
- No telemetry and no cloud accounts.

## Privacy and data protection

MAC addresses, device names and connection times can identify people (for example, a resident's phone). So:

- Data stays in a local file owned by the user. The project receives nothing.
- History retention will be configurable (M2).
- Whoever runs Atalaia on a third party's network (a company, a building) is the controller of that data and must inform the people affected.

The project's compliance documentation is done with **AI-assisted review** and is **not legal advice**.

## How to report a vulnerability

Use GitHub's [private vulnerability reporting](https://github.com/EttoreCSenatore/atalaia/security/advisories/new). Do not open a public issue with details and do not send real personal data. The repository's `SECURITY.md` has the same text.

## Supply chain

CI audits dependency vulnerabilities and licenses, and Dependabot keeps dependencies and GitHub Actions up to date.
