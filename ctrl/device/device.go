package device

import (
	"encoding/json"
	"log"
	"sdwan/common"
	"time"
)

type DeviceState struct {
	Esn       string `json:"esn"`
	Action    string `json:"action"`
	TimeStamp int64  `json:"timeStamp"`
}

type DeviceManager struct {
	dsChan chan DeviceState
	theMap map[string]*common.Device
}

func (dm *DeviceManager) CheckAlive() {
	ticker := time.NewTicker(time.Second * 30)
	for {
		<-ticker.C
		//log.Printf("CheckAlive...")
		for esn := range dm.theMap {
			item := dm.theMap[esn]
			if !item.Alive() {
				log.Printf("Device %s is dead, current time: %d, heartbeat time: %d",
					esn, time.Now().Unix(), item.LiveTime)
				dm.dsChan <- DeviceState{
					Esn:       esn,
					Action:    "dead",
					TimeStamp: time.Now().Unix(),
				}
			}
			/*else {
				log.Printf("Device %s is alive, heartbeat time: %d.", esn, item.liveTime)
			}*/
		}
	}
}

//检查指定ESN是否注册并存活
func (dm *DeviceManager) IsAlive(esn string) bool {
	dev, ok := dm.theMap[esn]
	if ok {
		return dev.Alive()
	}
	return false
}

//检查指定ESN是否注册并存活
func (dm *DeviceManager) GetDevice(esn string) common.Device {
	dev, ok := dm.theMap[esn]
	if ok {
		return *dev
	}
	return common.Device{}
}

//注册/更新指定ESN的心跳时间为当前
func (dm *DeviceManager) HeartBeat(msg common.Message) {
	esn := msg.Sender
	dev, ok := dm.theMap[esn]
	if !ok {
		dev = new(common.Device)
		dev.Esn = esn
		common.LoadBody(msg.Body, &dev.SysInfo)
		dm.theMap[esn] = dev
		dm.dsChan <- DeviceState{
			Esn:       esn,
			Action:    "register",
			TimeStamp: time.Now().Unix(),
		}
		log.Printf("Register new device %s", esn)
	} else {
		common.LoadBody(msg.Body, &dev.SysInfo)
		// if !dev.Alive() {
		dm.dsChan <- DeviceState{
			Esn:       esn,
			Action:    "alive",
			TimeStamp: time.Now().Unix(),
		}
		// }
	}
	dev.HeartBeat()
}

//角色改变
func (dm *DeviceManager) RoleChange(msg common.Message) {
	esn := msg.Sender
	dev := new(common.Device)
	dev.Esn = esn
	common.LoadBody(msg.Body, &dev.SysInfo)
	dm.theMap[esn] = dev
	log.Printf("Replace new device %s", esn)
	dev.HeartBeat()
}

func (dm *DeviceManager) StateNotify() {
	var buff [10]DeviceState
	idx := 0
	var firstTime int64
	for {
		if idx == 10 || (idx > 0 && time.Now().Unix()-firstTime >= 30) {
			data, _ := json.Marshal(buff[:idx])
			go common.HttpPost(common.GCTC.GetStateNotifyURL(), data, common.GCTC.DisableNotify())
			idx = 0
		} else {
			buff[idx] = <-dm.dsChan
			if idx == 0 {
				firstTime = time.Now().Unix()
			}
			idx++
		}
	}
}

var gdm *DeviceManager

func InitGDM() {
	theMap := make(map[string]*common.Device)
	gdm = new(DeviceManager)
	gdm.theMap = theMap
	gdm.dsChan = make(chan DeviceState, 10)
}

func GetGDM() *DeviceManager {
	return gdm
}

func RunObserver() {
	InitGDM()
	go gdm.CheckAlive()
	gdm.StateNotify()
}
