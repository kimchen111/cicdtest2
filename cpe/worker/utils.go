package worker

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"sdwan/common"
	"strconv"
	"strings"
	"time"

	uci "github.com/digineo/go-uci"
	"github.com/vishvananda/netlink"
)

func PostBoot() {
	preState := GetPreState()
	runState := GetVrrpState()
	if preState == "MASTER" && runState == "BACKUP" {
		HoldMaster()
	}
}

func GetIntfaddr(name string) (info common.IntfAddrInfo, ok bool) {
	infox := make([]common.IntfAddrInfo, 50)
	command := exec.Command("ip", "-j", "addr", "show", "dev", name)
	out, err := command.CombinedOutput()
	if err == nil {
		err = json.Unmarshal(out, &infox)
		if err == nil {
			return infox[0], true
		}
	}
	// log.Printf("out: %s, err: %s", out, err)
	return infox[0], false
}

func GetDefaultGateway() (info common.RouteVO, ok bool) {
	infox := make([]common.RouteVO, 50)
	command := exec.Command("ip", "-j", "ro", "show", "default")
	out, err := command.CombinedOutput()
	if err == nil {
		err = json.Unmarshal(out, &infox)
		if err == nil {
			return infox[0], true
		} else {
			log.Printf("out: %s, json err: %s", out, err)
		}
	}
	return infox[0], false
}

/*
uci set network.eth4=interface
uci set network.eth4.proto='none'
uci set network.eth4.device='eth4'
*/
/*
	network.lan=interface
	network.lan.device='br-lan'
	network.lan.proto='static'
	network.lan.ipaddr='192.168.3.1'
	network.lan.netmask='255.255.255.0'
*/
func ProcessAddrs(vrrp common.VrrpVO) bool {
	ipaddr, netmask := common.Cidr2AddrMask(vrrp.HBIpAddr)
	intfName := vrrp.HBIntfName
	uci.AddSection("network", intfName, "interface")
	uci.Set("network", intfName, "proto", "static")
	uci.Set("network", intfName, "device", intfName)
	uci.Set("network", intfName, "ipaddr", ipaddr)
	uci.Set("network", intfName, "netmask", netmask)
	for _, vipaddr := range vrrp.VipAddrs {
		if _, ok := uci.Get("network", vipaddr.LanName, "ipaddr"); !ok {
			log.Printf("Failed lan name")
			return false
		}
		if len(vipaddr.SolidAddr) > 10 {
			ipaddr, netmask := common.Cidr2AddrMask(vipaddr.SolidAddr)
			uci.Set("network", vipaddr.LanName, "proto", "static")
			uci.Set("network", vipaddr.LanName, "ipaddr", ipaddr)
			uci.Set("network", vipaddr.LanName, "netmask", netmask)
		}
	}
	uci.Commit()
	exec.Command("/etc/init.d/network", "reload").Run()
	return true
}

func LoadFrrConfig(content string) bool {
	//log.Println(content)
	tmppath := fmt.Sprintf("/tmp/%d", time.Now().UnixMilli())
	defer os.Remove(tmppath)
	if ok := common.WriteFile(tmppath, content); ok {
		err := exec.Command("vtysh", "-f", tmppath).Run()
		if err != nil {
			errpath := fmt.Sprintf("%s.err", tmppath)
			log.Printf("Load frr config Error: %v file: %s", err, errpath)
			exec.Command("cp", tmppath, errpath).Run()
		} else {
			return err == nil
		}
	}
	return false
}

func PrepareKeepalived() {
	if !common.FileExists(common.CpeKeepalivedDir) {
		exec.Command("mkdir", "-p", common.CpeKeepalivedDir).Run()
	}
	GenVrrpNotifyScript()
	GenCheckHoldScript()
	GenGetPreStateScript()
	GenPermanentMasterScript()
}

func GenVrrpNotifyScript() {
	template := `#!/bin/sh
# notify.sh
# for ANY state transition.
# "notify" script is called AFTER the
# notify_* script(s) and is executed
# with 3 arguments provided by keepalived
# (ie don’t include parameters in the notify line).
# arguments
# $1 = "GROUP"|"INSTANCE"
# $2 = name of group or instance
# $3 = target state of transition
# ("MASTER"|"BACKUP"|"FAULT")
echo $3 > /var/run/vrrp.state
`
	path := fmt.Sprintf("%s/notify.sh", common.CpeKeepalivedDir)
	common.WriteFile(path, template)
	exec.Command("chmod", "+x", path).Run()
}

func GenCheckHoldScript() {
	template := `#!/bin/sh
[ -f /var/run/vrrp.master ] && {
	exit 0
}
exit 1
`
	path := fmt.Sprintf("%s/check_hold.sh", common.CpeKeepalivedDir)
	common.WriteFile(path, template)
	exec.Command("chmod", "+x", path).Run()
}
func GenGetPreStateScript() {
	template := `#!/bin/sh
grep state /etc/keepalived/keepalived.conf|cut -d" " -f2
`
	path := fmt.Sprintf("%s/get_prestate.sh", common.CpeKeepalivedDir)
	common.WriteFile(path, template)
	exec.Command("chmod", "+x", path).Run()
}
func GenPermanentMasterScript() {
	template := `#!/bin/sh
sed -i "s/state BACKUP/state MASTER/" /etc/keepalived/keepalived.conf
sed -i "s/priority 45/priority 70/" /etc/keepalived/keepalived.conf
service keepalived reload
`
	path := fmt.Sprintf("%s/permanent_master.sh", common.CpeKeepalivedDir)
	common.WriteFile(path, template)
	exec.Command("chmod", "+x", path).Run()
}

func EnableKeepalived(vrrp common.VrrpVO) bool {
	content := `
	config globals 'globals'
    	option alt_config_file "/etc/keepalived/keepalived.conf"
`
	common.WriteFile("/etc/config/keepalived", content)
	common.WriteFile("/etc/keepalived/keepalived.conf", vrrp.GenEnableContent())
	err := exec.Command("service", "keepalived", "restart").Run()
	return err == nil
}

func DisableKeepalived() bool {
	exec.Command("service", "keepalived", "stop").Run()
	os.Remove("/etc/config/keepalived")
	os.Remove("/etc/keepalived/keepalived.conf")
	return true
}

func CheckMasterSignal() bool {
	return common.FileExists("/var/run/vrrp.master")
}

func GetPreState() string {
	path := fmt.Sprintf("%s/get_prestate.sh", common.CpeKeepalivedDir)
	command := exec.Command(path)
	if out, err := command.CombinedOutput(); err == nil {
		return strings.Replace(string(out), "\n", "", -1)
	}
	return ""
}
func PermanentMaster() bool {
	path := fmt.Sprintf("%s/permanent_master.sh", common.CpeKeepalivedDir)
	err := exec.Command(path).Run()
	return err == nil
}

func GetPreVrrpState() string {
	if common.FileExists("/etc/keepalived/keepalived.conf") {
		content, _ := common.ReadFile("/etc/keepalived/keepalived.conf")
		lines := strings.Split(content, "\n")

		for _, line := range lines {
			if strings.Contains(line, "state") {
				line = strings.TrimSpace(line)
				arr := strings.Split(line, " ")
				if len(arr) == 2 {
					return arr[1]
				}
			}
		}
	}
	return ""
}
func GetVrrpState() string {
	if common.FileExists("/var/run/vrrp.state") {
		content, _ := common.ReadFile("/var/run/vrrp.state")
		return strings.ReplaceAll(content, "\n", "")
	}
	return ""
}

func HoldMaster() bool {
	log.Printf("Grab keepalived MASTER state.")
	err := exec.Command("touch", "/var/run/vrrp.master").Run()
	return err == nil
}

func ReleaseMaster() bool {
	err := exec.Command("rm", "-f", "/var/run/vrrp.master").Run()
	return err == nil
}

func InterfaceUp(intf string) bool {
	err := exec.Command("ifup", intf).Run()
	return err == nil
}
func InterfaceDown(intf string) bool {
	err := exec.Command("ifdown", intf).Run()
	return err == nil
}
func SetupWireguard(name string) {
	s := time.After(time.Second * 1)
	<-s
	if ok := InterfaceUp(name); !ok {
		log.Printf("WireguardUp %s failed.", name)
	}
	if !IsHub() {
		prevWgname := GetAnotherLink(name)
		if prevWgname != "" {
			s1 := time.After(time.Second * 2)
			<-s1
			if ok := InterfaceUp(prevWgname); !ok {
				log.Printf("WireguardUp %s failed.", prevWgname)
			}
		}
	}
}
func AddIpaddr(intf string, ipaddr string) {
	nl, _ := netlink.LinkByName(intf)
	addr, _ := netlink.ParseAddr(ipaddr)
	netlink.AddrAdd(nl, addr)
}
func ReloadNetwork() {
	exec.Command("/etc/init.d/network", "reload").Run()
}
func ReloadFirewall() {
	exec.Command("/etc/init.d/firewall", "reload").Run()
}
func ReloadDnsmasq() {
	exec.Command("/etc/init.d/dnsmasq", "reload").Run()
}
func FirstLinkByOsInterface() bool {
	ifaces, _ := net.Interfaces()
	ptx, _ := regexp.Compile("^wg.*")
	count := 0
	for _, iface := range ifaces {
		if ptx.MatchString(iface.Name) {
			count++
		}
	}
	// log.Println("Current ifaces: ", ifaces)
	return count == 0
}

func GetAnotherLink(wgname string) string {
	ifaces, _ := net.Interfaces()
	ptx, _ := regexp.Compile("^wg.*")
	for _, iface := range ifaces {
		if !ptx.MatchString(iface.Name) {
			return iface.Name
		}
	}
	return ""
}

func BatchEnableBirdIncludes(filepath string, paths []string) bool {
	content, _ := common.ReadFile(filepath)

	includescontent := ""
	for _, path := range paths {
		//TODO CHECK 为什么文件存在才写？
		if common.FileExists(path) {
			//必须要有换行
			temp := `
include "%s";
`
			includeline := fmt.Sprintf(temp, path)
			clearLine := strings.ReplaceAll(includeline, "\n", "")
			if !strings.Contains(content, clearLine) {
				includescontent += includeline
			}
		}
	}
	common.AppendFile(filepath, includescontent)
	exec.Command("/etc/init.d/bird", "restart").Run()
	return true
}

func BatchDisableBirdIncludes(filepath string, paths []string) bool {
	content, _ := common.ReadFile(filepath)

	for _, path := range paths {
		temp := `include "%s";`
		includeline := fmt.Sprintf(temp, path)
		if strings.Contains(content, includeline) {
			content = strings.ReplaceAll(content, includeline, "")
		}
		if common.FileExists(path) {
			os.Remove(path)
		}
	}
	common.WriteFile(filepath, content)
	exec.Command("/etc/init.d/bird", "restart").Run()
	return true
}

func EnableBirdInclude(filepath string, path string) bool {
	if !common.FileExists(path) {
		return false
	}
	//必须要有换行
	temp := `
include "%s";
`
	includeline := fmt.Sprintf(temp, path)
	content, _ := common.ReadFile(filepath)
	if !strings.Contains(content, strings.ReplaceAll(includeline, "\n", "")) {
		common.AppendFile(filepath, includeline)
	}
	exec.Command("/etc/init.d/bird", "restart").Run()
	return true
}

func EnableBirdStaticInclude(path string) bool {
	return EnableBirdInclude(common.CpeBirdStaticPath, path)
}

func DisableBirdInclude(filepath string, path string) bool {
	temp := `include "%s";`
	includeline := fmt.Sprintf(temp, path)
	content, _ := common.ReadFile(filepath)
	if strings.Contains(content, includeline) {
		content = strings.ReplaceAll(content, includeline, "")
		common.WriteFile(filepath, content)
	}
	if common.FileExists(path) {
		os.Remove(path)
	}
	exec.Command("/etc/init.d/bird", "restart").Run()
	return true
}

func DisableBirdStaticInclude(path string) bool {
	return DisableBirdInclude(common.CpeBirdStaticPath, path)
}

func IsHub() bool {
	return common.FileExists(common.CpeHubConfPath)
}

func SetupCpeBird(link common.Link) {
	cl := link.CommonLinkVO()
	cl.Init()
	neighContent := link.GenBirdNeighContent()
	common.WriteFile(link.BirdNeighFileName(), neighContent)
	EnableBirdInclude(common.CpeBirdConfPath, link.BirdNeighFileName())
}

func ResetCpeVpnlink(rvl common.ResetVpnlinkVO) {
	plink := rvl.Plink
	err1 := exec.Command(common.CpeBirdResetPath, plink.IntfName(), plink.Filter(rvl.DevState())).Run()
	if err1 != nil {
		log.Printf("err1: %s", err1)
		return
	}
	slink := rvl.Slink
	err2 := exec.Command(common.CpeBirdResetPath, slink.IntfName(), slink.Filter(rvl.DevState())).Run()
	if err2 != nil {
		log.Printf("err2: %s", err2)
		return
	}
	err3 := exec.Command("/etc/init.d/bird", "restart").Run()
	if err3 != nil {
		log.Printf("err3: %s", err3)
	}
}

func AllowIncoming(ruleName string, remoteAddr string, localAddr string, localPort int) {
	//防火墙
	uci.AddSection("firewall", ruleName, "rule")
	uci.Set("firewall", ruleName, "name", ruleName)
	uci.Set("firewall", ruleName, "proto", "udp")
	uci.Set("firewall", ruleName, "src", "wan")
	if remoteAddr != "" {
		uci.Set("firewall", ruleName, "src_ip", remoteAddr)
	}
	uci.Set("firewall", ruleName, "target", "ACCEPT")
	uci.Set("firewall", ruleName, "dest_ip", localAddr)
	uci.Set("firewall", ruleName, "dest_port", strconv.Itoa(localPort))

}

func RemoveIncoming(id int) {
	//防火墙
	ruleName := fmt.Sprintf("vl%d", id)
	uci.DelSection("firewall", ruleName)
}

func AddLanZone(name string) {
	values, _ := uci.Get("firewall", "@zone[0]", "network")
	log.Printf("add firewall.network: %s", values)
	exists := false
	for _, v := range values {
		if v == name {
			exists = true
			break
		}
	}
	if !exists {
		lanlist := append(values, name)
		log.Printf("add new firewall.network: %s", lanlist)
		uci.Set("firewall", "@zone[0]", "network", lanlist...)
	}
}

func DelLanZone(name string) {
	lanlist := make([]string, 10)
	values, _ := uci.Get("firewall", "@zone[0]", "network")
	log.Printf("firewall.network: %s", values)
	i := 0
	for _, v := range values {
		if v != name {
			lanlist[i] = v
			i++
		}
	}
	log.Printf("new firewall.network: %s", lanlist)
	if i > 0 {
		uci.Set("firewall", "@zone[0]", "network", lanlist[:i]...)
	} else {
		uci.Del("firewall", "@zone[0]", "network")
	}
}

func AddVnetZone(name string) {
	values, _ := uci.Get("firewall", "@zone[2]", "network")
	log.Printf("add firewall.network: %s", values)
	exists := false
	for _, v := range values {
		if v == name {
			exists = true
			break
		}
	}
	if !exists {
		lanlist := append(values, name)
		log.Printf("add new firewall.network: %s", lanlist)
		uci.Set("firewall", "@zone[2]", "network", lanlist...)
	}
}

func DelVnetZone(name string) {
	list := make([]string, 10)
	values, _ := uci.Get("firewall", "@zone[2]", "network")
	log.Printf("firewall.network: %s", values)
	idx := 0
	for _, v := range values {
		if v != name {
			list[idx] = v
			idx++
		}
	}
	log.Printf("new firewall.network: %s", list)
	if idx > 0 {
		uci.Set("firewall", "@zone[2]", "network", list[:idx]...)
	} else {
		uci.Del("firewall", "@zone[2]", "network")
	}
}

func AddDevToBridge(brdevname string, devname string) {
	values, _ := uci.Get("network", brdevname, "ports")
	log.Printf("orig %s ports: %s", brdevname, values)
	exists := false
	for _, v := range values {
		if v == devname {
			exists = true
			break
		}
	}
	if !exists {
		portlist := append(values, devname)
		log.Printf("new %s ports: %s", brdevname, portlist)
		uci.Set("network", brdevname, "ports", portlist...)
	}
}

func DelDevFromBridge(brdevname string, devname string) {
	portlist := make([]string, 10)
	values, _ := uci.Get("network", brdevname, "ports")
	log.Printf("orig %s ports: %s", brdevname, values)
	idx := 0
	for _, v := range values {
		if v != devname {
			portlist[idx] = v
			idx++
		}
	}
	log.Printf("new %s ports: %s", brdevname, portlist)
	if idx > 0 {
		uci.Set("network", brdevname, "ports", portlist[:idx]...)
	} else {
		uci.Del("network", brdevname, "ports")
	}
}

func AttachVxlanToBridge(vlname string, brname string) {
	values, _ := uci.Get("network", brname, "ports")
	log.Printf("current network.%s.ports: %s", brname, values)
	exists := false
	for _, v := range values {
		if v == vlname {
			exists = true
			break
		}
	}
	if !exists {
		portlist := append(values, vlname)
		log.Printf("add new network.ports: %s", portlist)
		uci.Set("network", brname, "ports", portlist...)
	}
}

func DettachVxlanFromBridge(vlname string, brname string) {
	portlist := make([]string, 10)
	values, _ := uci.Get("network", brname, "ports")
	log.Printf("current network.%s.ports: %s", brname, values)
	i := 0
	for _, v := range values {
		if v != vlname {
			portlist[i] = v
			i++
		}
	}
	log.Printf("After proc network.%s.ports: %s", brname, portlist)
	uci.Set("network", brname, "ports", portlist[:i]...)
}

func EnsureRemoveVxlan(vlname string) {
	vxl, err := netlink.LinkByName(vlname)
	if err == nil {
		netlink.LinkDel(vxl)
	}
}

func PrepareBirdResetScript() {
	content := `#!/bin/sh
if [ $# != 2 ]; then
	echo "Usage: resetbirdstate wg211 master_primary"
	echo "Usage: resetbirdstate wg212 master_secondary"
	echo "Usage: resetbirdstate wg211 backup_primary"
	echo "Usage: resetbirdstate wg212 backup_secondary"
	exit
fi
IntfName=$1
Filter=$2
sed -i "s/import filter .*/import filter ${Filter}_i;/" /etc/bird-${IntfName}.conf
sed -i "s/export filter .*/export filter ${Filter}_o;/" /etc/bird-${IntfName}.conf
`
	common.WriteFile(common.CpeBirdResetPath, content)
	exec.Command("chmod", "+x", common.CpeBirdResetPath).Run()
}

func PrepareQos() {
	if !common.FileExists(common.CpeQosDir) {
		exec.Command("mkdir", "-p", common.CpeQosDir).Run()
	}
	GenNftDelRuleScript()
	GenIfbNewScript()
	GenIfbMirrScript()
	GenTcMarkScript()
	GenTcpMarkScript()
}

func GenIfbNewScript() {
	template := `#!/bin/sh
IFB=$1
ip li del ${IFB}
ip li add ${IFB} type ifb 1>/dev/null 2>&1
ip li set up dev ${IFB}
tc qdisc add dev ${IFB} root handle 1: prio

`
	path := fmt.Sprintf("%s/ifb-new.sh", common.CpeQosDir)
	common.WriteFile(path, template)
	exec.Command("chmod", "+x", path).Run()
}

func GenIfbMirrScript() {
	template := `#!/bin/sh
INTF=$1
IFB=$2
tc qdisc del dev ${INTF} ingress 1>/dev/null 2>&1
tc qdisc add dev ${INTF} ingress
tc filter add dev ${INTF} parent ffff: protocol all prio 10 u32 match u32 0 0 action connmark action mirred egress redirect dev ${IFB}

`
	path := fmt.Sprintf("%s/ifb-mirr.sh", common.CpeQosDir)
	common.WriteFile(path, template)
	exec.Command("chmod", "+x", path).Run()
}

func GenTcMarkScript() {
	template := `#!/bin/sh
INTF=$1
tc qdisc del dev ${INTF} root 1>/dev/null 2>&1
tc qdisc add dev ${INTF} root handle 1: prio
tc filter add dev ${INTF} parent 1: protocol ip handle 0x10 fw flowid 1:1
tc filter add dev ${INTF} parent 1: protocol ip handle 0x20 fw flowid 1:2
tc filter add dev ${INTF} parent 1: protocol ip handle 0x30 fw flowid 1:3

`
	path := fmt.Sprintf("%s/tc-mark.sh", common.CpeQosDir)
	common.WriteFile(path, template)
	exec.Command("chmod", "+x", path).Run()
}

func GenTcpMarkScript() {
	template := `#!/bin/sh
INTF=$1
PARENT=$1
tc qdisc add dev ${INTF} parent ${PARENT} handle 2: prio
tc filter add dev ${INTF} parent 2: protocol ip handle 0x10 fw flowid 2:1
tc filter add dev ${INTF} parent 2: protocol ip handle 0x20 fw flowid 2:2
tc filter add dev ${INTF} parent 2: protocol ip handle 0x30 fw flowid 2:3

`
	path := fmt.Sprintf("%s/tcp-mark.sh", common.CpeQosDir)
	common.WriteFile(path, template)
	exec.Command("chmod", "+x", path).Run()
}

func GenNftDelRuleScript() {
	template := `#!/bin/sh
ZONE=$1
PROTO=$2
line=$(nft -a list table inet qos_${ZONE}|grep "meta l4proto ${PROTO}")
handle=${line#*handle}
nft delete rule inet qos_${ZONE} mark_${ZONE} handle ${handle}

`
	path := fmt.Sprintf("%s/nft-delrule.sh", common.CpeQosDir)
	common.WriteFile(path, template)
	exec.Command("chmod", "+x", path).Run()
}

func GenNftable(baseQos common.BaseQosVO) bool {
	exec.Command("nft", strings.Split(baseQos.DeleteNftable(), " ")...).Run()
	if baseQos.Zone == "wan" {
		common.WriteFile(baseQos.ScriptPath(), baseQos.NftWan())
	} else if baseQos.Zone == "vnet" {
		common.WriteFile(baseQos.ScriptPath(), baseQos.NftVnet())
	}
	err2 := exec.Command("nft", "-f", baseQos.ScriptPath()).Run()
	if err2 != nil {
		log.Printf("Failed load nftable %s: %s", baseQos.TableName(), err2)
		return false
	}
	return true
}

func GetIntfs(bq common.BaseQosVO) ([]string, error) {
	if bq.Zone == "vnet" {
		if names, ok := uci.Get("firewall", "@zone[2]", "name"); ok && names[0] == "vnet" {
			if networks, ok := uci.Get("firewall", "@zone[2]", "network"); ok {
				return networks, nil
			}
		}
	} else if bq.Zone == "wan" {
		if names, ok := uci.Get("network", "wan", "device"); ok {
			return names, nil
		}
	}
	return nil, errors.New("error zone")
}

func setBaseHtb(intf string, rate int) {
	qtemp := "qdisc add dev %s root handle 1: htb default 1"
	ctemp := "class add dev %s parent 1: classid 1:1 htb rate %dMbit"
	qcmd := fmt.Sprintf(qtemp, intf)
	ccmd := fmt.Sprintf(ctemp, intf, rate)

	exec.Command("tc", strings.Split(qcmd, " ")...).Run()
	exec.Command("tc", strings.Split(ccmd, " ")...).Run()
}

func setIfb(ifb string, intf string) bool {
	ifbNewCmd := fmt.Sprintf("%s/ifb-new.sh", common.CpeQosDir)
	err1 := exec.Command(ifbNewCmd, ifb).Run()
	if err1 != nil {
		log.Printf("Failed ifb-new.sh %s: %s", ifb, err1)
		return false
	}
	ifbMirrCmd := fmt.Sprintf("%s/ifb-mirr.sh", common.CpeQosDir)
	err3 := exec.Command(ifbMirrCmd, intf, ifb).Run()
	if err3 != nil {
		log.Printf("Failed ifb-mirr.sh %s %s: %s", intf, ifb, err3)
		return false
	}
	return true
}

func SetBaseHtb(intf string, rate int) {
	setBaseHtb(intf, rate)
	ifb := "ifbvnet"
	setIfb(ifb, intf)
	setBaseHtb(ifb, rate)
}

func SetupQos(baseQos common.BaseQosVO) bool {
	ifb := baseQos.IfbName()

	if ok := GenNftable(baseQos); !ok {
		return false
	}
	intfs, err := GetIntfs(baseQos)
	if err != nil {
		log.Printf("Failed get interfaces %s", err)
		return false
	}

	for _, intf := range intfs {
		if baseQos.Zone == "vnet" {
			intfMarkCmd := fmt.Sprintf("%s/tcp-mark.sh", common.CpeQosDir)
			err4 := exec.Command(intfMarkCmd, intf).Run()
			if err4 != nil {
				log.Printf("Failed tcp-mark.sh %s: %s", intf, err4)
			}
			ifbMarkCmd := fmt.Sprintf("%s/tcp-mark.sh", common.CpeQosDir)
			err5 := exec.Command(ifbMarkCmd, ifb).Run()
			if err5 != nil {
				log.Printf("Failed tcp-mark.sh %s: %s", ifb, err5)
			}
		} else {
			setIfb(ifb, intf)
			intfMarkCmd := fmt.Sprintf("%s/tc-mark.sh", common.CpeQosDir)
			err4 := exec.Command(intfMarkCmd, intf).Run()
			if err4 != nil {
				log.Printf("Failed tc-mark.sh %s: %s", intf, err4)
			}

			ifbMarkCmd := fmt.Sprintf("%s/tc-mark.sh", common.CpeQosDir)
			err5 := exec.Command(ifbMarkCmd, ifb).Run()
			if err5 != nil {
				log.Printf("Failed tc-mark.sh %s: %s", ifb, err5)
			}
		}
		vmapOifCmd := baseQos.VmapOifCmd(intf)
		err6 := exec.Command("nft", strings.Split(vmapOifCmd, " ")...).Run()
		if err6 != nil {
			log.Printf("failed: nft %s %s", vmapOifCmd, err6)
		}

		vmapIifCmd := baseQos.VmapIifCmd(intf)
		err7 := exec.Command("nft", strings.Split(vmapIifCmd, " ")...).Run()
		if err7 != nil {
			log.Printf("failed: nft %s %s", vmapIifCmd, err7)
		}
	}
	return true
}

func AddQos(qos common.QosVO) bool {
	addrCmds := qos.AddrCmds()
	for _, cmd := range addrCmds {
		err := exec.Command("nft", strings.Split(cmd, " ")...).Run()
		if err != nil {
			log.Printf("CMD: nft %s error: %s", cmd, err)
		}
	}

	serviceCmds := qos.ServiceCmds()
	for _, cmd := range serviceCmds {
		err := exec.Command("nft", strings.Split(cmd, " ")...).Run()
		if err != nil {
			log.Printf("CMD: nft %s error: %s", cmd, err)
		}
	}

	protoCmds := qos.ProtocolMarkCmds()
	// protoCmds := qos.ProtocolPriorityCmds()
	for _, cmd := range protoCmds {
		err := exec.Command("nft", strings.Split(cmd, " ")...).Run()
		if err != nil {
			log.Printf("CMD: nft %s error: %s", cmd, err)
		}
	}
	return true
}

func DelQos(qos common.QosVO) bool {
	addrCmds := qos.AddrDelCmds()
	for _, cmd := range addrCmds {
		err := exec.Command("nft", strings.Split(cmd, " ")...).Run()
		if err != nil {
			log.Printf("CMD: nft %s error: %s", cmd, err)
		}
	}
	serviceCmds := qos.ServiceDelCmds()
	for _, cmd := range serviceCmds {
		err := exec.Command("nft", strings.Split(cmd, " ")...).Run()
		if err != nil {
			log.Printf("CMD: nft %s error: %s", cmd, err)
		}
	}
	delruleCmd := fmt.Sprintf("%s/nft-delrule.sh", common.CpeQosDir)
	for _, proto := range qos.ProtocolRules {
		if proto != "" {
			err := exec.Command(delruleCmd, qos.Zone, proto).Run()
			if err != nil {
				log.Printf("nft-delrule.sh %s %s error: %s", qos.Zone, proto, err)
			}
		}
	}
	return true
}

func ResetFirstBoot() {
	s := time.After(time.Second * 3)
	<-s
	exec.Command("/sbin/firstboot", "-r", "-y").Run()
}
