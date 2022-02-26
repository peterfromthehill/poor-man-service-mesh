module poor-man-service-mesh

go 1.15

require (
	github.com/google/gopacket v1.1.19
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/miekg/dns v1.1.38
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/paultag/sniff v0.0.0-20200207005214-cf7e4d167732
	github.com/peterfromthehill/autocertLego v0.0.0-20210209204954-f1561af40a7a
	golang.org/x/crypto v0.0.0-20220214200702-86341886e292
	golang.org/x/time v0.0.0-20220210224613-90d013bbcef8 // indirect
	honnef.co/go/netdb v0.0.0-20210921115105-e902e863d85d
	k8s.io/api v0.23.4
	k8s.io/apimachinery v0.23.4
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/klog v1.0.0
	k8s.io/klog/v2 v2.5.0
	k8s.io/utils v0.0.0-20210111153108-fddb29f9d009 // indirect

)

// Pin k8s deps to 1.17.6
replace (
	k8s.io/api => k8s.io/api v0.20.2
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.20.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.2
	k8s.io/apiserver => k8s.io/apiserver v0.20.2
	k8s.io/client-go => k8s.io/client-go v0.20.2
	k8s.io/code-generator => k8s.io/code-generator v0.20.2
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20200410145947-bcb3869e6f29
)
