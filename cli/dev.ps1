function devCertcenter {
    cd D:\Users\AmandaChou\git\itsower\certcenter
    backToDefault
    docker build -t its-certcenter:test .
    docker-compose up -d
}

function TestCertCenter {
    $url= "http://localhost:9250"
    Write-Host "now test $url"
    curl $url/register
}

function TestIssueCertCenter {
    curl -X POST "http://localhost:9250/cert?domain=*.itsower.com.tw"
}
function TestDownCertCenter {
    # curl -OJ "http://localhost:9250/cert?domain=*.itsower.com.tw"
}
 
# 查詢到期日：
# curl "http://localhost:9250/expire?domain=*.itsower.com.tw"
# 強制更新：
# curl -X POST "http://localhost:9250/renew?domain=*.itsower.com.tw"


