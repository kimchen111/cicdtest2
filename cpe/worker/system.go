package worker

import (
	"fmt"
	"log"
	"os/exec"

	"sdwan/common"
)

type System struct {
}

func (system System) Name() string {
	return "System"
}

func (system System) Reboot(msg common.Message) {
	command := exec.Command("reboot")
	err := command.Run()
	result := ""
	if err != nil {
		// log.Printf("cmd.run() failed with %s", err)
		result = fmt.Sprintf("failed: %s", err)
	} else {
		result = "success"
	}
	Response(msg.ToResult(result))
} //重启

func (system System) Enable() {

} //启用

func (system System) Disable() {

} //禁用

func (system System) Init(msg common.Message) {
	iv := common.InitVO{}
	if err := common.LoadBody(msg.Body, &iv); err == nil {
		go ResetFirstBoot()
		Response(msg.ToResult("success"))
		return
	}
	Response(msg.ToResult("failed: unknown error"))
} //初始化

func (system System) Upgrade() {

} //升级AGENT程序

func (system System) Restart() {

} //重启AGENT程序

func (system System) EnableVRRP(msg common.Message) {
	vrrp := common.VrrpVO{}
	if err := common.LoadBody(msg.Body, &vrrp); err == nil {
		if ok := ProcessAddrs(vrrp); !ok {
			log.Printf("Process ip address failed.")
			Response(msg.ToResult("failed: address"))
			return
		}
		PrepareKeepalived()
		if vrrp.PreState == "MASTER" {
			HoldMaster()
		} else {
			ReleaseMaster()
			//确保没有hold信号文件，虽然也没有什么用处
		}
		if ok := EnableKeepalived(vrrp); !ok {
			log.Printf("Enable keepalived failed.")
			Response(msg.ToResult("failed: keepalived"))
			return
		}
		Response(msg.ToResult("success"))
		return
	}
	Response(msg.ToResult("failed: unknown error"))
} //启用HA（VRRP）

func (system System) DisableVRRP(msg common.Message) {
	DisableKeepalived()
	Response(msg.ToResult("success"))
} //禁用HA（VRRP）

func (system System) SwitchVRRP(msg common.Message) {
	vrrp := common.VrrpStateVO{}
	if err := common.LoadBody(msg.Body, &vrrp); err == nil {
		if vrrp.State == "BACKUP" {
			if !CheckMasterSignal() {
				Response(msg.ToResult("failed: not master"))
				return
			}
			if ok := ReleaseMaster(); ok {
				Response(msg.ToResult("success"))
				return
			}
		} else {
			if vrrp.Permanent {
				prestate := GetPreState()
				log.Printf("Permanent change state: %s to %s, ", prestate, vrrp.State)
				if GetPreState() == "BACKUP" {
					PermanentMaster()
				}
			}
			if ok := HoldMaster(); ok {
				Response(msg.ToResult("success"))
				return
			}
		}
	}
	Response(msg.ToResult("failed: unknown error"))
} //切换HA（VRRP）

func (system System) EnableBFD() {

} //启用BFD

func (system System) DisableBFD() {

} //禁用BFD

func (system System) AddFwRule() {

} //添加防火墙规则

func (system System) DelFwRule() {

} //删除防火墙规则

func (system System) SetupQos(msg common.Message) {
	bq := common.BaseQosVO{}
	if err := common.LoadBody(msg.Body, &bq); err == nil {
		if !bq.Validate() {
			Response(msg.ToResult("failed: error zones"))
			return
		}
		if ok := SetupQos(bq); ok {
			Response(msg.ToResult("success"))
			return
		}
	}
	Response(msg.ToResult("failed: unknown error"))
} //设置QOS

func (system System) AddQosRule(msg common.Message) {
	qos := common.QosVO{}
	if err := common.LoadBody(msg.Body, &qos); err == nil {
		if !qos.Validate() {
			Response(msg.ToResult("failed: error zones"))
			return
		}
		if ok := AddQos(qos); ok {
			Response(msg.ToResult("success"))
			return
		}
	}
	Response(msg.ToResult("failed: unknown error"))
} //添加QOS规则

func (system System) DelQosRule(msg common.Message) {
	qos := common.QosVO{}
	if err := common.LoadBody(msg.Body, &qos); err == nil {
		if !qos.Validate() {
			Response(msg.ToResult("failed: error zones"))
			return
		}
		if ok := DelQos(qos); ok {
			Response(msg.ToResult("success"))
			return
		}
	}
	Response(msg.ToResult("failed: unknown error"))
} //删除QOS规则

func (system System) SysInfo(msg common.Message) {
	// log.Println("SysInfo Invoked")
	vo := common.SysInfoVO{
		Version: common.CpeAgentVersion,
		Esn:     common.GBC.GetESN(),
	}
	result := msg.ToResult(vo)
	Response(result)
}
