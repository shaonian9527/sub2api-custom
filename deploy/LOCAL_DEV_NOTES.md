# Local dev / local image deployment notes

## Current verified local deployment method

This workspace is using a **locally built image** instead of the upstream official image.

Verified running container image:
- `sub2api-local:checkin`

Verified compose file:
- `deploy/docker-compose.workspace.yml`

## Why this file exists

The stock `deploy/docker-compose.local.yml` points to:
- `weishaw/sub2api:latest`

That is fine for upstream releases, but it will **ignore local uncommitted / local custom code changes**.
For local feature validation, this caused the service to run the official image instead of the modified workspace code.

Also, the host currently uses legacy `docker-compose` v1.29, which has an interpolation compatibility issue with the Redis `command:` block in the original compose file. The compat file fixes that.

## Use this for local validation

```bash
cd /root/.openclaw/workspace/sub2api/deploy

docker-compose \
  -f docker-compose.workspace.yml \
  --env-file .env \
  up -d --build
```

## Confirm the correct image is running

```bash
docker inspect sub2api --format '{{.Config.Image}}'
```

Expected result:

```bash
sub2api-local:checkin
```

## If you accidentally switched back to official image

If `docker inspect` shows `weishaw/sub2api:latest`, you are not testing local code.
Re-run with the workspace compose file above.

## Canonical local workspace entrypoint

Use this file as the default local-dev / local-feature-validation entrypoint:
- `deploy/docker-compose.workspace.yml`

It already includes:
1. local `build:` enabled for `sub2api`
2. Redis command compatible with the installed compose version

That way local feature testing has a single canonical entrypoint.
