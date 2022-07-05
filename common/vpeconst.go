package common

//任务分类
type vpeTaskClass struct {
	Vnet  string //Vnet
	Link  string //网络隧道，VPN连接
	Route string //网络隧道，VPN连接
}

var VpeTaskClass vpeTaskClass

//虚拟网
type vpeVnetTaskType struct {
	SetVnetEndpoint          string //设置一个VNET的隧道端点
	SetVnetEndpointReflector string //设置一个VNET的RR隧道端点
	// AddCpeNeighbor string //添加对端CPE的配置
	// DelCpeNeighbor string //添加对端CPE的配置
}

var VpeVnetTaskType vpeVnetTaskType

//VPN，TUNNEL
type vpeLinkTaskType struct {
	// AddTunnelEndpoint string //创建VPE的Tunnel端点
	// DelTunnelEndpoint string //删除VPE的Tunnel端点
}

var VpeLinkTaskType vpeLinkTaskType

//Routing
type vpeRouteTaskType struct {
	// AddAllowedIps string //添加发布网段
	// DelAllowedIps string //删除发布网段
}

var VpeRouteTaskType vpeRouteTaskType

var VpeTaskTypeSet *TaskTypeSet

func InitVpeConst() {
	resetConst(&MsgType, relectFields(MsgType))
	resetConst(&CommonTaskType, relectFields(CommonTaskType))

	resetConst(&CommonTaskClass, relectFields(CommonTaskClass))
	resetConst(&CommonLinkTaskType, relectFields(CommonLinkTaskType))

	resetConst(&VpeTaskClass, relectFields(VpeTaskClass))

	resetConst(&VpeVnetTaskType, relectFields(VpeVnetTaskType))
	resetConst(&VpeLinkTaskType, relectFields(VpeLinkTaskType))
	resetConst(&VpeRouteTaskType, relectFields(VpeRouteTaskType))

	VpeTaskTypeSet = new(TaskTypeSet)
	VpeTaskTypeSet.Init()
	VpeTaskTypeSet.ConcatMap(relectFields(VpeVnetTaskType))
	VpeTaskTypeSet.ConcatMap(relectFields(VpeLinkTaskType))
	VpeTaskTypeSet.ConcatMap(relectFields(VpeRouteTaskType))
	VpeTaskTypeSet.ConcatMap(relectFields(CommonLinkTaskType))
}
