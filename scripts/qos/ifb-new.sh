#!/bin/sh
IFB=$1
ip li del ${IFB}
ip li add ${IFB} type ifb
ip li set up dev ${IFB}
tc qdisc add dev ${IFB} root handle 1: prio
