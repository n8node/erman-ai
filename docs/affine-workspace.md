# AFFiNE Workspace

Erman AI exposes AFFiNE at `/workspace-app` and embeds it in the dashboard at
`/dashboard/tools/workspace`. AFFiNE provides both documents and Edgeless
whiteboards.

## Required production environment

Set these values in the server `.env` before the first start:

```dotenv
AFFINE_REVISION=stable
AFFINE_DB_USER=affine
AFFINE_DB_PASSWORD=<strong-database-password>
AFFINE_DB_NAME=affine
AFFINE_SERVER_EXTERNAL_URL=https://erman.ai/workspace-app

WORKSPACE_OIDC_CLIENT_ID=erman-affine
WORKSPACE_OIDC_CLIENT_SECRET=<strong-oidc-client-secret>
WORKSPACE_OIDC_PRIVATE_KEY_B64=<base64-pkcs8-rsa-private-key>
WORKSPACE_OIDC_REDIRECT_URI=https://erman.ai/workspace-app/oauth/callback
```

Generate the OIDC values once and keep them stable:

```bash
openssl rand -base64 48
openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -outform DER \
  | base64 -w 0
```

Do not commit generated secrets.

## First deployment

```bash
git pull origin main
docker compose -f docker-compose.yml -f docker-compose.prod.yml pull affine affine-migration affine-redis affine-postgres
docker compose -f docker-compose.yml -f docker-compose.prod.yml up --build -d
docker compose -f docker-compose.yml -f docker-compose.prod.yml exec -T backend \
  goose -dir ./migrations postgres "$DATABASE_URL" up
docker compose -f docker-compose.yml -f docker-compose.prod.yml ps
```

Open `https://erman.ai/workspace-app/admin/setup` and create the first AFFiNE admin.

If the browser lands on a WordPress 404 after `/workspace-app/admin`, rebuild nginx
so `proxy_redirect` is applied, or open the setup URL directly.

## Configure seamless sign-in

In AFFiNE Admin → Settings → OAuth → OIDC, configure:

- Issuer: `https://erman.ai/api/v1/workspace/oidc` (discovery document:
  `https://erman.ai/api/v1/workspace/oidc/.well-known/openid-configuration`)
- Client ID: the value of `WORKSPACE_OIDC_CLIENT_ID`
- Client secret: the value of `WORKSPACE_OIDC_CLIENT_SECRET`
- Scopes: `openid email profile`
- ID claim: `sub`
- Email claim: `email`
- Name claim: `name`

AFFiNE discovers the following endpoints automatically:

- authorization: `/api/v1/workspace/oidc/authorize`
- token: `/api/v1/workspace/oidc/token`
- user info: `/api/v1/workspace/oidc/userinfo`
- JWKS: `/api/v1/workspace/oidc/jwks`

The authorization code is one-time, expires after 90 seconds, and supports
PKCE. ID and access tokens use RS256 and expire after 15 minutes.

## Configure OpenRouter

In AFFiNE Admin → Settings → AI:

1. Enable Copilot.
2. Configure the OpenAI-compatible provider.
3. Set Base URL to `https://openrouter.ai/api/v1`.
4. Set the API key from `OPENROUTER_API_KEY`.
5. Select the configured Erman AI model IDs for chat and generation.
6. Enable the legacy `/chat/completions` API style if the selected AFFiNE
   release exposes that option.

AFFiNE currently documents only partial feature parity for self-hosted AI.
Verify chat, document generation, and Edgeless generation separately after each
AFFiNE upgrade.

## Smoke test

1. Sign in to Erman AI.
2. Open `/dashboard/tools/workspace`.
3. Choose the Erman AI OIDC provider in AFFiNE; no second credentials should be
   requested.
4. Create a document.
5. Switch it to Edgeless and add a shape.
6. Refresh the page and verify both changes persist.
7. Run an AI prompt and confirm the request appears in OpenRouter usage.

## Backups

Back up all three AFFiNE data stores:

- `affine_postgres_data`
- `affine_storage`
- `affine_config`

Pin `AFFINE_REVISION` to a tested release tag before production upgrades. Run
the migration job and perform a restore test before changing that tag.
