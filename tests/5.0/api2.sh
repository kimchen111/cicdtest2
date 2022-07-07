#!/bin/bash

echo "HUB2A: *HUB2A"
echo "HUB2B: *HUB2B"
echo "CPE5: *CPE5"

create_lan() {
echo "HUB-2A 创建LAN"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/cpe/net/addlan/*HUB2A' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
"devices": [
"eth2",
"eth3"
],
"ipaddr": "192.168.150.1",
"name": "lan",
"netmask": "255.255.255.0",
"protocol": "static"
}'
echo
echo "========================="

curl -X 'POST' \
  'http://192.168.236.236:18080/v2/cpe/net/addlan/*HUB2B' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
"devices": [
"eth2",
"eth3"
],
"ipaddr": "192.168.150.2",
"name": "lan",
"netmask": "255.255.255.0",
"protocol": "static"
}'
echo
echo "========================="

curl -X 'POST' \
  'http://192.168.236.236:18080/v2/cpe/net/addlan/*CPE5' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
"devices": [
"eth2",
"eth3"
],
"ipaddr": "192.168.155.1",
"name": "lan",
"netmask": "255.255.255.0",
"protocol": "static"
}'
echo
echo "========================="

}

setup_hub() {
echo "创建HUB2A的HUB"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/hub/setup/*HUB2A' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "vni": 50,
  "role": "RR",
  "vtepAddr": "10.16.18.20/24",
  "cpeCidr": "10.254.0.0/16"
}'
echo
echo "=================="

echo "创建HUB2B的HUB"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/hub/setup/*HUB2B' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "vni": 50,
  "role": "HUB",
  "rrAddr": "10.16.18.20",
  "vtepAddr": "10.16.18.19/24",
  "cpeCidr": "10.254.0.0/16"
}'
echo
echo "=================="

}

create_link() {


echo "HUB-2A至HUB-2B createhubmstp"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/createhubmstp' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
    "id":310,
    "vni":50,
    "peerA":{
        "esn":"*HUB2A",
        "intfName":"eth4",
        "vlanId":2310
    },
    "peerB":{
        "esn":"*HUB2B",
        "intfName":"eth4",
        "vlanId":2310
    }
}'
echo
echo "=================="


echo "CPE-5至HUB-2A createcpemstp"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/createcpemstp' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
    "id":311,
    "vni":50,
    "state":"PRIMARY",
    "peerA":{
        "esn":"*CPE5",
        "intfName":"eth4",
        "vlanId":2311,
        "intfAddr":"10.254.0.41/30"
    },
    "peerB":{
        "esn":"*HUB2A",
        "intfName":"eth5",
        "vlanId":2311,
        "intfAddr":"10.254.0.42/30"
    }
}'
echo
echo "========================="


echo "CPE-5至HUB-2B createcpemstp"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/createcpemstp' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
    "id":312,
    "vni":50,
    "state":"SECONDARY",
    "peerA":{
        "esn":"*CPE5",
        "intfName":"eth5",
        "vlanId":2312,
        "intfAddr":"10.254.0.45/30"
    },
    "peerB":{
        "esn":"*HUB2B",
        "intfName":"eth5",
        "vlanId":2312,
        "intfAddr":"10.254.0.46/30"
    }
}'
echo
echo "========================="
}


remove_link() {

echo "CPE-5至HUB-2A removecpemstp"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/removecpemstp' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 311,
  "vni": 50,
  "peerA": {
    "esn": "*CPE5",
    "intfName":"eth4",
    "vlanId":2311
  },
  "peerB": {
    "esn": "*HUB2A",
    "intfName":"eth5",
    "vlanId":2311
  }
}'
echo
echo "========================="

echo "CPE-5至HUB-2B removecpemstp"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/removecpemstp' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 312,
  "vni": 50,
  "peerA": {
    "esn": "*CPE5",
    "intfName":"eth5",
    "vlanId":2312
  },
  "peerB": {
    "esn": "*HUB2B",
    "intfName":"eth5",
    "vlanId":2312
  }
}'
echo
echo "========================="

  
echo "删除HUB2A-HUB2B的专线"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/removehubmstp' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 310,
  "vni": 50,
  "peerA": {
    "esn": "*HUB2A",
    "intfName":"eth4",
    "vlanId":2310
  },
  "peerB": {
    "esn": "*HUB2B",
    "intfName":"eth4",
    "vlanId":2310
  }
}'
echo
echo "========================="

}

destroy_hub() {
echo "清理HUB2A的HUB"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/hub/destroy/*HUB2A' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "vni": 50
}'
echo
echo "========================="

echo "清理HUB2B的HUB"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/hub/destroy/*HUB2B' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "vni": 50
}'
echo
echo "========================="

}

remove_lan() {
echo "删除HUB2A的LAN"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/cpe/net/dellan/*HUB2A' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
   "name": "lan"
}'
echo
echo "========================="

echo "删除HUB2B的LAN"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/cpe/net/dellan/*HUB2B' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "name": "lan"
}'
echo
echo "========================="

echo "删除CPE5的LAN"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/cpe/net/dellan/*CPE5' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "name": "lan"
}'
echo
echo "========================="
}

if [ $# != 1 ]; then
    echo "Usage: "
    echo "sh 5.2.sh create_lan|setup_hub|create_link"
    echo "sh 5.2.sh remove_lan|destroy_hub|remove_link"
    exit 0
fi

case $1 in
    create_lan)
        create_lan
        ;;
    setup_hub)
        setup_hub
        ;;
    create_link)
        create_link
        ;;
    remove_lan)
        remove_lan
        ;;
    destroy_hub)
        destroy_hub
        ;;
    remove_link)
        remove_link
        ;;
    *)
        echo "?"
        ;;
esac
