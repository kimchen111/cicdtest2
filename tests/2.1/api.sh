#!/bin/bash

## 添加LAN
echo "CPE-1 addlan"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/cpe/net/addlan/*CPE1' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
"devices": [
"eth2","eth3"
],
"ipaddr": "192.168.110.1",
"name": "lan",
"netmask": "255.255.255.0",
"protocol": "static"
}'
echo

## 添加LAN
echo "CPE-2 addlan"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/cpe/net/addlan/*CPE2' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
"devices": [
"eth2","eth3"
],
"ipaddr": "192.168.120.1",
"name": "lan",
"netmask": "255.255.255.0",
"protocol": "static"
}'
echo

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

echo "CPE-4 enablepublan"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/cpe/route/enablepublan/*CPE4' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "name": "lan"
}'
echo

## 为CPE2发布一个静态路由
echo "CPE-2 addstaticroute"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/cpe/route/addstaticroute/*CPE2' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '[
  {
    "metric": 20,
    "publish": true,
    "target": "192.168.20.0/24",
    "via": "192.168.120.2"
  }
]'
echo

echo "finish"