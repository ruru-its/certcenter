#!/bin/bash
set -e

ACME_HOME="/root/.acme.sh"
REG_FILE="/its-certcenter/register.json"

# 初始化 acme 帳號 & 註冊 acme-dns
if [ ! -f "$REG_FILE" ]; then
  echo "[its-certcenter] Registering with acme-dns..."
  REG_JSON=$(curl -s -X POST "$ACME_DNS_API")
  echo "$REG_JSON" > "$REG_FILE"
  export REG_JSON="$REG_JSON"

  USERNAME=$(jq -r .username "$REG_FILE")
  SUBDOMAIN=$(jq -r .subdomain "$REG_FILE")
  FQDN=$(jq -r .fulldomain "$REG_FILE")

  echo "USERNAME=$USERNAME" >> /etc/environment
  echo "SUBDOMAIN=$SUBDOMAIN" >> /etc/environment
  echo "FQDN=$FQDN" >> /etc/environment
else
  echo "[its-certcenter] Using existing acme-dns registration"
fi

exec its-certcenter