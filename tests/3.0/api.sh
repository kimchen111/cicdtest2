#!/bin/bash
echo "CPE3A: *CPE3A"
echo "CPE3B: *CPE3B"

enableHA() {
echo "CPE3A与CPE3B的高可用"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/cpe/sys/enablehagroup' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "master": {
    "esn": "*CPE3A",
    "hbIntfName": "eth5",
    "vipAddrs": [
      {
        "lanName": "lan",
        "solidAddr": "192.168.130.11/24",
        "vipAddr": "192.168.130.1/24"
      }
    ]
  },
  "backup": {
    "esn": "*CPE3B",
    "hbIntfName": "eth5",
    "vipAddrs": [
      {
        "lanName": "lan",
        "solidAddr": "192.168.130.12/24",
        "vipAddr": "192.168.130.1/24"
      }
    ]
  }
}'
echo
}

AtoBackup() {
echo "把CPE3A变成BACKUP"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/cpe/sys/switchvrrp/*CPE3A' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "permanent": false,
  "state": "BACKUP"
}'
echo
}

AtoMaster() {
echo "把CPE3B永久变成MASTER"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/cpe/sys/switchvrrp/*CPE3A' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "permanent": true,
  "state": "MASTER"
}'
echo
}

BtoMaster() {
echo "把CPE3B永久变成MASTER"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/cpe/sys/switchvrrp/*CPE3B' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "permanent": true,
  "state": "MASTER"
}'
echo
}

BtoBackup() {
echo "把CPE3A变成BACKUP"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/cpe/sys/switchvrrp/*CPE3B' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "permanent": true,
  "state": "BACKUP"
}'
echo
}

BdisableVrrp() {
echo "禁用HA"
curl -X 'POST' \
  'http://192.168.236.236:18080/v2/cpe/sys/disablevrrp/*CPE3B' \
  -H 'accept: application/json' \
  -d ''
}

if [ $# != 1 ]; then
    echo "Usage: "
    echo "sh 3.0.sh enableHA"
    echo "sh 3.0.sh AtoBackup"
    echo "sh 3.0.sh BtoMaster"
    echo "sh 3.0.sh BtoBackup"
    echo "sh 3.0.sh AtoMaster"
    echo "sh 3.0.sh BdisableVrrp"
    exit 0
fi

case $1 in
    enableHA)
        enableHA
        ;;
    AtoBackup)
        AtoBackup
        ;;
    BtoMaster)
        BtoMaster
        ;;
    AtoMaster)
        AtoMaster
        ;;
    BtoBackup)
        BtoBackup
        ;;
    BdisableVrrp)
        BdisableVrrp
        ;;
    *)
        echo "?"
        ;;
esac
