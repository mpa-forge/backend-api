# Auth Implementation

This file is a compatibility entry point. The canonical auth requirements now
live in `openspec/specs/api-authentication/spec.md`.

## OpenSpec Capability

- `api-authentication`

## Quick Reference

- protected Connect procedures are enforced through a Connect interceptor
- bearer tokens are validated against Clerk JWKS at
  `AUTH_ISSUER_URL/.well-known/jwks.json`
- verification enforces signature, issuer, and audience
- the baseline claim contract requires `sub` and may consume `email`,
  `display_name`, `given_name`, `family_name`, `role`, and `roles`
- supported internal roles remain `user` and `admin`
- unsupported role claims return `403`; missing or invalid tokens return `401`
- authenticated profile procedures use the verified Clerk `sub` as the local
  identity key
- the auth interceptor now enriches the active request span with the Connect
  procedure identity, auth result, and Connect failure code so protected API
  traces stay correlated without creating a second server span

## Update Rule

When auth behavior changes, update the OpenSpec capability first and keep this
file as a lightweight reader-friendly summary.
