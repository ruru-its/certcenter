# save at etc/cert-sync.sh
# 每天凌晨 3 點跑一次
# 0 3 * * * /bin/bash /etc/cert-sync.sh >> /var/log/cert-sync.log 2>&1
#!/bin/bash
DOMAIN="*.itsower.com.tw"
API="http://192.168.100.63:9250/health?domain=$DOMAIN"
CERT_API="http://192.168.100.63:9250/cert?domain=$DOMAIN"
TARGET_DIR="/etc/carcare-cert/live"

resp=$(curl -s "$API")
status=$(echo "$resp" | jq -r .status)

echo "[INFO] Checking cert status for domain=$DOMAIN"

if [ "$status" = "OK" ] || [ "$status" = "WARN" ]; then
    echo "[INFO] Cert status=$status, downloading..."
    rm -f live.zip # 移除舊檔
    mkdir -p "$TARGET_DIR" # 確保目錄存在
    curl -OJ "$CERT_API"
    unzip -o live.zip -d "$TARGET_DIR"
elif [ "$status" = "ERROR" ]; then
    echo "[ERROR] Certificate expired, manual intervention required!"
else
    echo "[ERROR] Unknown status: $status"
fi