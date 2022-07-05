package monitor

import (
	"fmt"
	"os/exec"
	"sdwan/common"
)

func StartTrafficAnalysis() {
	if common.FileExists(common.CpeTaIdPath) {
		temp := `#!/bin/sh
line=$(lsmod |grep pkt_netflow) 
if [ "" == "${line}" ]; then
  ver=$(uname -r)
  insmod /lib/modules/${ver}/pkt_netflow.ko protocol=10 engine_id=%d destination=%s
fi
`
		ta := common.TrafficAnalysisVO{}
		common.LoadConfig(common.CpeTaIdPath, &ta)
		destination := common.GCC.GetNetflowCollector()
		content := fmt.Sprintf(temp, ta.DeviceId, destination)
		common.WriteFile("/tmp/modprobe_netflow", content)
		exec.Command("sh", "/tmp/modprobe_netflow").Run()
	}
}

func StopTrafficAnalysis() {
	exec.Command("rmmod", "pkt_netflow").Run()
}
