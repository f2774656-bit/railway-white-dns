# Railway VLESS + WebSocket

A small Railway-ready VLESS over WebSocket service using Xray-core.

## What this is

This is a Railway-native deployment. It does not require:
- SSH access to a VPS
- 3x-ui
- PostgreSQL
- Docker-in-Docker
- a separate server

Railway runs the container and provides the public HTTPS endpoint.

## Important

This is **not a full replacement for the original WhiteDNS Wizard stack**. It is a minimal VLESS + WebSocket service adapted for Railway's HTTP/HTTPS public networking model.

## Deploy from GitHub

1. Create a new GitHub repository.
2. Put these files at the repository root.
3. Deploy the repository to Railway.
4. Generate a public domain from the Railway service settings.
5. In Railway Variables, add a fixed UUID.
6. Redeploy.
7. Open the deployment logs and copy the printed `vless://` URI.

## Variables

Required:

`UUID`
- Use one fixed UUID.
- Do not change it after clients are configured.

Optional:

`WS_PATH`
- Default: `/whitedns`

`PUBLIC_DOMAIN`
- Optional override.
- Normally Railway's `RAILWAY_PUBLIC_DOMAIN` is used automatically.

`PORT`
- Railway supplies this automatically.

## Local Docker test

Build:

```bash
docker build -t railway-vless-ws .
```

Run:

```bash
docker run --rm \
  -e PORT=8080 \
  -e UUID=00000000-0000-4000-8000-000000000000 \
  -e WS_PATH=/whitedns \
  -p 8080:8080 \
  railway-vless-ws
```

## Expected startup log

```text
Railway VLESS + WebSocket
Listen: 0.0.0.0:8080
WebSocket path: /whitedns
```

After a Railway public domain exists, the log also prints a `vless://` client URI.

## Notes

- Railway's public endpoint provides HTTPS; Xray itself listens for plain WebSocket traffic inside the container.
- The generated URI uses port 443 on the Railway public domain.
- Keep the UUID private.
- This project intentionally avoids hardcoding secrets in source control.
