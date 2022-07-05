package vpe

import (
	"net/http"
	"sdwan/common"
	"sdwan/ctrl/agent"

	"github.com/gin-gonic/gin"
)

// @Summary 设置一个VNET的隧道端点
// @Description 示例
// @Description {
// @Description "vni": 50,
// @Description "cpeCidr": "10.254.0.0/16",
// @Description "hubCidr": "10.253.0.0/16",
// @Description "reflectors": [{
// @Description 	"esn": "0c7d0b880000",
// @Description 	"action": "ADD",
// @Description 	"vtepAddr": "10.16.16.10/24"
// @Description 	}],
// @Description "vteps": [
// @Description 	{
// @Description 	"esn": "0c616d6c0000",
// @Description 	"action": "ADD",
// @Description 	"vtepAddr": "10.16.16.11/24"
// @Description 	},
// @Description 	{
// @Description 	"esn": "0c516aef0000",
// @Description 	"action": "ADD",
// @Description 	"vtepAddr": "10.16.16.12/24"
// @Description 	},
// @Description 	{
// @Description 	"esn": "0cc1f2ce0000",
// @Description 	"action": "ADD",
// @Description 	"vtepAddr": "10.16.16.13/24"
// @Description 	},
// @Description 	{
// @Description 	"esn": "0c5e0a060000",
// @Description 	"action": "ADD",
// @Description 	"vtepAddr": "10.16.16.14/24"
// @Description 	}
// @Description ]
// @Description }
// @Tags VPE-Vnet
// @Accept  json
// @Produce  json
// @Param data body common.VnetVO true "虚拟网的端点参数"
// @Success 200  {string} string  "结果描述"
// @Router /v2/vpe/vnet/setvnetendpoint [post]
func SetVnetEndpoint(c *gin.Context) {
	vnet := common.VnetVO{}
	c.BindJSON(&vnet)
	UpdateRouterIP(&vnet)
	total := len(vnet.Reflectors) + len(vnet.Vteps)
	arrs := make([]common.ApiResult, total)
	idx := 0
	for i, vtep := range vnet.Vteps {
		task := common.NewRequestTaskWithBody(
			vtep.Esn,
			common.VpeTaskClass.Vnet,
			common.VpeVnetTaskType.SetVnetEndpoint,
			vnet.GenEndpoint(i))
		result := agent.Request(task)
		arrs[idx] = result
		idx++
	}

	for i, rr := range vnet.Reflectors {
		task := common.NewRequestTaskWithBody(
			rr.Esn,
			common.VpeTaskClass.Vnet,
			common.VpeVnetTaskType.SetVnetEndpointReflector,
			vnet.GenRR(i))
		result := agent.Request(task)
		arrs[idx] = result
		idx++
	}
	c.JSON(http.StatusOK, arrs[:total])
}
