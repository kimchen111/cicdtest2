package monitor

import (
	"log"
	"regexp"
	"sdwan/common"
	"time"

	"github.com/vishvananda/netlink"
)

func MonitorHub() {
	log.Printf("Start UdpEchoServer for HUB.")
	quit := make(chan int)
	common.UdpEchoServer(common.CpeMonListenAddr, quit)
}

type TrafficMonitor struct {
	ch  chan Point
	tcs map[string]*TrafficCollector
}

func (tm *TrafficMonitor) TC(intf string) *TrafficCollector {
	if obj, ok := tm.tcs[intf]; ok {
		return obj
	}
	tc := new(TrafficCollector)
	tc.Intf = intf
	tc.LastTime = 0
	tc.LastTraffic = Traffic{}
	tm.tcs[intf] = tc
	return tc
}

func (tm *TrafficMonitor) Start(intfs []string) {
	for _, intf := range intfs {
		log.Printf("Start TrafficMonitor for %s", intf)
		tc := tm.TC(intf)
		tc.Collect()
		log.Printf("End MonitorTraffic for %s inRate: %d outRate: %d ", intf, tc.InRate, tc.OutRate)
		tm.ch <- tc.ToPoint()
	}
}

type LinkMonitor struct {
	name  string
	raddr string
	ch    chan Point
	// quit     chan int
}

func (mi *LinkMonitor) Start() {
	log.Printf("Start LinkMonitor for %s", mi.name)
	// err_count := 0
	// for {
	curr := time.Now().Unix()
	rch := make(chan Result, 10)
	defer close(rch)
	up := UdpPing{
		Interf:   mi.name,
		Raddr:    mi.raddr,
		Count:    100,
		Interval: 300,
		Chan:     rch,
	}
	go up.StartPing()
	r := <-rch
	// log.Printf("Sent %d, Recv %d, Rtt %d ms, Loss: %d", r.Sent, r.Recv, r.Rtt, r.Loss)
	if r.Err == "" {
		mi.ch <- r.ToPoint()
		// err_count = 0
	} else {
		// err_count++
		// s := err_count * 5
		r.Loss = 100
		mi.ch <- r.ToPoint()
		s := 30 - (time.Now().Unix() - curr)
		log.Printf("%s, retry after %d seconds", r.Err, s)
	}
	log.Printf("End MonitorLInk for %s Sent %d, Recv %d, Rtt %d ms, Loss: %d", mi.name, r.Sent, r.Recv, r.Rtt, r.Loss)
	// }
}

type Monitor struct {
	ch chan Point
	tm *TrafficMonitor
}

func (m *Monitor) Init() {
	m.ch = make(chan Point, 10)
	tm := TrafficMonitor{
		ch:  m.ch,
		tcs: make(map[string]*TrafficCollector),
	}
	m.tm = &tm
}

func (m *Monitor) WaitMonitorData() {
	log.Printf("Start Monitor.WaitMonitorData")
	cache := make([]Point, 5)
	// for {
	begin := time.Now().Unix()
	// 	idx := 0
	idx := 0
	for {
		point := <-m.ch
		cache[idx] = point
		idx++
		log.Printf("Recv point: %s count: %d", point.Measurement, idx)
		dura := time.Now().Unix() - begin
		if idx > 2 || dura > 32 {
			PushMonitorData(cache[:idx])
			idx = 0
		}
	}
	// }
}

/*
func (m *Monitor) AlreadyRun(name string) bool {
	_, ok := m.instances[name]
	return ok
}

func (m *Monitor) Stop(name string) {
	mi := m.instances[name]
	mi.quit <- 1
	close(mi.quit)
	m.Delete(name)
	log.Printf("Netns %s removed", name)
}

func (m *Monitor) CheckRemove(peers map[string]string) []string {
	arr := make([]string, len(m.instances)+len(peers))
	count := 0
	for name := range m.instances {
		if _, ok := peers[name]; ok {
			break
		} else {
			arr[count] = name
			count++
		}
	}
	return arr[:count]
}
*/

func (m *Monitor) GetTrafficMonitorIntfs() []string {
	links, err := netlink.LinkList()
	if err != nil {
		log.Printf("failed get link list %s", err)
		return nil
	}
	arr := make([]string, len(links))
	reg, _ := regexp.Compile("wg.+")
	idx := 0
	for _, link := range links {
		name := link.Attrs().Name
		if reg.MatchString(name) && !IsWgServer(name) {
			// log.Printf("netlink name: %s, is Wg Server: %v", name, IsWgServer(name))
			arr[idx] = name
			idx++
		}
		if name == "eth0" || name == "eth1" {
			state := link.Attrs().OperState.String()
			if state == "up" {
				arr[idx] = name
				idx++
			}
		}
	}
	log.Printf("Intfs: %s", arr[:idx])
	return arr[:idx]
}

func GuessHubPeerAddrs() []string {
	return nil
}

func (m *Monitor) GetVpnlinkPeerAddrs() map[string]string {
	links, err := netlink.LinkList()
	if err != nil {
		log.Printf("failed get link list %s", err)
		return nil
	}
	reg, _ := regexp.Compile("wg.+")
	r := make(map[string]string)
	for _, link := range links {
		name := link.Attrs().Name
		// log.Printf("Link name: %s", name)
		if reg.MatchString(name) && !IsWgServer(name) {
			addrs, err1 := netlink.AddrList(link, netlink.FAMILY_V4)
			if err1 == nil && len(addrs) == 1 {
				cidr := addrs[0].IPNet.String()
				log.Printf("Link name: %s ipaddr: %s", name, cidr)
				r[name] = common.FirstIp(cidr)
			}
		}
	}
	return r
}

func (m *Monitor) MonitorLink() {
	peers := m.GetVpnlinkPeerAddrs()
	for name, addr := range peers {
		lm := LinkMonitor{
			name:  name,
			raddr: addr,
			ch:    m.ch,
		}
		go lm.Start()
	}
}

func (m *Monitor) MonitorTraffic() {
	intfs := m.GetTrafficMonitorIntfs()
	go m.tm.Start(intfs)
}

func (m *Monitor) StartMonitor() {
	monitor.Init()
	go m.WaitMonitorData()
	for {
		m.MonitorLink()
		m.MonitorTraffic()
		tchan := time.After(time.Second * 30)
		<-tchan
	}
}

var monitor *Monitor

func StartCpeMonitor() {
	if common.GCC.DisableMonitor() {
		return
	}
	s := time.After(time.Second * 5)
	<-s
	role := common.GBC.GetAgentType()
	if role == "HUB" {
		go MonitorHub()
	}
	monitor = new(Monitor)
	monitor.StartMonitor()
}
