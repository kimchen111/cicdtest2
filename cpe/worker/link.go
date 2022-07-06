package worker

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sdwan/common"
	"strconv"

	uci "github.com/digineo/go-uci"
)

type Link struct {
}

func (link Link) Name() string {
	return "Link"
}

func (link Link) SetupHub(msg common.Message) {
	hubvo := common.HubVO{}
	if err := common.LoadBody(msg.Body, &hubvo); err == nil {
		//bridge
		brdevname := hubvo.BridgeDevName()
		briname := hubvo.BridgeName()
		if _, ok := uci.Get("network", brdevname, "type"); !ok {
			uci.AddSection("network", brdevname, "device")
			uci.Set("network", brdevname, "type", "bridge")
			uci.Set("network", brdevname, "name", briname)
			uci.Set("network", brdevname, "bridge_empty", "1")
			uci.Set("network", brdevname, "ipv6", "0")

			uci.AddSection("network", briname, "interface")
			uci.Set("network", briname, "device", briname)
			uci.Set("network", briname, "proto", "static")
			ipaddr, netmask := common.Cidr2AddrMask(hubvo.VtepAddr)
			uci.Set("network", briname, "ipaddr", ipaddr)
			uci.Set("network", briname, "netmask", netmask)
		}
		AddVnetZone(briname)
		uci.Commit()
		ReloadFirewall()
		if ok := InterfaceUp(briname); !ok {
			log.Printf("Interface setup %s failed.", briname)
		}
		// if ok := hubvo.InitBridge(); !ok {
		// 	Response(msg.ToResult("failed: init bridge"))
		// 	return
		// }
		content := hubvo.GenBgpContent()
		if hubvo.Role == "RR" {
			content += hubvo.GenRrServerBgpContent()
		} else {
			content += hubvo.GenRrClientBgpContent()
		}
		common.WriteFile(common.CpeBirdConfPath, content)
		common.WriteFile(common.CpeBirdStaticPath, "") //预先写一个空的静态文件
		exec.Command("/etc/init.d/bird", "restart").Run()
		hubvo.GenHubConf() //写配置文件标记角色为HUB
		Response(msg.ToResult("success"))
		return
	}
	Response(msg.ToResult("failed: payload error"))
}

func (link Link) DestroyHub(msg common.Message) {
	hubvo := common.HubVO{}
	if err := common.LoadBody(msg.Body, &hubvo); err == nil {
		brdevname := hubvo.BridgeDevName()
		briname := hubvo.BridgeName()
		if _, ok := uci.Get("network", briname, "proto"); ok {
			uci.DelSection("network", briname)
		}
		if _, ok := uci.Get("network", brdevname, "type"); ok {
			uci.DelSection("network", brdevname)
		}
		DelVnetZone(briname)
		uci.Commit()
		if ok := InterfaceDown(briname); !ok {
			log.Println("set interface down failed .")
		}
		ReloadFirewall()
		// if ok := hubvo.Removeridge(); !ok {
		// 	Response(msg.ToResult("failed: delete bridge"))
		// 	return
		// }
		exec.Command("/etc/init.d/bird", "stop")
		os.Remove(common.CpeBirdConfPath)
		os.Remove(common.CpeBirdStaticPath)
		hubvo.RemoveHubConf()
		Response(msg.ToResult("success"))
		return
	}
	Response(msg.ToResult("failed: payload error"))
}

func (link Link) AddTunnelEndpoint(msg common.Message) {
	vtvo := common.VxlanTunnelVO{}
	if err := common.LoadBody(msg.Body, &vtvo); err == nil {
		name := vtvo.VxlanName()
		if _, ok := uci.Get("network", name, "proto"); !ok {
			log.Printf("Add %s type interface", name)
			//vxlan dev
			uci.AddSection("network", name, "interface")
			uci.Set("network", name, "proto", "vxlan")
			uci.Set("network", name, "name", name)
			uci.Set("network", name, "ipaddr", vtvo.SelfIpaddr)
			uci.Set("network", name, "peeraddr", vtvo.RemoteIpaddr)
			uci.Set("network", name, "port", strconv.Itoa(vtvo.Port()))
			uci.Set("network", name, "vid", strconv.Itoa(vtvo.Vni))
			brdevname := vtvo.BridgeDevName()
			AddDevToBridge(brdevname, name)

			// ruleName := fmt.Sprintf("vl%d", vtvo.Id)
			AllowIncoming(name, vtvo.RemoteIpaddr, vtvo.SelfIpaddr, vtvo.Port())
			uci.Commit()
			InterfaceUp(name)
			ReloadFirewall()
			Response(msg.ToResult("success"))
			return
		} else {
			Response(msg.ToResult("success: exits"))
			return
		}
	}
	Response(msg.ToResult("failed: payload error"))
} //创建HUB的TUNNEL端点

func (link Link) DelTunnelEndpoint(msg common.Message) {
	vtvo := common.VxlanTunnelVO{}
	if err := common.LoadBody(msg.Body, &vtvo); err == nil {
		name := vtvo.VxlanName()
		if _, ok := uci.Get("network", name, "proto"); ok {
			uci.DelSection("network", name)
			RemoveIncoming(vtvo.Id)
			uci.Commit()
			ReloadFirewall()
			brdevname := vtvo.BridgeDevName()
			DelDevFromBridge(brdevname, name)
			InterfaceDown(name)
			Response(msg.ToResult("success"))
			return
		} else {
			Response(msg.ToResult("success: removed"))
			return
		}
	}
	Response(msg.ToResult("failed: payload error"))
} //删除HUB的TUNNEL端点

func (link Link) AddHubMstpEndpoint(msg common.Message) {
	dl := common.MstpVO{}
	if err := common.LoadBody(msg.Body, &dl); err == nil {
		dle := dl.GetEndpoint()
		devname := dle.DevName()
		if _, ok := uci.Get("network", devname, "type"); !ok {
			if dle.WithVlanIntf() {
				uci.AddSection("network", devname, "device")
				uci.Set("network", devname, "type", "8021q")
				uci.Set("network", devname, "ifname", dle.IntfName)
				uci.Set("network", devname, "vid", strconv.Itoa(dle.VlanId))
				uci.Set("network", devname, "name", devname)
				uci.Set("network", devname, "ipv6", "0")
			}
			brdevname := dl.BridgeDevName()
			AddDevToBridge(brdevname, devname)
			uci.Commit()
			InterfaceUp(devname)
			Response(msg.ToResult("success"))
			return
		} else {
			Response(msg.ToResult("success: exits"))
			return
		}
	}
	Response(msg.ToResult("failed: payload error"))
} //创建HUB的TUNNEL端点

func (link Link) DelHubMstpEndpoint(msg common.Message) {
	dl := common.MstpVO{}
	if err := common.LoadBody(msg.Body, &dl); err == nil {
		dle := dl.GetEndpoint()
		devname := dle.DevName()
		if _, ok := uci.Get("network", devname, "type"); ok {
			uci.DelSection("network", devname)
			uci.Commit()
			brdevname := dl.BridgeDevName()
			DelDevFromBridge(brdevname, devname)
			InterfaceDown(devname)
			Response(msg.ToResult("success"))
			return
		} else {
			Response(msg.ToResult("success: removed"))
			return
		}
	}
	Response(msg.ToResult("failed: payload error"))
} //删除HUB的TUNNEL端点

/*
func (link Link) AddTunnelEndpoint(msg common.Message) {
	vtvo := common.VxlanTunnelVO{}
	if err := common.LoadBody(msg.Body, &vtvo); err == nil {
		vtname := vtvo.VxlanName()
		if !common.IntfExists(vtname) {
			vtvo.WriteConf()
			if ok := vtvo.Start(); !ok {
				Response(msg.ToResult("failed: Start tunnel"))
				return
			}
			ruleName := fmt.Sprintf("vl%d", vtvo.Id)
			AllowIncoming(ruleName, vtvo.RemoteIpaddr, vtvo.SelfIpaddr, vtvo.Port())
			uci.Commit()
			ReloadFirewall()
			Response(msg.ToResult("success"))
			return
		} else {
			Response(msg.ToResult("success: exits"))
			return
		}
	}
	Response(msg.ToResult("failed: payload error"))
} //创建HUB的TUNNEL端点

func (link Link) DelTunnelEndpoint(msg common.Message) {
	vtvo := common.VxlanTunnelVO{}
	if err := common.LoadBody(msg.Body, &vtvo); err == nil {
		vtname := vtvo.VxlanName()
		if exists := common.IntfExists(vtname); exists {
			vtvo.RemoveConf()
			vtvo.Stop()
			ruleName := fmt.Sprintf("vl%d", vtvo.Id)
			uci.DelSection("firewall", ruleName)
			uci.Commit()
			ReloadFirewall()
			Response(msg.ToResult("success"))
			return
		} else {
			Response(msg.ToResult("success"))
			return
		}
	}
	Response(msg.ToResult("failed: payload error"))
} //删除HUB的TUNNEL端点
*/

/*
network.wg0=interface
network.wg0.proto='wireguard'
network.wg0.private_key='IIt/PgJ1csMImvmOdYltuBWrKU+OlmgwHJXF6j9n6mw='
network.wg0.addresses='192.168.2.9'
network.@wireguard_wg0[0]=wireguard_wg0
network.@wireguard_wg0[0].description='VPE1'
network.@wireguard_wg0[0].public_key='upfXjD2o2VuCamiy1Oig7/oK6Iqd1OTqZ8KiV7cWhU0='
network.@wireguard_wg0[0].allowed_ips='10.0.0.0/16'
network.@wireguard_wg0[0].route_allowed_ips='1'
network.@wireguard_wg0[0].endpoint_host='222.73.0.9'
network.@wireguard_wg0[0].endpoint_port='5678'
network.@wireguard_wg0[0].persistent_keepalive='30'
*/
func (link Link) AddVpnEndpoint(msg common.Message) {
	vpnlink := common.VpnlinkVO{}
	if err := common.LoadBody(msg.Body, &vpnlink); err == nil {
		wgname := vpnlink.IntfName()
		peername := vpnlink.PeerName(vpnlink.Id)
		uci.AddSection("network", wgname, "interface")
		uci.Set("network", wgname, "proto", "wireguard")
		uci.AddSection("network", peername, fmt.Sprintf("wireguard_%s", wgname))
		uci.Set("network", peername, "description", peername)
		uci.Set("network", peername, "allowed_ips", "0.0.0.0/1", "128.0.0.0/1")
		uci.Set("network", peername, "route_allowed_ips", "0")
		uci.Set("network", peername, "persistent_keepalive", "30")
		if vpnlink.IsServer() {
			uci.Set("network", wgname, "private_key", vpnlink.Server.PrivKey)
			uci.Set("network", wgname, "listen_port", strconv.Itoa(vpnlink.ListenPort()))
			uci.Set("network", wgname, "addresses", vpnlink.Server.IntfAddr)
			uci.Set("network", peername, "public_key", vpnlink.Client.PubKey())
			AllowIncoming(vpnlink.FirewallRuleName(), "", vpnlink.Server.ListenAddr, vpnlink.ListenPort())
		} else {
			uci.Set("network", wgname, "private_key", vpnlink.Client.PrivKey)
			uci.Set("network", wgname, "addresses", vpnlink.Client.IntfAddr)
			uci.Set("network", peername, "public_key", vpnlink.Server.PubKey())
			uci.Set("network", peername, "endpoint_host", vpnlink.Server.ListenAddr)
			uci.Set("network", peername, "endpoint_port", fmt.Sprintf("%d", vpnlink.ListenPort()))

		}
		if !IsHub() {
			state := GetPreState()
			if state == "" {
				state = "MASTER"
			}
			vpnlink.SetDevState(state)
			SetupCpeBird(&vpnlink)
		}
		AddVnetZone(wgname)
		uci.Commit()
		ReloadFirewall()
		go SetupWireguard(wgname)
		// routemapContent := wglink.GenRouteMapContent(state)
		SetBaseHtb(wgname, vpnlink.Rate)
		Response(msg.ToResult("success"))
		return
	}
	Response(msg.ToResult("failed: unknown error"))
} //新建VPN隧道

func (link Link) DelVpnEndpoint(msg common.Message) {
	vpnlink := common.VpnlinkVO{}
	if err := common.LoadBody(msg.Body, &vpnlink); err == nil {
		wgname := vpnlink.IntfName()
		if _, ok := uci.Get("network", wgname, "proto"); ok {
			uci.DelSection("network", wgname)
			typename := fmt.Sprintf("wireguard_%s", wgname)
			if peers, ok2 := uci.GetSections("network", typename); ok2 {
				for _, peer := range peers {
					uci.DelSection("network", peer)
				}
			}
			if vpnlink.IsServer() {
				RemoveIncoming(vpnlink.Id)
			}
			DelVnetZone(wgname)
			uci.Commit()
			if ok := InterfaceDown(wgname); !ok {
				log.Println("Wireguard interface failed down.")
				Response(msg.ToResult("failed: cann't down"))
				return
			}
			ReloadFirewall()
			DisableBirdInclude(common.CpeBirdConfPath, vpnlink.BirdNeighFileName())
			Response(msg.ToResult("success"))
			return
		} else {
			Response(msg.ToResult("success"))
			return
		}
	}
	Response(msg.ToResult("failed: unknown error"))
} //删除VPN隧道

func (link Link) ResetLinkState(msg common.Message) {
	rvl := common.ResetVpnlinkVO{}
	if err := common.LoadBody(msg.Body, &rvl); err == nil {
		if !IsHub() {
			state := GetPreState()
			if state == "" {
				state = "MASTER"
			}
			rvl.SetDevState(state)
			ResetCpeVpnlink(rvl)
			Response(msg.ToResult("success"))
			return
		}
	}
	Response(msg.ToResult("failed: unknown error"))
} //切换链路

/*
uci add network device # =cfg070f15
uci set network.@device[-1].type='8021q'
uci set network.@device[-1].ifname='eth5'
uci set network.@device[-1].vid='2002'
uci set network.@device[-1].name='eth5_2002'
uci set network.@device[-1].ipv6='0'

uci set network.eth5_2002=interface
uci set network.eth5_2002.proto='static'
uci set network.eth5_2002.device='eth5_2002'
uci set network.eth5_2002.ipaddr='10.0.3.2'
uci set network.eth5_2002.netmask='255.255.255.252'

config device
        option type '8021q'
        option ifname 'eth5'
        option vid '2002'
        option name 'eth5_2002'
        option ipv6 '0'

config interface 'eth5_2002'
        option proto 'static'
        option device 'eth5_2002'
        option ipaddr '10.0.3.2'
        option netmask '255.255.255.252'
*/
func (link Link) AddMstpEndpoint(msg common.Message) {
	dl := common.MstpVO{}
	if err := common.LoadBody(msg.Body, &dl); err == nil {
		dle := dl.GetEndpoint()
		name := dle.InterfaceName()
		if dle.WithVlanIntf() {
			devname := dle.DevName()
			uci.AddSection("network", devname, "device")
			uci.Set("network", devname, "type", "8021q")
			uci.Set("network", devname, "ifname", dle.IntfName)
			uci.Set("network", devname, "vid", strconv.Itoa(dle.VlanId))
			uci.Set("network", devname, "name", name)
			uci.Set("network", devname, "ipv6", "0")
		}
		addr, netmask := common.Cidr2AddrMask(dle.IntfAddr)
		uci.AddSection("network", name, "interface")
		uci.Set("network", name, "proto", "static")
		uci.Set("network", name, "device", name)
		uci.Set("network", name, "ipaddr", addr)
		uci.Set("network", name, "netmask", netmask)
		if !IsHub() {
			state := GetPreState()
			if state == "" {
				state = "MASTER"
			}
			dl.SetDevState(state)
			SetupCpeBird(&dl)
		}
		AddVnetZone(name)
		uci.Commit()
		if ok := InterfaceUp(name); !ok {
			log.Printf("Interface setup %s failed.", name)
		}
		ReloadFirewall()
		Response(msg.ToResult("success"))
		return
	}
	Response(msg.ToResult("failed: unknown error"))
} //新建直连专线链路的网卡

func (link Link) DelDirEndpoint(msg common.Message) {
	mstp := common.MstpVO{}
	if err := common.LoadBody(msg.Body, &mstp); err == nil {
		dle := mstp.GetEndpoint()
		name := dle.InterfaceName()
		if _, ok := uci.Get("network", name, "proto"); ok {
			uci.DelSection("network", name)
			devname := dle.DevName()
			if _, ok := uci.Get("network", devname, "type"); ok {
				uci.DelSection("network", devname)
			}
			DelVnetZone(name)
			uci.Commit()
			if ok := InterfaceDown(name); !ok {
				log.Println("Set dirlink interface down failed.")
				Response(msg.ToResult("failed: cann't down"))
				return
			}
			ReloadFirewall()
			// DisableBirdInclude(common.CpeBirdConfPath, vpnlink.NeighFileName())
			Response(msg.ToResult("success"))
			return
		} else {
			Response(msg.ToResult("success"))
			return
		}
	}
	Response(msg.ToResult("failed: unknown error"))
} //删除直连专线链路的网卡

/*
config interface 'tunnel10002'
        option proto 'vxlan'
        option peeraddr '192.168.13.9'
        option ipaddr '192.168.3.2'
        option port '10002'
        option vid '10002'
network.tunnel10002=interface
network.tunnel10002.proto='vxlan'
network.tunnel10002.peeraddr='192.168.13.9'
network.tunnel10002.ipaddr='192.168.3.2'
network.tunnel10002.port='10002'
network.tunnel10002.vid='10002'

uci add network device # =cfg090f15
uci set network.@device[-1].type='bridge'
uci set network.@device[-1].name='br50'
uci add_list network.@device[-1].ports='vl301'
uci set network.@device[-1].ipv6='0'
uci set network.br50=interface
uci set network.br50.proto='static'
uci set network.br50.device='br50'
uci set network.br50.ipaddr='10.16.16.31'
uci set network.br50.netmask='255.255.255.0'

firewall.@rule[9]=rule
firewall.@rule[9].name='vxlan'
firewall.@rule[9].proto='udp'
firewall.@rule[9].src='wan'
firewall.@rule[9].src_ip='192.168.122.21'
firewall.@rule[9].dest_port='40301'
firewall.@rule[9].target='ACCEPT'
firewall.@rule[9].dest_ip='192.168.122.96'
*/
/*
func (link Link) AddTunnelEndpoint(msg common.Message) {
	vtvo := common.VxlanTunnelVO{}
	if err := common.LoadBody(msg.Body, &vtvo); err == nil {
		name := vtvo.VxlanName()
		if _, ok := uci.Get("network", name, "proto"); !ok {
			log.Printf("Add %s type interface", name)
			//vxlan dev
			uci.AddSection("network", name, "device")
			uci.Set("network", name, "proto", "vxlan")
			uci.Set("network", name, "ipaddr", vtvo.SelfIpaddr)
			uci.Set("network", name, "peeraddr", vtvo.RemoteIpaddr)
			uci.Set("network", name, "port", strconv.Itoa(vtvo.Port()))
			uci.Set("network", name, "vid", strconv.Itoa(vtvo.Vni))

			// //防火墙
			// ruleName := fmt.Sprintf("vl%d", tunnel.Id)
			// uci.AddSection("firewall", ruleName, "rule")
			// uci.Set("firewall", ruleName, "name", ruleName)
			// uci.Set("firewall", ruleName, "proto", "udp")
			// uci.Set("firewall", ruleName, "src", "wan")
			// uci.Set("firewall", ruleName, "src_ip", tunnel.RemoteIpaddr)
			// uci.Set("firewall", ruleName, "target", "ACCEPT")
			// uci.Set("firewall", ruleName, "dest_ip", tunnel.SelfIpaddr)
			// uci.Set("firewall", ruleName, "dest_port", strconv.Itoa(tunnel.Port()))

			// //bridge
			// brname := tunnel.BridgeName()
			// if _, ok := uci.Get("network", brname, "type"); !ok {
			// 	uci.AddSection("network", brname, "device")
			// 	uci.Set("network", brname, "type", "bridge")
			// 	uci.Set("network", brname, "name", brname)

			// 	briname := tunnel.BridgeIntfName()
			// 	uci.AddSection("network", briname, "interface")
			// 	uci.Set("network", briname, "device", brname)
			// 	uci.Set("network", briname, "proto", "static")
			// 	ipaddr, netmask := tunnel.SplitIpaddr()
			// 	uci.Set("network", briname, "ipaddr", ipaddr)
			// 	uci.Set("network", briname, "netmask", netmask)
			// }

			AttachVxlanToBridge(name, brname)

			uci.Commit()
			ReloadNetwork()
			ReloadFirewall()
			Response(msg.ToResult("success"))
			return
		} else {
			Response(msg.ToResult("failed: exits"))
			return
		}
	}
	Response(msg.ToResult("failed: unknown error"))
}

func (link Link) DelTunnelEndpoint(msg common.Message) {
	tunnel := common.VxlanTunnelVO{}
	if err := common.LoadBody(msg.Body, &tunnel); err == nil {
		vxlanName := tunnel.VxlanName()
		if _, ok := uci.Get("network", vxlanName, "proto"); ok {
			uci.DelSection("network", vxlanName)
			ruleName := fmt.Sprintf("vl%d", tunnel.Id)
			uci.DelSection("firewall", ruleName)
			DettachVxlanFromBridge(vxlanName, tunnel.BridgeName())
			uci.Commit()
			ReloadNetwork()
			ReloadFirewall()
			EnsureRemoveVxlan(vxlanName)
			Response(msg.ToResult("success"))
			return
		} else {
			Response(msg.ToResult("success"))
			return
		}
	}
	Response(msg.ToResult("failed: unknown error"))
}
*/
