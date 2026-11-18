---
id: modelo-de-ameacas
title: Modelo de ameaças
sidebar_position: 2
---

# Threat model

Atalaia is a security tool that is also a target: it knows every device on the network and can probe them. This document says what we protect, from whom, and what the project does **not** do.

## What we protect

- The network inventory (MACs, names, services), which is personal data in many contexts.
- The dashboard credentials and notification secrets (tokens).
- The network itself: Atalaia must not become an attack tool.

## Who can attack

| Actor | Example |
|---|---|
| Another device on the network | Tries to open the dashboard, guess the password or send forged packets to the agent |
| Someone on the internet | Reaches the dashboard if the user exposes it by mistake |
| Malicious user | Uses Atalaia to scan someone else's network |
| Supply chain | A compromised dependency or image |

## Threats and mitigations

| Threat | Mitigation |
|---|---|
| Dashboard reachable without login | Login required on first run, no default password |
| Dashboard exposed to the internet | Listens on `127.0.0.1` by default; even with `--listen` it refuses non-local sources; docs recommend a VPN or a TLS proxy |
| DNS rebinding and CSRF against the dashboard | `Host` check, `X-Atalaia` header and `Origin` check, `SameSite=Strict` cookie |
| Password guessing | argon2id, progressive lockout per address |
| Used to scan third-party networks | Scope fixed to the host's own private networks (`internal/scope`), no arbitrary target |
| Active scan without consent | Opt-in, needs an "this is my network" confirmation, rate-limited, written to an audit log |
| Compromised agent becomes a foothold | Runs without root, only `CAP_NET_RAW`; minimal image with no extra privileges |
| An alert channel's secret leaks (Telegram token, e-mail password) | Stored only in the local `0600` database, never returned by the API nor repeated in errors; e-mail refuses to send a password without TLS |
| A malicious device name used to attack the alert | Sent as plain text, no markup; line breaks removed from titles |
| Data leaks | Local storage, no telemetry, configurable retention |
| Malicious input from the network (names, banners) | Treated as untrusted data: escaped in the UI, never executed |
| Compromised dependency or image | CI with vulnerability and license audits, SBOM and Dependabot |

## Out of scope, on purpose

Blocking devices with ARP spoofing, Wi-Fi deauthentication, exploiting vulnerabilities and brute-forcing. The only planned credential check uses a short public list of factory-default passwords, is opt-in and is logged.

## Responsible use

Use Atalaia only on networks you administer or have written authorization for. Scanning other people's networks may be a crime. See also [Security and privacy](seguranca-e-privacidade.md).
