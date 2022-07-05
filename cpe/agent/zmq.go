package agent

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"sdwan/common"

	"github.com/go-zeromq/zmq4"
)

type ZeroMQ struct {
	id     int
	esn    string
	socket zmq4.Socket
	w2zc   chan common.Message
	z2wc   chan common.Message
	ctx    context.Context
	cancel context.CancelFunc
}

func (zmq *ZeroMQ) connect() {
	id := zmq4.SocketIdentity(zmq.esn)
	socket := zmq4.NewDealer(
		context.Background(),
		zmq4.WithID(id),
		zmq4.WithDialerRetry(time.Second*time.Duration(common.CpeDialerRetry)),
		zmq4.WithDialerTimeout(time.Second*time.Duration(common.CpeDialerTimeout)),
	)
	//defer socket.Close()
	err := socket.Dial(common.GetControllerURL())
	if err == nil {
		log.Printf("Connect to %s successfully.", common.GetControllerURL())
	}
	/*else {
		log.Printf("Failed connect to %s: %s", common.GetControllerURL(), err)
		return
	}*/
	zmq.socket = socket
}

func (zmq *ZeroMQ) close() {
	if zmq.socket != nil {
		zmq.socket.Close()
	}
}

func (zmq *ZeroMQ) recvMsg() (common.Message, error) {
	msg, err := zmq.socket.Recv()
	message := common.Message{}
	if err != nil {
		return message, err
	} else {
		content := msg.Frames[0]
		log.Printf("Recv msg: %s", string(content))
		json.Unmarshal(content, &message)
		return message, nil
	}
}

func (zmq *ZeroMQ) sendMsg(message common.Message) error {
	content, _ := json.Marshal(message)
	if message.TaskType != common.CommonTaskType.HeartBeat {
		log.Printf("Send msg: %s", string(content))
	}
	msg := zmq4.NewMsgFrom(content)
	err := zmq.socket.Send(msg)
	if err != nil {
		return err
	}
	return nil
}

//心跳
func (zmq *ZeroMQ) runHeartbeat() {
	log.Printf("Zmq-%d Heartbeat-Thread Running.", zmq.id)
	ticker := time.NewTicker(time.Second * common.CpeHeartBeatDuration)
	lastTimeRole := common.GBC.GetAgentType()
	for {
		select {
		case <-zmq.ctx.Done():
			log.Printf("Zmq-%d Heartbeat-Thread Stoped.", zmq.id)
			return
		default:
			role := common.GBC.GetAgentType()
			var err error
			if role == lastTimeRole {
				err = zmq.sendMsg(common.NewHeartBeatMsg())
			} else {
				lastTimeRole = role
				err = zmq.sendMsg(common.NewRoleChangeMsg())
			}
			if err != nil {
				log.Printf("Heartbeat error: %s, Stop Zmq-%d Heartbeat-Thread.", err, zmq.id)
				zmq.cancel()
				return
			}
			<-ticker.C
		}
	}
}

//接收消息
func (zmq *ZeroMQ) runRecv() {
	log.Printf("Zmq-%d Recv-Thread Running.", zmq.id)
	for {
		select {
		case <-zmq.ctx.Done():
			log.Printf("Zmq-%d Receive-Thread Stoped.", zmq.id)
			return
		default:
			msg, err := zmq.recvMsg()
			if err != nil {
				log.Printf("RecvMsg error: %s, Stop Zmq-%d Recv-Thread.", err, zmq.id)
				zmq.cancel()
				return
			} else if msg.IsEmpty() {
				log.Print("Ignore empty msg.")
			} else {
				zmq.z2wc <- msg
			}
		}
	}
}

//从worker接收反回并发送消息
func (zmq *ZeroMQ) runSend() {
	log.Printf("Zmq-%d Send-Thread Running.", zmq.id)
	for {
		msg := <-zmq.w2zc
		select {
		case <-zmq.ctx.Done():
			//当前协程的zmq的socket已经关闭，尝试把message还回去，让新的zmq的socket发送
			log.Printf("Zmq-%d Send-Thread Stoped.", zmq.id)
			zmq.w2zc <- msg
			return
		default:
			err := zmq.sendMsg(msg)
			if err != nil {
				log.Printf("SendMsg error: %s, Stop Zmq-%d Send-Thread.", err, zmq.id)
				log.Printf("Discard a message: %s", msg.MsgID)
				zmq.cancel()
				return
			}
		}
	}
}

func NewZeroMQ(id int,
	w2zc chan common.Message,
	z2wc chan common.Message) ZeroMQ {
	ctx, cancel := context.WithCancel(context.Background())
	zmq := ZeroMQ{
		id:     id,
		esn:    common.GBC.GetESN(),
		w2zc:   w2zc,
		z2wc:   z2wc,
		ctx:    ctx,
		cancel: cancel,
	}
	zmq.connect()
	return zmq
}

func RunZmq(w2zc chan common.Message, z2wc chan common.Message) {
	id := 0
	for {
		id++
		log.Printf("Start a new ZeroMQ with ESN: %s", common.GBC.GetESN())
		zmq := NewZeroMQ(id, w2zc, z2wc)
		go zmq.runHeartbeat()
		go zmq.runRecv()
		go zmq.runSend()
	NormalLoop:
		for {
			select {
			case <-zmq.ctx.Done():
				log.Println("Some error occurred, retrying...")
				zmq.close()
				break NormalLoop
			default:
				tchan := time.After(time.Second)
				<-tchan
			}
		}
	}
}
