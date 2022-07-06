package api

import (
	"net/http"
	"sdwan/common"
	"sdwan/ctrl/api/cpe"
	"sdwan/ctrl/api/link"
	"sdwan/ctrl/api/vpe"
	"sdwan/ctrl/device"

	"github.com/gin-gonic/gin"
)

func RunAPI() {
	r := gin.Default()
	r.Static("/bin", "./bin")
	r.Static("/tests", "./tests")

	RouteSwagger(r)
	//VPE 相关API
	RouteVPE(r)
	//CPE 相关API
	RouteCPE(r)
	//Agent 相关API
	RouteAgent(r)
	//HUB-HUB HUB-VPE TUNNEL 相关API
	//CPE-VPE CPE-HUB VPN 相关API
	RouteLink(r)
	r.Run(common.GCTC.GetApiListen())
}

func RouteAgent(r *gin.Engine) {
	agent := r.Group("/v2/agent")
	{
		agent.GET("/sysinfo/:esn", SysInfo)
	}
}

func RouteLink(r *gin.Engine) {
	linkg := r.Group("/v2/link")
	{
		linkg.POST("/createcpevpn", link.CreateCpeVpn)
		linkg.POST("/removecpevpn", link.RemoveCpeVpn)
		linkg.POST("/createcpemstp", link.CreateCpeMstp)
		linkg.POST("/removecpemstp", link.RemoveCpeMstp)

		linkg.POST("/createhubtunnel", link.CreateHubTunnel)
		linkg.POST("/removehubtunnel", link.RemoveHubTunnel)
		linkg.POST("/createhubmstp", link.CreateHubMstp)
		linkg.POST("/removehubmstp", link.RemoveHubMstp)

		linkg.POST("/resetlinkstate/:esn", link.ResetLinkState)
	}
	r.POST("/v2/hub/setup/:esn", link.SetupHub)
	r.POST("/v2/hub/destroy/:esn", link.DestroyHub)

}

func RouteCPE(r *gin.Engine) {
	sys := r.Group("/v2/cpe/sys")
	{
		sys.POST("/init/:esn", cpe.Init)
		sys.POST("/reboot/:esn", cpe.Reboot)
		sys.POST("/enablehagroup", cpe.EnableHAGroup)
		sys.POST("/disablevrrp/:esn", cpe.DisableVRRP)
		sys.POST("/switchvrrp/:esn", cpe.SwitchVRRP)
	}
	net := r.Group("/v2/cpe/net")
	{
		net.GET("/netinfo/:esn", cpe.NetworkInfo)
		net.POST("/addlan/:esn", cpe.AddLan)
		net.POST("/dellan/:esn", cpe.DelLan)
		net.POST("/enabledhcp/:esn", cpe.EnableDHCP)
		net.POST("/disabledhcp/:esn", cpe.DisableDHCP)
		net.POST("/detectvpe/:esn", cpe.DetectVpe)
		net.POST("/enabletrafficanalysis/:esn", cpe.EnableTrafficAnalysis)
		net.POST("/disabletrafficanalysis/:esn", cpe.DisableTrafficAnalysis)
	}
	route := r.Group("/v2/cpe/route")
	{
		route.POST("/enablepublan/:esn", cpe.EnablePubLan)
		route.POST("/disablepublan/:esn", cpe.DisablePubLan)
		route.POST("/addstaticroute/:esn", cpe.AddStaticRoute)
		route.POST("/delstaticroute/:esn", cpe.DelStaticRoute)
		route.POST("/assignoutport/:esn", cpe.AssignOutPort)
		route.POST("/unassignoutport/:esn", cpe.UnAssignOutPort)
	}
	/*
		firewall := r.Group("/v2/firewall")
		{
			firewall.POST("/addrule", cpe.AddFirewallRule)
			firewall.POST("/addwrule", cpe.DelFirewallRule)
		}
	*/
	qos := r.Group("/v2/cpe/qos")
	{
		qos.POST("/setup/:esn", cpe.SetupQos)
		// qos.POST("/destroy/:esn", cpe.DestroyQos)
		qos.POST("/addrule/:esn", cpe.AddQosRule)
		qos.POST("/delrule/:esn", cpe.DelQosRule)
	}
}

func RouteVPE(r *gin.Engine) {
	vpenet := r.Group("/v2/vpe/vnet")
	{
		vpenet.POST("/setvnetendpoint", vpe.SetVnetEndpoint)
		// vpenet.POST("/delvnetendpoint/:esn", vpe.DelVnetEndpoint)
		// vpenet.POST("/addcpeneighbor/:esn", vpe.AddCpeNeighbor)
		// vpenet.POST("/delcpeneighbor/:esn", vpe.DelCpeNeighbor)
	}
	/*
		route := r.Group("/v2/vpe/route")
		{
			route.POST("/addallowedips/:esn", vpe.AddAllowedIps)
			route.POST("/delallowedips/:esn", vpe.DelAllowedIps)
		}
	*/
	// vpelink := r.Group("/v2/vpe/link")
	// {
	// 	vpelink.POST("/addvpnclient/:esn", vpe.AddVpnClient)
	// }
}

// @Summary 系统信息
// @Description 获取系统信息
// @Tags Agent
// @Param esn path string true "Device ID"
// @Accept  json
// @Produce  json
// @Success 200  {object} common.Device  "查询成功的body"
// @Router /v2/agent/sysinfo/{esn} [get]
func SysInfo(c *gin.Context) {
	result := device.GetGDM().GetDevice(c.Param("esn"))
	c.JSON(http.StatusOK, result)
}
