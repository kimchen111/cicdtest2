package worker

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"sdwan/common"
	"time"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

func EnsureVrf(ep common.EndpointVO) (netlink.Link, error) {
	vrf, err := netlink.LinkByName(ep.VrfName())
	if err != nil {
		la := netlink.NewLinkAttrs()
		la.Name = ep.VrfName()
		err2 := netlink.LinkAdd(&netlink.Vrf{
			LinkAttrs: la,
			Table:     uint32(ep.VrfTableId())})
		if err2 != nil {
			log.Printf("add Vrf failed %v ", err)
			return nil, err2
		}
		vrf, err = netlink.LinkByName(ep.VrfName())
		netlink.LinkSetUp(vrf)
		return vrf, err
	} else {
		return vrf, nil
	}
}

func ensureVxlan(ep common.EndpointVO) (netlink.Link, error) {
	vxlan, err := netlink.LinkByName(ep.VxlanName())
	if err != nil {
		la := netlink.NewLinkAttrs()
		la.Name = ep.VxlanName()
		err3 := netlink.LinkAdd(&netlink.Vxlan{
			LinkAttrs: la,
			VxlanId:   ep.Vni,
			Port:      ep.VxlanPort(),
			SrcAddr:   net.IP(ep.RouterID),
		})
		if err3 != nil {
			log.Printf("add vxlan failed %v ", err3)
			return nil, err3
		}
		vxlan, err = netlink.LinkByName(ep.VxlanName())
		netlink.LinkSetUp(vxlan)
		return vxlan, err
	} else {
		return vxlan, err
	}
}

/*
ip li add vrf10 type vrf table 10010
ip li set up dev vrf10

ip li add vxlan10 type vxlan id 10 dstport 4789
ip li set up dev vxlan10

ip li add br10 type bridge
ip li set up dev br10

ip li set master br10 dev vxlan10

ip li set vrf vrf10 dev br10
ip li set vrf vrf10 dev eth7

ip a add 10.3.0.1/24 dev eth7
ip a add 172.16.10.1/24 dev br10
*/
func CreateInterface(ep common.EndpointVO) bool {
	// vrf, err1 := ensureVrf(ep)
	br, err2 := common.EnsureBridge(ep.BridgeName())
	vxlan, err3 := ensureVxlan(ep)
	if err2 != nil || err3 != nil {
		log.Printf("Create interface failed %v %v", err2, err3)
		return false
	}
	err4 := netlink.LinkSetMaster(vxlan, br)
	// err5 := netlink.LinkSetMaster(br, vrf)
	if err4 != nil {
		log.Printf("Attach vxlan interface to bridge failed %v ", err4)
	}
	if err2 == nil && err3 == nil && err4 == nil {
		return true
	}
	return false
}

func SetLinkNetns(linkname string, nsname string) bool {
	link, err := netlink.LinkByName(linkname)
	if err != nil {
		log.Printf("get interface %s Failed.", linkname)
		return false
	}
	file, err4 := OpenNetns(nsname)
	if err4 != nil {
		log.Printf("open ns failed.")
		return false
	}
	defer file.Close()
	fd := int(file.Fd())
	err5 := netlink.LinkSetNsFd(link, fd)
	if err5 != nil {
		log.Printf("set ns failed.")
		return false
	}
	return true
}

//evxx在外面，进VRF，en在BGP的ns里面
func CreateVeth(ep common.EndpointVO) bool {
	namev := ep.VethvName()
	namen := ep.VethnName()
	_, err0 := netlink.LinkByName(namev)
	if err0 != nil {
		vethx := netlink.Veth{
			LinkAttrs: netlink.LinkAttrs{
				Name: namev,
			},
			PeerName: namen,
		}
		err1 := netlink.LinkAdd(&vethx)
		if err1 != nil {
			log.Printf("Add veth pair %s %s Failed.", namev, namen)
			return false
		}
	}
	linkv, err2 := netlink.LinkByName(namev)
	if err2 != nil {
		log.Printf("get vethv %s Failed.", namev)
		return false
	}
	// if ok := AddIpaddr(linkv, ep.VethvAddr()); !ok {
	// 	log.Printf("set vethv ipaddr failed %s, %s ", ep.VethvName(), ep.VethvAddr())
	// }
	br, _ := netlink.LinkByName(ep.BridgeName())
	netlink.LinkSetMaster(linkv, br)
	netlink.LinkSetUp(linkv)
	return true
}

func CreateNetns(id int) bool {
	name := fmt.Sprintf("ns%d", id)
	if exists := common.NetnsExists(name); exists {
		return true
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	origns, _ := netns.Get()
	defer origns.Close()
	ns, err := netns.GetFromName(name)
	failed := false
	if err != nil {
		ns, err = netns.NewNamed(name)
		if err != nil {
			log.Printf("Add netns Failed.")
			failed = true
		}
	}
	defer ns.Close()
	netns.Set(origns)
	return !failed
}

func OpenNetns(name string) (*os.File, error) {
	path := fmt.Sprintf("/var/run/netns/%s", name)
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	// defer f.Close()
	if err != nil {
		log.Printf("Open fd Failed.")
		return nil, err
	}
	return f, nil
}

func AddRoute(dst string, via string, tableID int) bool {
	ipnet, _ := netlink.ParseIPNet(dst)
	viaIP := net.ParseIP(via)
	route := netlink.Route{
		Dst:   ipnet,
		Gw:    viaIP,
		Table: tableID}
	err := netlink.RouteAdd(&route)
	if err != nil {
		log.Printf("Add route %s via %s to table %d failed", dst, via, tableID)
		return false
	}
	return true
}

func AddMainRoute(dst string, via string) bool {
	return AddRoute(dst, via, 254)
}

func SetDefautGW(gwaddr string) bool {
	gw := net.ParseIP(gwaddr)
	defaultRoute := &netlink.Route{
		Dst: nil,
		Gw:  gw,
	}
	err := netlink.RouteAdd(defaultRoute)
	if err != nil {
		log.Printf("Set default gateway %s failed", gwaddr)
		return false
	}
	return true
}

func CheckNetnsLink(nsname string, linkname string) bool {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	origns, _ := netns.Get()
	defer origns.Close()
	newns, errx := netns.GetFromName(nsname)
	found := false
	if errx == nil {
		netns.Set(newns)
		ifaces, _ := net.Interfaces()
		for _, iface := range ifaces {
			if iface.Name == linkname {
				found = true
				break
			}
		}
	}
	defer newns.Close()
	netns.Set(origns)
	return found
}

func CleanNetnsLink(nsname string) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	origns, _ := netns.Get()
	defer origns.Close()
	newns, errx := netns.GetFromName(nsname)
	if errx == nil {
		netns.Set(newns)
		ifaces, _ := net.Interfaces()
		for _, iface := range ifaces {
			if iface.Name != "lo" {
				link, _ := netlink.LinkByName(iface.Name)
				netlink.LinkSetDown(link)
				if err := netlink.LinkDel(link); err != nil {
					log.Printf("Link %s remove failed", iface.Name)
				}
			}
		}
	}
	defer newns.Close()
	netns.Set(origns)
}

func RemoveNetnsLink(nsname string, linkname string) bool {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	origns, _ := netns.Get()
	defer origns.Close()
	newns, errx := netns.GetFromName(nsname)
	failed := false
	if errx == nil {
		netns.Set(newns)
		link, err := netlink.LinkByName(linkname)
		if err == nil {
			netlink.LinkSetDown(link)
			if err1 := netlink.LinkDel(link); err1 != nil {
				log.Printf("Link %s remove failed", linkname)
				failed = true
			}
		}
		netns.Set(origns)
	}
	defer newns.Close()

	return !failed
}

func SetNetnsDefautGW(via string, nsname string) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	origns, _ := netns.Get()
	defer origns.Close()
	newns, _ := netns.GetFromName(nsname)
	defer newns.Close()
	netns.Set(newns)
	SetDefautGW(via)
	netns.Set(origns)
}

func SetupVethNetns(ep common.EndpointVO) bool {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	// Save the current network namespace
	origns, _ := netns.Get()
	defer origns.Close()

	newns, _ := netns.GetFromName(ep.NetnsName())
	defer newns.Close()
	netns.Set(newns)

	llo, _ := netlink.LinkByName("lo")
	netlink.LinkSetUp(llo)

	linkn, _ := netlink.LinkByName(ep.VethnName())
	netlink.LinkSetUp(linkn)

	common.AddIpaddr(linkn, ep.VtepAddr)

	netns.Set(origns)
	// Switch back to the original namespace
	return true
}

/*
	confFile := fmt.Sprintf("/sdwan/bgp/%s.conf", ep.NetnsName())
	ctrlFile := fmt.Sprintf("/run/%s.ctl", ep.NetnsName())
	pidFile := fmt.Sprintf("/var/run/bird/%s.pid", ep.NetnsName())
	if common.FileExists(pidFile) {
		if pid, err1 := common.ReadFile(pidFile); err1 == nil {
			exec.Command("kill", pid).Run()
		}
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	// Save the current network namespace
	origns, _ := netns.Get()
	defer origns.Close()
	newns, _ := netns.GetFromName(ep.NetnsName())
	defer newns.Close()
	netns.Set(newns)
	exec.Command("/usr/sbin/bird", "-c", confFile, "-s", ctrlFile, "-P", pidFile).Run()
	netns.Set(origns)
*/
func StartBird(nsname string) bool {
	s := time.After(time.Second * 5)
	<-s //5秒后再启动，否则有可能会IP不可用
	err := exec.Command("sh", "/sdwan/bird", "start", nsname).Run()
	if err != nil {
		log.Printf("Start bird in %s failed: %s", nsname, err)
	}
	return true
}

func StopBird(nsname string) {
	err := exec.Command("sh", "/sdwan/bird", "stop", nsname).Run()
	if err != nil {
		log.Printf("Stop bird at %s failed: %s", nsname, err)
	}
}

/*
ip := net.ParseIP(peer)
hwAddr := net.HardwareAddr{0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
neigh := netlink.Neigh{
	LinkIndex:    linkv.Attrs().Index,
	State:        netlink.NUD_PERMANENT,
	Family:       syscall.AF_BRIDGE,
	Flags:        netlink.NTF_SELF,
	IP:           ip,
	HardwareAddr: hwAddr,
}
netlink.NeighAppend(&neigh)
// bridge fdb append to 00:00:00:00:00:00 dev vxlan50  dst 10.10.10.14
*/
func UpdateFDB(ep common.EndpointVO) bool {
	// linkv, _ := netlink.LinkByName(ep.VxlanName())
	for _, peer := range ep.Neighs {
		if peer.Action == "ADD" {
			log.Printf("bridge fdb append to 00:00:00:00:00:00 dev %s dst %s", ep.VxlanName(), peer.VpeAddr)
			exec.Command("bridge", "fdb", "append", "to", "00:00:00:00:00:00", "dev", ep.VxlanName(), "dst", peer.VpeAddr).Run()
		} else {
			log.Printf("bridge fdb append to 00:00:00:00:00:00 dev %s dst %s", ep.VxlanName(), peer.VpeAddr)
			exec.Command("bridge", "fdb", "del", "to", "00:00:00:00:00:00", "dev", ep.VxlanName(), "dst", peer.VpeAddr).Run()
		}
	}
	return true
}

func DeleteInterface(ep common.EndpointVO) bool {
	var derr1, derr2, derr3 error
	br, err1 := netlink.LinkByName(ep.BridgeName())
	if err1 == nil {
		derr1 = netlink.LinkDel(br)
	}
	vxlan, err2 := netlink.LinkByName(ep.VxlanName())
	if err2 == nil {
		derr2 = netlink.LinkDel(vxlan)
	}
	vethv, err3 := netlink.LinkByName(ep.VethvName())
	if err3 == nil {
		derr3 = netlink.LinkDel(vethv)
	}
	CleanNetnsLink(ep.NetnsName())
	err4 := netns.DeleteNamed(ep.NetnsName())
	if derr1 != nil || derr2 != nil || derr3 != nil || err4 != nil {
		log.Printf("delete interface errors: err1:%v err2:%v err3:%v err4:%v", derr1, derr2, derr3, err4)
		return false
	}
	return true
}

func WgintfExists(ns string, name string) bool {
	// log.Printf("wg show %s", name)
	err := exec.Command("ip", "-n", ns, "li", "show", "dev", name).Run()
	return err == nil
}

func WgLoadConf(name string) bool {
	err := exec.Command("wg-quick", "up", name).Run()
	return err == nil
}

/*
func WgReloadConf(name string) bool {
	exec.Command("wg-quick", "down", name).Run()
	err := exec.Command("wg-quick", "up", name).Run()
	return err == nil
}
*/
/*
func WgSetupNetns(vl common.VpnlinkVO) bool {
	if ok := SetLinkNetns(vl.IntfName(), vl.NetnsName()); !ok {
		log.Printf("set vethn into netns Failed.")
		return false
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	// Save the current network namespace
	origns, _ := netns.Get()
	defer origns.Close()
	newns, _ := netns.GetFromName(vl.NetnsName())
	defer newns.Close()
	netns.Set(newns)

	wglink, _ := netlink.LinkByName(vl.IntfName())
	netlink.LinkSetUp(wglink)
	// br, err = netlink.LinkByName(ep.BridgeName())
	// if ok := AddIpaddr(br, ep.BridgeAddr); !ok {
	// 	log.Printf("set bridge ipaddr failed %s, %s ", ep.BridgeName(), ep.BridgeAddr)
	// }
	common.AddIpaddr(wglink, vl.Server.IntfAddr)
	netns.Set(origns)
	// Switch back to the original namespace
	return true
}
*/

func WgDelete(ns string, name string) bool {
	err := exec.Command("ip", "-n", ns, "li", "del", name).Run()
	return err == nil
}

/*
func WgDelAllowedIps(aips common.AllowedIpsVO) bool {
	name := aips.IntfName()
	pubkey := common.GenWgPubKey(aips.PrivKey)
	origCidrs := wgGetAllowedIps(name, pubkey)
	// log.Printf("orig: %s", origCidrs)

	set := make(map[string]string)
	for _, cidr := range aips.Cidrs {
		set[cidr] = cidr
	}
	// log.Printf("set: %s", set)

	result := make([]string, 1024)
	idx := 0
	for _, cidr := range origCidrs {
		if _, ok := set[cidr]; !ok {
			result[idx] = cidr
			idx++
		}
	}

	retained := result[:idx]
	// log.Printf("retained: %s", retained)

	arrstr := strings.Join(retained, ",")
	log.Printf("allowed-ips: %s", arrstr)
	err := exec.Command("wg", "set", name, "peer", pubkey, "allowed-ips", arrstr).Run()
	if err == nil {
		err := exec.Command("wg-quick", "save", name).Run()
		return err == nil
	}
	log.Printf("Failed set allowed-ips for %s", name)
	return false
}

func wgGetAllowedIps(name string, pubkey string) []string {
	command := exec.Command("wg", "show", name, "allowed-ips")
	out, err := command.CombinedOutput()
	if err == nil {
		content := string(out)
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			str := strings.Replace(line, " ", ",", -1)
			str = strings.Replace(str, "\t", ",", -1)
			// log.Printf("replaced: %s", str)
			arr := strings.Split(str, ",")
			if arr[0] == pubkey {
				return arr[1:]
			}
		}
	}
	r := [...]string{}
	return r[:]
}
*/

func SetupDirlinkIntf(dl common.DirlinkVO) (string, bool) {
	dle := dl.Server
	if ok := common.CreateVlanIntf(dle.IntfName, dle.VlanId, dl.VlanifName()); !ok {
		log.Printf("Create vlan intface %s failed.", dl.VlanifName())
		return "failed: create vlan interface", false
	}

	if ok := SetLinkNetns(dl.VlanifName(), dl.NetnsName()); !ok {
		log.Printf("set link %s into netns Failed.", dl.VlanifName())
		return "failed: set vlan interface into netns", false
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	// Save the current network namespace
	origns, _ := netns.Get()
	defer origns.Close()

	newns, _ := netns.GetFromName(dl.NetnsName())
	defer newns.Close()
	netns.Set(newns)

	linkn, _ := netlink.LinkByName(dl.VlanifName())
	netlink.LinkSetUp(linkn)
	common.AddIpaddr(linkn, dle.IntfAddr)
	netns.Set(origns)
	return "success", true
}

func LoadDirlinkIntf(name string) (string, bool) {
	dl := common.DirlinkVO{}
	path := dl.Path()
	common.LoadConfig(path, &dl)
	return SetupDirlinkIntf(dl)
}
