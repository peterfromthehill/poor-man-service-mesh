module poor-man-service-mesh

go 1.15

require (
	github.com/go-acme/lego/v4 v4.14.2 // indirect
	github.com/go-jose/go-jose/v3 v3.0.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/gopacket v1.1.19
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/miekg/dns v1.1.57
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/paultag/sniff v0.0.0-20200207005214-cf7e4d167732
	github.com/peterfromthehill/autocertLego v0.0.0
	golang.org/x/crypto v0.15.0
	golang.org/x/oauth2 v0.14.0 // indirect
	golang.org/x/time v0.4.0 // indirect
	golang.org/x/tools v0.15.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
	honnef.co/go/netdb v0.0.0-20210921115105-e902e863d85d
	k8s.io/api v0.28.4
	k8s.io/apimachinery v0.28.4
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/klog v1.0.0
	k8s.io/klog/v2 v2.110.1
	k8s.io/utils v0.0.0-20231121161247-cf03d44ff3cf // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.1 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect

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
