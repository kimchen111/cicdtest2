# 创建VNET与VPN链路

# 拓朴
按1.0/topo.png

# 前置步骤
请先做完 1.0

***
# AGENT调试

### AGENT启动方法
#### VPE-AGENT VPE-1 VPE-2 VPE-3 VPE-4 
    wget http://192.168.236.236:18080/tests/startva -O startva
    sh startva

注意修改配置文件

#### CPE-AGENT CPE-1 CPE-2 CPE-3 CPE-4 
    wget http://192.168.236.236:18080/tests/startca -O startca
    sh startca

注意修改配置文件

# API调试，可以控制器上执行
    export RR=0c9c9e3b0000
    export VPE1=0c17f7020000
    export VPE2=0ce8fae70000
    export VPE3=0c6257e00000
    export VPE4=0c38e1500000
    export CPE1=0c1a29130000
    export CPE2=0c250bc90000
    export CPE4=0c79111d0000
    wget http://192.168.236.236:18080/tests/replvar -O replvar
    wget http://192.168.236.236:18080/tests/2.0/api.sh -O 2.0.sh
    sh replvar 2.0.sh
    sh 2.0.sh

***
# 清理
    export RR=0c9c9e3b0000
    export VPE1=0c17f7020000
    export VPE2=0ce8fae70000
    export VPE3=0c6257e00000
    export VPE4=0c38e1500000
    export CPE1=0c1a29130000
    export CPE2=0c250bc90000
    export CPE4=0c79111d0000
    wget http://192.168.236.236:18080/tests/2.0/clean.sh -O 2.0c.sh
    sh replvar 2.0c.sh
    sh 2.0c.sh
