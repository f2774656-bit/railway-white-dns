#!/bin/bash
set -euo pipefail

PORT="${PORT:-8080}"
UUID="${UUID:-}"
WS_PATH="${WS_PATH:-/whitedns}"
PUBLIC_DOMAIN="${PUBLIC_DOMAIN:-${RAILWAY_PUBLIC_DOMAIN:-}}"

if [[ -z "${UUID}" ]]; then
  echo "ERROR: UUID is not set."
  echo "Create a Railway variable named UUID with a fixed UUID value."
  exit 1
fi

if [[ "${WS_PATH}" != /* ]]; then
  WS_PATH="/${WS_PATH}"
fi

if ! [[ "${PORT}" =~ ^[0-9]+$ ]]; then
  echo "ERROR: PORT must be numeric. Current value: ${PORT}"
  exit 1
fi

cat > /app/config.json <<EOF
{
  "log": {
    "loglevel": "warning"
  },
  "inbounds": [
    {
      "listen": "0.0.0.0",
      "port": ${PORT},
      "protocol": "vless",
      "settings": {
        "clients": [
          {
            "id": "${UUID}",
            "email": "railway"
          }
        ],
        "decryption": "none"
      },
      "streamSettings": {
        "network": "ws",
        "wsSettings": {
          "path": "${WS_PATH}"
        }
      }
    }
  ],
  "outbounds": [
    {
      "protocol": "freedom",
      "tag": "direct"
    },
    {
      "protocol": "blackhole",
      "tag": "block"
    }
  ],
  "routing": {
    "domainStrategy": "AsIs",
    "rules": [
      {
        "type": "field",
        "protocol": ["bittorrent"],
        "outboundTag": "block"
      }
    ]
  }
}
EOF

echo "========================================"
echo "Railway VLESS + WebSocket"
echo "========================================"
echo "Listen: 0.0.0.0:${PORT}"
echo "WebSocket path: ${WS_PATH}"

if [[ -n "${PUBLIC_DOMAIN}" ]]; then
  echo
  echo "Client URI:"
  echo "vless://${UUID}@${PUBLIC_DOMAIN}:443?encryption=none&security=tls&type=ws&host=${PUBLIC_DOMAIN}&sni=${PUBLIC_DOMAIN}&path=${WS_PATH}#railway"
  echo
else
  echo
  echo "PUBLIC DOMAIN is not available yet."
  echo "Generate a Railway public domain, redeploy, and the VLESS URI will be printed."
  echo
fi

/opt/xray/xray -test -config /app/config.json
exec /opt/xray/xray run -config /app/config.json
