package main

import (
	"errors"
	"poor-man-service-mesh/pkg/cProxy"
	"poor-man-service-mesh/pkg/certpool"
	"poor-man-service-mesh/pkg/eproxy"
	"poor-man-service-mesh/pkg/mesh"
	"poor-man-service-mesh/pkg/proxyserver"
	"poor-man-service-mesh/pkg/tproxy"

	"k8s.io/klog/v2"
)

func run(config *MeshConfig) {

	cache := certpool.GetCache()
	if config.CA != "" {
		err := cache.AddCerts(config.CA)
		if err != nil {
			panic(err)
		}
	}
	var tproxyServer *tproxy.TProxy
	if !config.ExitNode {
		tproxyServer = tproxy.NewTProxy(config.OutgoingPort, config.ServiceCidr, config.PodCidr, config.ExitProxy, config.Secure)
	}

	var incomeProxyInstance proxyserver.HttpProxyInstance
	if config.ExitNode {
		klog.Info("Starting eProxy")
		incomeProxyInstance = eproxy.Build(config.ExitProxy)
	} else {
		klog.Info("Starting cProxy")
		incomeProxyInstance = cProxy.Build()
	}
	var cproxy *proxyserver.HttpProxyServer
	if config.Secure {
		if config.DirectoryURL != "" {
			cproxy = proxyserver.NewHttpProxyServerWithACME(config.IncomePort, config.DirectoryURL, incomeProxyInstance)
		} else {
			cproxy = proxyserver.NewHttpProxyServer(config.IncomePort, config.SslKeyPath, config.SslCertPath, incomeProxyInstance)
		}
	} else {
		cproxy = proxyserver.NewHttpProxyServerInsecure(config.IncomePort, incomeProxyInstance)
	}
	mesh := mesh.NewMesh(cproxy, tproxyServer)
	mesh.Listen()
	panic(errors.New("Something goes wrong"))
}
