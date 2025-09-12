FROM golang:1.25-bookworm AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY cmd/server ./cmd/server
RUN go build -o /app/certcenter ./cmd/server

FROM debian:bookworm-slim
# 基本工具
RUN apt-get update && \
# -y 自動回答所有互動式問題為 yes，也就是「不用再人工輸入 y 來確認安裝」。
apt-get install -y \
# 發 HTTP 請求（下載檔案、呼叫 API）
curl \
# DNS 查詢工具，用來檢查 DNS 設定
dnsutils \
# 處理 JSON，方便解析/格式化 API 回應
jq \
# 處理 Cer 檔案 zip 下載
zip \
# SSL/TLS 工具，用來檢查/轉換憑證
openssl \
# 排程工具，自動定期續期憑證
cron \
# socket 工具，部分 ACME 驗證流程會用到
socat && \
# 清除 apt 暫存清單，減少 image 體積
rm -rf /var/lib/apt/lists/*

# 安裝 acme.sh
RUN curl https://get.acme.sh | sh

WORKDIR /app
COPY --from=builder /app/certcenter /usr/local/bin/certcenter
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

EXPOSE 9250
ENTRYPOINT ["/entrypoint.sh"]