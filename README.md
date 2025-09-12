### project build
``` shell
# 專案目錄（certcenter/）初始化
go mod init certcenter
# go 執行
go run ./cmd/server
# go 建置
go build -o certcenter ./cmd/server
```

### docker build

``` shell
 docker build -t its-certcenter:test .
```

查詢當前 fqdn
curl http://localhost:9250/register

發行憑證
curl -X POST "http://localhost:9250/cert?domain=*.itsower.com.tw"

下載憑證
curl -OJ "http://localhost:9250/cert?domain=*.itsower.com.tw"

檢查到期日
curl "http://localhost:9250/expire?domain=*.itsower.com.tw"

檢查健康狀態
curl "http://localhost:9250/health?domain=*.itsower.com.tw"