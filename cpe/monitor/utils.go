package monitor

import "github.com/digineo/go-uci"

func IsWgServer(name string) bool {
	if res, ok := uci.Get("network", name, "listen_port"); ok {
		if len(res) > 0 {
			return true
		} else {
			return false
		}
	}
	return false
}
