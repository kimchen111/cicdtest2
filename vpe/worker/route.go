package worker

type Route struct {
}

func (routing Route) Name() string {
	return "Route"
}

/*
func (routing Route) AddAllowedIps(msg common.Message) {
	aips := common.AllowedIpsVO{}
	if err := common.LoadBody(msg.Body, &aips); err == nil {
		if ok := WgAddAllowedIps(aips); ok {
			Response(msg.ToResult("success"))
			return
		} else {
			log.Println("Can not set allow ips")
			Response(msg.ToResult("failed: wireguard config"))
			return
		}
	}
	Response(msg.ToResult("failed: unknown error"))
} //添加发布网段

func (routing Route) DelAllowedIps(msg common.Message) {
	aips := common.AllowedIpsVO{}
	if err := common.LoadBody(msg.Body, &aips); err == nil {
		if ok := WgDelAllowedIps(aips); ok {
			Response(msg.ToResult("success"))
			return
		}
	}
	Response(msg.ToResult("failed: unknown error"))
} //删除发布网段
*/
