package dns

import (
	"strings"
	"time"

	"github.com/peterfromthehill/autocertLego"
	"k8s.io/klog/v2"
)

type DNSWatcher struct {
	k8sclient      *Kubeclient
	currentDomains []string
}

func NewDNSWatcher() *DNSWatcher {
	dnswatcher := &DNSWatcher{}
	kubernetesClient, err := NewKubernetesClient("")
	if err == nil {
		dnswatcher.k8sclient = kubernetesClient
	} else {
		klog.Error("No kubernetes config found :(")
	}
	return dnswatcher
}

func (d DNSWatcher) GetWhitelistetDomains() []string {
	var dnsNames []string
	externalIP, err := externalIP()
	if err != nil {
		klog.Error(err.Error())
	}
	var serviceNames []string

	if d.k8sclient != nil {
		serviceNames = d.k8sclient.GetMatchedServices(externalIP)
		for _, serviceName := range serviceNames {
			err := d.k8sclient.AddChallencePortToService(strings.Split(serviceName, ".")[0])
			if err != nil {
				klog.Error(err.Error())
			}
		}
	}

	externalIPinAddr, _ := reverseaddr(externalIP)
	rDNSNames, err := reverseLookup(externalIPinAddr)
	if err != nil {
		klog.Error(err.Error())
	}
	dnsNames = append(dnsNames, externalIP)
	dnsNames = append(dnsNames, serviceNames...)
	dnsNames = append(dnsNames, rDNSNames...)
	for _, dnsName := range dnsNames {
		bla, _ := traverseUpside(dnsName)
		dnsNames = append(dnsNames, bla...)
	}
	dnsNames = unique(dnsNames)
	return dnsNames
}

func (d DNSWatcher) WatchAndUpdateAutocertManager(m *autocertLego.Manager) {

	for {
		domains := d.GetWhitelistetDomains()
		if !Equal(domains, d.currentDomains) {
			klog.Infof("-> %q", domains)
			m.HostPolicy = autocertLego.HostWhitelist(domains...)
			d.currentDomains = domains
		}
		time.Sleep(10 * time.Second)
	}
}
