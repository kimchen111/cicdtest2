# 发布LAN网段与静态路由

# 拓朴
按1.0/topo.png

# 前置步骤
请先做完 1.0 2.0 3.0

## LAN网段发布与静态路由发布测试
    export CPE1=0c1a29130000
    export CPE2=0c250bc90000
    export CPE4=0c79111d0000
    wget http://192.168.236.236:18080/tests/2.1/api.sh -O 2.1.sh
    sh replvar 2.1.sh
    sh 2.1.sh

***
# 测试结果

### VPE1 VPE2 与 VPE4
    ip netns exec ns50 ip ro 

VPE 应该可以看到从CPE收到的路由，以及通过 RR 转发的其他VPE收到的CPE的路由

VPE2应该可以看到一条192.168.20.0/24的静态路由，并且其它VPE应该也可以收得到

CPE4上执行

    ping -I 192.168.140.1 192.168.110.1

应该是可以通的

### CPE1 与 CPE4
    ip ro

CPE1应该可以获得CPE4的LAN网段：192.168.140.0/24

CPE4应该可以获得CPE1的LAN网段：192.168.110.0/24

并且应该可以看到CPE2上面发布的一个192.168.20.0/24的静态路由网段

### C1
    ip a add 192.168.110.2/24 dev eth0
    ip li set up dev eth0
    ip ro add default via 192.168.110.1

### C2
    ip a add 192.168.120.2/24 dev eth0
    ip li set up dev eth0
    ip ro add default via 192.168.120.1

### C4
    ip a add 192.168.140.2/24 dev eth0
    ip li set up dev eth0
    ip ro add default via 192.168.140.1

### C4
    ping 192.168.110.2


# 清理
    export CPE1=0c1a29130000
    export CPE2=0c250bc90000
    export CPE4=0c79111d0000
    wget http://192.168.236.236:18080/tests/2.1/clean.sh -O 2.1c.sh
    sh replvar 2.1c.sh
    sh 2.1c.sh
