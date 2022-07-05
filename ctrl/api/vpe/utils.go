package vpe

import (
	"sdwan/common"
	"sdwan/ctrl/device"
)

func UpdateRouterIP(vnet *common.VnetVO) {
	for _, rr := range vnet.Reflectors {
		dev := device.GetGDM().GetDevice(rr.Esn)
		rr.SetRouterID(dev.SysInfo.RouterID)
	}
	for _, vtep := range vnet.Vteps {
		dev := device.GetGDM().GetDevice(vtep.Esn)
		vtep.SetRouterID(dev.SysInfo.RouterID)
		// log.Printf("RouterID: %s", dev.SysInfo.RouterID)
	}
}
