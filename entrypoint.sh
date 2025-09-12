#!/bin/bash
set -e

ACME_HOME="/root/.acme.sh"
BASE_DIR="/certcenter"
REG_FILE="$BASE_DIR/register.json"

# 確保目錄存在
mkdir -p "$BASE_DIR"

# 設定預設 CA 為 Let's Encrypt
$ACME_HOME/acme.sh --set-default-ca --server letsencrypt
$ACME_HOME/acme.sh --register-account -m "$ACME_ACCOUNT" || true

# 初始化 acme 帳號 & 註冊 acme-dns
# 使憑證中心所有憑證申請都統一使用同一個帳號
if [ ! -f "$REG_FILE" ]; then
  echo "[certcenter] Registering with acme-dns..."
  REG_JSON=$(curl -s -X POST "$ACME_DNS_API")
  echo "$REG_JSON" > "$REG_FILE"
  export REG_JSON="$REG_JSON"

  USERNAME=$(jq -r .username "$REG_FILE")
  PASSWORD=$(jq -r .password "$REG_FILE")
  SUBDOMAIN=$(jq -r .subdomain "$REG_FILE")
  FQDN=$(jq -r .fulldomain "$REG_FILE")

  export USERNAME="$USERNAME"
  export PASSWORD="$PASSWORD"
  export SUBDOMAIN="$SUBDOMAIN"
  export FQDN="$FQDN"

  # 寫到 env 檔
  cat > "$BASE_DIR/register.env" <<EOF
USERNAME=$USERNAME
PASSWORD=$PASSWORD
SUBDOMAIN=$SUBDOMAIN
FQDN=$FQDN
EOF

else
  echo "[certcenter] Using existing acme-dns registration"
  if [ -f "$BASE_DIR/register.env" ]; then
    set -a
    source "$BASE_DIR/register.env"
    set +a
  fi
fi

exec /app/certcenter