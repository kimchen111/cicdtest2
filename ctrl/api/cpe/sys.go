package cpe

import (
	"net/http"
	"sdwan/common"
	"sdwan/ctrl/agent"

	"github.com/gin-gonic/gin"
)

// @Summary 根据参数配置HA
// @Description 示例
// @Description {
// @Description "master": {
// @Description "esn": "0c1210020000",
// @Description 	"hbIntfName": "eth5",
// @Description 	"vipAddrs": [
// @Description 	{
// @Description 		"lanName": "lan",
// @Description 		"solidAddr": "192.168.110.11/24",
// @Description 		"vipAddr": "192.168.110.1/24"
// @Description 	}
// @Description 	]
// @Description },
// @Description "backup": {
// @Description 	"esn": "0c1d814a0000",
// @Description 	"hbIntfName": "eth5",
// @Description 	"vipAddrs": [
// @Description 	{
// @Description 		"lanName": "lan",
// @Description 		"solidAddr": "192.168.110.12/24",
// @Description 		"vipAddr": "192.168.110.1/24"
// @Description 	}
// @Description 	]
// @Description }
// @Description }
// @Tags CPE-Sys
// @Accept  json
// @Produce  json
// @Param data body common.HaVO true "HA-Keepalived参数"
// @Success 200  {string} string  "结果描述"
// @Router /v2/cpe/sys/enablehagroup [post]
func EnableHAGroup(c *gin.Context) {
	ha := common.HaVO{}
	c.BindJSON(&ha)
	master, backup := ha.GenVrrp()
	mtask := common.NewRequestTaskWithBody(
		ha.Master.Esn,
		common.CpeTaskClass.System,
		common.CpeSystemTaskType.EnableVRRP,
		master,
	)
	mresult := agent.Request(mtask)

	btask := common.NewRequestTaskWithBody(
		ha.Backup.Esn,
		common.CpeTaskClass.System,
		common.CpeSystemTaskType.EnableVRRP,
		backup,
	)
	bresult := agent.Request(btask)
	c.JSON(http.StatusOK, [...]common.ApiResult{mresult, bresult})
}

// @Summary 禁用VRRP-HA
// @Description 无payload
// @Description 仅停止keepalived服务，并删除配置文件
// @Tags CPE-Sys
// @Accept  json
// @Produce  json
// @Param esn path string true "Device ID"
// @Success 200  {string} string  "结果描述"
// @Router /v2/cpe/sys/disablevrrp/{esn} [post]
func DisableVRRP(c *gin.Context) {
	task := common.NewRequestTask(
		c.Param("esn"),
		common.CpeTaskClass.System,
		common.CpeSystemTaskType.DisableVRRP,
	)
	result := agent.Request(task)
	c.JSON(http.StatusOK, result)
}

// @Summary 切换HA，按参数填入的State
// @Description 示例
// @Description {
// @Description   "state": "BACKUP",
// @Description   "permanent": faslse
// @Description }
// @Description permanent仅在原始状态是BACKUP，新状态是MASTER，设置为true才有意义
// @Description 若需还原，需手工配置或者重新设置HA
// @Tags CPE-Sys
// @Accept  json
// @Produce  json
// @Param esn path string true "Device ID"
// @Param data body common.VrrpStateVO true "状态参数，MASTER BACKUP"
// @Success 200  {string} string  "结果描述"
// @Router /v2/cpe/sys/switchvrrp/{esn} [post]
func SwitchVRRP(c *gin.Context) {
	vrrp := common.VrrpStateVO{}
	c.BindJSON(&vrrp)
	task := common.NewRequestTaskWithBody(
		c.Param("esn"),
		common.CpeTaskClass.System,
		common.CpeSystemTaskType.SwitchVRRP,
		vrrp,
	)
	result := agent.Request(task)
	c.JSON(http.StatusOK, result)
}

// @Summary 初始化CPE
// @Description 示例
// @Description {
// @Description   "keepWan": true
// @Description }
// @Tags CPE-Sys
// @Param esn path string true "Device ID"
// @Accept  json
// @Produce  json
// @Success 200  {string} string  "结果描述"
// @Router /v2/cpe/sys/init/{esn} [post]
func Init(c *gin.Context) {
	init := common.InitVO{}
	c.BindJSON(&init)
	task := common.NewRequestTaskWithBody(
		c.Param("esn"),
		common.CpeTaskClass.System,
		common.CpeSystemTaskType.Init,
		init,
	)
	result := agent.Request(task)
	c.JSON(http.StatusOK, result)
}

// @Summary 重启CPE
// @Description 无需payload
// @Tags CPE-Sys
// @Param esn path string true "Device ID"
// @Accept  json
// @Produce  json
// @Success 200  {string} string  "结果描述"
// @Router /v2/cpe/sys/reboot/{esn} [post]
func Reboot(c *gin.Context) {
	task := common.NewRequestTask(
		c.Param("esn"),
		common.CpeTaskClass.System,
		common.CpeSystemTaskType.Reboot)
	result := agent.Request(task)
	c.JSON(http.StatusOK, result)
}

// @Summary 设置基本参数
// @Description 示例1
// @Description {
// @Description "zone": "wan"
// @Description }
// @Description
// @Description 示例2
// @Description {
// @Description "zone": "vnet"
// @Description }
// @Description
// @Tags CPE-misc
// @Accept  json
// @Produce  json
// @Param esn path string true "Device ID"
// @Param data body common.BaseQosVO true "参数"
// @Success 200  {string} string  "结果描述"
// @Router /v2/cpe/qos/setup/{esn} [post]
func SetupQos(c *gin.Context) {
	bq := common.BaseQosVO{}
	c.BindJSON(&bq)
	task := common.NewRequestTaskWithBody(
		c.Param("esn"),
		common.CpeTaskClass.System,
		common.CpeSystemTaskType.SetupQos,
		bq,
	)
	result := agent.Request(task)
	c.JSON(http.StatusOK, result)
}

/*
// @Summary 彻底清理QOS
// @Description 示例1
// @Description {
// @Description "zone": "wan"
// @Description }
// @Description
// @Tags CPE-misc
// @Accept  json
// @Produce  json
// @Param esn path string true "Device ID"
// @Param data body common.BaseQosVO true "参数"
// @Success 200  {string} string  "结果描述"
// @Router /v2/cpe/qos/destroy/{esn} [post]
func DestroyQos(c *gin.Context) {
	bq := common.BaseQosVO{}
	c.BindJSON(&bq)
	task := common.NewRequestTaskWithBody(
		c.Param("esn"),
		common.CpeTaskClass.System,
		common.CpeSystemTaskType.DestroyQos,
		bq,
	)
	result := agent.Request(task)
	c.JSON(http.StatusOK, result)
}
*/

// @Summary 添加Qos规则
// @Description 示例
// @Description {
// @Description   "zone": "wan",
// @Description   "prio": 1,
// @Description   "addrRules": ["192.168.2.9","192.168.1.0/24"],
// @Description   "serviceRules": ["5201","5202"],
// @Description   "protocolRules": ["udp","icmp"]
// @Description }
// @Tags CPE-misc
// @Accept  json
// @Produce  json
// @Param esn path string true "Device ID"
// @Param data body common.QosVO true "参数"
// @Success 200  {string} string  "结果描述"
// @Router /v2/cpe/qos/addrule/{esn} [post]
func AddQosRule(c *gin.Context) {
	qos := common.QosVO{}
	c.BindJSON(&qos)
	task := common.NewRequestTaskWithBody(
		c.Param("esn"),
		common.CpeTaskClass.System,
		common.CpeSystemTaskType.AddQosRule,
		qos,
	)
	result := agent.Request(task)
	c.JSON(http.StatusOK, result)
}

// @Summary 删除Qos规则
// @Description 示例
// @Description {
// @Description   "zone": "wan",
// @Description   "prio": 1,
// @Description   "addrRules": ["192.168.1.0/24"],
// @Description   "serviceRules": ["5201","5202"],
// @Description   "protocolRules": ["udp","icmp"]
// @Description }
// @Tags CPE-misc
// @Accept  json
// @Produce  json
// @Param esn path string true "Device ID"
// @Param data body common.QosVO true "参数"
// @Success 200  {string} string  "结果描述"
// @Router /v2/cpe/qos/delrule/{esn} [post]
func DelQosRule(c *gin.Context) {
	qos := common.QosVO{}
	c.BindJSON(&qos)
	task := common.NewRequestTaskWithBody(
		c.Param("esn"),
		common.CpeTaskClass.System,
		common.CpeSystemTaskType.DelQosRule,
		qos,
	)
	result := agent.Request(task)
	c.JSON(http.StatusOK, result)
}
