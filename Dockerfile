# Railway-native WhiteDNS edition.
# No SSH, Docker-in-Docker, VPS provisioning, or PostgreSQL is required.

FROM ghcr.io/xtls/xray-core:26.9.9 AS xray

FROM golang:1.24-bookworm AS builder
WORKDIR /src
COPY cmd/railway-xray/main.go ./main.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o /out/railway-xray ./main.go

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=xray /usr/local/bin/xray /usr/local/bin/xray
COPY --from=builder /out/railway-xray /railway-xray
EXPOSE 8080
ENTRYPOINT ["/railway-xray"]
