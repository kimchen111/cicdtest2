#!/bin/bash
echo "RR1: *RR1"
echo "RR2: *RR2"
echo "VPE1: *VPE1"
echo "VPE2: *VPE2"
echo "HUB1A: *HUB1A"
echo "HUB1B: *HUB1B"
echo "CPE1: *CPE1"
echo "CPE2: *CPE2"
echo "CPE3: *CPE3"
echo "CPE4: *CPE4"

create() {
##创建VNET可使用的端点，产生VRF、VXLAN(含Bridge)
echo "Create Vnet endpoints."
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/vpe/vnet/setvnetendpoint' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "vni": 50,
  "hubCidr": "10.253.0.0/16",
  "cpeCidr": "10.254.0.0/16",
  "reflectors": [
    {
      "esn": "*RR1",
      "action": "ADD",
      "vtepAddr": "10.16.16.1/24"
    },
    {
      "esn": "*RR2",
      "action": "ADD",
      "vtepAddr": "10.16.16.2/24"
    }
  ],
  "vteps": [
    {
      "esn": "*VPE1",
      "action": "ADD",
      "vtepAddr": "10.16.16.11/24"
    },
    {
      "esn": "*VPE2",
      "action": "ADD",
      "vtepAddr": "10.16.16.12/24"
    }
  ]
}'
echo
echo "========================="

##创建VPN LINK
echo "CPE-3至VPE-1 createvpnlink"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/createvpnlink' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 211,
  "vni": 50,
  "state": "PRIMARY",
  "client": {
    "esn": "*CPE3",
    "intfAddr": "10.254.11.2/30"
  },
  "server": {
    "esn": "*VPE1",
    "listenAddr": "192.168.122.111",
    "listenIntf": "eth0",
    "intfAddr": "10.254.11.1/30"
  }
}
'
echo
echo "========================="

echo "CPE-3至VPE-2 createvpnlink"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/createvpnlink' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 212,
  "vni": 50,
  "state": "SECONDARY",
  "client": {
    "esn": "*CPE3",
    "intfAddr": "10.254.12.2/30"
  },
  "server": {
    "esn": "*VPE2",
    "listenAddr": "192.168.122.112",
    "listenIntf": "eth0",
    "intfAddr": "10.254.12.1/30"
  }
}
'
echo
echo "========================="

echo "CPE-2至VPE-2 createvpnlink"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/createvpnlink' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 213,
  "vni": 50,
  "state": "SECONDARY",
  "client": {
    "esn": "*CPE2",
    "intfAddr": "10.254.12.6/30"
  },
  "server": {
    "esn": "*VPE2",
    "listenAddr": "192.168.122.112",
    "listenIntf": "eth0",
    "intfAddr": "10.254.12.5/30"
  }
}
'
echo
echo "========================="

echo "CPE-1至VPE-2 createvpnlink"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/createvpnlink' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 214,
  "vni": 50,
  "state": "SECONDARY",
  "client": {
    "esn": "*CPE1",
    "intfAddr": "10.254.12.10/30"
  },
  "server": {
    "esn": "*VPE2",
    "listenAddr": "192.168.122.112",
    "listenIntf": "eth0",
    "intfAddr": "10.254.12.9/30"
  }
}
'
echo
echo "========================="

## 创建LAN
echo "CPE-2 创建LAN"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/cpe/net/addlan/*CPE2' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
"devices": [
"eth2",
"eth3"
],
"ipaddr": "192.168.120.1",
"name": "lan",
"netmask": "255.255.255.0",
"protocol": "static"
}'
echo
echo "========================="

echo "CPE-3 创建LAN"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/cpe/net/addlan/*CPE3' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
"devices": [
"eth2",
"eth3"
],
"ipaddr": "192.168.130.1",
"name": "lan",
"netmask": "255.255.255.0",
"protocol": "static"
}'
echo
echo "========================="

## 把LAN网段发布出去
echo "CPE-1 enablepublan"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/cpe/route/enablepublan/*CPE1' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "name": "lan"
}'
echo

echo "CPE-2 enablepublan"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/cpe/route/enablepublan/*CPE2' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "name": "lan"
}'
echo

echo "CPE-3 enablepublan"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/cpe/route/enablepublan/*CPE3' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "name": "lan"
}'
echo

}

###################################################################################################################
remove() {

## 把LAN网段发布删除
echo "CPE-1 disablepublan"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/cpe/route/disablepublan/*CPE1' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "name": "lan"
}'
echo

echo "CPE-2 disablepublan"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/cpe/route/disablepublan/*CPE2' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "name": "lan"
}'
echo

echo "CPE-3 disablepublan"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/cpe/route/disablepublan/*CPE3' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "name": "lan"
}'
echo

## 删除vpnlink
echo "CPE3-VPE1 removevpnlink"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/removevpnlink' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 211,
  "vni": 50,
  "client": {
    "esn": "*CPE3"
  },
  "server": {
    "esn": "*VPE1"
  }
}'
echo
echo "====="

echo "CPE3-VPE2 removevpnlink"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/removevpnlink' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 212,
  "vni": 50,
  "client": {
    "esn": "*CPE3"
  },
  "server": {
    "esn": "*VPE2"
  }
}'
echo
echo "====="

echo "CPE2-VPE2 removevpnlink"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/removevpnlink' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 213,
  "vni": 50,
  "client": {
    "esn": "*CPE2"
  },
  "server": {
    "esn": "*VPE2"
  }
}'
echo
echo "====="

echo "CPE1-VPE2 removevpnlink"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/removevpnlink' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 214,
  "vni": 50,
  "client": {
    "esn": "*CPE1"
  },
  "server": {
    "esn": "*VPE2"
  }
}'
echo
echo "====="


##清理端点，包括VRF、VXLAN(含Bridge)、NETNS
echo "Clean vnet Endpoints"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/vpe/vnet/setvnetendpoint' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "vni": 50,
  "reflectors": [
    {
      "esn": "*RR1",
      "action": "DEL",
      "vtepAddr": "10.16.16.1/24"
    },
    {
      "esn": "*RR2",
      "action": "DEL",
      "vtepAddr": "10.16.16.2/24"
    }
  ],
  "vteps": [
    {
      "esn": "*VPE1",
      "action": "DEL",
      "vtepAddr": "10.16.16.11/24"
    },
    {
      "esn": "*VPE2",
      "action": "DEL",
      "vtepAddr": "10.16.16.12/24"
    }
  ]
}'
echo
echo "====="


}


resetlink() {
##重设一台CPE上两条链路的优先级
echo "Clean vnet Endpoints"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/resetlinkstate/*CPE3' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
"plink":{
"id": 211,
"state": "SECONDARY"
},
"slink":{
"id": 212,
"state": "PRIMARY"
}
}'

echo
echo "====="
}

if [ $# != 1 ]; then
    echo "Usage: "
    echo "sh 4.0.1.sh create"
    echo "sh 4.0.1.sh remove"
    exit 0
fi

case $1 in
    create)
        create
        ;;
    remove)
        remove
        ;;
    *)
        echo "?"
        ;;
esac
