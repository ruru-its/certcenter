#!/bin/bash
set -e

ACME_HOME="/root/.acme.sh"
BASE_DIR="/its-certcenter"
REG_FILE="$BASE_DIR/register.json"

# 確保目錄存在
mkdir -p "$BASE_DIR"

# 設定預設 CA 為 Let's Encrypt
$ACME_HOME/acme.sh --set-default-ca --server letsencrypt
$ACME_HOME/acme.sh --register-account -m "$ACME_ACCOUNT" || true

# 初始化 acme 帳號 & 註冊 acme-dns
if [ ! -f "$REG_FILE" ]; then
  echo "[its-certcenter] Registering with acme-dns..."
  REG_JSON=$(curl -s -X POST "$ACME_DNS_API")
  echo "$REG_JSON" > "$REG_FILE"
  export REG_JSON="$REG_JSON"

  USERNAME=$(jq -r .username "$REG_FILE")
  PASSWORD=$(jq -r .password "$REG_FILE")
  SUBDOMAIN=$(jq -r .subdomain "$REG_FILE")
  FQDN=$(jq -r .fulldomain "$REG_FILE")

  echo "USERNAME=$USERNAME" >> /etc/environment
  echo "PASSWORD=$PASSWORD" >> /etc/environment
  echo "SUBDOMAIN=$SUBDOMAIN" >> /etc/environment
  echo "FQDN=$FQDN" >> /etc/environment
else
  echo "[its-certcenter] Using existing acme-dns registration"
fi

exec its-certcenter