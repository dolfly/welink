#!/bin/bash
set -e

DOCS_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(dirname "$DOCS_DIR")"
CERT_DIR="$DOCS_DIR/certs"
DOMAIN="welink.click"
EMAIL="runzhliu@163.com"

echo "==> 创建共享网络（已存在则跳过）"
docker network create welink-net 2>/dev/null || true

# ── 证书 ──────────────────────────────────────────────────────────
if [ ! -f "$CERT_DIR/fullchain.pem" ]; then
    echo "==> 未找到证书，停止 website 以释放 80 端口..."
    docker compose -f "$DOCS_DIR/docker-compose.yml" down 2>/dev/null || true

    echo "==> 申请证书（$DOMAIN, demo.$DOMAIN）..."
    certbot certonly --standalone \
        -d "$DOMAIN" \
        -d "demo.$DOMAIN" \
        --email "$EMAIL" \
        --agree-tos \
        --no-eff-email \
        --non-interactive

    mkdir -p "$CERT_DIR"
    cp /etc/letsencrypt/live/$DOMAIN/fullchain.pem "$CERT_DIR/"
    cp /etc/letsencrypt/live/$DOMAIN/privkey.pem  "$CERT_DIR/"
    chmod 644 "$CERT_DIR/fullchain.pem"
    chmod 600 "$CERT_DIR/privkey.pem"
    echo "==> 证书申请成功"

    # 注册续期 hook（只需执行一次，幂等）
    HOOK_FILE="/etc/letsencrypt/renewal-hooks/deploy/welink.sh"
    cat > "$HOOK_FILE" <<HOOK
#!/bin/bash
cp /etc/letsencrypt/live/$DOMAIN/fullchain.pem $CERT_DIR/
cp /etc/letsencrypt/live/$DOMAIN/privkey.pem  $CERT_DIR/
chmod 644 $CERT_DIR/fullchain.pem
chmod 600 $CERT_DIR/privkey.pem
docker exec welink-website nginx -s reload 2>/dev/null || true
HOOK
    chmod +x "$HOOK_FILE"
    echo "==> 续期 hook 已注册：$HOOK_FILE"
else
    echo "==> 证书已存在，跳过申请"
fi

# ── 启动服务 ──────────────────────────────────────────────────────
echo "==> 启动 WeLink Demo..."
docker compose -f "$ROOT_DIR/docker-compose.demo.yml" pull
docker compose -f "$ROOT_DIR/docker-compose.demo.yml" up -d

echo "==> 启动官网..."
docker compose -f "$DOCS_DIR/docker-compose.yml" pull
docker compose -f "$DOCS_DIR/docker-compose.yml" up -d

echo ""
echo "完成！"
echo "  官网：https://$DOMAIN"
echo "  Demo：https://demo.$DOMAIN"
