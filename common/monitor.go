package common

import (
	"log"
	"net"
	"time"
)

func UdpEchoServer(listenAddr string, quit chan int) {
	addr, _ := net.ResolveUDPAddr("udp", listenAddr)
	// addr := net.UDPAddr{IP: net.ParseIP(CpeMonListenAddr), Port: CpeMonListenPort}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Printf("Failed Listen: %s error: %s.", listenAddr, err)
		return
	}
	defer conn.Close()
	log.Printf("Listening on %s", listenAddr)
	buff := make([]byte, 2048)
	for {
		select {
		case <-quit:
			log.Printf("Quit UdpEchoServer.")
			return
		default:
			// log.Println("To read, blocked 1 seconds.")
			conn.SetReadDeadline(time.Now().Add(time.Second))
			count, remote, rderr := conn.ReadFrom(buff)
			if rderr == nil {
				// log.Printf("Recv data %d from %s.", count, remote.String())
				_, wrerr := conn.WriteTo(buff[0:count], remote)
				if wrerr != nil {
					log.Printf("Write error: %s.", wrerr)
				}
			}
		}
	}
}
