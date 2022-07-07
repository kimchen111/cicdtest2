package link

import (
	"net/http"
	"sdwan/common"
	"sdwan/ctrl/agent"
	"sdwan/ctrl/device"

	"github.com/gin-gonic/gin"
)

func initMstpRole(dl *common.MstpVO) {
	peerA := device.GetGDM().GetDevice(dl.PeerA.Esn)
	dl.PeerA.Role = peerA.SysInfo.AgentType
	peerB := device.GetGDM().GetDevice(dl.PeerB.Esn)
	dl.PeerB.Role = peerB.SysInfo.AgentType
}

// @Summary 在两个节点之间建立专线链路（CPE/HUB-VPE，CPE-HUB）
// @Description 示例：
// @Description CPE-VPE/HUB
// @Description {
// @Description "id": 102,
// @Description	"vni": 50,
// @Description "state": "PRIMARY",
// @Description "peerA": {
// @Description 	"esn": "0c6661fd0000",
// @Description 	"intfName": "eth4",
// @Description 	"vlanId": 1,
// @Description 	"intfAddr": "10.0.11.2/24"
// @Description },
// @Description "PeerB": {
// @Description 	"esn": "0cf391b30000",
// @Description 	"intfName": "enp3s0",
// @Description 	"vlanId": 2022,
// @Description 	"intfAddr": "10.0.11.1/24"
// @Description }
// @Description }
// @Description
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

	// if dl.PeerA.Role == "VPE" || (dl.PeerA.Role == "HUB" && dl.PeerB.Role == "HUB") {
	// 	c.JSON(http.StatusOK, common.ApiResult{Status: "error", Body: "failed: Role error"})
	// 	return
	// }

	peerA := common.NewRequestTaskWithBody(
		dl.PeerA.Esn,
		common.CommonTaskClass.Link,
		common.CommonLinkTaskType.AddMstpEndpoint,
		dl,
	)
	r_A := agent.Request(peerA)

	peerB := common.NewRequestTaskWithBody(
		dl.PeerB.Esn,
		common.CommonTaskClass.Link,
		common.CommonLinkTaskType.AddMstpEndpoint,
		dl,
	)
	r_B := agent.Request(peerB)

	c.JSON(http.StatusOK, [...]common.ApiResult{r_A, r_B})
	// c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// @Summary 删除指定的专线链路
// @Description 示例：
// @Description {
// @Description "id": 102,
// @Description	"vni": 50,
// @Description "peerA": {
// @Description 	"esn": "0c6661fd0000"
// @Description },
// @Description "peerB": {
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
	A := common.NewRequestTaskWithBody(
		dl.PeerA.Esn,
		common.CommonTaskClass.Link,
		common.CommonLinkTaskType.DelMstpEndpoint,
		dl,
	)
	rA := agent.Request(A)

	B := common.NewRequestTaskWithBody(
		dl.PeerB.Esn,
		common.CommonTaskClass.Link,
		common.CommonLinkTaskType.DelMstpEndpoint,
		dl,
	)
	rB := agent.Request(B)

	c.JSON(http.StatusOK, [...]common.ApiResult{rA, rB})
}
