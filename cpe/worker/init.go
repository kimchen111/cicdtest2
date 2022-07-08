package worker

import (
	"fmt"
	"log"
	"os/exec"
	"sdwan/common"
)

func InitConfig() {
	log.Printf("AgentType %s", common.GBC.GetAgentType())
	InitSystem()
	InitVnetZone()
	// InitVxlanConfDir()
	InitMonitorConfDir()
	PrepareQos()
	PrepareBirdResetScript()
	// RenameLandev()
}

func InitSystem() {
	// system.@system[0].hostname
	// system.@system[0].zonename='Asia/Shanghai'
	// system.@system[0].timezone='CST-8'
	if !common.FileExists("/etc/system_inited") {
		exec.Command("uci", "set", fmt.Sprintf("system.@system[0].hostname=%s", common.GBC.GetESN())).Run()
		exec.Command("uci", "set", "system.@system[0].zonename=Asia/Shanghai").Run()
		exec.Command("uci", "set", "system.@system[0].timezone=CST-8").Run()
		exec.Command("uci", "commit", "system").Run()
		exec.Command("touch", "/etc/system_inited").Run()
	}
}

/*
func InitVxlanConfDir() {
	if !common.FileExists(common.CpeVxlanConfDir) {
		exec.Command("mkdir", common.CpeVxlanConfDir).Run()
	}
}
*/
func InitMonitorConfDir() {
	if !common.FileExists(common.CpeMonitorConfDir) {
		exec.Command("mkdir", common.CpeMonitorConfDir).Run()
	}
}
func InitVnetZone() {
	if !common.FileExists("/etc/fwzone_inited") {
		exec.Command("uci", "add", "firewall", "zone").Run()
		exec.Command("uci", "set", "firewall.@zone[-1].name=vnet").Run()
		exec.Command("uci", "set", "firewall.@zone[-1].input=ACCEPT").Run()
		exec.Command("uci", "set", "firewall.@zone[-1].output=ACCEPT").Run()
		exec.Command("uci", "set", "firewall.@zone[-1].forward=ACCEPT").Run()
		exec.Command("uci", "add", "firewall", "forwarding").Run()
		exec.Command("uci", "set", "firewall.@forwarding[-1].src=vnet").Run()
		exec.Command("uci", "set", "firewall.@forwarding[-1].dest=lan").Run()
		exec.Command("uci", "add", "firewall", "forwarding").Run()
		exec.Command("uci", "set", "firewall.@forwarding[-1].src=lan").Run()
		exec.Command("uci", "set", "firewall.@forwarding[-1].dest=vnet").Run()
		exec.Command("uci", "commit", "firewall").Run()
		exec.Command("touch", "/etc/fwzone_inited").Run()
	}
}

/*
	uci rename network.@device[0]=brlan
	uci set network.brlan.name='brlan'
	uci set network.lan.device='brlan'
	uci commit network
*/
/*
func RenameLandev() {
	if _, ok := uci.Get("network", "landev", "name"); !ok {
		exec.Command("uci", "rename", "network.@device[0]=landev").Run()
		exec.Command("uci", "commit", "network").Run()
		exec.Command("/etc/init.d/network", "restart").Run()
	}
}
*/
