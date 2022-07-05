#!/bin/bash

##创建VPN LINK
echo "CPE-3A至VPE-1 createvpnlink"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/createvpnlink' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 211,
  "vni": 50,
  "state": "PRIMARY",
  "client": {
    "esn": "*CPE3A",
    "intfAddr": "10.0.11.6/30"
  },
  "server": {
    "esn": "*VPE1",
    "listenAddr": "10.2.0.73",
    "listenIntf": "eth6",
    "intfAddr": "10.0.11.5/30"
  }
}
'
echo

echo "CPE-3A至VPE-3 createvpnlink"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/createvpnlink' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 212,
  "vni": 50,
  "state": "SECONDARY",
  "client": {
    "esn": "*CPE3A",
    "intfAddr": "10.0.13.2/30"
  },
  "server": {
    "esn": "*VPE3",
    "listenAddr": "10.2.0.61",
    "listenIntf": "eth6",
    "intfAddr": "10.0.13.1/30"
  }
}
'
echo

echo "CPE-3B至VPE-1 createvpnlink"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/createvpnlink' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 213,
  "vni": 50,
  "state": "PRIMARY",
  "client": {
    "esn": "*CPE3B",

    "intfAddr": "10.0.11.10/30"
  },
  "server": {
    "esn": "*VPE1",
    "listenAddr": "10.2.0.73",
    "listenIntf": "eth6",
	  "intfAddr": "10.0.11.9/30"
  }
}
'
echo 

echo "CPE-3B至VPE-3 createvpnlink"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/createvpnlink' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 214,
  "vni": 50,
  "state": "SECONDARY",
  "client": {
    "esn": "*CPE3B",
    "intfAddr": "10.0.13.6/30"
  },
  "server": {
    "esn": "*VPE3",
    "listenAddr": "10.2.0.61",
    "listenIntf": "eth6",
	  "intfAddr": "10.0.13.5/30"
  }
}
'
echo 

## 把LAN网段发布出去
echo "CPE-3A enablepublan"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/cpe/route/enablepublan/*CPE3A' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "name": "lan"
}'
echo

echo "CPE-3B enablepublan"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/cpe/route/enablepublan/*CPE3B' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "name": "lan"
}'
echo
