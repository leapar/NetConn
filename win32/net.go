package win32

import (
	"strings"
	"strconv"
	"net"
	"syscall"
	"unsafe"
	"os"
	"fmt"
	"errors"
)

var (
	iphlpapi             = syscall.NewLazyDLL("iphlpapi.dll")
	procGetNetworkParams = iphlpapi.NewProc("GetNetworkParams")
)

const MAX_HOSTNAME_LEN = 128    // arb.
const MAX_DOMAIN_NAME_LEN = 128 // arb.
const MAX_SCOPE_ID_LEN = 256    // arb.

type IpAddapterParams struct {
	HostName         [MAX_HOSTNAME_LEN + 4]byte
	DomainName       [MAX_DOMAIN_NAME_LEN + 4]byte
	CurrentDnsServer *syscall.IpAddrString
	DnsServerList    syscall.IpAddrString
	NodeType         uint32
	ScopeId          [MAX_SCOPE_ID_LEN + 4]byte
	EnableRouting    uint32
	EnableProxy      uint32
	EnableDns        uint32
}

type NSServer []string

type NSServers struct {
	NSServer
	NicNS map[string]NSServer `json:"nicns" yaml:"nicns"` //ns set by nic
}

func htons(port uint16) uint16 {
	var (
		lowbyte  uint8  = uint8(port)
		highbyte uint8  = uint8(port << 8)
		ret      uint16 = uint16(lowbyte)<<8 + uint16(highbyte)
	)
	return ret
}

func inet_addr(ipaddr string) uint32 {
	var (
		segments []string = strings.Split(ipaddr, ".")
		ip       [4]uint64
		ret      uint64
	)
	for i := 0; i < 4; i++ {
		ip[i], _ = strconv.ParseUint(segments[i], 10, 64)
	}
	ret = ip[3]<<24 + ip[2]<<16 + ip[1]<<8 + ip[0]
	return uint32(ret)
}


func ipStringToI32(a string) uint32 {
	return ipToI32(net.ParseIP(a))
}
func ipToI32(ip net.IP) uint32 {
	ip = ip.To4()
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}

func i32ToIP(a uint32) net.IP {
	return net.IPv4(byte(a>>24), byte(a>>16), byte(a>>8), byte(a))
}



func getAdapterList() (*syscall.IpAdapterInfo, error) {
	b := make([]byte, 2048)
	l := uint32(len(b))
	a := (*syscall.IpAdapterInfo)(unsafe.Pointer(&b[0]))

	// TODO(mikio): GetAdaptersInfo returns IP_ADAPTER_INFO that
	// contains IPv4 address list only. We should use another API
	// for fetching IPv6 stuff from the kernel.

	err := syscall.GetAdaptersInfo(a, &l)
	if err == syscall.ERROR_BUFFER_OVERFLOW {
		b = make([]byte, l)
		a = (*syscall.IpAdapterInfo)(unsafe.Pointer(&b[0]))
		err = syscall.GetAdaptersInfo(a, &l)
	}
	if err != nil {
		return nil, os.NewSyscallError("GetAdaptersInfo", err)
	}
	return a, nil
}

//返回是否是公网IP
func IsPublicIP(IP net.IP) bool {
	if IP.IsLoopback() || IP.IsLinkLocalMulticast() || IP.IsLinkLocalUnicast() {
		return false
	}
	if ip4 := IP.To4(); ip4 != nil {
		switch true {
		case ip4[0] == 10:
			return false
		case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
			return false
		case ip4[0] == 192 && ip4[1] == 168:
			return false
		default:
			return true
		}
	}
	return false
}

func bytePtrToString(p *uint8) string {
	a := (*[10000]uint8)(unsafe.Pointer(p))
	i := 0
	for a[i] != 0 {
		i++
	}
	return string(a[:i])
}

func localMac() (net.HardwareAddr, error) {
	adapterList, err := getAdapterList()
	if err != nil {
		return nil, err
	}

	for adapter := adapterList; adapter != nil; adapter = adapter.Next {
		mac := make([]byte,8)
		copy(mac,adapter.Address[:adapter.AddressLength])
		macAddr := fmt.Sprintf("%02X-%02X-%02X-%02X-%02X-%02X-%02X-%02X",mac[0],mac[1],mac[2],mac[3],mac[4],mac[5],mac[6],mac[7])
		fmt.Println(macAddr)
		ipl := &adapter.IpAddressList
		for ipl != nil {
			ips := bytePtrToString(&ipl.IpAddress.String[0])
			fmt.Println(ips)
			ipl = ipl.Next
		}
		fmt.Println()

		/*
		printf("%x\n", j);  //输出结果为:    2f
		printf("%X\n", j);  //输出结果为:    2F
		printf("%#x\n", j); //输出结果为:    0x2f
		printf("%#X\n", j); //输出结果为:    0X2F    %#X推荐使用
		*/
		//if bytes.Contains([]byte(dev), adapter.AdapterName[:bytes.IndexRune(adapter.AdapterName[:], 0)]) {
		//	return adapter.Address[:adapter.AddressLength], nil
		//}
	}

	return nil, errors.New("Could not find adapter")
}



func getNetworkParams() (*IpAddapterParams, error) {
	b := make([]byte, 2048)
	l := uint32(len(b))
	a := (*IpAddapterParams)(unsafe.Pointer(&b[0]))
	localMac()
	// TODO(mikio): GetAdaptersInfo returns IP_ADAPTER_INFO that
	// contains IPv4 address list only. We should use another API
	// for fetching IPv6 stuff from the kernel.
	r0, _, _ := syscall.Syscall(procGetNetworkParams.Addr(), 2, uintptr(unsafe.Pointer(a)), uintptr(unsafe.Pointer(&l)), 0)
	if r0 != 0 {
		return nil, syscall.Errno(r0)
	}
	return a, nil
}

//https://github.com/chennqqi/goutils/blob/6de57397a59b91e7104ad1ac946f97b867cc8d40/net/net.go
func GetLocalNS() (NSServers, error) {
	var ns NSServers
	var rerr error

	//windows only support one couple dns
	iphelper, err := getNetworkParams()
	if err != nil {
		return ns, err
	}
	nsip := strings.Trim(fmt.Sprintf(`%s`, iphelper.DnsServerList.IpAddress.String), "\t ")
	ns.NSServer = append(ns.NSServer, nsip)
	pIPAddr := iphelper.DnsServerList.Next
	for pIPAddr != nil {
		nsip = strings.Trim(fmt.Sprintf(`%s`, pIPAddr.IpAddress.String), "\t ")
		ns.NSServer = append(ns.NSServer, nsip)
		pIPAddr = pIPAddr.Next
	}
	return ns, rerr
}