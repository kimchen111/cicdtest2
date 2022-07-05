package monitor

import (
	"encoding/json"
	"log"
	"sdwan/common"
	"time"

	"github.com/go-ping/ping"
)

type DetectResult struct {
	Cpe  string     `json:"cpe"`
	Vpes []IcmpPing `json:"vpes"`
}

type IcmpPing struct {
	ch     chan IcmpPing
	Esn    string `json:"esn"`
	Ipaddr string `json:"ipaddr"`
	AvgRtt int    `json:"avgRtt"`
	MaxRtt int    `json:"maxRtt"`
	MinRtt int    `json:"minRtt"`
	Loss   int    `json:"loss"`
}

func (ip IcmpPing) Run() {
	pinger, _ := ping.NewPinger(ip.Ipaddr)
	pinger.Count = 100
	pinger.Interval = time.Millisecond * 300
	err := pinger.Run() // Blocks until finished.
	if err != nil {
		log.Printf("ping %s failed, err: %s", ip.Ipaddr, err)
	}
	stats := pinger.Statistics() // get send/receive/duplicate/rtt stats
	ip.AvgRtt = int(stats.AvgRtt / time.Millisecond)
	ip.MaxRtt = int(stats.MaxRtt / time.Millisecond)
	ip.MinRtt = int(stats.MinRtt / time.Millisecond)
	ip.Loss = int(stats.PacketLoss * 100)
	ip.ch <- ip
}

func StartVpeDetect(vd common.VpeDetectVO) {
	ipch := make(chan IcmpPing)
	defer close(ipch)
	count := 0
	for _, vpe := range vd.Vpes {
		for _, ipaddr := range vpe.IpAddrs {
			ip := IcmpPing{
				ch:     ipch,
				Esn:    vpe.Esn,
				Ipaddr: ipaddr,
			}
			go ip.Run()
			count++
		}
	}

	rest := make([]IcmpPing, count)
	for i := 0; i < count; i++ {
		ip := <-ipch
		rest[i] = ip
	}
	dr := DetectResult{
		Cpe:  common.GBC.GetESN(),
		Vpes: rest,
	}
	data, _ := json.Marshal(dr)
	common.HttpPost(common.GCC.GetVpeDetectURL(), data, common.GCC.DisablePost())
}
