package device

import (
	"context"
	"encoding/json"
	"log"

	"sdwan/common"

	"github.com/go-zeromq/zmq4"
)

type ZeroMQ struct {
	socket zmq4.Socket
	a2zc   chan common.Message
	z2ac   chan common.Message
}

func (zmq *ZeroMQ) listen() {
	zmq.socket = zmq4.NewRouter(
		context.Background(),
		zmq4.WithID(zmq4.SocketIdentity(common.Controller)),
	)
	err := zmq.socket.Listen(common.GCTC.GetZmqListen())
	if err != nil {
		log.Fatalf("could not listen %q: %s\n", common.GCTC.GetZmqListen(), err)
	} else {
		log.Printf("Listen at %s successfully.\n", common.GCTC.GetZmqListen())
	}
}

func (zmq *ZeroMQ) recvMsg() (common.Message, error) {
	msg, err := zmq.socket.Recv()
	if err == nil {
		id := string(msg.Frames[0])
		content := msg.Frames[1]
		message := common.Message{}
		err = json.Unmarshal(content, &message)
		if err != nil {
			log.Printf("unmarshal failed: %s", err)
		}
		if message.TaskType != common.CommonTaskType.HeartBeat {
			log.Printf("Recv msg: %s from %s", string(content), id)
		}
		if id != message.Sender {
			log.Printf("ID not equal sender of message: %v", err)
		}
		return message, nil
	} else {
		log.Printf("failed recv message: %s", err)
		return common.Message{}, err
	}
}

func (zmq *ZeroMQ) sendMsg(message common.Message) error {
	content, err := json.Marshal(message)
	if err == nil {
		log.Printf("Send to %s a msg: %s", message.Receiver, string(content))
		msg := zmq4.NewMsgFrom([]byte(message.Receiver), content)
		err = zmq.socket.Send(msg)
		if err == nil {
			return nil
		} else {
			log.Printf("failed send message: %s", err)
			return err
		}
	} else {
		log.Printf("marshal failed: %s", err)
		return nil
	}
}

func (zmq *ZeroMQ) runRecv() {
	log.Println("Receive thread Running.")
	for {
		msg, err := zmq.recvMsg()
		if err == nil {
			switch msg.MsgType {
			case common.MsgType.Notice:
				if msg.TaskType == common.CommonTaskType.HeartBeat {
					gdm.HeartBeat(msg)
				} else if msg.TaskType == common.CommonTaskType.RoleChange {
					gdm.RoleChange(msg)
				}

			case common.MsgType.Response:
				zmq.z2ac <- msg
			case common.MsgType.Request:
				//A Request
			}
		} else {
			log.Printf("RecvMsg Error: %s", err)
		}
	}
}

func (zmq *ZeroMQ) runSend() {
	for {
		task := <-zmq.a2zc
		esn := task.Receiver
		if gdm.IsAlive(esn) {
			err := zmq.sendMsg(task)
			if err != nil {
				log.Printf("Send message error: %s", err)
			}
		} else {
			log.Printf("Got task %s for %s, it's dead, discard task.", task.MsgID, esn)
		}
	}
}

func NewZeroMQ(a2zc chan common.Message, z2ac chan common.Message) ZeroMQ {
	zmq := ZeroMQ{
		a2zc: a2zc,
		z2ac: z2ac,
	}
	zmq.listen()
	return zmq
}

func RunZmq(a2zc chan common.Message, z2ac chan common.Message) {
	zmq := NewZeroMQ(a2zc, z2ac)
	go zmq.runRecv()
	zmq.runSend()
}
