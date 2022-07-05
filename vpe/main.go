package main

import (
	"log"
	"sdwan/vpe/agent"
	"sdwan/vpe/monitor"
	"sdwan/vpe/worker"

	"sdwan/common"
)

func main() {
	log.Println("================================================================")
	log.Printf("Starting VPE Agent, Version: %s", common.VpeAgentVersion)
	if r, ok := common.CheckTools(); !ok {
		log.Fatalf("Miss tools: %s.", r)
	}
	common.InitConfig("VPE")
	worker.InitConfig()

	common.Preparebgp()

	common.InitVpeConst()

	w2zc := make(chan common.Message)
	z2wc := make(chan common.Message)
	worker.InitAgent(w2zc)

	go agent.RunZmq(w2zc, z2wc)

	go monitor.StartVpeMonitorServer()

	agent.RunWorker(z2wc)
}
