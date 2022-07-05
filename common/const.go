package common

//消息类型
type msgType struct {
	Request  string //请求（等待响应）
	Response string //响应
	Notice   string //通知（不等待响应）
}

var MsgType msgType

type commonTaskType struct {
	HeartBeat  string //心跳
	RoleChange string //角色变化
	Init       string //初始化信息
}

var CommonTaskType commonTaskType

type TaskTypeSet struct {
	mapx map[string]string
}

func (tts *TaskTypeSet) Init() {
	// log.Printf("Init ...... ")
	tts.mapx = make(map[string]string)
}

func (tts *TaskTypeSet) ConcatMap(mapy map[string]string) {
	for k, v := range mapy {
		if _, ok := tts.mapx[k]; !ok {
			tts.mapx[k] = v
		}
	}
}

func (tts *TaskTypeSet) MatchTaskType(taskType string) bool {
	// for k, v := range tts.mapx {
	// 	log.Printf(" k: %s, v: %s", k, v)
	// }
	_, ok := tts.mapx[taskType]
	// log.Printf("Match: %v", ok)
	return ok
}

//任务分类
type commonTaskClass struct {
	Link string //网络隧道，VPN连接
}

var CommonTaskClass commonTaskClass

//VPN，TUNNEL
type commonLinkTaskType struct {
	AddVpnEndpoint string //创建VPN端点
	DelVpnEndpoint string //删除VPN端点
	AddDirEndpoint string //创建直连链路端点
	DelDirEndpoint string //删除直连链路端点
}

var CommonLinkTaskType commonLinkTaskType
