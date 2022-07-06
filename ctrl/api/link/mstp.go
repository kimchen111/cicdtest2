package link

import (
	"net/http"
	"sdwan/common"
	"sdwan/ctrl/agent"
	"sdwan/ctrl/device"

	"github.com/gin-gonic/gin"
)

func initMstpRole(dl *common.MstpVO) {
	server := device.GetGDM().GetDevice(dl.Server.Esn)
	dl.Server.Role = server.SysInfo.AgentType
	client := device.GetGDM().GetDevice(dl.Client.Esn)
	dl.Client.Role = client.SysInfo.AgentType
}

// @Summary 在两个节点之间建立专线链路（CPE/HUB-VPE，CPE-HUB）
// @Description 示例：
// @Description CPE-VPE/HUB
// @Description {
// @Description "id": 102,
// @Description	"vni": 50,
// @Description "state": "PRIMARY",
// @Description "client": {
// @Description 	"esn": "0c6661fd0000",
// @Description 	"intfName": "eth4",
// @Description 	"vlanId": 1,
// @Description 	"intfAddr": "10.0.11.2/24"
// @Description },
// @Description "server": {
// @Description 	"esn": "0cf391b30000",
// @Description 	"intfName": "enp3s0",
// @Description 	"vlanId": 2022,
// @Description 	"intfAddr": "10.0.11.2/24"
// @Description }
// @Description }
// @Tags Link
// @Accept  json
// @Produce  json
// @Param data body common.MstpVO true "专线信息"
// @Success 200  {string} string  "结果描述"
// @Router /v2/link/createcpemstp [post]
func CreateCpeMstp(c *gin.Context) {
	dl := common.MstpVO{}
	c.BindJSON(&dl)
	initMstpRole(&dl)

	if dl.Server.Role == "CPE" {
		c.JSON(http.StatusOK, common.ApiResult{Status: "error", Body: "failed: Role error"})
		return
	}

	server := common.NewRequestTaskWithBody(
		dl.Server.Esn,
		common.CommonTaskClass.Link,
		common.CommonLinkTaskType.AddMstpEndpoint,
		dl,
	)
	r_server := agent.Request(server)

	client := common.NewRequestTaskWithBody(
		dl.Client.Esn,
		common.CommonTaskClass.Link,
		common.CommonLinkTaskType.AddMstpEndpoint,
		dl,
	)
	r_client := agent.Request(client)

	c.JSON(http.StatusOK, [...]common.ApiResult{r_server, r_client})
	// c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// @Summary 删除指定的专线链路
// @Description 示例：
// @Description {
// @Description "id": 102,
// @Description	"vni": 50,
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
// @Param data body common.MstpVO true "专线信息"
// @Success 200  {string} string  "结果描述"
// @Router /v2/link/removecpemstp [post]
func RemoveCpeMstp(c *gin.Context) {
	dl := common.MstpVO{}
	c.BindJSON(&dl)
	server := common.NewRequestTaskWithBody(
		dl.Server.Esn,
		common.CommonTaskClass.Link,
		common.CommonLinkTaskType.DelMstpEndpoint,
		dl,
	)
	vpe_r := agent.Request(server)

	client := common.NewRequestTaskWithBody(
		dl.Client.Esn,
		common.CommonTaskClass.Link,
		common.CommonLinkTaskType.DelMstpEndpoint,
		dl,
	)
	cpe_r := agent.Request(client)

	c.JSON(http.StatusOK, [...]common.ApiResult{vpe_r, cpe_r})
}
