#!/bin/sh
set -e

CERT="/etc/certs/fullchain.pem"
TMPL_HTTP="/etc/nginx/templates/http-only.conf"
TMPL_HTTPS="/etc/nginx/templates/default.conf"
CONF="/etc/nginx/conf.d/default.conf"

if [ -f "$CERT" ]; then
    echo "[entrypoint] 证书已存在，使用 HTTPS 配置"
    cp "$TMPL_HTTPS" "$CONF"
else
    echo "[entrypoint] 未找到证书，使用 HTTP-only 配置"
    cp "$TMPL_HTTP" "$CONF"
fi

# 每天 reload 一次，证书续期后自动生效
(while true; do
    sleep 24h
    if [ -f "$CERT" ]; then
        cp "$TMPL_HTTPS" "$CONF"
        nginx -s reload && echo "[entrypoint] $(date): nginx reloaded"
    fi
done) &

exec nginx -g 'daemon off;'
