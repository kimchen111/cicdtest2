package monitor

import (
	"os"
	"sdwan/common"
)

func SaveCustomMonitor(cm common.CustomMonitorVO) {
	for _, mi := range cm.Monitors {
		ct := CustomTask{
			Measurement: cm.Measurement,
			MonitorItem: mi,
		}
		common.WriteConfig(mi.Path(), ct)
	}
}

func RemoveCustomMonitor(cm common.CustomMonitorVO) {
	for _, mi := range cm.Monitors {
		if common.FileExists(mi.Path()) {
			os.Remove(mi.Path())
		}
	}
}
