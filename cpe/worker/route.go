package worker

import (
	"sdwan/common"
)

type Route struct {
}

func (routing Route) Name() string {
	return "Route"
}

func (routing Route) SwitchDefaultRoute(body string) {

} //切换默认路由（Default Route）

/*
uci add network route # =cfg09c8b4
uci set network.@route[-1].target='222.73.99.0/24'
uci set network.@route[-1].gateway='172.18.16.1'
uci set network.@route[-1].metric='20'

network.@route[0]=route
network.@route[0].target='222.73.9.0/24'
network.@route[0].gateway='192.168.3.9'
network.@route[0].onlink='1'
network.@route[0].interface='lan'
network.@route[0].metric='20'
*/
/*
func (routing Route) AddStaticRoute(msg common.Message) {
	srList := []common.StaticRouteVO{}
	if err := common.LoadBody(msg.Body, &srList); err == nil {
		for _, sr := range srList {
			if _, ok := uci.Get("network", sr.Name, "target"); ok {
				uci.DelSection("network", sr.Name)
			}
			uci.AddSection("network", sr.Name, "route")
			uci.Set("network", sr.Name, "interface", sr.LanName)
			uci.Set("network", sr.Name, "target", sr.Target)
			uci.Set("network", sr.Name, "gateway", sr.Gateway)
			uci.Set("network", sr.Name, "onlink", "1")
		}
		uci.Commit()
		ReloadNetwork()
		Response(msg.ToResult("success"))
		return
	}
	Response(msg.ToResult("failed: unknown error"))
} //添加静态路由

func (routing Route) DelStaticRoute(msg common.Message) {
	srList := []common.StaticRouteVO{}
	if err := common.LoadBody(msg.Body, &srList); err == nil {
		for _, sr := range srList {
			if _, ok := uci.Get("network", sr.Name, "target"); ok {
				uci.DelSection("network", sr.Name)
			}
		}
		uci.Commit()
		ReloadNetwork()
		Response(msg.ToResult("success"))
		return
	}
	Response(msg.ToResult("failed: unknown error"))
} //添加静态路由
*/
func (routing Route) EnableReservRoute(msg common.Message) {

} //允许/添加保留静态路由

func (routing Route) DisableReservRoute(msg common.Message) {

} //禁用/删除保留静态路由

func (routing Route) EnablePubLAN(msg common.Message) {
	pl := common.PubLanVO{}
	if err := common.LoadBody(msg.Body, &pl); err == nil {
		content := pl.GenContent()
		common.WriteFile(pl.IncludeFilePath(), content)
		EnableBirdInclude(common.CpeBirdConfPath, pl.IncludeFilePath())
		Response(msg.ToResult("success"))
		return
	}
	Response(msg.ToResult("failed: unknown error"))
} //允许/添加保留静态路由

func (routing Route) DisablePubLAN(msg common.Message) {
	pl := common.PubLanVO{}
	if err := common.LoadBody(msg.Body, &pl); err == nil {
		DisableBirdInclude(common.CpeBirdConfPath, pl.IncludeFilePath())
		Response(msg.ToResult("success"))
		return
	}
	Response(msg.ToResult("failed: unknown error"))
} //禁用/删除保留静态路由

func (routing Route) AddStaticRoute(msg common.Message) {
	srList := make([]common.StaticRouteVO, 100)
	if err := common.LoadBody(msg.Body, &srList); err == nil {
		len := len(srList)
		paths := make([]string, len)
		for i, sr := range srList {
			common.WriteFile(sr.IncludeFilePath(), sr.Gen())
			paths[i] = sr.IncludeFilePath()
		}
		BatchEnableBirdIncludes(common.CpeBirdStaticPath, paths[:len])
		Response(msg.ToResult("success"))
		return
	}
	Response(msg.ToResult("failed: unknown error"))
} //添加静态路由

func (routing Route) DelStaticRoute(msg common.Message) {
	srList := make([]common.StaticRouteVO, 100)
	if err := common.LoadBody(msg.Body, &srList); err == nil {
		len := len(srList)
		paths := make([]string, len)
		for i, sr := range srList {
			paths[i] = sr.IncludeFilePath()
		}
		BatchDisableBirdIncludes(common.CpeBirdStaticPath, paths[:len])
		Response(msg.ToResult("success"))
		return
	}
	Response(msg.ToResult("failed: unknown error"))
} //删除静态路由

func (routing Route) AssignOutPort(msg common.Message) {
	op := common.OutPort{}
	if err := common.LoadBody(msg.Body, &op); err == nil {
		content := op.GenContent()
		common.WriteFile(op.IncludeFilePath(), content)
		EnableBirdStaticInclude(op.IncludeFilePath())
		Response(msg.ToResult("success"))
		return
	}
	Response(msg.ToResult("failed: unknown error"))
} //添加静态路由

func (routing Route) UnAssignOutPort(msg common.Message) {
	op := common.OutPort{}
	if err := common.LoadBody(msg.Body, &op); err == nil {
		DisableBirdStaticInclude(op.IncludeFilePath())
		Response(msg.ToResult("success"))
		return
	}
	Response(msg.ToResult("failed: unknown error"))
} //删除静态路由
