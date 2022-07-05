#!/bin/sh
INTF=$1
IFB=$2
tc qdisc del dev ${INTF} ingress
tc qdisc add dev ${INTF} ingress
tc filter add dev ${INTF} parent ffff: protocol all prio 10 u32 match u32 0 0 action connmark action mirred egress redirect dev ${IFB}
