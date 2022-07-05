#!/bin/sh
INTF=$1
tc qdisc del dev ${INTF} root
tc qdisc add dev ${INTF} root handle 1: prio
tc filter add dev ${INTF} parent 1: protocol ip handle 0x10 fw flowid 1:1
tc filter add dev ${INTF} parent 1: protocol ip handle 0x20 fw flowid 1:2
tc filter add dev ${INTF} parent 1: protocol ip handle 0x30 fw flowid 1:3
#tc filter list dev ${INTF}
#tc -s class show dev ${INTF}
