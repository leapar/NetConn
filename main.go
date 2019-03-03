package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"./mmap"
	"./win32"
	_ "github.com/davyxu/cellnet/peer/tcp" // 注册TCP Peer
	_ "github.com/davyxu/cellnet/proc/tcp" // 注册TCP Processor
	"encoding/binary"
	"./server"
	"./network"
)



func main()  {

	bytes0 := []byte{0x5D,0x2B,0x0B,0xCA,0x01,0x00,0x00,0x00,0x4C,0x04,0x00,0x00,0x00,0x48,0x02,0x40,0x00,0x00,0x00,0x00,0x08,0x00,0x00,0x00,0x65,0xFB,0xCA,0x52,0x22,0x4E,0x00,0x00}

	jj := network.JJProtoCodec{}

	jj.Decode(bytes0)
	fmt.Println(jj)
	by,err := jj.Encode()
	fmt.Println(by)

	mmap, err := mmap.MapRegion(nil,4096, mmap.RDWR, mmap.ANON,0,"{EBDA59A5-57B6-4d62-B0D0-EF9CC5E07F71}")
	if err != nil {
		fmt.Println(err)
	}
	defer mmap.Unmap()

	ip := win32.Inet_addr("192.168.0.105")

	binary.LittleEndian.PutUint32(mmap,ip)
	//0C-54-15-D4-E6-CB-00-00
	copy(mmap[4:],[]byte("0C-54-15-D4-E6-CB-00-00"))

	mmap[0x40] = 0x65
	mmap[0x41] = 0x7d

	//network.Server()
	server.RawServer()
	//network.ClientAsyncCallback()
	//mmap.Flush()
	done := make(chan bool, 1)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()
	// The program will wait here until it gets the
	// expect
	<-done
}

