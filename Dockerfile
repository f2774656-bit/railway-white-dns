FROM alpine:3.22

RUN apk add --no-cache ca-certificates curl unzip bash \
    && mkdir -p /opt/xray /app \
    && curl -fL --retry 5 --retry-delay 2 \
       -o /tmp/xray.zip \
       https://github.com/XTLS/Xray-core/releases/latest/download/Xray-linux-64.zip \
    && unzip -q /tmp/xray.zip -d /opt/xray \
    && chmod +x /opt/xray/xray \
    && rm -f /tmp/xray.zip

COPY app/entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

ENV WS_PATH=/whitedns

ENTRYPOINT ["/app/entrypoint.sh"]
