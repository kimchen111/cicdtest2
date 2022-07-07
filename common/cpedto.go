package common

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"

	"github.com/digineo/go-uci"
	"github.com/florianl/go-tc"
	"github.com/florianl/go-tc/core"
	"golang.org/x/sys/unix"
)

type WanVO struct {
	Intfname string `json:"intfname"` //网卡名
	Protocol string `json:"protocol"` //协议
	Ipaddr   string `json:"ipaddr"`   //IP地址
	Netmask  string `json:"netmask"`  //子网掩码
	Gateway  string `json:"gateway"`  //网关
	Enable   bool   `json:"enable"`   //启用/禁用
}

type PubLanVO struct {
	Name string `json:"name,omitempty"` //LAN的名称，必须为lan,lan1,lan2,lan3这样
}

func (pl *PubLanVO) IncludeFilePath() string {
	return fmt.Sprintf("/etc/bird-br%s.onf", pl.Name)
}

func (pl *PubLanVO) GenContent() string {
	temp := `
protocol direct br%s {
	ipv4;
	interface "br-%s";
}
`
	return fmt.Sprintf(temp, pl.Name, pl.Name)
}

type LanVO struct {
	Name     string   `json:"name,omitempty"`     //LAN的名称，必须为lan,lan1,lan2,lan3这样
	Devices  []string `json:"devices,omitempty"`  //网卡列表
	Protocol string   `json:"protocol,omitempty"` //协议
	Ipaddr   string   `json:"ipaddr,omitempty"`   //IP地址
	Netmask  string   `json:"netmask,omitempty"`  //子网掩码
}

func (lan LanVO) BridgeName() string {
	return fmt.Sprintf("br-%s", lan.Name)
}
func (lan LanVO) DevName() string {
	return fmt.Sprintf("%sdev", lan.Name)
}

type BaseQosVO struct {
	Zone string `json:"zone"` //Zone
}

func (bq *BaseQosVO) Validate() bool {
	return bq.Zone == "vnet" || bq.Zone == "wan"
}

func (bq *BaseQosVO) IfbName() string {
	return fmt.Sprintf("ifb%s", bq.Zone)
}

func (bq *BaseQosVO) TableName() string {
	return fmt.Sprintf("qos_%s", bq.Zone)
}

func (bq *BaseQosVO) DeleteNftable() string {
	return fmt.Sprintf("delete table inet %s", bq.TableName())
}

func (bq *BaseQosVO) VmapIifCmd(intf string) string {
	temp := "add rule inet qos_%s inc meta iifname %s counter goto mark_%s"
	return fmt.Sprintf(temp, bq.Zone, intf, bq.Zone)
}

func (bq *BaseQosVO) VmapOifCmd(intf string) string {
	temp := "add rule inet qos_%s outc meta oifname %s counter goto mark_%s"
	return fmt.Sprintf(temp, bq.Zone, intf, bq.Zone)
}

func (bq *BaseQosVO) ScriptPath() string {
	return fmt.Sprintf("/tmp/%s.nft", bq.TableName())
}

func (bq *BaseQosVO) NftWan() string {
	return `table inet qos_wan {
	set band1_service {
		type inet_service
	}

	set band2_service {
		type inet_service
	}

	set band3_service {
		type inet_service
	}

	set band1_addr {
		type ipv4_addr
		flags interval
	}

	set band2_addr {
		type ipv4_addr
		flags interval
	}

	set band3_addr {
		type ipv4_addr
		flags interval
	}

	chain mark_wan {
		meta l4proto { tcp, udp } th sport @band1_service meta mark set 0x10 ct mark set 0x10 counter return
		meta l4proto { tcp, udp } th dport @band1_service meta mark set 0x10 ct mark set 0x10 counter return
		ip saddr @band1_addr meta mark set 0x10 ct mark set 0x10 counter return
		ip daddr @band1_addr meta mark set 0x10 ct mark set 0x10 counter return

		meta l4proto { tcp, udp } th sport @band2_service meta mark set 0x20 ct mark set 0x20 counter return
		meta l4proto { tcp, udp } th dport @band2_service meta mark set 0x20 ct mark set 0x20 counter return
		ip saddr @band2_addr meta mark set 0x20 ct mark set 0x20 counter return
		ip daddr @band2_addr meta mark set 0x20 ct mark set 0x20 counter return

		meta l4proto { tcp, udp } th sport @band3_service meta mark set 0x30 ct mark set 0x30 counter return
		meta l4proto { tcp, udp } th dport @band3_service meta mark set 0x30 ct mark set 0x30 counter return
		ip saddr @band3_addr meta mark set 0x30 ct mark set 0x30 counter return
		ip daddr @band3_addr meta mark set 0x30 ct mark set 0x30 counter return
	}
	
	chain inc {
		type filter hook prerouting priority mangle; policy accept;
	}

	chain outc {
		type filter hook postrouting priority mangle; policy accept;
	}

}
`
}

func (bq *BaseQosVO) NftVnet() string {
	return `table inet qos_vnet {
	set band1_service {
		type inet_service
	}

	set band2_service {
		type inet_service
	}

	set band3_service {
		type inet_service
	}

	set band1_addr {
		type ipv4_addr
		flags interval
	}

	set band2_addr {
		type ipv4_addr
		flags interval
	}

	set band3_addr {
		type ipv4_addr
		flags interval
	}

	chain mark_vnet {
		meta l4proto { tcp, udp } th sport @band1_service meta mark set 0x10 ct mark set 0x10 counter return
		meta l4proto { tcp, udp } th dport @band1_service meta mark set 0x10 ct mark set 0x10 counter return
		ip saddr @band1_addr meta mark set 0x10 ct mark set 0x10 counter return
		ip daddr @band1_addr meta mark set 0x10 ct mark set 0x10 counter return

		meta l4proto { tcp, udp } th sport @band2_service meta mark set 0x20 ct mark set 0x20 counter return
		meta l4proto { tcp, udp } th dport @band2_service meta mark set 0x20 ct mark set 0x20 counter return
		ip saddr @band2_addr meta mark set 0x20 ct mark set 0x20 counter return
		ip daddr @band2_addr meta mark set 0x20 ct mark set 0x20 counter return

		meta l4proto { tcp, udp } th sport @band3_service meta mark set 0x30 ct mark set 0x30 counter return
		meta l4proto { tcp, udp } th dport @band3_service meta mark set 0x30 ct mark set 0x30 counter return
		ip saddr @band3_addr meta mark set 0x30 ct mark set 0x30 counter return
		ip daddr @band3_addr meta mark set 0x30 ct mark set 0x30 counter return
	}

	chain inc {
		type filter hook prerouting priority mangle; policy accept;
	}

	chain outc {
		type filter hook postrouting priority mangle; policy accept;
	}

}
`
}

func (bq *BaseQosVO) GetIntfs() ([]string, error) {
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

func (bq *BaseQosVO) Prepare(intfName string) bool {
	devID, err := net.InterfaceByName(intfName)
	if err != nil {
		log.Printf("could not get interface %s err: %v", intfName, err)
		return false
	}

	qdisc := tc.Object{
		Msg: tc.Msg{
			Family:  unix.AF_UNSPEC,
			Ifindex: uint32(devID.Index),
			Handle:  core.BuildHandle(0x1, 0x0),
			Parent:  tc.HandleRoot,
			Info:    0,
		},
		// configure a very basic hierarchy token bucket (htb) qdisc
		Attribute: tc.Attribute{
			Kind: "htb",
			Htb: &tc.Htb{
				Init: &tc.HtbGlob{
					Version:      0x3,
					Rate2Quantum: 0xa,
				},
			},
		},
	}

	tcnl, _ := tc.Open(&tc.Config{})
	defer tcnl.Close()

	err1 := tcnl.Qdisc().Add(&qdisc)
	if err1 != nil {
		fmt.Fprintf(os.Stderr, "could not assign htb to interface %s: %v\n", intfName, err1)
		return false
	}
	return true
}

func (bq *BaseQosVO) Setup() {
	intfs, err := bq.GetIntfs()
	if err != nil {
		log.Printf("Failed get interfaces %s", err)
		return
	}
	for _, intf := range intfs {
		bq.Prepare(intf)
	}
}

// func (bq *BaseQosVO) InRateStr() string {
// 	return fmt.Sprintf("%dkbit", bq.InRate)
// }
// func (bq *BaseQosVO) OutRateStr() string {
// 	return fmt.Sprintf("%dkbit", bq.OutRate)
// }
// func (bq *BaseQosVO) DeftInRateStr() string {
// 	defBW := int64(0.2 * float64(bq.InRate))
// 	return fmt.Sprintf("%dkbit", defBW)
// }
// func (bq *BaseQosVO) DeftOutRateStr() string {
// 	defBW := int64(0.5 * float64(bq.OutRate))
// 	return fmt.Sprintf("%dkbit", defBW)
// }

type QosVO struct {
	Zone          string   `json:"Zone"`                    //区域
	Prio          int      `json:"prio"`                    //优先级，值是 1 2 3，分别对应 band1(1:1) band2(1:2) band3(1:3)
	AddrRules     []string `json:"addrRules,omitempty"`     //地址规则列表，统配saddr和daddr
	ServiceRules  []string `json:"serviceRules,omitempty"`  //服务规则列表，统配sport和dport
	ProtocolRules []string `json:"protocolRules,omitempty"` //协议，需独立使用，不支持与端口一起配合使用
}

func (qos *QosVO) Validate() bool {
	return qos.Prio == 1 || qos.Prio == 2 || qos.Prio == 3
}

func (qos *QosVO) AddrCmds() []string {
	temp := "add element inet qos_%s band%d_addr { %s }"
	cmds := make([]string, len(qos.AddrRules))
	count := 0
	for i, rule := range qos.AddrRules {
		if rule != "" {
			cmds[i] = fmt.Sprintf(temp, qos.Zone, qos.Prio, rule)
			count++
		}
	}
	return cmds[:count]
	// strings.Join(temp, ",")
}

func (qos *QosVO) AddrDelCmds() []string {
	temp := "delete element inet qos_%s band%d_addr { %s }"
	cmds := make([]string, len(qos.AddrRules))
	count := 0
	for i, rule := range qos.AddrRules {
		if rule != "" {
			cmds[i] = fmt.Sprintf(temp, qos.Zone, qos.Prio, rule)
			count++
		}
	}
	return cmds[:count]
	// strings.Join(temp, ",")
}

func (qos *QosVO) ServiceCmds() []string {
	temp := "add element inet qos_%s band%d_service { %s }"
	cmds := make([]string, len(qos.ServiceRules))
	count := 0
	for i, rule := range qos.ServiceRules {
		if rule != "" {
			cmds[i] = fmt.Sprintf(temp, qos.Zone, qos.Prio, rule)
			count++
		}
	}
	return cmds[:count]
	// strings.Join(temp, ",")
}

func (qos *QosVO) ServiceDelCmds() []string {
	temp := "delete element inet qos_%s band%d_service { %s }"
	cmds := make([]string, len(qos.ServiceRules))
	count := 0
	for i, rule := range qos.ServiceRules {
		if rule != "" {
			cmds[i] = fmt.Sprintf(temp, qos.Zone, qos.Prio, rule)
			count++
		}
	}
	return cmds[:count]
	// strings.Join(temp, ",")
}

func (qos *QosVO) ProtocolMarkCmds() []string {
	temp := "add rule inet qos_%s mark_%s meta l4proto %s meta mark set 0x%d0 ct mark set mark counter accept"
	cmds := make([]string, len(qos.ProtocolRules))
	count := 0
	for i, rule := range qos.ProtocolRules {
		if rule != "" {
			cmds[i] = fmt.Sprintf(temp, qos.Zone, qos.Zone, rule, qos.Prio)
			count++
		}
	}
	return cmds[:count]
}

func (qos *QosVO) ProtocolPriorityCmds() []string {
	temp := "add rule inet qos_%s mark_%s meta l4proto %s meta priority set 1:%d counter accept"
	cmds := make([]string, len(qos.ProtocolRules))
	count := 0
	for i, rule := range qos.ProtocolRules {
		if rule != "" {
			cmds[i] = fmt.Sprintf(temp, qos.Zone, qos.Zone, rule, qos.Prio)
			count++
		}
	}
	return cmds[:count]
	// strings.Join(temp, ",")
}

type RouteVO struct {
	Dst      string `json:"dst"`               //目标
	Gateway  string `json:"gateway"`           //网关
	Dev      string `json:"dev"`               //设备
	Protocol string `json:"protocol"`          //协议
	Prefsrc  string `json:"prefsrc,omitempty"` //源IP地址
}

type NetworkInfoVO struct {
	Gateway RouteVO `json:"gateway"` //默认路由，默认网关
	Wan     []WanVO `json:"wan"`     //WAN
	Lan     []LanVO `json:"lan"`     //LAN
}

type LanDhcpVO struct {
	LanName string `json:"lanName"`         //LAN名称
	Start   int    `json:"start,omitempty"` //起始IP
	End     int    `json:"end,omitempty"`   //结束IP
	// Gateway string `json:"gateway"` //默认网关
	// Dns     string `json:"dns"`     //DNS
}

type TrafficAnalysisVO struct {
	DeviceId int `json:"deviceId"` //DeviceID
}

type VpeVO struct {
	Esn     string   `json:"esn"`     //ESN
	IpAddrs []string `json:"ipAddrs"` //IP地址列表
}

type VpeDetectVO struct {
	Vpes []VpeVO `json:"vpes"`
}

type StaticRouteVO struct {
	Target  string `json:"target"`            //网段列表
	Via     string `json:"via,omitempty"`     //下一跳
	Metric  int    `json:"metric,omitempty"`  //跃点数
	Publish bool   `json:"publish,omitempty"` //是否发布
}

func (sr *StaticRouteVO) Key() string {
	temp := strings.ReplaceAll(sr.Target, ".", "")
	return strings.ReplaceAll(temp, "/", "")
}

func (sr *StaticRouteVO) IncludeFilePath() string {
	return fmt.Sprintf("/etc/bird-st%s.onf", sr.Key())
}

func (sr *StaticRouteVO) GenContent() string {
	temp := `
protocol static nost%s {
	ipv4;
	route %s via %s;
}
`
	return fmt.Sprintf(temp, sr.Key(), sr.Target, sr.Via)
}

func (sr *StaticRouteVO) GenPubContent() string {
	temp := `
protocol static st%s {
	ipv4;
	route %s via %s;
}
`
	return fmt.Sprintf(temp, sr.Key(), sr.Target, sr.Via)
}

func (sr *StaticRouteVO) Gen() string {
	if sr.Publish {
		return sr.GenPubContent()
	} else {
		return sr.GenContent()
	}
}

type OutPort struct {
	IntfName string `json:"intfName"` //哪个端口的名称
	Target   string `json:"target"`   //目标地址
}

func (op *OutPort) Key() string {
	temp := strings.ReplaceAll(op.Target, ".", "")
	return strings.ReplaceAll(temp, "/", "")
}

func (op *OutPort) IncludeFilePath() string {
	return fmt.Sprintf("/etc/bird-st%s.onf", op.Key())
}

func (op *OutPort) IntfAddr() string {
	return GetInterfaceAddr(op.IntfName)
}

func (op *OutPort) GenContent() string {
	temp := `
protocol static nost%s {
	ipv4;
	route %s via %s;
}
`
	return fmt.Sprintf(temp, op.Key(), op.Target, op.IntfAddr())
}

type InitVO struct {
	KeepWan bool `json:"keepWan"` //是否保留WAN的配置
}

type VrrpStateVO struct {
	State     string `json:"state"`               //期望切换成的状态
	Permanent bool   `json:"permanent,omitempty"` //是否永久抢占
}

type VipAddrVO struct {
	LanName   string `json:"lanName"`             //网卡名称
	VipAddr   string `json:"vipAddr"`             //IP地址，带掩码
	SolidAddr string `json:"solidAddr,omitempty"` //保留的固定IP地址，带掩码
}

func (va *VipAddrVO) LanIntfName() string {
	return fmt.Sprintf("br-%s", va.LanName)
}

type VrrpVO struct {
	Rid        int         `json:"rid"`        //RouterID
	PreState   string      `json:"preState"`   //原始角色 MASTER BACKUP
	HBIntfName string      `json:"hbIntfName"` //心跳网卡
	HBIpAddr   string      `json:"hbIpAddr"`   //心跳网卡的IP地址，自动产生
	Priority   string      `json:"priority"`   //优先级
	VipAddrs   []VipAddrVO `json:"vipAddrs"`   //虚拟IP列表
}

/*
! Configuration File for keepalived

global_defs {
    router_id LVS_DEVEL
}

vrrp_instance VI_1 {
    state MASTER
    interface br-lan
    virtual_router_id 51
    priority 100
    advert_int 1
    authentication {
        auth_type PASS
        auth_pass 1111
    }

    virtual_ipaddress {
    192.168.1.1/24 dev br-lan
    172.23.48.2/20 dev eth0
    }
}
*/
func (vrrp *VrrpVO) GenEnableContent() string {
	weight := "-5"
	if vrrp.PreState == "MASTER" {
		weight = "-20"
	}
	template := `! Configuration File for keepalived

global_defs {
	router_id LVS_DEVEL
}
vrrp_script checkHold
{
	script "/sdwan/keepalived/check_hold.sh"
	interval 1
	weight %s
}
vrrp_instance VI_1 {
	state %s

	interface %s
	virtual_router_id %d
	priority %s
	advert_int 1
	notify /sdwan/keepalived/notify.sh
	authentication {
		auth_type PASS
		auth_pass SYSCLOUD
	}
    track_script 
    {
        checkHold
    }
	virtual_ipaddress {
	%s
	}
}	
`
	vips := ""
	for _, vip := range vrrp.VipAddrs {
		vips += fmt.Sprintf("%s dev %s\n", vip.VipAddr, vip.LanIntfName())
	}
	content := fmt.Sprintf(template,
		weight,
		vrrp.PreState,
		vrrp.HBIntfName,
		vrrp.Rid,
		vrrp.Priority,
		vips)
	return content
}

type HaMemberVO struct {
	Esn        string      `json:"esn"`        //ESN
	HBIntfName string      `json:"hbIntfName"` //心跳网卡
	VipAddrs   []VipAddrVO `json:"vipAddrs"`   //虚拟IP列表
}

type HaVO struct {
	Master HaMemberVO `json:"master"` //主机
	Backup HaMemberVO `json:"backup"` //备机
}

func (ha *HaVO) GenVrrp() (master VrrpVO, backup VrrpVO) {
	rid := rand.Intn(200)
	master = VrrpVO{
		Rid:        rid,
		PreState:   "MASTER",
		HBIntfName: ha.Master.HBIntfName,
		HBIpAddr:   "169.254.254.253/30",
		Priority:   "50",
		VipAddrs:   ha.Master.VipAddrs,
	}

	backup = VrrpVO{
		Rid:        rid,
		PreState:   "BACKUP",
		HBIntfName: ha.Backup.HBIntfName,
		HBIpAddr:   "169.254.254.254/30",
		Priority:   "45",
		VipAddrs:   ha.Backup.VipAddrs,
	}
	return
}
