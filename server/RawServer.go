package server

import (
	"fmt"
	"net"
	"log"
	"../network"
)

var proxyMap map[ net.Conn]net.Conn = make(map[ net.Conn]net.Conn,0)
func connectServers(lobbyConn net.Conn) net.Conn {
	conn, err := net.Dial("tcp", "127.0.0.1:22101")
	if err != nil {
		log.Println(err)
		return nil
	}
	tcp := conn.(*net.TCPConn)
	jj := network.JJParser{
		Flag:fmt.Sprintf("%s==>%s\t",tcp.LocalAddr().String(),tcp.RemoteAddr().String()),
	}
	go func(conn net.Conn,lobbyConn net.Conn) {
		data := make([]byte, 4096)
		for {

			n, err := conn.Read(data)
			if err != nil {
				fmt.Printf("read message from lotus failed")
				if lobbyConn != nil {
					lobbyConn.Close()
				}
				conn.Close()
				conn = nil


				return
			}
			lobbyConn.Write(data[:n])
			//fmt.Printf("facebook:%d\n",n)


			//jj := network.JJProtoCodec{}
			//jj.Decode(data[:n])
			jj.ParseJJ(data[:n])

			//fmt.Println(data[:n])
		}

	}(conn,lobbyConn)

	return conn
}

func getSocket(conn net.Conn) net.Conn  {
	_,has := proxyMap[conn]
	if !has {
		proxyMap[conn] = connectServers(conn)
	}

	return proxyMap[conn]

}

func Echo(c net.Conn) {
	data := make([]byte, 4096)
	defer c.Close()

	tcp := c.(*net.TCPConn)
	jj := network.JJParser{
		Flag:fmt.Sprintf("%s==>%s\t",tcp.LocalAddr().String(),tcp.RemoteAddr().String()),
	}
	jjConn := getSocket(c)
	for {
		n, err := c.Read(data)
		if err != nil {
			fmt.Printf("read message from lotus failed")
			return
		}
		//fmt.Printf("==>: %d\n",n)
		//fmt.Println(time.Unix(int64(binary.LittleEndian.Uint32(data)),0).Format("2006-01-02 15:04:05"),time.Now().Format("2006-01-02 15:04:05"))
		if jjConn != nil {
			jjConn.Write(data[:n])
		}

		jj.ParseJJ(data[:n])
	}
}

func RawServer()  {
	fmt.Printf("Server is ready...\n")
	l, err := net.Listen("tcp", ":32101")
	if err != nil {
		fmt.Printf("Failure to listen: %s\n", err.Error())
	}

	for {
		if c, err := l.Accept(); err == nil {

			go Echo(c) //new thread
		}
	}
}