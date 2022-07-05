#!/bin/bash

create() {

echo "创建HUB1A的HUB"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/hub/setup/*HUB1A' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "vni": 50,
  "role": "RR",
  "vtepAddr": "10.16.17.20/24",
  "cpeCidr": "10.254.0.0/16"
}'
echo
echo "=================="

echo "创建HUB1B的HUB"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/hub/setup/*HUB1B' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "vni": 50,
  "role": "HUB",
  "rrAddr": "10.16.17.20",
  "vtepAddr": "10.16.17.19/24",
  "cpeCidr": "10.254.0.0/16"
}'
echo
echo "=================="

echo "创建HUB1A-HUB1B的隧道"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/createtunnel' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 220,
  "vni": 50,
  "peerA": {
    "esn": "*HUB1A",
    "ipaddr": "192.168.122.220"
  },
  "peerB": {
    "esn": "*HUB1B",
    "ipaddr": "192.168.122.219"
  }
}'
echo
echo "=================="

##创建VPN LINK
echo "HUB-1A至VPE-1 createvpnlink"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/createvpnlink' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 221,
  "vni": 50,
  "state": "PRIMARY",
  "client": {
    "esn": "*HUB1A",
    "intfAddr": "10.253.20.2/30"
  },
  "server": {
    "esn": "*VPE1",
    "listenAddr": "192.168.122.111",
    "listenIntf": "eth0",
    "intfAddr": "10.253.20.1/30"
  }
}
'
echo
echo "========================="

echo "HUB-1B至VPE-2 createvpnlink"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/createvpnlink' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 222,
  "vni": 50,
  "state": "PRIMARY",
  "client": {
    "esn": "*HUB1B",
    "intfAddr": "10.253.19.2/30"
  },
  "server": {
    "esn": "*VPE2",
    "listenAddr": "192.168.122.112",
    "listenIntf": "eth0",
    "intfAddr": "10.253.19.1/30"
  }
}
'
echo
echo "========================="

echo "CPE-4至HUB-1A createvpnlink"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/createvpnlink' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 223,
  "vni": 50,
  "state": "PRIMARY",
  "client": {
    "esn": "*CPE4",
    "intfAddr": "10.254.20.2/30"
  },
  "server": {
    "esn": "*HUB1A",
    "listenAddr": "192.168.122.220",
    "listenIntf": "eth0",
    "intfAddr": "10.254.20.1/30"
  }
}
'
echo
echo "========================="

echo "CPE-4至HUB-1B createvpnlink"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/createvpnlink' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 224,
  "vni": 50,
  "state": "PRIMARY",
  "client": {
    "esn": "*CPE4",
    "intfAddr": "10.254.19.2/30"
  },
  "server": {
    "esn": "*HUB1B",
    "listenAddr": "192.168.122.219",
    "listenIntf": "eth0",
    "intfAddr": "10.254.19.1/30"
  }
}
'
echo
echo "========================="

## 把LAN网段发布出去
echo "HUB1A enablepublan"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/cpe/route/enablepublan/*HUB1A' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "name": "lan"
}'
echo

echo "CPE4 enablepublan"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/cpe/route/enablepublan/*CPE4' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "name": "lan"
}'
echo

}

###################################################################################################################
remove() {

## 把LAN网段取消发布
echo "HUB1A disablepublan"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/cpe/route/disablepublan/*HUB1A' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "name": "lan"
}'
echo

echo "删除HUB1A-HUB1B的隧道"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/removetunnel' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 220,
  "vni": 50,
  "peerA": {
    "esn": "*HUB1A"
  },
  "peerB": {
    "esn": "*HUB1B"
  }
}'
echo
echo "=================="


##删除VPN LINK
echo "HUB-1A至VPE-1 removevpnlink"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/removevpnlink' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 221,
  "vni": 50,
  "client": {
    "esn": "*HUB1A"
  },
  "server": {
    "esn": "*VPE1"
  }
}
'
echo
echo "========================="

echo "HUB-1B至VPE-2 removevpnlink"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/removevpnlink' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 222,
  "vni": 50,
  "client": {
    "esn": "*HUB1B"
  },
  "server": {
    "esn": "*VPE2"
  }
}
'
echo
echo "========================="

echo "CPE-4至HUB-1A removevpnlink"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/removevpnlink' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 223,
  "vni": 50,
  "client": {
    "esn": "*CPE4"
  },
  "server": {
    "esn": "*HUB1A"
  }
}
'
echo
echo "========================="

echo "HUB-1B至VPE-2 removevpnlink"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/removevpnlink' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 224,
  "vni": 50,
  "client": {
    "esn": "*CPE4"
  },
  "server": {
    "esn": "*HUB1B"
  }
}
'
echo
echo "========================="

echo "清理HUB1A的HUB"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/hub/destroy/*HUB1A' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "vni": 50
}'
echo
echo "========================="

echo "清理HUB1B的HUB"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/hub/destroy/*HUB1B' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "vni": 50
}'
echo
echo "========================="

}


if [ $# != 1 ]; then
    echo "Usage: "
    echo "sh 4.0.2.sh create"
    echo "sh 4.0.2.sh remove"
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
