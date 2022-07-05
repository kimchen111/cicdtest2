# CPE链路高可用

# 拓朴
按3.0/topo.png

# 前置步骤
请先做完1.0和2.0 2.1 3.0

# API调试，可以控制器上执行

    export CPE3A=0c5b2cc60000
    export CPE3B=0cd213a90000
    export VPE1=0c17f7020000
    export VPE3=0c6257e00000
    wget http://192.168.236.236:18080/tests/replvar -O replvar
    wget http://192.168.236.236:18080/tests/3.1/api.sh -O 3.1a.sh
    sh replvar 3.1a.sh
    sh 3.1a.sh

### 关于主备链路问题

在创建链路的时候需要标记State，这个会影响local-preference与metric

但是frr有某种BUG，导致现在无法正常测试

# 清理
    export CPE3A=0c5b2cc60000
    export CPE3B=0cd213a90000
    export VPE1=0c17f7020000
    export VPE3=0c6257e00000
    wget http://192.168.236.236:18080/tests/replvar -O replvar
    wget http://192.168.236.236:18080/tests/3.1/clean.sh -O 3.1c.sh
    sh replvar 3.1c.sh
    sh 3.1c.sh
