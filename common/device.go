package common

import "time"

type Device struct {
	Esn      string    `json:"esn"`      //ESN
	LiveTime int64     `json:"liveTime"` //最后心跳时间
	SysInfo  SysInfoVO `json:"sysInfo"`  //SysInfo
}

func (dev *Device) HeartBeat() {
	dev.LiveTime = time.Now().Unix()

	//log.Printf("Device %s heartbeat at %d.", dev.esn, dev.liveTime)
}

func (dev *Device) IsCPE() bool {
	return dev.SysInfo.AgentType == "CPE"
}

func (dev *Device) IsVPE() bool {
	return dev.SysInfo.AgentType == "VPE"

}

func (dev *Device) Alive() bool {
	if (time.Now().Unix() - dev.LiveTime) < 60 {
		return true
	} else {
		return false
	}
}
