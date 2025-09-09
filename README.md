### project build
``` shell
# 專案目錄（its-certcenter/）初始化
go mod init its-certcenter
# go 執行
go run ./cmd/server
# go 建置
go build -o its-certcenter ./cmd/server
```

### docker build

``` shell
 docker build -t its-certcenter:test .
```
查詢當前 fqdn
curl $url/register

發行憑證
curl -X POST "http://localhost:9250/cert?domain=*.itsower.com.tw"