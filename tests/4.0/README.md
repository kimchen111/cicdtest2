# HUB组网


# 基础网络

## RR1
    拷贝粘贴 4.0/rr1.txt的内容

## RR2
    拷贝粘贴 4.0/rr2.txt的内容

## VPE-1
    拷贝粘贴 4.0/vpe1.txt的内容

## VPE-2
    拷贝粘贴 4.0/vpe2.txt的内容

## 其它的HUB CPE
    拷贝粘贴 4.0/cpex.txt的内容
    拷贝粘贴 4.0/hubx.txt的内容

### api1.sh 
创建VPE1、VPE2、RR的VNET端点

    export RR1=0c4a7fe70000
    export RR2=0c4e07d20000
    export VPE1=0c1a501e0000
    export VPE2=0c7e94680000
    export HUB1A=0cd899110000
    export HUB1B=0ce6a0390000
    export CPE1=0c4ac8110000
    export CPE2=0cd7de760000
    export CPE3=0caf7da00000
    export CPE4=0cee0a980000
    wget http://192.168.236.236:18080/tests/replvar -O replvar
    wget http://192.168.236.236:18080/tests/4.0/api1.sh -O 4.0.1.sh
    sh replvar 4.0.1.sh
    sh 4.0.1.sh

创建CPE1 CPE2 CPE3 到VPE的VPN链路


### api2.sh 

创建HUB1A至VPE1、HUB1B至VPE2的隧道
    export RR1=0c4a7fe70000
    export RR2=0c4e07d20000
    export VPE1=0c1a501e0000
    export VPE2=0c7e94680000
    export HUB1A=0cd899110000
    export HUB1B=0ce6a0390000
    export CPE1=0c4ac8110000
    export CPE2=0cd7de760000
    export CPE3=0caf7da00000
    export CPE4=0cee0a980000
    wget http://192.168.236.236:18080/tests/replvar -O replvar
    wget http://192.168.236.236:18080/tests/4.0/api2.sh -O 4.0.2.sh
    sh replvar 4.0.2.sh
    sh 4.0.2.sh

创建HUB1A至HUB1B的隧道

### api3.sh
创建CPE4至HUB1A、HUB1B的VPN链路


