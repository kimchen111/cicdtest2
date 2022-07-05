package agent

import (
	"log"
	"sdwan/common"
	"sdwan/ctrl/device"
	"time"
)

type Task struct {
	req     common.Message
	rc      chan common.Message
	timeout int
}

type Agent struct {
	a2zc chan common.Message
	z2ac chan common.Message

	//ResultQueue 保存任务和对应的channel
	rq map[string]Task
}

func (agent *Agent) notice(req common.Message) {
	agent.a2zc <- req
}

func (agent *Agent) request(req common.Message, timeout int) chan common.Message {
	agent.a2zc <- req
	rc := make(chan common.Message)
	agent.rq[req.MsgID] = Task{
		req:     req,
		rc:      rc,
		timeout: timeout,
	}
	return rc
}

func (agent *Agent) RunSeparate() {
	for {
		msg := <-agent.z2ac
		//log.Printf("Got a response: %v", msg)
		task, ok := agent.rq[msg.MsgID]
		if ok {
			task.rc <- msg
			// log.Printf("finish task %s", msg.MsgID)
			close(task.rc)
			delete(agent.rq, msg.MsgID)
		}
	}
}

func (agent *Agent) RunClean() {
	ticker := time.NewTicker(time.Millisecond * 10)
	for {
		<-ticker.C
		nowTime := time.Now().Unix()
		for id := range agent.rq {
			// log.Printf("Checking task %s", id)
			item, ok := agent.rq[id]
			if ok {
				spendTime := int(nowTime - item.req.MsgTime)
				// log.Printf("spend time %d", spendTime)
				if spendTime > item.timeout {
					close(item.rc)
					delete(agent.rq, id)
				}
			}
		}
	}
}

var GlobalAgent *Agent

func RunSeparator(a2zc chan common.Message, z2ac chan common.Message) {
	GlobalAgent = new(Agent)
	GlobalAgent.rq = make(map[string]Task)
	GlobalAgent.a2zc = a2zc
	GlobalAgent.z2ac = z2ac
	go GlobalAgent.RunSeparate()
	GlobalAgent.RunClean()
}

//timeout in second
func RequestWithTimeout(task common.Message, timeout int) common.ApiResult {
	// log.Printf("task: %v", task)
	if task.MsgType == common.MsgType.Request {
		if dev := device.GetGDM().GetDevice(task.Receiver); dev.Alive() {
			if dev.IsVPE() && !common.VpeTaskTypeSet.MatchTaskType(task.TaskType) {
				log.Printf("Device %s is VPE, TaskType is %s", task.Receiver, task.TaskType)
				return common.ApiResult{Esn: task.Receiver, Status: "error", Body: "device dismatch"}
			}
			if dev.IsCPE() && !common.CpeTaskTypeSet.MatchTaskType(task.TaskType) {
				log.Printf("Device %s is CPE, TaskType is %s", task.Receiver, task.TaskType)
				return common.ApiResult{Esn: task.Receiver, Status: "error", Body: "device dismatch"}
			}
			rc := GlobalAgent.request(task, timeout)
			result, ok := <-rc
			if ok {
				// log.Printf("rc->result: %v", result)
				return common.ApiResult{Esn: result.Sender, Status: "ok", Body: result.Body}
			} else {
				return common.ApiResult{Esn: result.Sender, Status: "error", Body: "timed out"}
			}
		} else {
			log.Printf("Dead target: %s", task.Receiver)
			return common.ApiResult{Esn: task.Receiver, Status: "error", Body: "not exists or dead"}
		}
	} else if task.MsgType == common.MsgType.Notice {
		GlobalAgent.notice(task)
		return common.ApiResult{Esn: task.Receiver, Status: "ok", Body: "no response"}
	}
	return common.ApiResult{Esn: task.Receiver, Status: "error", Body: "unknown error"}
}

func Request(task common.Message) common.ApiResult {
	return RequestWithTimeout(task, common.CtrlAPIDefaultTimeout)
}
