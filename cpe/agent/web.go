package agent

import (
	"sdwan/common"

	"github.com/gin-gonic/gin"
)

func RunWeb() {
	r := gin.Default()
	//本地管理
	RouteWeb(r)
	r.Run(common.GCC.GetLocalListen())
}

func RouteWeb(r *gin.Engine) {

}
