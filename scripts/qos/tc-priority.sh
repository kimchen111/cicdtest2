#!/bin/sh
INTF=$1
tc qdisc del dev ${INTF} root
tc qdisc add dev ${INTF} root handle 1: prio
#tc filter list dev ${INTF}
#tc -s class show dev ${INTF}
