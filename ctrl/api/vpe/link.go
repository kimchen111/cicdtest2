package vpe

/*
import (
	"net/http"
	"sdwan/common"
	"sdwan/ctrl/agent"

	"github.com/gin-gonic/gin"
)

// @Summary 添加一条VPN客户端
// @Description 示例
// @Description {
// @Description "allowedIPs": [
// @Description  "10.0.3.2"
// @Description ],
// @Description "pubKey": "9GqOs9IAl3IvEkmv87/vhrCp6XqW062NCluZ3zMqGlA=",
// @Description "vni": 50
// @Description }
// @Tags VPE-Link
// @Accept  json
// @Produce  json
// @Param esn path string true "Device ID"
// @Param data body common.WgClientVO true "Vpn客户端信息"
// @Success 200  {string} string  "结果描述"
// @Router /v2/vpe/link/addvpnclient/{esn} [post]
func AddVpnClient(c *gin.Context) {
	client := common.WgClientVO{}
	c.BindJSON(&client)
	task := common.NewRequestTaskWithBody(
		c.Param("esn"),
		common.VpeTaskClass.Link,
		common.VpeLinkTaskType.AddVpnClient,
		client,
	)
	result := agent.Request(task)
	c.JSON(http.StatusOK, result)
}

*/
