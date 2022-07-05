package monitor

import (
	"fmt"
	"strings"
	"time"

	"github.com/vishvananda/netlink"
)

type Traffic struct {
	RecvBytes   int64
	RecvPackets int64
	TranBytes   int64
	TranPackets int64
}

func (traf *Traffic) String() string {
	return fmt.Sprintf("RecvBytes: %d RecvPackets: %d TranBytes: %d TranPackets: %d", traf.RecvBytes, traf.RecvPackets, traf.TranBytes, traf.TranPackets)
}

type TrafficCollector struct {
	Intf        string
	LastTime    int64
	LastTraffic Traffic
	InRate      int64
	OutRate     int64
}

func (tc *TrafficCollector) Same(f string) bool {
	t := fmt.Sprintf("%s:", tc.Intf)
	return f == t
}

func (tc *TrafficCollector) Collect() {
	currTime := time.Now().Unix()
	lastTime := tc.LastTime
	duration := currTime - lastTime
	link, _ := netlink.LinkByName(tc.Intf)
	statis := link.Attrs().Statistics
	ct := Traffic{
		RecvBytes:   int64(statis.RxBytes),
		RecvPackets: int64(statis.RxPackets),
		TranBytes:   int64(statis.TxBytes),
		TranPackets: int64(statis.TxPackets),
	}
	lt := tc.LastTraffic
	// log.Printf("lastTime: %d lastRecv: %d lastTran: %d", lastTime, lt.RecvBytes, lt.TranBytes)
	tc.InRate = (ct.RecvBytes - lt.RecvBytes) * 8 / duration
	tc.OutRate = (ct.TranBytes - lt.TranBytes) * 8 / duration
	if tc.InRate < 0 || tc.OutRate < 0 {
		tc.InRate = 0
		tc.OutRate = 0
	}
	tc.LastTime = currTime
	tc.LastTraffic = ct
}

func (tc *TrafficCollector) ToPoint() Point {
	ptag := make(map[string]string)
	ptag["if_name"] = tc.Intf
	if strings.Contains(tc.Intf, "wg") {
		ptag["link_id"] = tc.Intf[2:]
	}
	fields := make(map[string]int)
	fields["in_rate"] = int(tc.InRate)
	fields["out_rate"] = int(tc.OutRate)
	return Point{
		Measurement: "cpe_traffic",
		Time:        fmt.Sprintf("%d", time.Now().UnixMicro()),
		Tags:        ptag,
		Fields:      fields,
	}
}
