package worker

import (
	"fmt"
	"os"
	"sdwan/common"
	"sdwan/cpe/monitor"

	uci "github.com/digineo/go-uci"
)

type Network struct {
}

func (network Network) Name() string {
	return "Network"
}

/*
	network.lan=interface
	network.lan.device='br-lan'
	network.lan.proto='static'
	network.lan.ipaddr='192.168.3.1'
	network.lan.netmask='255.255.255.0'

	network.@device[0]=device
	network.@device[0].name='br-lan'
	network.@device[0].type='bridge'
	network.@device[0].ports='eth2' 'eth3'

	uci add network device # =cfg080f15
	uci set network.@device[-1].type='bridge'
	uci set network.@device[-1].name='br-lan1'
	uci add_list network.@device[-1].ports='eth4'
	uci add_list network.@device[-1].ports='eth5'
	uci set network.@device[-1].ipv6='0'
*/
func (network Network) AddLAN(msg common.Message) {
	lan := common.LanVO{}
	if err := common.LoadBody(msg.Body, &lan); err == nil {
		if _, ok := uci.Get("network", lan.Name, "device"); !ok {
			brname := lan.BridgeName()
			devname := lan.BridgeDevName()
			uci.AddSection("network", devname, "device")
			uci.Set("network", devname, "name", brname)
			uci.Set("network", devname, "type", "bridge")
			uci.Set("network", devname, "ports", lan.Devices...)

			uci.AddSection("network", lan.Name, "interface")
			uci.Set("network", lan.Name, "device", brname)
			uci.Set("network", lan.Name, "proto", "static")
			uci.Set("network", lan.Name, "ipaddr", lan.Ipaddr)
			uci.Set("network", lan.Name, "netmask", lan.Netmask)

			AddLanZone(lan.Name)
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
} //添加LAN网段

func (network Network) DelLAN(msg common.Message) {
	lan := common.LanVO{}
	if err := common.LoadBody(msg.Body, &lan); err == nil {
		if _, ok := uci.Get("network", lan.Name, "device"); ok {
			uci.DelSection("network", lan.Name)
			uci.DelSection("network", lan.BridgeName())
			DelLanZone(lan.Name)
			uci.Commit()
			ReloadNetwork()
			ReloadFirewall()
			Response(msg.ToResult("success"))
			return
		} else {
			Response(msg.ToResult("failed: no exits"))
			return
		}
	}
	Response(msg.ToResult("failed: unknown error"))
} //删除LAN网段

func (network Network) EnableDHCP(msg common.Message) {
	/*
	   uci set dhcp.lan1=dhcp
	   uci set dhcp.lan1.interface='lan1'
	   uci set dhcp.lan1.start='100'
	   uci set dhcp.lan1.limit='150'
	   uci set dhcp.lan1.leasetime='12h'
	*/
	lanDhcp := common.LanDhcpVO{}
	if err := common.LoadBody(msg.Body, &lanDhcp); err == nil {
		if _, ok := uci.Get("dhcp", lanDhcp.LanName, "interface"); ok {
			uci.DelSection("dhcp", lanDhcp.LanName)
		}
		uci.AddSection("dhcp", lanDhcp.LanName, "dhcp")
		uci.Set("dhcp", lanDhcp.LanName, "interface", lanDhcp.LanName)
		uci.Set("dhcp", lanDhcp.LanName, "leasetime", "12h")
		uci.Set("dhcp", lanDhcp.LanName, "start", fmt.Sprintf("%d", lanDhcp.Start))
		uci.Set("dhcp", lanDhcp.LanName, "limit", fmt.Sprintf("%d", (lanDhcp.End-lanDhcp.Start)))
		uci.Commit()
		ReloadDnsmasq()
		Response(msg.ToResult("success"))
		return
	}
	Response(msg.ToResult("failed: unknown error"))
} //启用DHCP服务

func (network Network) DisableDHCP(msg common.Message) {
	lanDhcp := common.LanDhcpVO{}
	if err := common.LoadBody(msg.Body, &lanDhcp); err == nil {
		if _, ok := uci.Get("dhcp", lanDhcp.LanName, "interface"); ok {
			uci.DelSection("dhcp", lanDhcp.LanName)
			uci.Commit()
			ReloadDnsmasq()
			Response(msg.ToResult("success"))
			return
		}
	}
	Response(msg.ToResult("failed: unknown error"))
} //禁用DHCP服务

func getWanInfo(name string) (r common.WanVO, ok bool) {
	wan := common.WanVO{}
	if values, ok := uci.Get("network", name, "device"); ok {
		wan.Intfname = values[0]
		values, _ = uci.Get("network", name, "proto")
		wan.Protocol = values[0]
		if wan.Protocol == "dhcp" {
			intfinfos, ok2 := GetIntfaddr(wan.Intfname)
			if ok2 {
				for _, addrinfo := range intfinfos.Addrinfos {
					if addrinfo.Family == "inet" {
						wan.Ipaddr = addrinfo.Ipaddr
						wan.Netmask = common.LenToNetMask(addrinfo.Prefixlen)
					}
				}
			}
		} else {
			values, _ = uci.Get("network", name, "ipaddr")
			wan.Ipaddr = values[0]
			values, _ = uci.Get("network", name, "netmask")
			wan.Netmask = values[0]
			values, _ = uci.Get("network", name, "gateway")
			wan.Gateway = values[0]
		}
		values, ok = uci.Get("network", name, "auto")
		if ok && len(values) == 1 {
			v := values[0]
			if v == "0" {
				wan.Enable = false
			} else {
				wan.Enable = true
			}
		} else {
			wan.Enable = true
		}
		return wan, true
	}
	return wan, false
}

func lanDeviceName(name string) string {
	return fmt.Sprintf("%sdev", name)
}

func getLanInfo(name string, idx int) (r common.LanVO, ok bool) {
	lan := common.LanVO{Name: name}
	devname := lanDeviceName(name)
	if _, ok := uci.Get("network", devname, "device"); ok {
		values, _ := uci.Get("network", name, "proto")
		lan.Protocol = values[0]
		if lan.Protocol == "dhcp" {
			intfinfos, _ := GetIntfaddr(lan.BridgeName())
			for _, addrinfo := range intfinfos.Addrinfos {
				if addrinfo.Family == "inet" {
					lan.Ipaddr = addrinfo.Ipaddr
					lan.Netmask = common.LenToNetMask(addrinfo.Prefixlen)
				}
			}
		} else {
			values, _ = uci.Get("network", name, "ipaddr")
			lan.Ipaddr = values[0]
			values, _ = uci.Get("network", name, "netmask")
			lan.Netmask = values[0]
		}
		if values, ok = uci.Get("network", devname, "ports"); ok {
			lan.Devices = values
		}
		return lan, true
	}
	return lan, false
}

func (network Network) NetworkInfo(msg common.Message) {
	networkInfo := common.NetworkInfoVO{}
	wans := make([]common.WanVO, 2)
	lans := make([]common.LanVO, 5)

	idx := 0
	if wan, ok := getWanInfo("wan"); ok {
		wans[0] = wan
		idx++
	}
	if wan, ok := getWanInfo("wan1"); ok {
		wans[1] = wan
		idx++
	}
	networkInfo.Wan = wans[:idx]

	idx = 0
	for i := 0; i < 5; i++ {
		name := "lan"
		if i != 0 {
			name = fmt.Sprintf("%s%d", name, i)
		}
		if lan, ok := getLanInfo(name, i); ok {
			lans[idx] = lan
			idx++
		}
	}
	networkInfo.Lan = lans[:idx]

	if routing, ok := GetDefaultGateway(); ok {
		networkInfo.Gateway = routing
	}
	result := msg.ToResult(networkInfo)
	Response(result)
}

func (network Network) EnableTrafficAnalysis(msg common.Message) {
	ta := common.TrafficAnalysisVO{}
	if err := common.LoadBody(msg.Body, &ta); err == nil {
		common.WriteConfig(common.CpeTaIdPath, ta)
		monitor.StartTrafficAnalysis()
		Response(msg.ToResult("success"))
		return
	}
	Response(msg.ToResult("failed: unknown error"))
}

func (network Network) DisableTrafficAnalysis(msg common.Message) {
	ta := common.TrafficAnalysisVO{}
	if err := common.LoadBody(msg.Body, &ta); err == nil {
		os.Remove(common.CpeTaIdPath)
		monitor.StopTrafficAnalysis()
		Response(msg.ToResult("success"))
		return
	}
	Response(msg.ToResult("failed: unknown error"))
}

func (network Network) DetectVpe(msg common.Message) {
	vd := common.VpeDetectVO{}
	if err := common.LoadBody(msg.Body, &vd); err == nil {
		go monitor.StartVpeDetect(vd)
		Response(msg.ToResult("success"))
		return
	}
	Response(msg.ToResult("failed: unknown error"))
}
