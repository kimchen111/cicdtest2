package link

import (
	"net/http"
	"sdwan/common"
	"sdwan/ctrl/agent"
	"sdwan/ctrl/device"

	"github.com/gin-gonic/gin"
)

func initTunnelRole(tunnel *common.TunnelVO) {
	a := device.GetGDM().GetDevice(tunnel.PeerA.Esn)
	tunnel.PeerA.Role = a.SysInfo.AgentType
	b := device.GetGDM().GetDevice(tunnel.PeerB.Esn)
	tunnel.PeerB.Role = b.SysInfo.AgentType
}

func initHubDirlinkRole(hdl *common.HubMstpVO) {
	a := device.GetGDM().GetDevice(hdl.PeerA.Esn)
	hdl.PeerA.Role = a.SysInfo.AgentType
	b := device.GetGDM().GetDevice(hdl.PeerB.Esn)
	hdl.PeerB.Role = b.SysInfo.AgentType
}

// @Summary 在两个站点之间建立TUNNEL隧道（HUB-HUB）
// @Description 示例：
// @Description {
// @Description "id": 321,
// @Description "vni": 50,
// @Description "peerA": {
// @Description 	"esn": "00155d380114",
// @Description 	"ipaddr": "192.168.3.254"
// @Description },
// @Description "peerB": {
// @Description 	"esn": "00155d38010a",
// @Description 	"ipaddr": "192.168.3.1"
// @Description }
// @Description }
// @Description 仅在HUB与HUB之间使用，因此必须先setuphub
// @Tags Link
// @Accept  json
// @Produce  json
// @Param data body common.TunnelVO true "Tunnel信息"
// @Success 200  {string} string  "结果描述"
// @Router /v2/link/createhubtunnel [post]
func CreateHubTunnel(c *gin.Context) {
	tunnel := common.TunnelVO{}
	c.BindJSON(&tunnel)

	initTunnelRole(&tunnel)

	if tunnel.PeerA.Role == "CPE" || tunnel.PeerB.Role == "CPE" {
		c.JSON(http.StatusOK, common.ApiResult{Status: "error", Body: "failed: Role error"})
		return
	}

	vta := tunnel.MixVOA()
	vtb := tunnel.MixVOB()
	ra := createTunnel(vta)
	rb := createTunnel(vtb)

	c.JSON(http.StatusOK, [...]common.ApiResult{ra, rb})
}

func createTunnel(vt common.VxlanTunnelVO) common.ApiResult {
	task := common.NewRequestTaskWithBody(
		vt.Esn,
		common.CpeTaskClass.Link,
		common.CpeLinkTaskType.AddTunnelEndpoint,
		vt,
	)
	return agent.Request(task)
}

// @Summary 删除两个站点之间的TUNNEL隧道
// @Description 示例：
// @Description {
// @Description "id": 321,
// @Description "peerA": {
// @Description 	"esn": "00155d380114"
// @Description },
// @Description "peerB": {
// @Description 	"esn": "00155d38010a"
// @Description }
// @Description }
// @Tags Link
// @Accept  json
// @Produce  json
// @Param data body common.TunnelVO true "Tunnel信息"
// @Success 200  {string} string  "结果描述"
// @Router /v2/link/removehubtunnel [post]
func RemoveHubTunnel(c *gin.Context) {
	tunnel := common.TunnelVO{}
	c.BindJSON(&tunnel)

	initTunnelRole(&tunnel)
	if tunnel.PeerA.Role == "CPE" || tunnel.PeerB.Role == "CPE" {
		c.JSON(http.StatusOK, common.ApiResult{Status: "error", Body: "failed: Role error"})
		return
	}

	vta := tunnel.MixVOA()
	vtb := tunnel.MixVOB()
	ra := removeTunnel(vta)
	rb := removeTunnel(vtb)
	c.JSON(http.StatusOK, [...]common.ApiResult{ra, rb})
}

func removeTunnel(vt common.VxlanTunnelVO) common.ApiResult {
	task := common.NewRequestTaskWithBody(
		vt.Esn,
		common.CpeTaskClass.Link,
		common.CpeLinkTaskType.DelTunnelEndpoint,
		vt,
	)
	return agent.Request(task)

}

// @Summary 在两个站点之间设置专线链路（HUB-HUB）
// @Description 示例：
// @Description {
// @Description "id": 321,
// @Description "vni": 50,
// @Description "peerA": {
// @Description 	"esn": "00155d380114",
// @Description 	"intfName": "enp2s0",
// @Description 	"vlanId": 2202
// @Description },
// @Description "peerB": {
// @Description 	"esn": "00155d38010a",
// @Description 	"intfName": "enp2s0",
// @Description 	"vlanId": 2202
// @Description }
// @Description }
// @Description 仅在HUB与HUB之间使用，因此，必须先setuphub
// @Tags Link
// @Accept  json
// @Produce  json
// @Param data body common.HubMstpVO true "专线信息"
// @Success 200  {string} string  "结果描述"
// @Router /v2/link/createhubmstp [post]
func CreateHubMstp(c *gin.Context) {
	hdl := common.HubMstpVO{}
	c.BindJSON(&hdl)
	initHubDirlinkRole(&hdl)

	if hdl.PeerA.Role == "CPE" || hdl.PeerB.Role == "CPE" {
		c.JSON(http.StatusOK, common.ApiResult{Status: "error", Body: "failed: Role error"})
		return
	}

	peera := common.NewRequestTaskWithBody(
		hdl.PeerA.Esn,
		common.CpeTaskClass.Link,
		common.CpeLinkTaskType.AddHubMstpEndpoint,
		hdl,
	)
	r_peera := agent.Request(peera)

	peerb := common.NewRequestTaskWithBody(
		hdl.PeerB.Esn,
		common.CpeTaskClass.Link,
		common.CpeLinkTaskType.AddHubMstpEndpoint,
		hdl,
	)
	r_peerb := agent.Request(peerb)

	c.JSON(http.StatusOK, [...]common.ApiResult{r_peera, r_peerb})
	// c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// @Summary 删除指定的专线链路配置
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
// @Param data body common.HubMstpVO true "专线信息"
// @Success 200  {string} string  "结果描述"
// @Router /v2/link/removehubmstp [post]
func RemoveHubMstp(c *gin.Context) {
	hdl := common.HubMstpVO{}
	c.BindJSON(&hdl)
	peera := common.NewRequestTaskWithBody(
		hdl.PeerA.Esn,
		common.CpeTaskClass.Link,
		common.CpeLinkTaskType.DelHubMstpEndpoint,
		hdl,
	)
	ra := agent.Request(peera)

	peerb := common.NewRequestTaskWithBody(
		hdl.PeerB.Esn,
		common.CpeTaskClass.Link,
		common.CpeLinkTaskType.DelHubMstpEndpoint,
		hdl,
	)
	rb := agent.Request(peerb)

	c.JSON(http.StatusOK, [...]common.ApiResult{ra, rb})
}

// func createTunnel(vt common.VxlanTunnelVO) common.ApiResult {
// 	task := common.NewRequestTaskWithBody(
// 		vt.Esn,
// 		common.CpeTaskClass.Link,
// 		common.CpeLinkTaskType.AddTunnelEndpoint,
// 		vt,
// 	)
// 	return agent.Request(task)
// }

// @Summary 设置HUB端点信息
// @Description 示例：
// @Description {
// @Description "vni": 50,
// @Description "role": "HUB",
// @Description "rrAddrs": ["10.16.16.10"],
// @Description "vtepAddr": "10.16.16.20/24",
// @Description "cpeCidr": "10.254.0.0/16"
// @Description }
// @Description 本地的vtepAddr必须要CIDR形式
// @Tags Link
// @Accept  json
// @Produce  json
// @Param esn path string true "Device ID"
// @Param data body common.HubVO true "Hub信息"
// @Success 200  {string} string  "结果描述"
// @Router /v2/hub/setup/{esn} [post]
func SetupHub(c *gin.Context) {
	hvo := common.HubVO{}
	c.BindJSON(&hvo)
	task := common.NewRequestTaskWithBody(
		c.Param("esn"),
		common.CpeTaskClass.Link,
		common.CpeLinkTaskType.SetupHub,
		hvo,
	)
	result := agent.Request(task)
	c.JSON(http.StatusOK, result)
}

// @Summary 清理HUB端点信息
// @Description 示例：
// @Description {
// @Description "vni": 50
// @Description }
// @Tags Link
// @Accept  json
// @Produce  json
// @Param esn path string true "Device ID"
// @Param data body common.HubVO true "Hub信息"
// @Success 200  {string} string  "结果描述"
// @Router /v2/hub/destroy/{esn} [post]
func DestroyHub(c *gin.Context) {
	hvo := common.HubVO{}
	c.BindJSON(&hvo)
	task := common.NewRequestTaskWithBody(
		c.Param("esn"),
		common.CpeTaskClass.Link,
		common.CpeLinkTaskType.DestroyHub,
		hvo,
	)
	result := agent.Request(task)
	c.JSON(http.StatusOK, result)
}
