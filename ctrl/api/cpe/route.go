package cpe

import (
	"net/http"
	"sdwan/common"
	"sdwan/ctrl/agent"

	"github.com/gin-gonic/gin"
)

// @Summary 发布LAN网络
// @Description 示例
// @Description { "name": "lan2" }
// @Tags CPE-Route
// @Accept  json
// @Produce  json
// @Param esn path string true "Device ID"
// @Param data body common.PubLanVO true "需要1个参数：name"
// @Success 200  {string} string  "结果描述"
// @Router /v2/cpe/route/enablepublan/{esn} [post]
func EnablePubLan(c *gin.Context) {
	pl := common.PubLanVO{}
	c.BindJSON(&pl)
	task := common.NewRequestTaskWithBody(
		c.Param("esn"),
		common.CpeTaskClass.Route,
		common.CpeRouteTaskType.EnablePubLAN,
		pl,
	)
	result := agent.Request(task)
	c.JSON(http.StatusOK, result)
}

// @Summary 删除LAN网络的发布
// @Description 示例
// @Description { "name": "lan2" }
// @Tags CPE-Route
// @Accept  json
// @Produce  json
// @Param esn path string true "Device ID"
// @Param data body common.PubLanVO true "需要1个参数：name"
// @Success 200  {string} string  "结果描述"
// @Router /v2/cpe/route/disablepublan/{esn} [post]
func DisablePubLan(c *gin.Context) {
	pl := common.PubLanVO{}
	c.BindJSON(&pl)
	task := common.NewRequestTaskWithBody(
		c.Param("esn"),
		common.CpeTaskClass.Route,
		common.CpeRouteTaskType.DisablePubLAN,
		pl,
	)
	result := agent.Request(task)
	c.JSON(http.StatusOK, result)
}

// @Summary 添加静态路由
// @Description 示例
// @Description [
// @Description   {
// @Description     "target": "222.9.2.0/24"
// @Description     "via": "192.168.3.9",
// @Description     "metric": 0,,
// @Description     "publish": true
// @Description   }
// @Description ]
// @Tags CPE-Route
// @Accept  json
// @Produce  json
// @Param esn path string true "Device ID"
// @Param data body []common.StaticRouteVO true "静态路由列表"
// @Success 200  {string} string  "结果描述"
// @Router /v2/cpe/route/addstaticroute/{esn} [post]
func AddStaticRoute(c *gin.Context) {
	staticRoute := []common.StaticRouteVO{}
	c.BindJSON(&staticRoute)
	task := common.NewRequestTaskWithBody(
		c.Param("esn"),
		common.CpeTaskClass.Route,
		common.CpeRouteTaskType.AddStaticRoute,
		staticRoute,
	)
	result := agent.Request(task)
	c.JSON(http.StatusOK, result)
}

// @Summary 删除静态路由
// @Description 示例
// @Description [
// @Description   {
// @Description     "target": "222.9.2.0/24"
// @Description   }
// @Description ]
// @Tags CPE-Route
// @Accept  json
// @Produce  json
// @Param esn path string true "Device ID"
// @Param data body []common.StaticRouteVO true "静态路由列表，只需要填定Name"
// @Success 200  {string} string  "结果描述"
// @Router /v2/cpe/route/delstaticroute/{esn} [post]
func DelStaticRoute(c *gin.Context) {
	staticRoute := []common.StaticRouteVO{}
	c.BindJSON(&staticRoute)
	task := common.NewRequestTaskWithBody(
		c.Param("esn"),
		common.CpeTaskClass.Route,
		common.CpeRouteTaskType.DelStaticRoute,
		staticRoute,
	)
	result := agent.Request(task)
	c.JSON(http.StatusOK, result)
}

// @Summary 指定到一个目标IP的出口
// @Description 示例
// @Description {
// @Description  "intfName": "eth1",
// @Description  "target": "220.97.220.88/32"
// @Description }
// @Tags CPE-Route
// @Accept  json
// @Produce  json
// @Param esn path string true "Device ID"
// @Param data body common.OutPort true "出口信息"
// @Success 200  {string} string  "结果描述"
// @Router /v2/cpe/route/assignoutport/{esn} [post]
func AssignOutPort(c *gin.Context) {
	op := common.OutPort{}
	c.BindJSON(&op)
	task := common.NewRequestTaskWithBody(
		c.Param("esn"),
		common.CpeTaskClass.Route,
		common.CpeRouteTaskType.AssignOutPort,
		op,
	)
	result := agent.Request(task)
	c.JSON(http.StatusOK, result)
}

// @Summary 删除指定到一个目标IP的出口
// @Description 示例
// @Description {
// @Description  "intfName": "eth1",
// @Description  "target": "220.97.220.88/32"
// @Description }
// @Tags CPE-Route
// @Accept  json
// @Produce  json
// @Param esn path string true "Device ID"
// @Param data body common.OutPort true "出口信息"
// @Success 200  {string} string  "结果描述"
// @Router /v2/cpe/route/unassignoutport/{esn} [post]
func UnAssignOutPort(c *gin.Context) {
	op := common.OutPort{}
	c.BindJSON(&op)
	task := common.NewRequestTaskWithBody(
		c.Param("esn"),
		common.CpeTaskClass.Route,
		common.CpeRouteTaskType.UnAssignOutPort,
		op,
	)
	result := agent.Request(task)
	c.JSON(http.StatusOK, result)
}
