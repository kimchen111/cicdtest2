package monitor

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"sdwan/common"
	"time"
)

type Result struct {
	Err    string
	Interf string
	Sent   int
	Recv   int
	Rtt    int
	Jitter int
	Loss   int
}

func (r *Result) ToPoint() Point {
	ptag := make(map[string]string)
	ptag["if_name"] = r.Interf
	ptag["link_id"] = r.Interf[2:]
	fields := make(map[string]int)
	fields["rtt"] = r.Rtt
	fields["loss"] = r.Loss
	fields["jitter"] = r.Jitter
	return Point{
		Measurement: "cpe_losrtt",
		Time:        fmt.Sprintf("%d", time.Now().UnixMicro()),
		Tags:        ptag,
		Fields:      fields,
	}
}

type UdpPing struct {
	Interf   string
	Raddr    string
	Count    uint32
	Interval int
	Chan     chan Result
}

/*
raddr: 远端地址
count: 一组共连续发送多少个（次）数据包
interval: 两个包之间间隔多少毫秒
ch: 返回结果的管道
*/
func (up *UdpPing) StartPing() {
	udpaddr := net.UDPAddr{IP: net.ParseIP(up.Raddr), Port: common.CpeMonitorPort}
	conn, err := net.DialUDP("udp", nil, &udpaddr)
	if err != nil {
		log.Printf("DailUDP failed: %s", err)
	}
	// log.Printf("DailUDP success to send packets.")
	defer conn.Close()
	buff := common.GenRandBuff()
	seq := uint32(0)
	rerr := ""
	sent := 0
	recv := 0
	rttSum := int64(0)
	rttMax := int64(0)
	rttMin := int64(0)
	for {
		if seq >= up.Count {
			break
		} else {
			seq++
		}
		sendTime := time.Now().UnixMicro()
		binary.LittleEndian.PutUint32(buff[:4], uint32(seq))
		_, wrerr := conn.Write(buff)
		if wrerr != nil {
			rerr = fmt.Sprintf("Write error: %s", wrerr)
			break
		} else {
			sent++
		}
		conn.SetReadDeadline(time.Now().Add(time.Millisecond * 300))
		recvCount, rderr := conn.Read(buff)
		if rderr != nil {
			rerr = fmt.Sprintf("Read error: %s", rderr)
			break
		}
		if recvCount < 64 {
			rerr = "Recieved data error"
			break
		}
		recv++
		recvSeq := binary.LittleEndian.Uint32(buff[:4])
		if recvSeq != seq {
			rerr = fmt.Sprintf("Bad seq or Timetout seq: %d", seq)
			break
		} else {
			rtt := time.Now().UnixMicro() - sendTime
			rttSum += rtt
			if rtt > rttMax {
				rttMax = rtt
			} else if rtt < rttMin {
				rttMin = rtt
			}
			// log.Printf("Recv: %d SEQ: %d RTT: %d us", recvCount, recvSeq, rtt)
		}
		tchan := time.After(time.Millisecond * time.Duration(up.Interval))
		<-tchan
	}
	if recv == 0 || sent == 0 {
		r := Result{
			Err:    rerr,
			Interf: up.Interf,
			Sent:   0,
			Recv:   0,
			Rtt:    0,
			Loss:   0,
		}
		up.Chan <- r
	} else {
		r := Result{
			Err:    rerr,
			Interf: up.Interf,
			Sent:   sent,
			Recv:   recv,
			Rtt:    int(rttSum / int64(recv) / 1000),
			Jitter: int(rttMax) - int(rttMin),
			Loss:   (sent - recv) / sent,
		}
		up.Chan <- r
	}
	// log.Printf("Finish UdpPing for %s", up.Interf)
}
