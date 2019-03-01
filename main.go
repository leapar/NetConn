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
)



func main()  {
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

