package common

import (
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strings"

	"github.com/digineo/go-uci"
	"github.com/vishvananda/netlink"
)

//========================================Vxlan Tunnel=================================================================
// Port         string `json:"port"`         //监听哪个端口，很神奇，linux的ip 使用dstport作为监听端口，两边一致
// Intf         string `json:"intf"`         //使用哪个网卡出去

type VxlanTunnelVO struct {
	Esn          string `json:"esn"`          //ESN
	Id           int    `json:"id"`           //id
	Vni          int    `json:"vni"`          //VNI
	SelfIpaddr   string `json:"selfIpaddr"`   //本机地址
	RemoteIpaddr string `json:"remoteIpaddr"` //远程地址
}

func (vto *VxlanTunnelVO) VxlanName() string {
	return fmt.Sprintf("vl%d", vto.Id)
}

func (vto *VxlanTunnelVO) HubBridgeName() string {
	return fmt.Sprintf("hub%d", vto.Vni)
}

func (vto *VxlanTunnelVO) VpeBridgeName() string {
	return fmt.Sprintf("br%d", vto.Vni)
}

func (vto *VxlanTunnelVO) NetnsName() string {
	return fmt.Sprintf("ns%d", vto.Vni)
}

/*
func (vto *VxlanTunnelVO) SplitIpaddr() (ipaddr string, netmask string) {
	ipaddr = PureIp(vto.VtepAddr)
	netmask = NetmaskStr(vto.VtepAddr)
	return
}
*/

func (vto *VxlanTunnelVO) Port() int {
	return 40000 + vto.Id
}

func (vto *VxlanTunnelVO) Path() (string, bool) {
	return fmt.Sprintf("%s/%s", CpeVxlanConfDir, vto.VxlanName()), true
}

func (vto *VxlanTunnelVO) WriteConf() {
	if path, ok := vto.Path(); ok {
		WriteConfig(path, vto)
	}
}

func (vto *VxlanTunnelVO) RemoveConf() {
	if path, ok := vto.Path(); ok {
		os.Remove(path)
	}
}

/*
ip link add vt10001 type vxlan id 10001 dstport 10001 remote 192.168.3.1 local 192.168.3.254 #dev eth1
ip li set up dev vt10001
ip a add 10.0.22.1/24 dev vt10001
*/
func (vto *VxlanTunnelVO) Start() bool {
	laddr := net.ParseIP(vto.SelfIpaddr)
	raddr := net.ParseIP(vto.RemoteIpaddr)
	name := vto.VxlanName()
	la := netlink.NewLinkAttrs()
	la.Name = name
	err1 := netlink.LinkAdd(&netlink.Vxlan{
		LinkAttrs: la,
		VxlanId:   vto.Vni,
		Port:      vto.Port(),
		SrcAddr:   laddr,
		Group:     raddr,
	})
	if err1 != nil {
		log.Printf("add vxlan failed %v ", err1)
		return false
	}
	vxlan, _ := netlink.LinkByName(name)
	netlink.LinkSetUp(vxlan)
	// brname := vto.VpeBridgeName()
	brname := vto.HubBridgeName()
	br, err2 := netlink.LinkByName(brname)
	if err2 != nil {
		log.Printf("get bridge failed %v ", err2)
		return false
	}
	err4 := netlink.LinkSetMaster(vxlan, br)
	if err4 != nil {
		log.Printf("Attach vxlan interface to bridge failed %v ", err4)
		return false
	}
	return true
}

func (vto *VxlanTunnelVO) Stop() {
	vxl, err := netlink.LinkByName(vto.VxlanName())
	if err == nil {
		netlink.LinkDel(vxl)
	}
}

type HubVO struct {
	Vni      int      `json:"vni"`      //VNI，用于在HUB上创建Vtep
	Role     string   `json:"role"`     //HUB节点的角色 RR HUB
	VtepAddr string   `json:"vtepAddr"` //HUB节点的VTEP地址
	RrAddrs  []string `json:"rrAddrs"`  //HUB虚拟网的RR的地址
	CpeCidr  string   `json:"cpeCidr"`  //连接当前HUB站点的CPE所有用互联地址的网段
}

func (hub *HubVO) VtepNetwork() string {
	return Network(hub.VtepAddr)
}

func (hub *HubVO) VtepPureAddr() string {
	return PureIp(hub.VtepAddr)
}

func (hub *HubVO) GenHubConf() {
	roleContent := fmt.Sprintf("ROLE=%s", hub.Role)
	WriteFile(CpeHubConfPath, roleContent)
}
func (hub *HubVO) RemoveHubConf() {
	if FileExists(CpeHubConfPath) {
		os.Remove(CpeHubConfPath)
	}
}

func (hub *HubVO) BridgeName() string {
	return fmt.Sprintf("hub%d", hub.Vni)
}

func (hub *HubVO) InitBridge() bool {
	_, err := netlink.LinkByName(hub.BridgeName())
	if err != nil {
		br, err := EnsureBridge(hub.BridgeName())
		if err != nil {
			return false
		}
		AddIpaddr(br, hub.VtepAddr)
	}
	return true
}

func (hub *HubVO) Removeridge() bool {
	link, err := netlink.LinkByName(hub.BridgeName())
	if err != nil {
		return true
	}
	err = netlink.LinkDel(link)
	if err != nil {
		log.Printf("Delete bridge failed: %s", err)
		return false
	}
	return true
}

func (hub *HubVO) GenBgpContent() string {
	temp := `
log syslog all;

router id %s;

define VPE_ASN = %d;
define HUB_ASN = %d;
define CPE_ASN = %d;

define CPE_CIDR = %s;

protocol device {
	scan time 10;
}

protocol direct hub {
    interface "hub*";
    ipv4 {
        import all;
    };
}

protocol direct wgi {
    interface "wg*";
    ipv4 {
        import all;
    };
}

filter ignore_local
{
	if proto = "wgi" then {
		reject;
  	}
	if proto = "hub" then {
		reject;
  	}
	if proto ~ "br*" then {
		reject;
  	}
	accept;
}

protocol kernel {
	ipv4 {
		export filter ignore_local;
	};
	scan time 15;
}

function repl_cpe_asn() {
	if bgp_path.first = CPE_ASN then {
		bgp_path.delete(CPE_ASN);
		bgp_path.prepend(HUB_ASN);
	}
}

filter ignore_local_and_repl_asn {
	if proto = "wgi" then {
		reject;
  	}
	if proto = "hub" then {
		reject;
  	}
	repl_cpe_asn();
	accept;
}
include "%s";

template bgp VPE_TEMP {
    debug all;
    description "VPES";
    local as HUB_ASN;
    multihop 8;
    connect retry time 10;
    hold time 30;
    allow bgp_local_pref on;
}

protocol bgp CPES {
	debug all;
	description "CPES";
	local as HUB_ASN;
	multihop 8;
	connect retry time 10;
	hold time 30;
	allow bgp_local_pref on;
	ipv4 {
		next hop self;
		import all;
		export filter ignore_local_and_repl_asn;
	};
	neighbor range CPE_CIDR as CPE_ASN;
}
`
	result := fmt.Sprintf(temp,
		GBC.GetRouterID(),
		VpeAsNO(hub.Vni),
		HubAsNO(hub.Vni),
		CpeAsNO(hub.Vni),
		hub.CpeCidr,
		CpeBirdStaticPath)
	return result
}

func (hub *HubVO) GenRrServerBgpContent() string {
	temp := `
protocol bgp HUBS { 
	debug all;
	description "HUBS";
	local as HUB_ASN;
	multihop 8;
	rr client;
	connect retry time 10;
	hold time 30;
	allow bgp_local_pref on;
	source address %s;
	ipv4 {
		import all;
		export filter ignore_local;
	};
	neighbor range %s internal;
}
`
	result := fmt.Sprintf(temp,
		hub.VtepPureAddr(),
		hub.VtepNetwork())
	return result
}

func (hub *HubVO) GenRrClientBgpContent() string {
	temp := `
protocol bgp RR%d {
	debug all;
	description "RR%d-igp";
	local as HUB_ASN;
	multihop;
	connect retry time 10;
	hold time 30;
	source address %s;
	neighbor %s internal;
	ipv4 {
		next hop self;
		import all;
		export filter ignore_local;
	};
}
`
	content := ""
	for i, rrAddr := range hub.RrAddrs {
		content += fmt.Sprintf(temp,
			i,
			i,
			hub.VtepPureAddr(),
			rrAddr)
	}
	return content
}

type TunnelPeerVO struct {
	Esn    string `json:"esn"`              //ESN
	Ipaddr string `json:"ipaddr,omitempty"` //IP地址
	Role   string `json:"role,omitempty"`   //角色，不用填写，自动获取
}

type TunnelVO struct {
	Id    int          `json:"id"`    //用于系统标识网卡名称，需要全局唯一，可以使用隧道的自增长字段
	Vni   int          `json:"vni"`   //VNI，用于在VPE上放入VRF
	PeerA TunnelPeerVO `json:"peerA"` //A端
	PeerB TunnelPeerVO `json:"peerB"` //B端
}

func (tunnel TunnelVO) MixVOA() VxlanTunnelVO {
	r := VxlanTunnelVO{
		Esn:          tunnel.PeerA.Esn,
		Id:           tunnel.Id,
		Vni:          tunnel.Vni,
		SelfIpaddr:   tunnel.PeerA.Ipaddr,
		RemoteIpaddr: tunnel.PeerB.Ipaddr,
	}
	return r
}
func (tunnel TunnelVO) MixVOB() VxlanTunnelVO {
	r := VxlanTunnelVO{
		Esn:          tunnel.PeerB.Esn,
		Id:           tunnel.Id,
		Vni:          tunnel.Vni,
		SelfIpaddr:   tunnel.PeerB.Ipaddr,
		RemoteIpaddr: tunnel.PeerA.Ipaddr,
	}
	return r
}

//========================================Wireguard Link=================================================================
/*
func CalcPreferenceMed(devState string, linkState string) (int, int) {
	lp := 100
	med := 0
	if devState == "MASTER" {
		if linkState == "PRIMARY" {
			lp = 100
			med = 0
		} else {
			lp = 90
			med = 10
		}
	} else {
		if linkState == "PRIMARY" {
			lp = 80
			med = 20
		} else {
			lp = 70
			med = 30
		}
	}
	return lp, med
}
*/
type CpeWgPeerVO struct {
	Id       int    `json:"id"`                 //ID，用于标识peer的名称，示例peer-vpeid-vni，可以直接使用vpe的自增长ID
	Host     string `json:"host,omitempty"`     //IP地址
	Port     int    `json:"port,omitempty"`     //端口
	PubKey   string `json:"pubKey,omitempty"`   //对端的公钥
	IntfAddr string `json:"intfAddr,omitempty"` //IP地址，虚拟网卡上的IP地址
}

func (peer *CpeWgPeerVO) PurIntfAddr() string {
	arr := strings.Split(peer.IntfAddr, "/")
	return arr[0]
}

//========================================Vpn Link=================================================================

type VpnServerVO struct {
	Esn        string `json:"esn"`                  //ESN
	Role       string `json:"role,omitempty"`       //角色，不用填写，自动计算
	ListenAddr string `json:"listenAddr,omitempty"` //多公网环境下监听哪个地址，需要和网卡对应
	IntfAddr   string `json:"intfAddr,omitempty"`   //虚拟网卡的IP地址
	ListenIntf string `json:"listenIntf,omitempty"` //多公网环境下监听哪个网卡
	PrivKey    string `json:"privKey,omitempty"`    //私钥
}

func (vs *VpnServerVO) IntfPureAddr() string {
	return PureIp(vs.IntfAddr)
}
func (vs *VpnServerVO) PubKey() string {
	return GenWgPubKey(vs.PrivKey)
}
func (vs *VpnServerVO) GenPrivKey() {
	vs.PrivKey = GenWgPrivKey()
}

type VpnClientVO struct {
	Esn      string `json:"esn"`                //ESN
	Role     string `json:"role"`               //角色，不用填写，自动计算
	IntfAddr string `json:"intfAddr,omitempty"` //虚拟网卡的IP地址
	PrivKey  string `json:"privKey,omitempty"`  //私钥
}

func (vc *VpnClientVO) IntfPureAddr() string {
	return PureIp(vc.IntfAddr)
}
func (vc *VpnClientVO) PubKey() string {
	return GenWgPubKey(vc.PrivKey)
}
func (vc *VpnClientVO) GenPrivKey() {
	vc.PrivKey = GenWgPrivKey()
}

type Link interface {
	CommonLinkVO() LinkVO        //生成LinkVO
	GenBirdNeighContent() string //生成当前interface的bird配置文件内容
	BirdNeighFileName() string   //当前interface的bird配置文件的路径
}

type LinkVO struct {
	Vni int //VNI
}

//因为没有合适的IP地址做为ROUTER ID，所以使用MAC地址计算为IP地址，基本上可以唯一
func (link *LinkVO) RouterID() string {
	return Esn2RID(GBC.GetESN())
}

/*
if devState == "MASTER" {
	if linkState == "PRIMARY" {
		lp = 100
		med = 0
	} else {
		lp = 90
		med = 10
	}
} else {
	if linkState == "PRIMARY" {
		lp = 80
		med = 20
	} else {
		lp = 70
		med = 30
	}
}
*/
func (link *LinkVO) GenBirdContent() string {
	temp := `
log syslog all;

router id %s;

define VPE_ASN = %d;
define HUB_ASN = %d;
define CPE_ASN = %d;

protocol device {
    scan time 10;
}
protocol direct wgi {
    ipv4;
    interface "wg*";
}
filter ignore_local {
  if proto = "wgi" then {
	reject;
  }
  if proto ~ "br*" then {
    reject;
  }
  accept;
}
protocol kernel {
    ipv4 {
        export filter ignore_local;
    };
    scan time 15;
}
#function ignore_local_and_bgp(int link_med) {
#    if proto = "bgppeer" then {
#	    return false;
#    }
#    if proto = "wgi" then {
#        return false;
#    }
#    if proto ~ "nost*" then {
#        return false;
#    }
#	bgp_med=link_med;
#    return true;
#}

filter master_primary_i {
	bgp_local_pref=100;
	accept;
}
filter master_secondary_i {
	bgp_local_pref=90;
	accept;
}
filter backup_primary_i {
	bgp_local_pref=80;
	accept;
}
filter backup_secondary_i {
	bgp_local_pref=70;
	accept;
}

filter master_primary_o {
	bgp_med  = 0;
	accept;
}
filter master_secondary_o {
	bgp_med  = 10;
	accept;
}
filter backup_primary_o {
	bgp_med  = 20;
	accept;
}
filter backup_secondary_o {
	bgp_med  = 30;
	accept;
}

include "%s";

template bgp VPE_TEMP {
    debug all;
    multihop 8;
    connect retry time 10;
    hold time 30;
    allow bgp_local_pref on;
	local as CPE_ASN;
}

template bgp HUB_TEMP {
    debug all;
    multihop 8;
    connect retry time 10;
    hold time 30;
    allow bgp_local_pref on;
	local as CPE_ASN;
}

`
	result := fmt.Sprintf(temp,
		link.RouterID(),
		VpeAsNO(link.Vni),
		HubAsNO(link.Vni),
		CpeAsNO(link.Vni),
		CpeBirdStaticPath)
	return result
}

func (link *LinkVO) FirstLink() bool {
	ifaces, _ := uci.GetSections("network", "interface")
	wgregexp, _ := regexp.Compile("^wg.*")
	dirregexp, _ := regexp.Compile("^dir.*")
	count := 0
	for _, iface := range ifaces {
		if wgregexp.MatchString(iface) || dirregexp.MatchString(iface) {
			count++
		}
	}
	// log.Println("Current ifaces: ", ifaces)
	return count == 0
}

func (link *LinkVO) Init() {
	if link.FirstLink() {
		log.Printf("Create new bird config file.")
		content := link.GenBirdContent()
		WriteFile(CpeBirdConfPath, content)
		WriteFile(CpeBirdStaticPath, "") //预先写一个空的静态文件
	}
}

type VpnlinkVO struct {
	Id       int         `json:"id"`     //ID，用于生成网卡名称，wg0 0就是id，可以使用链路表的自增长ID
	Vni      int         `json:"vni"`    //VNI，用于标识网卡名称
	Rate     int         `json:"rate"`   //速率，Mbps
	State    string      `json:"state"`  //是否主链路 PRIMARY SECONDARY
	Server   VpnServerVO `json:"server"` //服务端信息
	Client   VpnClientVO `json:"client"` //客户端信息
	devState string      //设备的状态，MASTER/BACKUP
}

func (vpnlink *VpnlinkVO) CommonLinkVO() LinkVO {
	return LinkVO{Vni: vpnlink.Vni}
}

func (vpnlink *VpnlinkVO) IsServer() bool {
	return GBC.GetESN() == vpnlink.Server.Esn
}

func (vpnlink *VpnlinkVO) FirewallRuleName() string {
	return fmt.Sprintf("wg%d", vpnlink.Id)
}
func (vpnlink *VpnlinkVO) BirdNeighFileName() string {
	return fmt.Sprintf("/etc/bird-wg%d.conf", vpnlink.Id)
}
func (vpnlink *VpnlinkVO) ListenPort() int {
	return vpnlink.Id + 10000
}
func (vpnlink *VpnlinkVO) PeerName(peerId int) string {
	return fmt.Sprintf("c%d_%d", peerId, vpnlink.Vni)
}
func (vpnlink *VpnlinkVO) IntfName() string {
	return fmt.Sprintf("wg%d", vpnlink.Id)
}
func (vpnlink *VpnlinkVO) SelfRole() string {
	if GBC.GetESN() == vpnlink.Server.Esn {
		return vpnlink.Server.Role
	} else {
		return vpnlink.Client.Role
	}
}
func (vpnlink *VpnlinkVO) PeerRole() string {
	if GBC.GetESN() == vpnlink.Server.Esn {
		return vpnlink.Client.Role
	} else {
		return vpnlink.Server.Role
	}
}
func (vpnlink *VpnlinkVO) NeighASN() string {
	return fmt.Sprintf("%s_ASN", vpnlink.PeerRole())
}
func (vpnlink *VpnlinkVO) ImportFilter() string {
	return fmt.Sprintf("%s_%s_i", strings.ToLower(vpnlink.devState), strings.ToLower(vpnlink.State))
}
func (vpnlink *VpnlinkVO) ExportFilter() string {
	return fmt.Sprintf("%s_%s_o", strings.ToLower(vpnlink.devState), strings.ToLower(vpnlink.State))
}

func (vpnlink *VpnlinkVO) SelfASN() string {
	return fmt.Sprintf("%s_ASN", vpnlink.SelfRole())
}

func (vpnlink *VpnlinkVO) SelfIntfPureAddr() string {
	if GBC.GetESN() == vpnlink.Server.Esn {
		return vpnlink.Server.IntfPureAddr()
	} else {
		return vpnlink.Client.IntfPureAddr()
	}
}

func (vpnlink *VpnlinkVO) SelfIntfAddr() string {
	if GBC.GetESN() == vpnlink.Server.Esn {
		return vpnlink.Server.IntfAddr
	} else {
		return vpnlink.Client.IntfAddr
	}
}

func (vpnlink *VpnlinkVO) PeerIntfPureAddr() string {
	if GBC.GetESN() == vpnlink.Server.Esn {
		return vpnlink.Client.IntfPureAddr()
	} else {
		return vpnlink.Server.IntfPureAddr()
	}
}
func (vpnlink *VpnlinkVO) NetnsName() string {
	return fmt.Sprintf("ns%d", vpnlink.Vni)
}
func (vpnlink *VpnlinkVO) InitKey() {
	vpnlink.Client.GenPrivKey()
	vpnlink.Server.GenPrivKey()
}
func (vpnlink *VpnlinkVO) SetDevState(state string) {
	vpnlink.devState = state
}
func (vpnlink *VpnlinkVO) DevState() string {
	return vpnlink.devState
}

//PEER内容
func (vpnlink *VpnlinkVO) GenBirdNeighContent() string {
	temp := `
protocol bgp %s from %s {
	source address %s;
	neighbor %s as %s;
	ipv4 {
		next hop self;
		import filter %s;
		export filter %s;
	};
}
`
	tempName := fmt.Sprintf("%s_TEMP", vpnlink.PeerRole())
	result := fmt.Sprintf(temp,
		vpnlink.IntfName(),
		tempName,
		vpnlink.SelfIntfPureAddr(),
		vpnlink.PeerIntfPureAddr(),
		vpnlink.NeighASN(),
		vpnlink.ImportFilter(),
		vpnlink.ExportFilter(),
	)
	return result
}

/*
[Interface]
PrivateKey = %s
Address = %s
ListenPort = %d
PostUp = ip li set netns %s dev %s && ip netns exec %s ip addr add %s dev %s && ip netns exec %s ip li set up dev %s

[Peer]
PublicKey = %s
AllowedIPs = 0.0.0.0/1, 128.0.0.0/1
PersistentKeepalive = 30
*/
func (vpnlink *VpnlinkVO) GenWgContent() string {
	template := `
[Interface]
PrivateKey = %s
ListenPort = %d
PostUp = ip li set %s netns %s; ip -n %s addr add %s dev %s; ip -n %s li set up dev %s

[Peer]
PublicKey = %s
AllowedIPs = 0.0.0.0/1, 128.0.0.0/1
PersistentKeepalive = 30
`
	ns := vpnlink.NetnsName()
	intf := vpnlink.IntfName()
	result := fmt.Sprintf(template,
		vpnlink.Server.PrivKey,
		vpnlink.ListenPort(),
		intf, ns, ns, vpnlink.SelfIntfAddr(), intf, ns, intf,
		vpnlink.Client.PubKey())
	return result
}

// func (vpnlink *VpnlinkVO) GenCpeWg() CpeWgVO {
// 	peer := CpeWgPeerVO{
// 		Id:       vpnlink.Id,
// 		Host:     vpnlink.Server.ListenAddr,
// 		Port:     vpnlink.Id + 10000,
// 		IntfAddr: vpnlink.Server.IntfAddr,
// 	}
// 	if len(vpnlink.Server.PrivKey) > 0 {
// 		peer.PubKey = vpnlink.Server.PubKey()
// 	}
// 	wgEndpoint := CpeWgVO{
// 		Id:       vpnlink.Id,
// 		Vni:      vpnlink.Vni,
// 		State:    vpnlink.State,
// 		PrivKey:  vpnlink.Client.PrivKey,
// 		IntfAddr: vpnlink.Client.IntfAddr,
// 		Server:   peer,
// 	}
// 	return wgEndpoint
// }

type DirlinkEndpointVO struct {
	Esn      string `json:"esn"`                //ESN
	Role     string `json:"role,omitempty"`     //角色，不用填写，自动计算
	IntfName string `json:"intfName,omitempty"` //专线接入到哪个网卡
	VlanId   int    `json:"vlanId,omitempty"`   //VLAN ID
	IntfAddr string `json:"intfAddr,omitempty"` //专线网卡的IP地址
}

func (dle *DirlinkEndpointVO) WithVlanIntf() bool {
	return dle.VlanId > 1
}

//For CPE/HUB
func (dle *DirlinkEndpointVO) InterfaceName() string {
	if dle.WithVlanIntf() {
		return fmt.Sprintf("%s_%d", dle.IntfName, dle.VlanId)
	} else {
		return dle.IntfName
	}
}

//For CPE/HUB uci config
func (dle *DirlinkEndpointVO) DevName() string {
	return fmt.Sprintf("dev%s_%d", dle.IntfName, dle.VlanId)
}
func (dle *DirlinkEndpointVO) IntfPureAddr() string {
	return PureIp(dle.IntfAddr)
}

type DirlinkVO struct {
	Id       int               `json:"id"`            //ID
	Vni      int               `json:"vni,omitempty"` //VNI，用于标识网卡名称
	State    string            `json:"state"`         //是否主链路 PRIMARY SECONDARY
	Server   DirlinkEndpointVO `json:"server"`        //服务端信息（VPE、HUB）
	Client   DirlinkEndpointVO `json:"client"`        //客户端信息（CPE-VPE、CPE-HUB，HUB-VPE）
	devState string            //设备的状态，MASTER/BACKUP
}

func (dl *DirlinkVO) GetEndpoint() DirlinkEndpointVO {
	if dl.IsServer() {
		return dl.Server
	} else {
		return dl.Client
	}
}
func (dl *DirlinkVO) BirdNeighFileName() string {
	return fmt.Sprintf("/etc/bird-dir%d.conf", dl.Id)
}

//For VPE
func (dl *DirlinkVO) VlanifName() string {
	return fmt.Sprintf("dir%d_%d", dl.Vni, dl.Server.VlanId)
}
func (dl *DirlinkVO) Path() string {
	return fmt.Sprintf("%s/%s", VpeDirlinkConfDir, dl.VlanifName())
}
func (dl *DirlinkVO) NetnsName() string {
	return fmt.Sprintf("ns%d", dl.Vni)
}
func (dl *DirlinkVO) IsServer() bool {
	return GBC.GetESN() == dl.Server.Esn
}

func (dl *DirlinkVO) SetDevState(state string) {
	dl.devState = state
}
func (dl *DirlinkVO) DevState() string {
	return dl.devState
}

func (dl *DirlinkVO) CommonLinkVO() LinkVO {
	return LinkVO{Vni: dl.Vni}
}

func (dl *DirlinkVO) PeerRole() string {
	if GBC.GetESN() == dl.Server.Esn {
		return dl.Client.Role
	} else {
		return dl.Server.Role
	}
}

func (dl *DirlinkVO) SelfIntfPureAddr() string {
	if GBC.GetESN() == dl.Server.Esn {
		return dl.Server.IntfPureAddr()
	} else {
		return dl.Client.IntfPureAddr()
	}
}

func (dl *DirlinkVO) PeerIntfPureAddr() string {
	if GBC.GetESN() == dl.Server.Esn {
		return dl.Client.IntfPureAddr()
	} else {
		return dl.Server.IntfPureAddr()
	}
}

func (dl *DirlinkVO) NeighASN() string {
	return fmt.Sprintf("%s_ASN", dl.PeerRole())
}

func (dl *DirlinkVO) ImportFilter() string {
	return fmt.Sprintf("%s_%s_i", strings.ToLower(dl.devState), strings.ToLower(dl.State))
}
func (dl *DirlinkVO) ExportFilter() string {
	return fmt.Sprintf("%s_%s_o", strings.ToLower(dl.devState), strings.ToLower(dl.State))
}

//PEER内容
func (dl *DirlinkVO) GenBirdNeighContent() string {
	temp := `
protocol bgp %s from %s {
	source address %s;
	neighbor %s as %s;
	ipv4 {
		next hop self;
		import filter %s;
		export filter %s;
	};
}
`
	tempName := fmt.Sprintf("%s_TEMP", dl.PeerRole())
	dle := dl.GetEndpoint()
	result := fmt.Sprintf(temp,
		dle.InterfaceName(), tempName,
		dl.SelfIntfPureAddr(),
		dl.PeerIntfPureAddr(),
		dl.NeighASN(),
		dl.ImportFilter(),
		dl.ExportFilter(),
	)
	return result
}

type HubDirlinkVO struct {
	Id       int               `json:"id"`            //ID
	Vni      int               `json:"vni,omitempty"` //VNI，用于标识网卡名称
	PeerA    DirlinkEndpointVO `json:"peerA"`         //服务端信息（VPE、HUB）
	PeerB    DirlinkEndpointVO `json:"peerB"`         //客户端信息（CPE-VPE、CPE-HUB，HUB-VPE）
	devState string            //设备的状态，MASTER/BACKUP
}

type VpnlinkStateVO struct {
	Id    int    `json:"id"`    //ID，用于生成网卡名称，wg0 0就是id，可以使用链路表的自增长ID
	State string `json:"state"` //是否主链路 PRIMARY SECONDARY
}

func (vls *VpnlinkStateVO) IntfName() string {
	return fmt.Sprintf("wg%d", vls.Id)
}
func (vls *VpnlinkStateVO) Filter(devState string) string {
	return fmt.Sprintf("%s_%s", strings.ToLower(devState), strings.ToLower(vls.State))
}

type ResetVpnlinkVO struct {
	Plink    VpnlinkStateVO `json:"plink"` //原来的primary link的信息
	Slink    VpnlinkStateVO `json:"slink"` //原来的secondary link的信息
	devState string         //设备的状态，MASTER/BACKUP
}

func (rvl *ResetVpnlinkVO) SetDevState(state string) {
	rvl.devState = state
}

func (rvl *ResetVpnlinkVO) DevState() string {
	return rvl.devState
}
