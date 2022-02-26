package main

import (
	"flag"
)

var config *MeshConfig

func init() {
	config = &MeshConfig{}
	flag.StringVar(&config.ExitProxy, "exitProxy", "pmsm-exit.cloud-svc.svc.cluster.local:15006", "hostname:port of the exit proxy (default TLS). If this is an exit node exitProxy is the HTTP/CONNECT Proxy (without TLS)")
	flag.StringVar(&config.ServiceCidr, "service-cidr", "10.96.0.0/12", "service Cidr")
	flag.StringVar(&config.PodCidr, "pod-cidr", "10.244.0.0/16", "pod Cidr")
	flag.BoolVar(&config.Secure, "secure", false, "use TLS")
	flag.StringVar(&config.SslKeyPath, "key", "/cert/tls.key", "Key file")
	flag.StringVar(&config.SslCertPath, "cert", "/cert/tls.pem", "Cert file")
	flag.IntVar(&config.IncomePort, "incoming-port", 15006, "inbound traffic port")
	flag.IntVar(&config.OutgoingPort, "outgoing-port", 15001, "outbound traffic port")
	flag.StringVar(&config.CA, "cacert", "", "CA file")
	flag.StringVar(&config.DirectoryURL, "directory-url", "", "ACME Directory URL (https://acme.proxy.svc.cluster.local/acme/development/directory)")
	flag.BoolVar(&config.ExitNode, "exitNode", false, "if true this node act as the exit node (the node spezified with -exitProxy")
}

func main() {
	flag.Parse()

	run(config)
}
