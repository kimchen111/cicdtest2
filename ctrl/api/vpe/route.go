package vpe

/*
// @Summary 添加BGP发布路由进入Allowed-Ips
// @Description 示例
// @Description {
// @Description "cidrs": [
// @Description  "192.168.19.0/24"
// @Description ],
// @Description "privKey": "aKgNcMovLG/cF9B/+O/POTqqnGYoy21xynhtH+C5NVU=",
// @Description "vni": 50
// @Description }
// @Tags VPE-Route
// @Accept  json
// @Produce  json
// @Param esn path string true "Device ID"
// @Param data body common.AllowedIpsVO true "BGP发布的网段列表"
// @Success 200  {string} string  "结果描述"
// @Router /v2/vpe/route/addallowedips/{esn} [post]
func AddAllowedIps(c *gin.Context) {
	pubRoute := common.AllowedIpsVO{}
	c.BindJSON(&pubRoute)
	task := common.NewRequestTaskWithBody(
		c.Param("esn"),
		common.VpeTaskClass.Route,
		common.VpeRouteTaskType.AddAllowedIps,
		pubRoute,
	)
	result := agent.Request(task)
	c.JSON(http.StatusOK, result)
}

// @Summary 删除BGP发布路由从Allowed-Ips
// @Description 示例
// @Description {
// @Description "cidrs": [
// @Description  "192.168.19.0/24"
// @Description ],
// @Description "privKey": "aKgNcMovLG/cF9B/+O/POTqqnGYoy21xynhtH+C5NVU=",
// @Description "vni": 50
// @Description }
// @Tags VPE-Route
// @Accept  json
// @Produce  json
// @Param esn path string true "Device ID"
// @Param data body common.PublishRouteVO true "BGP发布的网段列表"
// @Success 200  {string} string  "结果描述"
// @Router /v2/vpe/route/delallowedips/{esn} [post]
func DelAllowedIps(c *gin.Context) {
	pubRoute := common.AllowedIpsVO{}
	c.BindJSON(&pubRoute)
	task := common.NewRequestTaskWithBody(
		c.Param("esn"),
		common.VpeTaskClass.Route,
		common.VpeRouteTaskType.DelAllowedIps,
		pubRoute,
	)
	result := agent.Request(task)
	c.JSON(http.StatusOK, result)
}
*/
