package monitor

import (
	"encoding/json"
	"log"
	"sdwan/common"
)

type Point struct {
	Measurement string            `json:"measurement"` //相当于表
	Time        string            `json:"time"`        //时间
	Tags        map[string]string `json:"tags"`        //标签
	Fields      map[string]int    `json:"fields"`      //监拧
}

type MonitorData struct {
	Database string            `json:"database"` //数据库
	Tags     map[string]string `json:"tags"`     //标签
	Points   []Point           `json:"points"`   //监测点
}

func PushMonitorData(points []Point) {
	log.Printf("PushMonitorData to Center count: %d", len(points))
	mtag := make(map[string]string)
	mtag["esn_id"] = common.GBC.GetESN()
	md := MonitorData{
		Database: "nsdwan",
		Tags:     mtag,
		Points:   points,
	}
	data, _ := json.Marshal(md)
	common.HttpPost(common.GCC.GetMonitorURL(), data, common.GCC.DisablePost())
}
