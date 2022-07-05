package agent

import (
	"log"
	"reflect"
	"sdwan/vpe/worker"

	"sdwan/common"
)

type Worker interface {
	Name() string
}

func process(msg common.Message) {
	if msg.MsgType == common.MsgType.Request {
		var executor Worker
		switch msg.TaskClass {
		case common.VpeTaskClass.Vnet:
			executor = worker.Vnet{}
		case common.VpeTaskClass.Link:
			executor = worker.Link{}
		case common.VpeTaskClass.Route:
			executor = worker.Route{}
		}
		sv := reflect.ValueOf(executor)
		mv := sv.MethodByName(msg.TaskType)
		// params := []reflect.Value{}
		params := make([]reflect.Value, 1)
		//params[0] = reflect.ValueOf(msg.MsgType)
		params[0] = reflect.ValueOf(msg)
		log.Printf("Call %s.%s\n", sv.Type().Name(), msg.TaskType)
		mv.Call(params)
	}
}

func RunWorker(z2wc chan common.Message) {
	for {
		msg := <-z2wc
		process(msg)
	}
}
