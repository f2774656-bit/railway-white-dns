# WhiteDNS Railway Edition

This is a Railway-native adaptation of the WhiteDNS Wizard architecture.

The original WhiteDNS Wizard is a local Go CLI that provisions a Docker-based 3x-ui/Xray stack on a VPS over SSH. This edition removes the VPS/SSH/Docker-in-Docker/PostgreSQL provisioning path and runs a minimal Xray VLESS-WebSocket service directly as one Railway service.

## What this version does

- Uses Railway's injected `PORT`.
- Uses a single VLESS + WebSocket inbound.
- Lets Railway terminate HTTPS/TLS at its public edge.
- Automatically uses `RAILWAY_PUBLIC_DOMAIN` once a Railway service domain exists.
- Accepts a stable `UUID` through a Railway secret.
- Prints a VLESS import string to the deployment logs.
- Requires no VPS and no SSH.

## What it intentionally does not do

It does not reproduce the original Wizard's full 3x-ui stack, PostgreSQL, Tor sidecar, Hysteria2 UDP, Reality, Shadowsocks UDP/TCP, certificate provisioning, or remote SSH management. Railway's public networking model is not equivalent to a VPS with arbitrary inbound ports.

## Deploy

1. Create a new Railway project and deploy this repository.
2. Generate a Railway service domain in **Settings -> Networking -> Public Networking**.
3. Set a Railway variable named `UUID` to a stable UUID.
4. Optional: set `WS_PATH` (default `/whitedns`).
5. Redeploy.
6. Open deployment logs and copy the printed `vless://...` import URL.
7. Import that URL into a client that supports VLESS + WebSocket + TLS.

### Notes about the generated link

The link uses:

- Host: Railway public domain (or `DOMAIN` if set)
- Port: `443`
- TLS: enabled
- Transport: WebSocket
- Path: `WS_PATH`
- SNI/Host: the same public domain

Railway terminates the public HTTPS connection and forwards the WebSocket traffic to the service port. Therefore Xray is deliberately configured with `security: none` behind the Railway edge.

## Resource and pricing note

Railway's current Free plan provides a small monthly usage credit rather than unlimited compute. New accounts also receive a one-time trial credit. Actual cost depends on usage; check your Railway Usage page.

## Xray version

The Dockerfile pins the official Xray container image to `26.9.9`. Update the tag deliberately when you want to move to a newer Xray release.
