package common

type SysInfoVO struct {
	Esn       string `json:"esn"`       //ESN
	Version   string `json:"version"`   //版本
	AgentType string `json:"agentType"` //AGENT类型 CPE/VPE/HUB
	RouterID  string `json:"routerID"`  //Loop back ip
}

type ApiResult struct {
	Esn    string `json:"esn"` //ESN
	Status string `json:"status"`
	Body   any    `json:"body"`
}

type AddrInfo struct {
	Family    string `json:"family"`    //IPV4/IPV6
	Ipaddr    string `json:"local"`     //本地IP地址
	Prefixlen int    `json:"prefixlen"` //掩码长度
}

type IntfAddrInfo struct {
	Addrinfos []AddrInfo `json:"addr_info"`
}
