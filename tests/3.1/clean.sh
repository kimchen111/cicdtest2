#!/bin/bash

echo "CPE3A-VPE1 removecpevpn"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/removecpevpn' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 211,
  "vni": 50,
  "client": {
    "esn": "*CPE3A"
  },
  "server": {
    "esn": "*VPE1"
  }
}'
echo
echo "====="

echo "CPE3A-VPE3 removecpevpn"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/removecpevpn' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 212,
  "vni": 50,
  "client": {
    "esn": "*CPE3A"
  },
  "server": {
    "esn": "*VPE3"
  }
}'
echo
echo "====="

echo "CPE3B-VPE1 removecpevpn"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/removecpevpn' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 213,
  "vni": 50,
  "client": {
    "esn": "*CPE3B"
  },
  "server": {
    "esn": "*VPE1"
  }
}'
echo
echo "====="

echo "CPE3B-VPE3 removecpevpn"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/link/removecpevpn' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": 214,
  "vni": 50,
  "client": {
    "esn": "*CPE3B"
  },
  "server": {
    "esn": "*VPE3"
  }
}'
echo
echo "====="

## 把LAN网段发布删除
echo "CPE-3A disablepublan"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/cpe/route/disablepublan/*CPE3A' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "name": "lan"
}'
echo

## 把LAN网段发布删除
echo "CPE-3B disablepublan"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/cpe/route/disablepublan/*CPE3B' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "name": "lan"
}'
echo
