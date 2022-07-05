package monitor

import (
	"log"
	"runtime"
	"sdwan/common"
	"time"

	"github.com/vishvananda/netns"
)

type MonitorServer struct {
	name string
	quit chan int
}

/*
name: Netns name
*/
func (ms *MonitorServer) ServNetns() {
	log.Printf("Start MonitorServer.ServNetns on %s", ms.name)
	// ms.Running[name] = true
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	origns, _ := netns.Get()
	defer origns.Close()
	newns, errx := netns.GetFromName(ms.name)
	if errx != nil {
		log.Printf("Failed get netns.")
		return
	}
	defer newns.Close()
	err := netns.Set(newns)
	if err != nil {
		log.Printf("Failed to switch netns.")
		return
	}
	common.UdpEchoServer(common.VpeMonListenAddr, ms.quit)
	netns.Set(origns)
	log.Printf("Quit from netns %s", ms.name)
}

type Monitor struct {
	servers map[string]MonitorServer
}

func (m *Monitor) AlreadyRun(name string) bool {
	_, ok := m.servers[name]
	return ok
}

func (m *Monitor) Delete(name string) {
	delete(m.servers, name)
}

func (m *Monitor) CheckRemove(names []string) []string {
	arr := make([]string, len(m.servers)+len(names))
	count := 0
	for name := range m.servers {
		remove := true
		for _, n := range names {
			if n == name {
				remove = false
				break
			}
		}
		if remove {
			arr[count] = name
			count++
		}
	}
	return arr[:count]
}

func (m *Monitor) Stop(name string) {
	ms := m.servers[name]
	ms.quit <- 1
	close(ms.quit)
	m.Delete(name)
}

func (m *Monitor) Start() {
	m.servers = make(map[string]MonitorServer)
	ticker := time.NewTicker(time.Second * 30)
	for {
		names := common.Netns()
		for _, name := range names {
			if name != "" {
				if m.AlreadyRun(name) {
					continue
				}
				ms := MonitorServer{
					name: name,
					quit: make(chan int),
				}
				m.servers[name] = ms
				go ms.ServNetns()
			}
		}
		removed := m.CheckRemove(names)
		for _, name := range removed {
			m.Stop(name)
		}
		<-ticker.C
	}
}

var monitor *Monitor

func StartVpeMonitorServer() {
	s := time.After(time.Second * 5)
	<-s
	monitor = new(Monitor)
	monitor.Start()
}
