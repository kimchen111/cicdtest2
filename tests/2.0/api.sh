#!/bin/bash
echo "RR: *RR"
echo "VPE1: *VPE1"
echo "VPE2: *VPE2"
echo "VPE3: *VPE3"
echo "VPE4: *VPE4"
echo "CPE1: *CPE1"
echo "CPE2: *CPE2"
echo "CPE3: *CPE3"
echo "CPE4: *CPE4"

##创建VNET可使用的端点，产生VRF、VXLAN(含Bridge)
echo "Create Vnet endpoints."
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/vpe/vnet/setvnetendpoint' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "vni": 50,
  "hubCidr": "10.253.0.0/16",
  "cpeCidr": "10.0.0.0/16",
  "reflectors": [
    {
      "esn": "*RR",
      "action": "ADD",
      "vtepAddr": "10.16.16.10/24"
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
    },
    {
      "esn": "*VPE3",
      "action": "ADD",
      "vtepAddr": "10.16.16.13/24"
    },
    {
      "esn": "*VPE4",
      "action": "ADD",
      "vtepAddr": "10.16.16.14/24"
    }
  ]
}'
echo
echo "====="

##创建VPN LINK
echo "CPE-1至VPE-1 createcpevpn"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/createcpevpn' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 101,
  "vni": 50,
  "client": {
    "esn": "*CPE1",
    "intfAddr": "10.0.11.2/30";
    "phyIntf":"eth0"
  },
  "server": {
    "esn": "*VPE1",
    "listenAddr": "10.2.0.53",
    "listenIntf": "eth7",
    "intfAddr": "10.0.11.1/30"
  }
}
'
echo
echo "====="

echo "CPE-2至VPE-2 createcpevpn"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/createcpevpn' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 102,
  "vni": 50,
  "client": {
    "esn": "*CPE2",
    "intfAddr": "10.0.12.2/30"
  },
  "server": {
    "esn": "*VPE2",
    "listenAddr": "10.2.0.57",
    "listenIntf": "eth7",
    "intfAddr": "10.0.12.1/30"
  }
}
'
echo
echo "====="

echo "CPE-4至VPE-4 createcpevpn"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/createcpevpn' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 104,
  "vni": 50,
  "client": {
    "esn": "*CPE4",
    "intfAddr": "10.0.14.2/30"
  },
  "server": {
    "esn": "*VPE4",
    "listenAddr": "10.2.0.69",
    "listenIntf": "eth7",
	  "intfAddr": "10.0.14.1/30"
  }
}
'
echo
echo "====="

# echo "finish"
