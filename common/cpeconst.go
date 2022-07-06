package common

//任务分类
type cpeTaskClass struct {
	System  string //系统类任务
	Network string //网络类
	Route   string //路由类
	Link    string //链路类
}

var CpeTaskClass cpeTaskClass

// System任务类型

type cpeSystemTaskType struct {
	SysInfo string //系统信息

	Reboot  string //重启
	Enable  string //启用
	Disable string //禁用
	Init    string //初始化
	Upgrade string //升级AGENT程序
	Restart string //重启AGENT程序

	EnableVRRP  string //启用HA（VRRP）
	DisableVRRP string //禁用HA（VRRP）
	SwitchVRRP  string //切换HA（VRRP）
	EnableBFD   string //启用BFD
	DisableBFD  string //禁用BFD

	AddFwRule string //添加防火墙规则
	DelFwRule string //删除防火墙规则

	SetupQos   string //QOS基本设置
	DestroyQos string //清除QOS
	AddQosRule string //添加QOS规则
	DelQosRule string //删除QOS规则
}

var CpeSystemTaskType cpeSystemTaskType

//Network
type cpeNetworkTaskType struct {
	AddLAN      string //添加LAN网段
	DelLAN      string //删除LAN网段
	EnableDHCP  string //启用DHCP服务
	DisableDHCP string //禁用DHCP服务
	NetworkInfo string //网络信息
	DetectVpe   string

	EnableTrafficAnalysis  string //开启流量分析
	DisableTrafficAnalysis string //开启流量分析
}

var CpeNetworkTaskType cpeNetworkTaskType

//Routing
type cpeRouteTaskType struct {
	SwitchDefaultRoute string //切换默认路由（Default Route）
	AddStaticRoute     string //添加静态路由
	DelStaticRoute     string //添加静态路由
	EnablePubLAN       string //发布LAN网段
	DisablePubLAN      string //取消发布LAN网段
	AssignOutPort      string //指定出口
	UnAssignOutPort    string //取消指定出口
}

var CpeRouteTaskType cpeRouteTaskType

//Link
type cpeLinkTaskType struct {
	ResetLinkState     string //切换链路
	AddTunnelEndpoint  string //新建HUB之间的TUNNEL隧道
	DelTunnelEndpoint  string //删除HUB之间的TUNNEL隧道
	AddHubMstpEndpoint string //新建HUB之间的专线链路
	DelHubMstpEndpoint string //删除HUB之间的专线链路
	SetupHub           string //新建一个HUB端点
	DestroyHub         string //删除一个HUB端点
}

var CpeLinkTaskType cpeLinkTaskType

var CpeTaskTypeSet *TaskTypeSet

func InitCpeConst() {
	resetConst(&MsgType, relectFields(MsgType))
	resetConst(&CommonTaskType, relectFields(CommonTaskType))

	resetConst(&CommonTaskClass, relectFields(CommonTaskClass))
	resetConst(&CommonLinkTaskType, relectFields(CommonLinkTaskType))

	resetConst(&CpeTaskClass, relectFields(CpeTaskClass))

	resetConst(&CpeSystemTaskType, relectFields(CpeSystemTaskType))
	resetConst(&CpeNetworkTaskType, relectFields(CpeNetworkTaskType))
	resetConst(&CpeRouteTaskType, relectFields(CpeRouteTaskType))
	resetConst(&CpeLinkTaskType, relectFields(CpeLinkTaskType))

	CpeTaskTypeSet = new(TaskTypeSet)
	CpeTaskTypeSet.Init()
	CpeTaskTypeSet.ConcatMap(relectFields(CpeSystemTaskType))
	CpeTaskTypeSet.ConcatMap(relectFields(CpeNetworkTaskType))
	CpeTaskTypeSet.ConcatMap(relectFields(CpeRouteTaskType))
	CpeTaskTypeSet.ConcatMap(relectFields(CpeLinkTaskType))
	CpeTaskTypeSet.ConcatMap(relectFields(CommonLinkTaskType))
}
