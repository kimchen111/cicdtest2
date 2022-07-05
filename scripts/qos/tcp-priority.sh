#!/bin/sh
INTF=$1
tc qdisc add dev ${INTF} parent 1:1 handle 2: prio
#tc filter list dev ${INTF}
#tc -s class show dev ${INTF}
