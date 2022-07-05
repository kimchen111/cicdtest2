package common

import (
	"fmt"
	"strings"
)

type Neigh struct {
	VpeAddr  string `json:"vpeAddr"`  //地址
	VtepAddr string `json:"vtepAddr"` //VTEP地址
	Action   string `json:"action"`   //动作，删除还是添加
}

func (nei *Neigh) VtepPureAddr() string {
	arr := strings.Split(nei.VtepAddr, "/")
	return arr[0]
}

type EndpointVO struct {
	Vni        int      `json:"vni"`               //虚拟网络ID
	RouterID   string   `json:"routerID"`          //本节点的IP地址
	VtepAddr   string   `json:"VtepAddr"`          //本节点的终结网卡（bridge）IP地址，用于和其它节点互联转发数据
	Action     string   `json:"action"`            //删除还是添加
	Reflectors []string `json:"reflectors"`        //VNET中RR的Vtep地址
	Neighs     []Neigh  `json:"neighs"`            //邻居
	CpeCidr    string   `json:"cpeCidr,omitempty"` //连接当前VPE站点的CPE使用互联地址的网段
	HubCidr    string   `json:"hubCidr,omitempty"` //连接当前VPE站点的HUB使用互联地址的网段
}

func (ep *EndpointVO) VrfName() string {
	return fmt.Sprintf("vrf%d", ep.Vni)
}
func (ep *EndpointVO) VxlanName() string {
	return fmt.Sprintf("vxlan%d", ep.Vni)
}
func (ep *EndpointVO) VxlanPort() int {
	return 10000 + ep.Vni
}
func (ep *EndpointVO) VrfTableId() int {
	return 10000 + ep.Vni
}
func (ep *EndpointVO) BridgeName() string {
	return fmt.Sprintf("br%d", ep.Vni)
}
func (ep *EndpointVO) NetnsName() string {
	return fmt.Sprintf("ns%d", ep.Vni)
}
func (ep *EndpointVO) VtepPureAddr() string {
	return PureIp(ep.VtepAddr)
}
func (ep *EndpointVO) VtepNetwork() string {
	return Network(ep.VtepAddr)
}
func (ep *EndpointVO) VethvName() string {
	return fmt.Sprintf("ev%d", ep.Vni)
}
func (ep *EndpointVO) VethnName() string {
	return fmt.Sprintf("en%d", ep.Vni)
}

/*
func (ep *EndpointVO) BgpRD() string {
	return fmt.Sprintf("%d:%d", GlobalBgpAS, ep.Vni)
}
func (ep *EndpointVO) BgpRT() string {
	return fmt.Sprintf("%d:%d", ep.AsNO(), ep.Vni)
}
*/
func (ep *EndpointVO) ConfPath() string {
	return fmt.Sprintf("%s/%s.conf", VpeBgpConfDir, ep.NetnsName())
}
func (ep *EndpointVO) RelfectorConfPath() string {
	return fmt.Sprintf("%s/%s-rrs.conf", VpeBgpConfDir, ep.NetnsName())
}
func (ep *EndpointVO) GenRRBgpContent() string {
	temp := `
log syslog all;

define VPE_ASN = %d;
define VPE_CIDR = %s;
define LOCAL_ADDR = %s;
router id LOCAL_ADDR;
protocol device {
    scan time 10;
}
protocol direct {
	interface "*";
	ipv4 {
		import all;
	};
}
filter ignore_local
{
	if proto = "direct1" then {
		reject;
  	}
	accept;
}
protocol bgp VPES { 
	debug all;
    description "VPES";
    local LOCAL_ADDR as VPE_ASN;
    multihop 8;
    rr client;
    connect retry time 10;
    hold time 30;
    allow bgp_local_pref on;
    ipv4 {
        import all;
        export filter ignore_local;
    };
    neighbor range VPE_CIDR internal;
}
`
	result := fmt.Sprintf(temp,
		VpeAsNO(ep.Vni),
		ep.VtepNetwork(),
		ep.VtepPureAddr())
	return result
}

func (ep *EndpointVO) GenBgpContentRrs() string {
	temp := `
protocol bgp RR%d {
	debug all;
	description "RR%d-igp";
	local as VPE_ASN;
	multihop;
	connect retry time 10;
	hold time 30;
	source address LOCAL_VPE_ADDR;
	neighbor %s internal;
	ipv4 {
		next hop self;
		import all;
		export filter ignore_local;
	};
}
`
	content := ""
	for i, rr := range ep.Reflectors {
		content += fmt.Sprintf(temp, i, i, PureIp(rr))
	}
	return content
}

func (ep *EndpointVO) GenBgpContent() string {
	temp := `
log syslog all;

define VPE_ASN = %d;
define HUB_ASN = %d;
define CPE_ASN = %d;
define LOCAL_VPE_ADDR = %s;
define HUB_CIDR = %s;
define CPE_CIDR = %s;

router id LOCAL_VPE_ADDR;

protocol device {
	scan time 10;
}
protocol direct {
    interface "en*", "wg*";
    ipv4 {
        import all;
    };
}

filter ignore_local
{
	if proto = "direct1" then {
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

include "%s";

function repl_cpe_asn() {
	if bgp_path.first = CPE_ASN then {
		bgp_path.delete(CPE_ASN);
		bgp_path.prepend(VPE_ASN);
	}
}
filter ignore_local_and_repl_asn {
	if proto = "direct1" then {
		reject;
  	}
	repl_cpe_asn();
	accept;
}
protocol bgp CPES { 
	debug all;
	description "CPES";
	local as VPE_ASN;
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
protocol bgp HUBS { 
	debug all;
	description "HUBS";
	local as VPE_ASN;
	multihop 8;
	connect retry time 10;
	hold time 30;
    allow bgp_local_pref on;
	ipv4 {
		next hop self;
		import all;
		export filter ignore_local_and_repl_asn;
	};
	neighbor range HUB_CIDR as HUB_ASN;
}
`
	result := fmt.Sprintf(temp,
		VpeAsNO(ep.Vni),
		HubAsNO(ep.Vni),
		CpeAsNO(ep.Vni),
		ep.VtepPureAddr(),
		ep.HubCidr,
		ep.CpeCidr,
		ep.RelfectorConfPath())
	return result
}

//VXLAN Tunnel End Point
type VtepVO struct {
	Esn      string `json:"esn"`      //ESN
	VtepAddr string `json:"vtepAddr"` //本节点的终结网卡（bridge）IP地址，用于和其它节点互联转发数据
	Action   string `json:"action"`   //删除还是添加此节点对于本VNI的服务
	RouterID string `json:"routerID"` //ROUTER ID，API不用填写，自动从AGENT上报
}

func (vtep *VtepVO) SetRouterID(routerID string) {
	vtep.RouterID = routerID
}

type VnetVO struct {
	Vni        int       `json:"vni"`        //虚拟网络ID
	CpeCidr    string    `json:"cpeCidr"`    //CPE所用的CIDR
	HubCidr    string    `json:"hubCidr"`    //HUB所用的CIDR
	Reflectors []*VtepVO `json:"reflectors"` //运行本VNET的RR的信息
	Vteps      []*VtepVO `json:"vteps"`      //包含的VTEP信息
}

func (vnet *VnetVO) Neighs() []Neigh {
	neighs := make([]Neigh, len(vnet.Vteps))
	for i, vtepx := range vnet.Vteps {
		n := Neigh{
			VpeAddr:  vtepx.RouterID,
			VtepAddr: vtepx.VtepAddr,
			Action:   vtepx.Action,
		}
		neighs[i] = n
	}
	return neighs
}

func (vnet *VnetVO) GenRR(idx int) EndpointVO {
	rr := vnet.Reflectors[idx]
	endpoint := EndpointVO{
		Vni:      vnet.Vni,
		RouterID: rr.RouterID,
		VtepAddr: rr.VtepAddr,
		Action:   rr.Action,
		Neighs:   vnet.Neighs(),
	}
	return endpoint
}

func (vnet *VnetVO) GenRrs() []string {
	arr := make([]string, len(vnet.Reflectors))
	for i, rr := range vnet.Reflectors {
		arr[i] = rr.VtepAddr
	}
	return arr
}

func (vnet *VnetVO) ExcludeNeighs(idx int) []Neigh {
	neighs := make([]Neigh, len(vnet.Vteps)+len(vnet.Reflectors))
	i := 0
	for j, vtepx := range vnet.Vteps {
		if j != idx {
			neighs[i] = Neigh{
				VpeAddr:  vtepx.RouterID,
				VtepAddr: vtepx.VtepAddr,
				Action:   vtepx.Action,
			}
			i++
		}
	}

	for _, vtepx := range vnet.Reflectors {
		neighs[i] = Neigh{
			VpeAddr:  vtepx.RouterID,
			VtepAddr: vtepx.VtepAddr,
			Action:   vtepx.Action,
		}
		i++
	}
	return neighs[:i]
}

func (vnet *VnetVO) GenEndpoint(idx int) EndpointVO {
	vtep := vnet.Vteps[idx]
	endpoint := EndpointVO{
		Vni:        vnet.Vni,
		RouterID:   vtep.RouterID,
		VtepAddr:   vtep.VtepAddr,
		Action:     vtep.Action,
		Reflectors: vnet.GenRrs(),
		CpeCidr:    vnet.CpeCidr,
		HubCidr:    vnet.HubCidr,
		Neighs:     vnet.ExcludeNeighs(idx),
	}
	return endpoint
}

/*
type AllowedIpsVO struct {
	Vni     int      `json:"vni"`               //VNI
	PrivKey string   `json:"privKey"`           //添加到哪个Peer上
	Cidrs   []string `string:"cidrs,omitempty"` //BGP发布的网段列表
}

func (aips *AllowedIpsVO) IntfName() string {
	return fmt.Sprintf("wg%d", aips.Vni)
}
*/
