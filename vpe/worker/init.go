package worker

import (
	"os/exec"
	"sdwan/common"
)

func InitConfig() {
	InitVxlanConfDir()
	InitDirlinkConfDir()
}

func InitVxlanConfDir() {
	if !common.FileExists(common.VpeVxlanConfDir) {
		exec.Command("mkdir", common.VpeVxlanConfDir).Run()
	}
}

func InitDirlinkConfDir() {
	if !common.FileExists(common.VpeDirlinkConfDir) {
		exec.Command("mkdir", common.VpeDirlinkConfDir).Run()
	}
}
