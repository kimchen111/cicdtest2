package cpe

import (
	"net/http"
	"sdwan/common"
	"sdwan/ctrl/agent"

	"github.com/gin-gonic/gin"
)

// @Summary CPE的网络信息
// @Description 获取CPE的网络信息
// @Tags CPE-Net
// @Param esn path string true "Device ID"
// @Accept  json
// @Produce  json
// @Success 200  {object} common.NetworkInfoVO  "查询成功的body"
// @Router /v2/cpe/net/netinfo/{esn} [get]
func NetworkInfo(c *gin.Context) {
	task := common.NewRequestTask(
		c.Param("esn"),
		common.CpeTaskClass.Network,
		common.CpeNetworkTaskType.NetworkInfo)
	result := agent.Request(task)
	c.JSON(http.StatusOK, result)
}

// @Summary 添加LAN网络
// @Description 示例
// @Description {
// @Description  "devices": [
// @Description    "eth5"
// @Description  ],
// @Description  "ipaddr": "192.168.5.1",
// @Description  "name": "lan2",
// @Description  "netmask": "255.255.255.0",
// @Description "protocol": "static"
// @Description }
// @Tags CPE-Net
// @Accept  json
// @Produce  json
// @Param esn path string true "Device ID"
// @Param data body common.LanVO true "需要以下参数：name ipaddr netmask devices"
// @Success 200  {string} string  "结果描述"
// @Router /v2/cpe/net/addlan/{esn} [post]
func AddLan(c *gin.Context) {
	lan := common.LanVO{}
	c.BindJSON(&lan)
	lan.Protocol = "static"
	task := common.NewRequestTaskWithBody(
		c.Param("esn"),
		common.CpeTaskClass.Network,
		common.CpeNetworkTaskType.AddLAN,
		lan,
	)
	result := agent.Request(task)
	c.JSON(http.StatusOK, result)
}

// @Summary 删除LAN网络
// @Description 示例
// @Description { "name": "lan2" }
// @Tags CPE-Net
// @Accept  json
// @Produce  json
// @Param esn path string true "Device ID"
// @Param data body common.LanVO true "需要1个参数：name，其余的不用填"
// @Success 200  {string} string  "结果描述"
// @Router /v2/cpe/net/dellan/{esn} [post]
func DelLan(c *gin.Context) {
	lan := common.LanVO{}
	c.BindJSON(&lan)
	task := common.NewRequestTaskWithBody(
		c.Param("esn"),
		common.CpeTaskClass.Network,
		common.CpeNetworkTaskType.DelLAN,
		lan,
	)
	result := agent.Request(task)
	c.JSON(http.StatusOK, result)
}

// @Summary 为一个LAN网络启用DHCP
// @Description 示例
// @Description {
// @Description  "lanName": "lan2",
// @Description  "start": 100,
// @Description  "end": 200
// @Description }
// @Tags CPE-Net
// @Accept  json
// @Produce  json
// @Param esn path string true "Device ID"
// @Param data body common.LanDhcpVO true "DHCP参数"
// @Success 200  {string} string  "结果描述"
// @Router /v2/cpe/net/enabledhcp/{esn} [post]
func EnableDHCP(c *gin.Context) {
	lanDhcp := common.LanDhcpVO{}
	c.BindJSON(&lanDhcp)
	task := common.NewRequestTaskWithBody(
		c.Param("esn"),
		common.CpeTaskClass.Network,
		common.CpeNetworkTaskType.EnableDHCP,
		lanDhcp,
	)
	result := agent.Request(task)
	c.JSON(http.StatusOK, result)
}

// @Summary 为一个LAN网络禁用DHCP
// @Description 示例
// @Description { "lanName": "lan2" }
// @Tags CPE-Net
// @Accept  json
// @Produce  json
// @Param esn path string true "Device ID"
// @Param data body common.LanDhcpVO true "DHCP参数，只需要填写 lanName"
// @Success 200  {string} string  "结果描述"
// @Router /v2/cpe/net/disabledhcp/{esn} [post]
func DisableDHCP(c *gin.Context) {
	lanDhcp := common.LanDhcpVO{}
	c.BindJSON(&lanDhcp)
	task := common.NewRequestTaskWithBody(
		c.Param("esn"),
		common.CpeTaskClass.Network,
		common.CpeNetworkTaskType.DisableDHCP,
		lanDhcp,
	)
	result := agent.Request(task)
	c.JSON(http.StatusOK, result)
}

// @Summary 开启流量分析
// @Description 示例
// @Description { "deviceId": 2495 }
// @Tags CPE-Net
// @Accept  json
// @Produce  json
// @Param esn path string true "Device ID"
// @Param data body common.TrafficAnalysisVO true "指定CPE的数字ID"
// @Success 200  {string} string  "结果描述"
// @Router /v2/cpe/net/enabletrafficanalysis/{esn} [post]
func EnableTrafficAnalysis(c *gin.Context) {
	ta := common.TrafficAnalysisVO{}
	c.BindJSON(&ta)
	task := common.NewRequestTaskWithBody(
		c.Param("esn"),
		common.CpeTaskClass.Network,
		common.CpeNetworkTaskType.EnableTrafficAnalysis,
		ta,
	)
	result := agent.Request(task)
	c.JSON(http.StatusOK, result)
}

// @Summary 关闭流量分析
// @Description 示例
// @Description { "deviceId": 2495 }
// @Tags CPE-Net
// @Accept  json
// @Produce  json
// @Param esn path string true "Device ID"
// @Param data body common.TrafficAnalysisVO true "指定CPE的数字ID"
// @Success 200  {string} string  "结果描述"
// @Router /v2/cpe/net/disabletrafficanalysis/{esn} [post]
func DisableTrafficAnalysis(c *gin.Context) {
	ta := common.TrafficAnalysisVO{}
	c.BindJSON(&ta)
	task := common.NewRequestTaskWithBody(
		c.Param("esn"),
		common.CpeTaskClass.Network,
		common.CpeNetworkTaskType.DisableTrafficAnalysis,
		ta,
	)
	result := agent.Request(task)
	c.JSON(http.StatusOK, result)
}

// @Summary 开始VPE探测
// @Description 示例
// @Description { "vpes": [
// @Description  {
// @Description  "esn": "0c17f7020000",
// @Description  "ipAddrs": ["222.222.222.97","223.223.223.73"]
// @Description  },
// @Description  {
// @Description  "esn": "0ce8fae70000",
// @Description  "ipAddrs": ["222.222.222.97","223.223.223.73"]
// @Description  }
// @Description ] }
// @Tags CPE-Net
// @Accept  json
// @Produce  json
// @Param esn path string true "Device ID"
// @Param data body common.VpeDetectVO true "对哪些VPE进行探测的参数"
// @Success 200  {string} string  "结果描述"
// @Router /v2/cpe/net/detectvpe/{esn} [post]
func DetectVpe(c *gin.Context) {
	vd := common.VpeDetectVO{}
	c.BindJSON(&vd)
	task := common.NewRequestTaskWithBody(
		c.Param("esn"),
		common.CpeTaskClass.Network,
		common.CpeNetworkTaskType.DetectVpe,
		vd,
	)
	result := agent.Request(task)
	c.JSON(http.StatusOK, result)
}

// @Summary 添加自定义监控任务
// @Description 示例
// @Description  {
// @Description  "measurement":"customer1_m1",
// @Description  "monitors":[{
// @Description  	"id":22204,
// @Description  	"srcDev":"eth0",
// @Description  	"srcAddr":"192.168.0.1",
// @Description  	"dstAddr":"192.168.211.9"
// @Description     },{
// @Description  	"id":22205,
// @Description  	"srcDev":"eth0",
// @Description  	"srcAddr":"192.168.0.2",
// @Description  	"dstAddr":"192.168.211.10"
// @Description     }]
// @Description  }
// @Tags CPE-Net
// @Accept  json
// @Produce  json
// @Param esn path string true "Device ID"
// @Param data body common.CustomMonitorVO true "自定义监控的参数"
// @Success 200  {string} string  "结果描述"
// @Router /v2/cpe/net/addcustommonitor/{esn} [post]
func AddCustomMonitor(c *gin.Context) {
	cms := common.CustomMonitorVO{}
	c.BindJSON(&cms)
	task := common.NewRequestTaskWithBody(
		c.Param("esn"),
		common.CpeTaskClass.Network,
		common.CpeNetworkTaskType.AddCustomMonitor,
		cms,
	)
	result := agent.Request(task)
	c.JSON(http.StatusOK, result)
}

// @Summary 删除自定义监控任务
// @Description 示例
// @Description  {
// @Description  "measurement":"customer1_m1",
// @Description  "monitors":[{
// @Description  	"id":22204
// @Description     },{
// @Description  	"id":22205
// @Description     }]
// @Description  }
// @Tags CPE-Net
// @Accept  json
// @Produce  json
// @Param esn path string true "Device ID"
// @Param data body common.CustomMonitorVO true "自定义监控的参数"
// @Success 200  {string} string  "结果描述"
// @Router /v2/cpe/net/delcustommonitor/{esn} [post]
func DelCustomMonitor(c *gin.Context) {
	cms := common.CustomMonitorVO{}
	c.BindJSON(&cms)
	task := common.NewRequestTaskWithBody(
		c.Param("esn"),
		common.CpeTaskClass.Network,
		common.CpeNetworkTaskType.DelCustomMonitor,
		cms,
	)
	result := agent.Request(task)
	c.JSON(http.StatusOK, result)
}
