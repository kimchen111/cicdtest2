#!/bin/bash

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

## 把LAN网段发布删除
echo "CPE-2 disablepublan"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/cpe/route/disablepublan/*CPE2' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "name": "lan"
}'
echo

## 把LAN网段发布删除
echo "CPE-4 disablepublan"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/cpe/route/disablepublan/*CPE4' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "name": "lan"
}'
echo

## 把静态路由删除
echo "CPE-2 delstaticroute"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/cpe/route/delstaticroute/*CPE2' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "target": "192.168.20.0/24"
}'
echo

