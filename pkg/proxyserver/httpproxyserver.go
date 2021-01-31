package proxyserver

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"poor-man-service-mesh/pkg/certpool"
	"poor-man-service-mesh/pkg/dns"
	"time"

	"github.com/peterfromthehill/autocertLego"
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
	"k8s.io/klog/v2"
)

type HttpProxyInstance interface {
	HandleTunneling(w http.ResponseWriter, r *http.Request)
}

type HttpProxyServer struct {
	proxyInstance   HttpProxyInstance
	port            int
	sslKeyFilePath  string
	sslCertFilePath string
	directoryURL    string
	secure          bool
}

func NewHttpProxyServer(port int, sslKeyFilePath string, sslCertFilePath string, httpProxyInstance HttpProxyInstance) *HttpProxyServer {
	return &HttpProxyServer{
		port:            port,
		sslKeyFilePath:  sslKeyFilePath,
		sslCertFilePath: sslCertFilePath,
		secure:          true,
		proxyInstance:   httpProxyInstance,
	}
}

func NewHttpProxyServerInsecure(port int, httpProxyInstance HttpProxyInstance) *HttpProxyServer {
	return &HttpProxyServer{
		port:          port,
		secure:        false,
		proxyInstance: httpProxyInstance,
	}
}

func NewHttpProxyServerWithACME(port int, directoryUrl string, httpProxyInstance HttpProxyInstance) *HttpProxyServer {
	return &HttpProxyServer{
		port:          port,
		directoryURL:  directoryUrl,
		secure:        true,
		proxyInstance: httpProxyInstance,
	}
}

func (proxyServer *HttpProxyServer) listen0(server *http.Server) error {
	if proxyServer.directoryURL != "" && proxyServer.sslKeyFilePath == "" {
		tlsConfig, _ := proxyServer.acmeSetup()
		server.TLSConfig = tlsConfig
		klog.Infof("Starting secure HTTPS w/ACME Server: %d\n", proxyServer.port)
		return server.ListenAndServeTLS("", "")
	} else if proxyServer.sslKeyFilePath != "" && proxyServer.directoryURL == "" {
		klog.Infof("Starting secure HTTPS Server: %d\n", proxyServer.port)
		return server.ListenAndServeTLS(proxyServer.sslCertFilePath, proxyServer.sslKeyFilePath)
	} else if proxyServer.secure {
		return fmt.Errorf("Not ACME/No certs? Come on...")
	}
	klog.Infof("Starting insecure HTTP Server: %d\n", proxyServer.port)
	return server.ListenAndServe()
}

func (proxyServer *HttpProxyServer) Listen() error {
	tlsConfig := &tls.Config{}
	server := &http.Server{
		Addr:      fmt.Sprintf(":%d", proxyServer.port),
		Handler:   http.HandlerFunc(proxyServer.proxyInstance.HandleTunneling),
		TLSConfig: tlsConfig,
		// Disable HTTP/2.
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}
	return proxyServer.listen0(server)
}

func (proxyServer *HttpProxyServer) acmeSetup() (*tls.Config, *autocertLego.Manager) {
	dnswatcher := dns.NewDNSWatcher()

	manager := &autocertLego.Manager{
		Cache:      autocert.DirCache("secret-dir"),
		HostPolicy: autocertLego.HostWhitelist(""),
		EMail:      "",
		Directory:  proxyServer.directoryURL,
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				TLSHandshakeTimeout:   15 * time.Second,
				ResponseHeaderTimeout: 15 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				TLSClientConfig: &tls.Config{
					RootCAs: certpool.GetCache().RootCAs,
				},
			},
		},
	}
	go dnswatcher.WatchAndUpdateAutocertManager(manager)

	tlsConfig := manager.TLSConfig()
	// tlsConfig.RootCAs = certpool.GetCache().RootCAs
	tlsConfig.NextProtos = []string{
		"http/1.1",     // enable HTTP/1.1 // Do not allow HTTP/2!!!
		acme.ALPNProto, // enable tls-alpn ACME challenges
	}
	return tlsConfig, manager
}
