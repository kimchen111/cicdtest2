package monitor

import (
	"encoding/json"
	"fmt"
	"log"
	"sdwan/common"
	"strconv"
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

func (ip IcmpPing) ToPoint(id int, measurement string) Point {
	ptag := make(map[string]string)
	ptag["id"] = strconv.Itoa(id)
	ptag["Ipaddr"] = ip.Ipaddr
	fields := make(map[string]int)
	fields["rtt"] = ip.AvgRtt
	fields["loss"] = ip.Loss
	fields["jitter"] = ip.MaxRtt - ip.MinRtt
	return Point{
		Measurement: measurement,
		Time:        fmt.Sprintf("%d", time.Now().UnixMicro()),
		Tags:        ptag,
		Fields:      fields,
	}
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

/*
func StartCustomMonitor(cm common.CustomMonitorVO) {
	ipch := make(chan IcmpPing)
	defer close(ipch)
	for _, item := range cm.Monitors {
		ip := IcmpPing{
			ch:     ipch,
			Ipaddr: item.DstAddr,
		}
		go ip.Run()
	}
}

func StopCustomMonitor(cm common.CustomMonitorVO) {

}
*/
