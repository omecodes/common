package network

import (
	"github.com/zoenion/common/log"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"
)

/*
Taken functions from gists
https://gist.github.com/andres-erbsen/62d7defe8dce2e182bd9a90da2e1f659
*/

func localIPV4Addresses() []string {
	var addresses []string
	ifaces, err := net.Interfaces()
	if err != nil {
		return addresses
	}

	// handle err
	for _, i := range ifaces {
		if i.Index > 3 {
			continue
		}

		addrs, err := i.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip != nil && ip.To4() != nil && !strings.Contains(ip.String(), ":") {
				if addressPingTest(ip.String()) {
					addresses = append(addresses, ip.String())
				}
			}
		}
	}
	return addresses
}

func addressPingTest(addr string) bool {
	result := make(chan bool, 1)
	listener, err := net.Listen("tcp", addr+":0")
	if err != nil {
		return false
	}
	defer func() {
		if err := listener.Close(); err != nil {
			log.E("net.ip.ping", err, "listener.close() caused error")
		}
	}()

	listenAddr := listener.Addr().String()
	go func() {
		con, err := listener.Accept()
		if err != nil {
			result <- false
			return
		}
		defer func() {
			if err := con.Close(); err != nil {
				log.E("net.ip.ping", err, "con.close() caused error")
			}
		}()

		buffer := make([]byte, 1)
		_, err = con.Read(buffer)
		if err != nil {
			result <- false
		}
		if _, err := con.Write(buffer[:1]); err != nil {
			log.E("network", err)
		}
	}()
	go func() {
		con, err := net.Dial("tcp", listenAddr)
		if err != nil {
			log.E("net.ip.ping", err)
			result <- false
			return
		}

		defer func() {
			_ = con.Close()
		}()

		buffer := []byte{12}
		if _, err := con.Write(buffer); err != nil {
			result <- false
			return
		}

		buffer[0] = 0
		if _, err := con.Read(buffer); err != nil {
			log.E("net.ip.ping", err)
		}
		result <- buffer[0] == 12
	}()

	select {
	case r := <-result:
		return r
	case <-time.After(time.Millisecond * 500):
	}
	return false
}

func rfc1918private(ip net.IP) bool {
	for _, cidr := range []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"} {
		_, subnet, err := net.ParseCIDR(cidr)
		if err != nil {
			panic("failed to parse hardcoded rfc1918 cidr: " + err.Error())
		}
		if subnet.Contains(ip) {
			return true
		}
	}
	return false
}

func rfc4193private(ip net.IP) bool {
	_, subnet, err := net.ParseCIDR("fd00::/8")
	if err != nil {
		panic("failed to parse hardcoded rfc4193 cidr: " + err.Error())
	}
	return subnet.Contains(ip)
}

func isLoopback(ip net.IP) bool {
	for _, cidr := range []string{"127.0.0.0/8", "::1/128"} {
		_, subnet, err := net.ParseCIDR(cidr)
		if err != nil {
			panic("failed to parse hardcoded loopback cidr: " + err.Error())
		}
		if subnet.Contains(ip) {
			return true
		}
	}
	return false
}

type Address struct {
	string
	net.IP
}

func heuristic(ni Address) (ret int) {
	a := strings.ToLower(ni.string)
	ip := ni.IP
	if isLoopback(ip) {
		ret += 1000
	}
	if rfc1918private(ip) || rfc4193private(ip) {
		ret += 500
	}
	if strings.Contains(a, "dyn") {
		ret += 100
	}
	if strings.Contains(a, "dhcp") {
		ret += 99
	}
	for i := 0; i < len(ip); i++ {
		if strings.Contains(a, strconv.Itoa(int(ip[i]))) {
			ret += 5
		}
	}
	return ret
}

type nameAndIPByStabilityHeuristic []Address

func (nis nameAndIPByStabilityHeuristic) Len() int { return len(nis) }

func (nis nameAndIPByStabilityHeuristic) Swap(i, j int) { nis[i], nis[j] = nis[j], nis[i] }

func (nis nameAndIPByStabilityHeuristic) Less(i, j int) bool {
	return heuristic(nis[i]) < heuristic(nis[j])
}

func getOutboundIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = conn.Close()
	}()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP, nil
}

func publicAddresses() ([]Address, error) {
	var ret []Address
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	for _, addr := range addrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err != nil {
			return nil, err
		}
		// ignore unresolvable addresses
		names, err := net.LookupAddr(ip.String())
		if err != nil {
			continue
		}
		for _, name := range names {
			ret = append(ret, Address{name, ip})
		}
	}
	sort.Sort(nameAndIPByStabilityHeuristic(ret))
	return ret, nil
}

func PublicAddresses() []string {
	var publics []string

	ip, err := getOutboundIP()
	if err == nil {
		publics = append(publics, ip.String())
	}

	ips, err := publicAddresses()
	if err != nil {
		return publics
	}

	for _, ip := range ips {
		addr := ip.String()
		if addr == "127.0.0.1" {
			continue
		}
		publics = append(publics, addr)
	}
	return publics
}

func LocalAddresses() []string {
	return localIPV4Addresses()
}
