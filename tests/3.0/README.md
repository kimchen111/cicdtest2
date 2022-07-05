# CPE主机高可用

# 拓朴
按3.0/topo.png

# 前置步骤
请先做完1.0和2.0 2.1

# 基础网络

## CPE-3A
    拷贝粘贴 3.0/cpe3A.txt的内容

## CPE-3B
    拷贝粘贴 3.0/cpe3B.txt的内容

# 启动AGENT CPE-3A CPE-3B
    wget http://192.168.236.236:18080/tests/startca -O startca
    sh startca


# API调试，可以控制器上执行

    export CPE3A=0c5b2cc60000
    export CPE3B=0cd213a90000
    wget http://192.168.236.236:18080/tests/replvar -O replvar
    wget http://192.168.236.236:18080/tests/3.0/api.sh -O 3.0.sh
    sh replvar 3.0.sh
    sh 3.0.sh
    sh 3.0.sh enableHA
    sh 3.0.sh AtoBackup
    sh 3.0.sh BtoMaster
    sh 3.0.sh BtoBackup
    sh 3.0.sh AtoMaster
    sh 3.0.sh BdisableVrrp

CPE3A的brlan绑定了虚拟IP地址 192.168.110.1/24和固定IP地址192.168.110.11/24

CPE3B的brlan只绑定了固定IP地址192.168.110.12/24

