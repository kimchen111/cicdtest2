#!/bin/bash

echo "VPE1: *VPE1"
echo "VPE2: *VPE2"
echo "VPE3: *VPE3"
echo "VPE4: *VPE4"
echo "CPE1: *CPE1"
echo "CPE2: *CPE2"
echo "CPE4: *CPE4"

echo "CPE1-VPE1 removecpevpn"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/removecpevpn' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 101,
  "vni": 50,
  "client": {
    "esn": "*CPE1"
  },
  "server": {
    "esn": "*VPE1"
  }
}'
echo
echo "====="

echo "CPE2-VPE2 removecpevpn"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/removecpevpn' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 102,
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

echo "CPE1-VPE4 removecpevpn"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/removecpevpn' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 104,
  "vni": 50,
  "client": {
    "esn": "*CPE4"
  },
  "server": {
    "esn": "*VPE4"
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
      "esn": "*RR",
      "action": "DEL",
      "vtepAddr": "10.16.16.10/24"
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
    },
    {
      "esn": "*VPE3",
      "action": "DEL",
      "vtepAddr": "10.16.16.13/24"
    },
    {
      "esn": "*VPE4",
      "action": "DEL",
      "vtepAddr": "10.16.16.14/24"
    }
  ]
}'
echo
echo "====="

echo "finish"