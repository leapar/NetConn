package network

import (
	"fmt"
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/peer"
	"github.com/davyxu/cellnet/proc"
	"net"
	"log"
)


const peerAddress = "0.0.0.0:32101"

var proxyMap map[ cellnet.Session]net.Conn = make(map[ cellnet.Session]net.Conn,0)
func connectServers(session cellnet.Session) net.Conn {
	conn, err := net.Dial("tcp", "127.0.0.1:22101")
	if err != nil {
		log.Println(err)
		return nil
	}


	go func(conn net.Conn,session cellnet.Session) {
		data := make([]byte, 4096)
		for {

			n, err := conn.Read(data)
			if err != nil {
				fmt.Printf("read message from lotus failed")
				if session != nil {
					session.Close()
				}
				conn.Close()
				conn = nil


				return
			}
			session.Send(data[:n])
			fmt.Printf("lobby:%d\n",n)
			fmt.Println(data[:n])
		}

	}(conn,session)

	return conn
}

func getSocket(session cellnet.Session) net.Conn  {
	_,has := proxyMap[session]
	if !has {
		 proxyMap[session] = connectServers(session)
	}

	return proxyMap[session]

}



func Server() {

	queue := cellnet.NewEventQueue()

	peerIns := peer.NewGenericPeer("tcp.Acceptor", "server", peerAddress, queue)

	proc.BindProcessorHandler(peerIns, "jj.ltv", func(ev cellnet.Event) {

		switch msg := ev.Message().(type) {
		case *cellnet.SessionAccepted: // 接受一个连接

			getSocket(ev.Session())

			fmt.Println("server accepted")
		case *JJProtoCodec:
			conn := getSocket(  ev.Session())
			fmt.Printf("server recv %+v\n", msg)
			n,err := conn.Write(msg.RawData)
			if err != nil {
				if  ev.Session() != nil{
					if conn != nil {
						conn.Close()
						conn = nil
					}

					ev.Session().Close()
				}

			}
			fmt.Println(n,err)

		case *cellnet.SessionClosed: // 连接断开
			fmt.Println("session closed: ", ev.Session().ID())

			conn := getSocket(ev.Session())
			if conn != nil {
				conn.Close()
			}

		}

	})

	peerIns.Start()

	queue.StartLoop()
}
