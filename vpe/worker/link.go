package worker

import (
	"fmt"
	"log"
	"os"
	"sdwan/common"
)

type Link struct {
}

func (link Link) Name() string {
	return "Link"
}

func (link Link) AddVpnEndpoint(msg common.Message) {
	vpnlink := common.VpnlinkVO{}
	if err := common.LoadBody(msg.Body, &vpnlink); err == nil {
		ns := vpnlink.NetnsName()
		if exists := common.NetnsExists(ns); !exists {
			log.Println("Netns not exists.")
			Response(msg.ToResult("failed: Netns not exists"))
			return
		}
		wgname := vpnlink.IntfName()
		if exists := WgintfExists(ns, wgname); exists {
			if ok := WgDelete(ns, wgname); !ok {
				log.Println("Delete WG interface failed.")
			}
			path := fmt.Sprintf("/etc/wireguard/%s.conf", wgname)
			os.Remove(path)
		}
		content := vpnlink.GenWgContent()
		path := fmt.Sprintf("/etc/wireguard/%s.conf", wgname)
		if ok := common.WriteFile(path, content); !ok {
			log.Printf("Can not write config file: %s", path)
		}
		if ok := WgLoadConf(wgname); !ok {
			log.Println("Failed reload wireguard.")
			Response(msg.ToResult("failed: wireguard reload"))
			return
		}
		// if ok := WgSetupNetns(vpnlink); !ok {
		// 	log.Println("Failed setup netns.")
		// 	Response(msg.ToResult("failed: setup netns"))
		// 	return
		// }
		Response(msg.ToResult("success"))
		return
	}
	Response(msg.ToResult("failed: payload error"))
} //创建VPE的VPN客户端配置

/*
func (link Link) DelVpnEndpoint(msg common.Message) {
	endpoint := common.WgServerVO{}
	if err := common.LoadBody(msg.Body, &endpoint); err == nil {
		wgname := endpoint.IntfName()
		if WgintfExists(wgname) {
			path := fmt.Sprintf("/etc/wireguard/%s.conf", wgname)
			os.Remove(path)
			WgDelete(wgname)
			Response(msg.ToResult("success"))
			return
		}
		Response(msg.ToResult("success: not exits"))
		return
	}
	Response(msg.ToResult("failed: payload error"))
} //删除VPE的VPN端点
*/

func (link Link) DelVpnEndpoint(msg common.Message) {
	vpnlink := common.VpnlinkVO{}
	if err := common.LoadBody(msg.Body, &vpnlink); err == nil {
		wgname := vpnlink.IntfName()
		if ok := RemoveNetnsLink(vpnlink.NetnsName(), wgname); !ok {
			log.Println("Delete WG interface failed.")
			Response(msg.ToResult("failed: remove wg"))
			return
		}
		path := fmt.Sprintf("/etc/wireguard/%s.conf", wgname)
		os.Remove(path)
		Response(msg.ToResult("success"))
		return
	}
	Response(msg.ToResult("failed: payload error"))
} //删除VPE的VPN客户端配置

func (link Link) AddDirEndpoint(msg common.Message) {
	dl := common.DirlinkVO{}
	if err := common.LoadBody(msg.Body, &dl); err == nil {
		ns := dl.NetnsName()
		if exists := common.NetnsExists(ns); !exists {
			log.Println("Netns not exists.")
			Response(msg.ToResult("failed: Netns not exists"))
			return
		}
		if ok := common.WriteConfig(dl.Path(), dl); !ok {
			log.Printf("Can not write config file: %s", dl.Path())
		}
		if msgx, ok := SetupDirlinkIntf(dl); !ok {
			log.Println("Failed setup dirlink interface.")
			Response(msg.ToResult(msgx))
			return
		}
		Response(msg.ToResult("success"))
		return
	}
	Response(msg.ToResult("failed: payload error"))
} //创建VPE的直连链路端点的网卡相关配置

func (link Link) DelDirEndpoint(msg common.Message) {
	dl := common.DirlinkVO{}
	if err := common.LoadBody(msg.Body, &dl); err == nil {
		if ok := RemoveNetnsLink(dl.NetnsName(), dl.VlanifName()); !ok {
			log.Println("Delete Direct link interface failed.")
			Response(msg.ToResult("failed: remove dirlink"))
			return
		}
		os.Remove(dl.Path())
		Response(msg.ToResult("success"))
		return
	}
	Response(msg.ToResult("failed: payload error"))
} //删除VPE的VPN客户端配置

/*
func (link Link) AddTunnelEndpoint(msg common.Message) {
	vtvo := common.VxlanTunnelVO{}
	if err := common.LoadBody(msg.Body, &vtvo); err == nil {
		vtname := vtvo.VxlanName()
		if !common.IntfExists(vtname) {
			vtvo.WriteConf()
			if ok := vtvo.Start("VPE"); !ok {
				Response(msg.ToResult("failed: Start tunnel"))
				return
			}
			Response(msg.ToResult("success"))
			return
		} else {
			Response(msg.ToResult("success: exits"))
			return
		}
	}
	Response(msg.ToResult("failed: payload error"))
} //创建VPE的VPN客户端配置

func (link Link) DelTunnelEndpoint(msg common.Message) {
	vto := common.VxlanTunnelVO{}
	if err := common.LoadBody(msg.Body, &vto); err == nil {
		vtname := vto.VxlanName()
		if exists := common.IntfExists(vtname); exists {
			vto.RemoveConf()
			vto.Stop()
			Response(msg.ToResult("success"))
			return
		} else {
			Response(msg.ToResult("success"))
			return
		}
	}
	Response(msg.ToResult("failed: payload error"))
} //删除VPE的VPN客户端配置
*/
