#!/bin/sh
ZONE=$1
PROTO=$2
line=$(nft -a list table inet qos_${ZONE}|grep "meta l4proto ${PROTO}")
handle=${line#*handle}
nft delete rule inet qos_${ZONE} mark_{ZONE} handle ${handle}
