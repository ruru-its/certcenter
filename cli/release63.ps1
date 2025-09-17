function releaseIts-certcenter {
    cd D:\Users\AmandaChou\git\itsower\certcenter
    docker build -t 192.168.100.63:5000/its-certcenter:latest .
    docker push 192.168.100.63:5000/its-certcenter:latest
    cd D:\Users\AmandaChou\git\itsower\certcenter\its
    createAndUse63
    docker-compose up -d
}

function Test63CertCenter {
    $url= "http://192.168.100.63:9250"
    Write-Host "now test $url"
    curl $url/register
}

function Test63IssueCertCenter {
    curl -X POST "http://192.168.100.63:9250/cert?domain=*.itsower.com.tw"
}
function Test63DownCertCenter {
    # curl -OJ "http://192.168.100.63:9250/cert?domain=*.itsower.com.tw"
}
 
# 查詢到期日：
# curl "http://192.168.100.63:9250/expire?domain=*.itsower.com.tw"
# 強制更新：
# curl -X POST "http://192.168.100.63:9250/renew?domain=*.itsower.com.tw"


function devCarcareNginx {
    cd D:\Users\AmandaChou\git\itsower\LineCRM.CarCare.Devops\VMSolution\nginx
    docker build -t 192.168.100.63:5000/carcare-nginx .
}

#手動更新證書 手動更新憑證 手動憑證更新 Carcare 手動更新Carcare證書 手動更新Carcare憑證
function releaseCarcareNginx {
    backToDefault
    cd D:\Users\AmandaChou\git\itsower\LineCRM.CarCare.Devops\VMSolution\nginx
    docker build -t 192.168.100.63:5000/carcare-nginx:1.0.2 .
    docker push 192.168.100.63:5000/carcare-nginx:1.0.2

    createAndUse41
    cd D:\Users\AmandaChou\git\itsower\LineCRM.CarCare.Devops\VMSolution\DevDocker\DevNginx\carcare
    docker-compose up -d
    cd D:\Users\AmandaChou\git\itsower\LineCRM.CarCare.Devops\VMSolution\DevDocker\DevNginx\carcare-sit
    docker-compose up -d

    createAndUse101
    cd D:\Users\AmandaChou\git\itsower\LineCRM.CarCare.Devops\VMSolution\DevDocker\DevNginx\prod\carcare
    docker-compose up -d
}

function PushAndCompose101 {
    cd D:\Users\AmandaChou\git\itsower\LineCRM.CarCare.Devops\VMSolution\nginx
    docker build -t 192.168.100.63:5000/carcare-nginx:latest .
    docker push 192.168.100.63:5000/carcare-nginx:latest
}