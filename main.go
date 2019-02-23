package main

import (
	"./windows"
	"fmt"
)
func main()  {
	a,b := windows.GetLocalNS()
	fmt.Println(a,b)
}
