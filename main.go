package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"./mmap"
)

func main()  {
	mmap, err := mmap.MapRegion(nil,4096, mmap.RDWR, mmap.ANON,0,"{EBDA59A5-57B6-4d62-B0D0-EF9CC5E07F71}")
	if err != nil {
		fmt.Println(err)
	}
	defer mmap.Unmap()

	mmap[0x40] = 0x55
	mmap[0x41] = 0x55


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

