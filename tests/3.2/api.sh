#!/bin/bash

# 指定到一个目标IP的公网出口
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/cpe/route/assignoutport/*CPE3A' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "intfName": "eth1",
  "target": "10.2.0.53/32"
}'