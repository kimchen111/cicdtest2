package main

import (
	"log"
	"sdwan/common"
	"sdwan/ctrl/agent"
	"sdwan/ctrl/api"
	"sdwan/ctrl/device"
)

// gin-swagger必须在项目的根目录下
// @title 犀思云SDWAN系统控制器API服务
// @version 2.0
// @description 本服务提供犀思云SDWAN系统的控制器API服务，接受CPE、VPE的注册，接受中心管理控制台的API调用，从技术上来讲类似于是集合了API网关、业务中间件、frps的功能
// @host 192.168.236.236:18080
// @BasePath /
func main() {
	if r, ok := common.CheckTools(); !ok {
		log.Fatalf("Miss tools: %s.", r)
	}
	common.InitConfig("CTRL")
	common.InitVpeConst()
	common.InitCpeConst()

	go device.RunObserver()

	a2zc := make(chan common.Message)
	z2ac := make(chan common.Message)
	go device.RunZmq(a2zc, z2ac)
	go agent.RunSeparator(a2zc, z2ac)

	api.RunAPI()
}
