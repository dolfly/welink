#!/bin/sh
# 在宿主机上运行一次，申请初始证书
# 用法：./init-letsencrypt.sh your@email.com
# 前提：docker compose up -d 已经启动

if [ -z "$1" ]; then
  echo "用法：./init-letsencrypt.sh your@email.com"
  exit 1
fi
EMAIL="$1"
DOMAIN="welink.click"

set -e

echo "==> 申请证书（domain: $DOMAIN, email: $EMAIL）"
docker exec welink-certbot certbot certonly \
  --webroot \
  -w /var/www/certbot \
  -d "$DOMAIN" \
  -d "www.$DOMAIN" \
  --email "$EMAIL" \
  --agree-tos \
  --no-eff-email

echo "==> 重启 website，切换到 HTTPS 配置"
docker compose restart website

echo ""
echo "完成！访问 https://$DOMAIN 验证"
