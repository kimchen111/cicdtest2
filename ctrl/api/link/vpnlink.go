package link

import (
	"net/http"
	"sdwan/common"
	"sdwan/ctrl/agent"
	"sdwan/ctrl/device"

	"github.com/gin-gonic/gin"
)

func initVpnlinkRole(vpnlink *common.VpnlinkVO) {
	server := device.GetGDM().GetDevice(vpnlink.Server.Esn)
	vpnlink.Server.Role = server.SysInfo.AgentType
	client := device.GetGDM().GetDevice(vpnlink.Client.Esn)
	vpnlink.Client.Role = client.SysInfo.AgentType
}

// @Summary 在两个节点之间建立Vpn链路
// @Description 示例：
// @Description {
// @Description "id": 102,
// @Description "vni": 50,
// @Description "rate": 10,
// @Description "state": "PRIMARY",
// @Description "client": {
// @Description 	"esn": "0c6661fd0000",
// @Description 	"intfAddr": "10.0.11.2/24"
// @Description },
// @Description "server": {
// @Description 	"esn": "0cf391b30000",
// @Description 	"listenAddr": "10.2.0.53",
// @Description 	"listenIntf": "eth7",
// @Description 	"intfAddr": "10.0.11.1/24"
// @Description }
// @Description }
// @Tags Link
// @Accept  json
// @Produce  json
// @Param data body common.VpnlinkVO true "Tunnel信息"
// @Success 200  {string} string  "结果描述"
// @Router /v2/link/createvpnlink [post]
func CreateVpnlink(c *gin.Context) {
	vpnlink := common.VpnlinkVO{}
	c.BindJSON(&vpnlink)
	vpnlink.InitKey()
	initVpnlinkRole(&vpnlink)

	if vpnlink.Server.Role == "CPE" {
		c.JSON(http.StatusOK, common.ApiResult{Status: "error", Body: "failed: Role error"})
		return
	}

	server := common.NewRequestTaskWithBody(
		vpnlink.Server.Esn,
		common.CommonTaskClass.Link,
		common.CommonLinkTaskType.AddVpnEndpoint,
		vpnlink,
	)
	r_server := agent.Request(server)

	client := common.NewRequestTaskWithBody(
		vpnlink.Client.Esn,
		common.CommonTaskClass.Link,
		common.CommonLinkTaskType.AddVpnEndpoint,
		vpnlink,
	)
	r_client := agent.Request(client)

	c.JSON(http.StatusOK, [...]common.ApiResult{r_server, r_client})
	// c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// @Summary 删除指定的vpnlink
// @Description 示例：
// @Description {
// @Description "id": 102,
// @Description "vni": 50,
// @Description "client": {
// @Description 	"esn": "0c6661fd0000"
// @Description },
// @Description "server": {
// @Description 	"esn": "0cf391b30000"
// @Description }
// @Description }
// @Tags Link
// @Accept  json
// @Produce  json
// @Param data body common.VpnlinkVO true "Tunnel信息"
// @Success 200  {string} string  "结果描述"
// @Router /v2/link/removevpnlink [post]
func RemoveVpnlink(c *gin.Context) {
	vpnlink := common.VpnlinkVO{}
	c.BindJSON(&vpnlink)
	server := common.NewRequestTaskWithBody(
		vpnlink.Server.Esn,
		common.CommonTaskClass.Link,
		common.CommonLinkTaskType.DelVpnEndpoint,
		vpnlink,
	)
	vpe_r := agent.Request(server)

	client := common.NewRequestTaskWithBody(
		vpnlink.Client.Esn,
		common.CommonTaskClass.Link,
		common.CommonLinkTaskType.DelVpnEndpoint,
		vpnlink,
	)
	cpe_r := agent.Request(client)

	c.JSON(http.StatusOK, [...]common.ApiResult{vpe_r, cpe_r})
}

// @Summary 设置指定CPE上两条VPNLINK的优先级
// @Description 示例
// @Description {
// @Description "plink":{
// @Description  "id": 213,
// @Description  "state": "SECONDARY"
// @Description },
// @Description "slink":{
// @Description  "id": 214,
// @Description  "state": "PRIMARY"
// @Description }
// @Description }
// @Tags Link
// @Accept  json
// @Produce  json
// @Param esn path string true "Device ID"
// @Param data body common.VpnlinkStateVO true "LINK的优先级描述"
// @Success 200  {string} string  "结果描述"
// @Router /v2/link/resetlinkstate/{esn} [post]
func ResetLinkState(c *gin.Context) {
	rv := common.ResetVpnlinkVO{}
	c.BindJSON(&rv)
	task := common.NewRequestTaskWithBody(
		c.Param("esn"),
		common.CpeTaskClass.Link,
		common.CpeLinkTaskType.ResetLinkState,
		rv,
	)
	result := agent.Request(task)
	c.JSON(http.StatusOK, result)
}
