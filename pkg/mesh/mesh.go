package mesh

import (
	"poor-man-service-mesh/pkg/proxyserver"
	"poor-man-service-mesh/pkg/tproxy"

	"k8s.io/klog/v2"
)

type Mesh struct {
	cProxy *proxyserver.HttpProxyServer
	tProxy *tproxy.TProxy
	finish chan bool
}

func NewMesh(cProxy *proxyserver.HttpProxyServer, tProxy *tproxy.TProxy) *Mesh {
	return &Mesh{
		cProxy: cProxy,
		tProxy: tProxy,
		finish: make(chan bool),
	}
}

func (mesh *Mesh) Listen() {
	go mesh.listenTProxy()
	go mesh.listenCProxy()
	<-mesh.finish
}

func (mesh *Mesh) listenTProxy() error {
	if mesh.tProxy == nil {
		klog.Info("No TProxy configured")
		return nil
	}
	panic(mesh.tProxy.Listen())
}

func (mesh *Mesh) listenCProxy() error {
	panic(mesh.cProxy.Listen())
}
