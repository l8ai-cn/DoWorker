# Core Access Token RS256 Rollout

- **Date:** 2026-07-12
- **Scope:** Core Auth、Marketplace identity、Oilan deployment

## Contract

Core Auth signs access tokens with RS256 and publishes the active public key at
`/.well-known/jwks.json`. REST and Connect require issuer, audience, expiration,
RS256, and an exact `kid`. Marketplace validates the `marketplace-api` audience
through JWKS and never receives the private key.

`JWT_SECRET` remains limited to Relay tokens and credential encryption. It is
not an Access Token signing or verification key.

## Required Configuration

Core Backend:

```text
ACCESS_TOKEN_PRIVATE_KEY_FILE
ACCESS_TOKEN_PUBLIC_KEY_FILE
ACCESS_TOKEN_KEY_ID
ACCESS_TOKEN_ISSUER
ACCESS_TOKEN_AUDIENCES
ACCESS_TOKEN_CORE_AUDIENCE
ACCESS_TOKEN_EXPIRATION_HOURS
```

Marketplace API:

```text
MARKETPLACE_IDENTITY_ISSUER
MARKETPLACE_IDENTITY_AUDIENCE
MARKETPLACE_IDENTITY_JWKS_URL
MARKETPLACE_DATABASE_URL
MARKETPLACE_MIGRATION_DATABASE_URL
```

Missing or invalid key files stop Core Backend startup. Missing Marketplace
identity configuration stops Marketplace API startup.

## Rollout

1. Generate one RSA-2048 key pair and store it outside Git.
2. Mount the private and public key only into Core Backend.
3. Configure Marketplace with Core issuer, `marketplace-api`, and the public
   JWKS URL.
4. Deploy Core Backend and verify JWKS before deploying Marketplace API.
5. Require users to sign in again. Existing HS256 access tokens are rejected.

Rotation uses a new key and a new `kid`. Keep the previous public key in JWKS
until all tokens signed by it expire; multi-key publication is required before
performing the first production rotation.
