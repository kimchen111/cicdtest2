#!/bin/bash
echo "RR: *RR"
echo "VPE1: *VPE1"
echo "VPE2: *VPE2"
echo "HUB1A: *HUB1A"
echo "HUB1B: *HUB1B"
echo "CPE1: *CPE1"
echo "CPE2: *CPE2"
echo "CPE3: *CPE3"

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
      "esn": "*RR",
      "action": "ADD",
      "vtepAddr": "10.16.16.1/24"
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
  'http://192.168.236.236:18080/v2/link/createhubtunnel' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 220,
  "vni": 50,
  "peerA": {
    "esn": "*HUB1A",
    "ipaddr": "192.168.122.211"
  },
  "peerB": {
    "esn": "*HUB1B",
    "ipaddr": "192.168.122.212"
  }
}'
echo
echo "=================="

##创建链路
echo "HUB-1A至VPE-1 createcpemstp"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/createcpemstp' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
    "id":211,
    "vni":50,
    "state":"PRIMARY",
    "peerA":{
        "esn":"*HUB1A",
        "intfName":"eth4",
        "vlanId":2211,
        "intfAddr":"10.253.0.1/30"
    },
    "peerB":{
        "esn":"*VPE1",
        "intfName":"eth2",
        "vlanId":2211,
        "intfAddr":"10.253.0.2/30"
    }
}'
echo
echo "=================="

echo "HUB-1A至VPE-2 createcpevpn"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/createcpevpn' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 212,
  "vni": 50,
  "state": "SECONDARY",
  "client": {
    "esn": "*HUB1A",
    "intfAddr": "10.253.0.5/30"
  },
  "server": {
    "esn": "*VPE2",
    "listenAddr": "192.168.122.112",
    "listenIntf": "eth0",
    "intfAddr": "10.253.0.6/30"
  }
}
'
echo
echo "========================="

echo "HUB-1B至VPE-1 createcpemstp"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/createcpemstp' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
    "id":213,
    "vni":50,
    "state":"PRIMARY",
    "peerA":{
        "esn":"*HUB1B",
        "intfName":"eth4",
        "vlanId":2213,
        "intfAddr":"10.253.0.9/30"
    },
    "peerB":{
        "esn":"*VPE1",
        "intfName":"eth1",
        "vlanId":2213,
        "intfAddr":"10.253.0.10/30"
    }
}'

echo
echo "=================="

echo "HUB-1B至VPE-2 createcpemstp"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/createcpemstp' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
    "id":214,
    "vni":50,
    "state":"SECONDARY",
    "peerA":{
        "esn":"*HUB1B",
        "intfName":"eth4",
        "vlanId":2214,
        "intfAddr":"10.253.0.13/30"
    },
    "peerB":{
        "esn":"*VPE2",
        "intfName":"eth3",
        "vlanId":2214,
        "intfAddr":"10.253.0.14/30"
    }
}'

echo
echo "=================="



##创建专线链路
echo "CPE-1至VPE-2 createcpemstp"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/createcpemstp' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
    "id":215,
    "vni":50,
    "state":"PRIMARY",
    "peerA":{
        "esn":"*CPE1",
        "intfName":"eth4",
        "vlanId":2215,
        "intfAddr":"10.254.0.17/30"
    },
    "peerB":{
        "esn":"*VPE2",
        "intfName":"eth1",
        "vlanId":2215,
        "intfAddr":"10.254.0.18/30"
    }
}'
echo
echo "=================="

echo "CPE-1至HUB-1B createcpevpn"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/createcpevpn' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 216,
  "vni": 50,
  "state": "SECONDARY",
  "client": {
    "esn": "*CPE1",
    "intfAddr": "10.254.0.21/30"
  },
  "server": {
    "esn": "*HUB1B",
    "listenAddr": "192.168.122.212",
    "listenIntf": "eth0",
    "intfAddr": "10.254.0.22/30"
  }
}
'
echo
echo "========================="

echo "CPE-2至VPE-2 createcpevpn"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/createcpevpn' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 217,
  "vni": 50,
  "state": "PRIMARY",
  "client": {
    "esn": "*CPE2",
    "intfAddr": "10.254.0.25/30"
  },
  "server": {
    "esn": "*VPE2",
    "listenAddr": "192.168.122.112",
    "listenIntf": "eth0",
    "intfAddr": "10.254.0.26/30"
  }
}
'
echo
echo "========================="

echo "CPE-2至HUB-1B createcpemstp"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/createcpemstp' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
    "id":218,
    "vni":50,
    "state":"SECONDARY",
    "peerA":{
        "esn":"*CPE2",
        "intfName":"eth4",
        "vlanId":2218,
        "intfAddr":"10.254.0.29/30"
    },
    "peerB":{
        "esn":"*HUB1B",
        "intfName":"eth6",
        "vlanId":2218,
        "intfAddr":"10.254.0.30/30"
    }
}'
echo
echo "========================="

echo "CPE-3至VPE-2 createcpemstp"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/createcpemstp' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
    "id":219,
    "vni":50,
    "state":"PRIMARY",
    "peerA":{
        "esn":"*CPE3",
        "intfName":"eth4",
        "vlanId":2219,
        "intfAddr":"10.254.0.33/30"
    },
    "peerB":{
        "esn":"*VPE2",
        "intfName":"eth2",
        "vlanId":2219,
        "intfAddr":"10.254.0.34/30"
    }
}'
echo
echo "=================="

echo "CPE-3至HUB-1B createcpemstp"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/createcpemstp' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
    "id":220,
    "vni":50,
    "state":"SECONDARY",
    "peerA":{
        "esn":"*CPE3",
        "intfName":"eth5",
        "vlanId":2220,
        "intfAddr":"10.254.0.37/30"
    },
    "peerB":{
        "esn":"*HUB1B",
        "intfName":"eth6",
        "vlanId":2220,
        "intfAddr":"10.254.0.38/30"
    }
}'
echo
echo "=================="

}
remove(){

echo "CPE-3至HUB-1B removecpemstp"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/removecpemstp' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 220,
  "vni": 50,
  "peerA": {
    "esn": "*CPE3",
    "intfName":"eth5",
    "vlanId" : 2220
  },
  "peerB": {
    "esn": "*HUB1B",
    "intfName":"eth6",
    "vlanId" : 2220
  }
}'
echo
echo "==============="


echo "CPE-3至VPE-2 removecpemstp"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/removecpemstp' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 219,
  "vni": 50,
  "peerA": {
    "esn": "*CPE3",
    "intfName":"eth4",
    "vlanId" : 2219
  },
  "peerB": {
    "esn": "*VPE2",
    "intfName":"eth2",
    "vlanId" : 2219
  }
}'
echo
echo "==============="


echo "CPE-2至HUB-1B removecpemstp"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/removecpemstp' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 218,
  "vni": 50,
  "peerA": {
    "esn": "*CPE2",
    "intfName":"eth4",
    "vlanId" : 2218
  },
  "peerB": {
    "esn": "*HUB1B",
    "intfName":"eth6",
    "vlanId" : 2218
  }
}'
echo
echo "==============="



echo "CPE-2至VPE-2 removecpevpn"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/removecpevpn' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 217,
  "vni": 50,
  "client": {
    "esn": "*CPE2"
  },
  "server": {
    "esn": "*VPE2"
  }
}'
echo
echo "==============="


echo "CPE-1至HUB-1B removecpevpn"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/removecpevpn' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 216,
  "vni": 50,
  "client": {
    "esn": "*CPE1"
  },
  "server": {
    "esn": "*HUB1B"
  }
}'
echo
echo "==============="


echo "CPE-1至VPE-2 removecpemstp"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/removecpemstp' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 215,
  "vni": 50,
  "peerA": {
    "esn": "*CPE1",
    "intfName":"eth4",
    "vlanId" : 2215
  },
  "peerB": {
    "esn": "*VPE2",
    "intfName":"eth1",
    "vlanId" : 2215
  }
}'
echo
echo "==============="



echo "HUB1B至VPE-2 removecpemstp"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/removecpemstp' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 214,
  "vni": 50,
  "peerA": {
    "esn": "*HUB1B",
    "intfName":"eth4",
    "vlanId" : 2214
  },
  "peerB": {
    "esn": "*VPE2",
    "intfName":"eth3",
    "vlanId" : 2214
  }
}'
echo
echo "==============="

echo "HUB1B至VPE-1 removecpemstp"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/removecpemstp' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 213,
  "vni": 50,
  "peerA": {
    "esn": "*HUB1B",
    "intfName":"eth4",
    "vlanId" : 2213
  },
  "peerB": {
    "esn": "*VPE1",
    "intfName":"eth1",
    "vlanId" : 2213
  }
}'
echo
echo "==============="

echo "HUB-1A至VPE-2 removecpevpn"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/removecpevpn' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 212,
  "vni": 50,
  "client": {
    "esn": "*HUB1A"
  },
  "server": {
    "esn": "*VPE2"
  }
}'
echo
echo "==============="


echo "删除HUB1A 至 VPE-1 专线链路"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/removecpemstp' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 211,
  "vni": 50,
  "peerA": {
    "esn": "*HUB1A",
    "intfName":"eth4",
    "vlanId" : 2211
  },
  "peerB": {
    "esn": "*VPE1",
    "intfName":"eth2",
    "vlanId" : 2211
  }
}'
echo
echo "==============="


echo "删除HUB1A-HUB1B的隧道"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/removehubtunnel' \
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
