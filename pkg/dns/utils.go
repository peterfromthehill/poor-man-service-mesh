package dns

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"regexp"
	"strings"

	"k8s.io/klog/v2"
)

func reverseaddr(addr string) (arpa string, err error) {
	ip := net.ParseIP(addr)
	if ip == nil {
		return "", &net.DNSError{Err: "unrecognized address", Name: addr}
	}
	if ip.To4() != nil {
		return uitoa(uint(ip[15])) + "." + uitoa(uint(ip[14])) + "." + uitoa(uint(ip[13])) + "." + uitoa(uint(ip[12])) + ".in-addr.arpa.", nil
	}
	return "", fmt.Errorf("Given IP is not an IPv4")
}

func reverseLookup(addr string) ([]string, error) {
	candidateResolvConf, err := ioutil.ReadFile("/etc/resolv.conf")
	if err != nil {
		// silencing error as it will resurface at next calls trying to read defaultPath
		return []string{}, err
	}
	ns := getNameservers(candidateResolvConf, IP)
	if len(ns) == 1 && ns[0] == "127.0.0.53" {
		klog.Warning("detected 127.0.0.53 nameserver, assuming systemd-resolved, so using resolv.conf")
	}
	for _, server := range ns {
		if klog.V(4).Enabled() {
			klog.Infof("try server %s", server)
		}
		rbl := NewRbl(server)
		rr, err := rbl.GetPTRRecords(addr)
		if err != nil {
			klog.Error(err.Error())
			continue
		}
		var list []string
		for _, ans := range rr {
			list = append(list, strings.Trim(ans, "."))
		}
		return list, nil
	}
	return []string{}, nil
}

// constants for the IP address type
const (
	IP = iota // IPv4 and IPv6
	IPv4
	IPv6
)

// GetNameservers returns nameservers (if any) listed in /etc/resolv.conf
func getNameservers(resolvConf []byte, kind int) []string {
	ipv4NumBlock := `(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)`
	ipv4Address := `(` + ipv4NumBlock + `\.){3}` + ipv4NumBlock
	ipv6Address := `([0-9A-Fa-f]{0,4}:){2,7}([0-9A-Fa-f]{0,4})(%\w+)?`
	nsRegexp := regexp.MustCompile(`^\s*nameserver\s*((` + ipv4Address + `)|(` + ipv6Address + `))\s*$`)
	nsIPv6Regexpmatch := regexp.MustCompile(`^\s*nameserver\s*((` + ipv6Address + `))\s*$`)
	nsIPv4Regexpmatch := regexp.MustCompile(`^\s*nameserver\s*((` + ipv4Address + `))\s*$`)
	nameservers := []string{}
	for _, line := range getLines(resolvConf, []byte("#")) {
		var ns [][]byte
		if kind == IP {
			ns = nsRegexp.FindSubmatch(line)
		} else if kind == IPv4 {
			ns = nsIPv4Regexpmatch.FindSubmatch(line)
		} else if kind == IPv6 {
			ns = nsIPv6Regexpmatch.FindSubmatch(line)
		}
		if len(ns) > 0 {
			nameservers = append(nameservers, string(ns[1]))
		}
	}
	return nameservers
}

// getLines parses input into lines and strips away comments.
func getLines(input []byte, commentMarker []byte) [][]byte {
	lines := bytes.Split(input, []byte("\n"))
	var output [][]byte
	for _, currentLine := range lines {
		var commentIndex = bytes.Index(currentLine, commentMarker)
		if commentIndex == -1 {
			output = append(output, currentLine)
		} else {
			output = append(output, currentLine[:commentIndex])
		}
	}
	return output
}

func traverseUpside(fqdn string) ([]string, error) {
	var list []string
	fqdnList := strings.Split(fqdn, ".")
	for i, _ := range fqdnList {
		if klog.V(4).Enabled() {
			klog.Infof("try  %s", strings.Join(fqdnList[:i+1], "."))
		}
		_, err := net.LookupHost(strings.Join(fqdnList[:i+1], "."))
		if err != nil {
			continue
		}
		list = append(list, strings.Join(fqdnList[:i+1], "."))
	}
	return list, nil
}

func unique(intSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// Equal tells whether a and b contain the same elements.
// A nil argument is equivalent to an empty slice.
func Equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func externalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}

// Convert unsigned integer to decimal string.
func uitoa(val uint) string {
	if val == 0 { // avoid string allocation
		return "0"
	}
	var buf [20]byte // big enough for 64bit value base 10
	i := len(buf) - 1
	for val >= 10 {
		q := val / 10
		buf[i] = byte('0' + val - q*10)
		i--
		val = q
	}
	// val < 10
	buf[i] = byte('0' + val)
	return string(buf[i:])
}
