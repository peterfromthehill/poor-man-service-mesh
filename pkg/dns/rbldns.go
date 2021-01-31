package dns

import (
	"net"

	"github.com/miekg/dns"
	"k8s.io/klog/v2"
)

// Rblresult extends godnsbl and adds RBL name
type Rblresult struct {
	Address   string
	Listed    bool
	Text      string
	Error     bool
	ErrorType error
	Rbl       string
	Target    string
}

// Rbl ... object
type Rbl struct {
	Resolver  string
	Results   []Rblresult
	DNSClient *dns.Client
}

// NewRbl ... factory
func NewRbl(resolver string) Rbl {
	client := new(dns.Client)

	rbl := Rbl{
		Resolver:  resolver,
		DNSClient: client,
	}

	return rbl
}

func (rbl *Rbl) createQuestion(target string, record uint16) *dns.Msg {
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(target), record)

	return msg
}

func (rbl *Rbl) makeQuery(msg *dns.Msg) (*dns.Msg, error) {
	// fixme: we should inject the port as well
	result, rt, err := rbl.DNSClient.Exchange(msg, net.JoinHostPort(rbl.Resolver, "53"))
	//log.Println("Roundtrip", rt) // fixme -> histogram
	_ = rt
	return result, err
}

func (rbl *Rbl) GetARecords(target string) ([]string, error) {
	msg := rbl.createQuestion(target, dns.TypeA)

	result, err := rbl.makeQuery(msg)

	var list []string

	if err == nil && len(result.Answer) > 0 {
		for _, ans := range result.Answer {
			if t, ok := ans.(*dns.A); ok {
				if klog.V(3).Enabled() {
					klog.Infof("We have an A-Record %s for %s", t.A.String(), target)
				}
				list = append(list, t.A.String())
			}
		}
	}
	return list, err
}

func (rbl *Rbl) GetPTRRecords(target string) ([]string, error) {
	msg := rbl.createQuestion(target, dns.TypePTR)

	result, err := rbl.makeQuery(msg)

	var list []string

	if err == nil && len(result.Answer) > 0 {
		for _, ans := range result.Answer {
			if t, ok := ans.(*dns.PTR); ok {
				if klog.V(3).Enabled() {
					klog.Infof("We have an PTR-Record %s for %s; %q", t.String(), target, t.Ptr)
				}
				list = append(list, t.Ptr)
			}
		}
	}
	return list, err
}
