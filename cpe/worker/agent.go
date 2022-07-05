package worker

import "sdwan/common"

type Agent struct {
	w2zc chan common.Message
}

func (agent *Agent) response(msg common.Message) {
	agent.w2zc <- msg
}

var GlobalAgent *Agent

func InitAgent(w2zc chan common.Message) {
	GlobalAgent = new(Agent)
	GlobalAgent.w2zc = w2zc
}

func Response(result common.Message) {
	GlobalAgent.response(result)
}
