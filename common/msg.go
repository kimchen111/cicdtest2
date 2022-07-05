package common

import (
	"time"

	"github.com/go-basic/uuid"
)

type Message struct {
	MsgID   string `json:"msgID"`   //消息ID
	MsgType string `json:"msgType"` //消息类型
	MsgTime int64  `json:"msgTime"` //消息发送时的时间戳

	Sender   string `json:"sender"`   //发送者
	Receiver string `json:"receiver"` //接收者

	TaskClass string `json:"taskClass"` //任务的分类
	TaskType  string `json:"taskType"`  //消息内含的任务的类型

	Body any `json:"body"` //消息的body，json格式
}

func (msg Message) IsEmpty() bool {
	return msg.MsgID == "" && msg.MsgType == "" && msg.TaskClass == "" && msg.TaskType == ""
}

func (msg Message) ToResult(v any) Message {
	result := Message{
		MsgID:   msg.MsgID,
		MsgType: MsgType.Response,
		MsgTime: time.Now().Unix(),

		Sender:   msg.Receiver,
		Receiver: msg.Sender,

		TaskClass: msg.TaskClass,
		TaskType:  msg.TaskType,

		Body: v,
	}
	return result
}

func newNoticeMsg(taskType string) Message {
	vo := SysInfoVO{
		Esn:       GBC.GetESN(),
		Version:   VpeAgentVersion,
		AgentType: GBC.GetAgentType(),
		RouterID:  GBC.GetRouterID()}

	msg := Message{
		MsgType: MsgType.Notice,
		MsgTime: time.Now().Unix(),

		Sender:   GBC.GetESN(),
		Receiver: Controller,

		TaskType: taskType,
		Body:     vo,
	}
	return msg
}

func NewHeartBeatMsg() Message {
	return newNoticeMsg(CommonTaskType.HeartBeat)
}

func NewRoleChangeMsg() Message {
	return newNoticeMsg(CommonTaskType.RoleChange)
}

func NewMessage(esn string, taskClass string, taskType string) Message {
	return Message{
		MsgID:   uuid.New(),
		MsgTime: time.Now().Unix(),

		Sender:   Controller,
		Receiver: esn,

		TaskClass: taskClass,
		TaskType:  taskType,
	}
}

func NewRequestTask(esn string, taskClass string, taskType string) Message {
	msg := NewMessage(esn, taskClass, taskType)
	msg.MsgType = MsgType.Request
	return msg
}

func NewRequestTaskWithBody(esn string, taskClass string, taskType string, body any) Message {
	msg := NewRequestTask(esn, taskClass, taskType)
	msg.Body = body
	return msg
}

func NewNoticeTask(esn string, taskClass string, taskType string) Message {
	msg := NewMessage(esn, taskClass, taskType)
	msg.MsgType = MsgType.Notice
	return msg
}

func NewNoticeTaskWithBody(esn string, taskClass string, taskType string, body any) Message {
	msg := NewNoticeTask(esn, taskClass, taskType)
	msg.Body = body
	return msg
}
