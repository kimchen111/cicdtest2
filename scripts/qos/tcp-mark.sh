#!/bin/sh
INTF=$1
PARENT=$1
tc qdisc add dev ${INTF} parent ${PARENT} handle 2: prio
tc filter add dev ${INTF} parent 2: protocol ip handle 0x10 fw flowid 2:1
tc filter add dev ${INTF} parent 2: protocol ip handle 0x20 fw flowid 2:2
tc filter add dev ${INTF} parent 2: protocol ip handle 0x30 fw flowid 2:3
#tc filter list dev ${INTF}
#tc -s class show dev ${INTF}
