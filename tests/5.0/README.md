
# AGENT启动方法
#### VPE-AGENT VPE-1 VPE-2 VPE-3 VPE-4 
    wget http://192.168.236.236:18080/tests/startva -O startva
    sh startva

注意修改配置文件

#### CPE-AGENT CPE-1 CPE-2 CPE-3 CPE-5
    wget http://192.168.236.236:18080/tests/startca -O startca
    sh startca


# 基础网络

## RR VPE1 VPE2 HUB1A HUB1B CPE1 CPE2 CPE3 HUB2A HUB2B CPE5
    分别拷贝粘贴 5.0/对应的x.txt内容

### api1.sh 
    export RR=0c4b2d630000
    export VPE1=0cce2cee0000
    export VPE2=0caac3af0000
    export HUB1A=0c7cbb290000
    export HUB1B=0c88fd1d0000
    export CPE1=0cff26800000
    export CPE2=0c0be50d0000
    export CPE3=0cb2cebd0000

    export HUB2A=0cf5b50d0000
    export HUB2B=0c33fcd80000
    export CPE5=0c0422680000
    wget http://192.168.236.236:18080/tests/replvar -O replvar
    wget http://192.168.236.236:18080/tests/5.0/api1.sh -O 5.1.sh
    sh replvar 5.1.sh
    sh 5.1.sh


### api2.sh 
    export RR=0c4b2d630000
    export VPE1=0cce2cee0000
    export VPE2=0caac3af0000
    export HUB1A=0c7cbb290000
    export HUB1B=0c88fd1d0000
    export CPE1=0cff26800000
    export CPE2=0c0be50d0000
    export CPE3=0cb2cebd0000

    export HUB2A=0cf5b50d0000
    export HUB2B=0c33fcd80000
    export CPE5=0c0422680000
    wget http://192.168.236.236:18080/tests/replvar -O replvar
    wget http://192.168.236.236:18080/tests/5.0/api2.sh -O 5.2.sh
    sh replvar 5.2.sh
    sh 5.2.sh
