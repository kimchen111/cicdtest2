package common

import (
	"log"
	"strings"
)

//CTRL
const (
	Controller            = "Controller"
	CtrlAPIDefaultTimeout = 5 //默认等待返回时间
)

//CPE
const (
	// CpeAgentType         = "CPE"
	CpeAgentVersion      = "1.0.1"
	CpeDialerRetry       = 10
	CpeDialerTimeout     = 2
	CpeHeartBeatDuration = 30
	CpeEsnDev            = "eth0"
	CpeMonListenAddr     = ":10789"
	CpeMonitorPort       = 10789
	CpeQosDir            = "/sdwan/qos"
	CpeKeepalivedDir     = "/sdwan/keepalived"
	CpeVxlanConfDir      = "/sdwan/vxlan"
	CpeBirdResetPath     = "/sdwan/resetbirdstate"
	CpeBirdConfPath      = "/etc/bird.conf"
	CpeBirdStaticPath    = "/etc/bird-static.onf"
	CpeHubConfPath       = "/etc/hub.conf"
	CpeTaIdPath          = "/etc/taid"
)

//VPE
const (
	VpeAgentType         = "VPE"
	VpeAgentVersion      = "1.0.1"
	VpeDialerRetry       = 10
	VpeDialerTimeout     = 2
	VpeHeartBeatDuration = 30
	VpeMonListenAddr     = ":10789"
	VpeEsnDev            = "eth0"
	VpeBgpConfDir        = "/sdwan/bgp"
	VpeVxlanConfDir      = "/sdwan/vxlan"
	VpeDirlinkConfDir    = "/sdwan/dirlink"
)

type BaseConfig struct {
	configFile string
	esn        string
	routerID   string
	agentClass string
}

func (bd *BaseConfig) GetESN() string {
	return bd.esn
}

func (bd *BaseConfig) GetRouterID() string {
	return bd.routerID
}

func (bd *BaseConfig) GetAgentType() string {
	if bd.agentClass == "CPE" {
		if FileExists(CpeHubConfPath) {
			return "HUB"
		} else {
			return "CPE"
		}
	} else {
		return bd.agentClass
	}
}

type CpeConfig struct {
	baseConfig       *BaseConfig
	ctrlUrl          string
	monitorURL       string
	vpeDetectURL     string
	netflowCollector string
	localListen      string
	disablePost      string
}

func (cc *CpeConfig) Init() {
	cc.baseConfig.configFile = "/etc/cpe.json"
	if !FileExists(cc.baseConfig.configFile) {
		log.Fatalf("config %s not exists", cc.baseConfig.configFile)
	}
	config := make(map[string]string)
	LoadConfig(cc.baseConfig.configFile, &config)
	cc.ctrlUrl = config["ctrlUrl"]
	cc.monitorURL = config["monitorUrl"]
	cc.vpeDetectURL = config["vpeDetectUrl"]
	cc.netflowCollector = config["netflowCollector"]
	cc.localListen = config["localListen"]
	cc.disablePost = config["disablePost"]
	mac := GetIntfMacaddr(CpeEsnDev)
	cc.baseConfig.esn = strings.ReplaceAll(mac, ":", "")
	cc.baseConfig.routerID = Esn2RID(cc.baseConfig.esn)
	cc.baseConfig.agentClass = "CPE"
	log.Printf("ESN: %s, RouterID: %s, CtrlUrl: %s, Agent: %s", cc.baseConfig.esn, cc.baseConfig.routerID, cc.ctrlUrl, cc.baseConfig.agentClass)
}

func (cc *CpeConfig) GetCtrlURL() string {
	return cc.ctrlUrl
}

func (cc *CpeConfig) GetMonitorURL() string {
	return cc.monitorURL
}

func (cc *CpeConfig) GetVpeDetectURL() string {
	return cc.vpeDetectURL
}

func (cc *CpeConfig) GetNetflowCollector() string {
	return cc.netflowCollector
}

func (cc *CpeConfig) GetLocalListen() string {
	return cc.localListen
}
func (cc *CpeConfig) DisablePost() bool {
	return cc.disablePost == "true"
}

type VpeConfig struct {
	baseConfig *BaseConfig
	ctrlUrl    string
}

func (vc *VpeConfig) Init() {
	vc.baseConfig.configFile = "/etc/vpe.json"
	if !FileExists(vc.baseConfig.configFile) {
		log.Fatalf("config %s not exists", vc.baseConfig.configFile)
	}
	config := make(map[string]string)
	LoadConfig(vc.baseConfig.configFile, &config)
	vc.ctrlUrl = config["ctrlUrl"]

	mac := GetIntfMacaddr(VpeEsnDev)
	vc.baseConfig.esn = strings.ReplaceAll(mac, ":", "")
	vc.baseConfig.routerID = GetLoopBackAddr()
	vc.baseConfig.agentClass = "VPE"
	log.Printf("ESN: %s, RouterID: %s, ctrlUrl: %s, Agent: %s", vc.baseConfig.esn, vc.baseConfig.routerID, vc.ctrlUrl, vc.baseConfig.agentClass)
}

func (vc *VpeConfig) GetCtrlURL() string {
	return vc.ctrlUrl
}

type CtrlConfig struct {
	baseConfig     *BaseConfig
	zmqListen      string
	apiListen      string
	stateNotifyUrl string
	disableMonitor string
	disableNotify  string
}

func (cc *CtrlConfig) Init() {
	cc.baseConfig.configFile = "/etc/ctrl.json"
	if !FileExists(cc.baseConfig.configFile) {
		log.Fatalf("config %s not exists", cc.baseConfig.configFile)
	}
	config := make(map[string]string)
	LoadConfig(cc.baseConfig.configFile, &config)
	cc.zmqListen = config["zmqListen"]
	cc.apiListen = config["apiListen"]
	cc.stateNotifyUrl = config["stateNotifyUrl"]
	cc.disableMonitor = config["disableMonitor"]
	cc.disableNotify = config["disableNotify"]
	log.Printf("zmqListen: %s, apiListen: %s, Agent: %s", cc.zmqListen, cc.apiListen, cc.baseConfig.agentClass)
}

func (cc *CtrlConfig) GetZmqListen() string {
	return cc.zmqListen
}

func (cc *CtrlConfig) GetApiListen() string {
	return cc.apiListen
}

func (cc *CtrlConfig) GetStateNotifyURL() string {
	return cc.stateNotifyUrl
}

func (cc *CtrlConfig) DisablePosMonitor() bool {
	return cc.disableMonitor == "true"
}

func (cc *CtrlConfig) DisableNotify() bool {
	return cc.disableNotify == "true"
}

var GBC *BaseConfig
var GCC *CpeConfig
var GVC *VpeConfig
var GCTC *CtrlConfig

func InitConfig(agentClass string) {
	GBC = new(BaseConfig)
	if agentClass == "CPE" {
		GCC = new(CpeConfig)
		GCC.baseConfig = GBC
		GCC.Init()
	} else if agentClass == "VPE" {
		GVC = new(VpeConfig)
		GVC.baseConfig = GBC
		GVC.Init()
	} else if agentClass == "CTRL" {
		GCTC = new(CtrlConfig)
		GCTC.baseConfig = GBC
		GCTC.Init()
	}
}

func GetControllerURL() string {
	if GBC.agentClass == "CPE" {
		return GCC.ctrlUrl
	} else if GBC.agentClass == "VPE" {
		return GVC.ctrlUrl
	}
	return ""
}
