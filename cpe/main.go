package main

import (
	"log"
	"sdwan/common"
	"sdwan/cpe/agent"
	"sdwan/cpe/monitor"
	"sdwan/cpe/worker"
)

func main() {
	log.Println("================================================================")
	log.Printf("Starting CPE Agent, Version: %s", common.CpeAgentVersion)
	if r, ok := common.CheckTools(); !ok {
		log.Fatalf("Miss tools: %s.", r)
	}
	common.InitConfig("CPE")
	worker.InitConfig()

	// common.Preparebgp(common.CpeBgpConfDir)
	common.InitCpeConst()

	w2zc := make(chan common.Message)
	z2wc := make(chan common.Message)
	worker.InitAgent(w2zc)

	go agent.RunZmq(w2zc, z2wc)
	go agent.RunWorker(z2wc)
	monitor.StartCpeMonitor()
	monitor.StartTrafficAnalysis()

	worker.PostBoot()

	agent.RunWeb()
}
