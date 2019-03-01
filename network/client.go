package network

import (
	"fmt"
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/peer"
	"github.com/davyxu/cellnet/proc"
)

func ClientAsyncCallback() {
	queue := cellnet.NewEventQueue()
	p := peer.NewGenericPeer("tcp.Connector", "clientAsyncCallback", "127.0.0.1:21845", queue)
	proc.BindProcessorHandler(p, "jj.ltv", func(ev cellnet.Event) {
		switch msg := ev.Message().(type) {
		case *cellnet.SessionConnected: // 已经连接上
			fmt.Println("clientAsyncCallback connected")
		case *JJProtoCodec: //收到服务器发送的消息
			fmt.Printf("clientAsyncCallback recv %+v\n", msg)
		case *cellnet.SessionClosed:
			fmt.Println("clientAsyncCallback closed")
		}
	})
	p.Start()
	queue.StartLoop()
}
