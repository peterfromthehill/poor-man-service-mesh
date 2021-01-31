package dns

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	"honnef.co/go/netdb"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilv1 "k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
)

type Kubeclient struct {
	clientset *kubernetes.Clientset
	namespace string
}

func NewKubernetesClient(kubeconfig string) (*Kubeclient, error) {
	namespace, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return nil, err
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &Kubeclient{
		clientset: clientset,
		namespace: string(namespace),
	}, nil
}

func (k *Kubeclient) GetMatchedServices(ipaddr string) []string {
	endpointsClient := k.clientset.CoreV1().Endpoints(k.namespace)
	endpoints, err := endpointsClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Kubernetes: Error list Endpoint objects: %v", err)
		return []string{}
	}
	var list []string
	for _, e := range endpoints.Items {
		for _, s := range e.Subsets {
			for _, a := range s.Addresses {
				if ipaddr == a.IP {
					if klog.V(3).Enabled() {
						klog.Infof(" * %s %s %s %s/%s", e.Name, a.IP, *a.NodeName, a.TargetRef.Kind, a.TargetRef.Name)
					}
					serviceIP, err := k.GetServiceIPByServiceName(e.Name)
					if err != nil {
						continue
					}
					serviceIPAddr, err := reverseaddr(serviceIP)
					if err != nil {
						klog.Warningf("Error resolving %s: %s", serviceIP, err.Error())
						continue
					}
					serviceNames, err := reverseLookup(serviceIPAddr)
					if err != nil {
						klog.Warningf("Error resolving %s: %s", serviceIP, err.Error())
						continue
					}
					list = append(list, serviceNames...)
				}
			}
		}
	}
	return list
}

func (k *Kubeclient) GetServiceIPByServiceName(serviceName string) (string, error) {
	servicesClient := k.clientset.CoreV1().Services(k.namespace)
	services, err := servicesClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Kubernetes: Error list Services objects: %s", err.Error())
		return "", err
	}
	for _, s := range services.Items {
		if serviceName == s.ObjectMeta.Name {
			return s.Spec.ClusterIP, nil
		}
	}
	return "", nil
}

// AddChallencePortToService adds port 443 for the smallsteps ACME server :/
func (k *Kubeclient) AddChallencePortToService(serviceName string) error {
	servicesClient := k.clientset.CoreV1().Services(k.namespace)
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Service before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		result, getErr := servicesClient.Get(context.TODO(), serviceName, metav1.GetOptions{})
		if getErr != nil {
			return fmt.Errorf("Failed to get latest version of Service: %v", getErr)
		}

		for _, p := range result.Spec.Ports {
			if p.Port == 443 {
				//service port ist vorhanden
				return nil
			}
		}
		if len(result.Spec.Ports) == 1 {
			proto := netdb.GetProtoByName(string(result.Spec.Ports[0].Protocol))
			servent := netdb.GetServByPort(int(result.Spec.Ports[0].Port), proto)
			if servent != nil {
				result.Spec.Ports[0].Name = servent.Name
			} else {
				result.Spec.Ports[0].Name = fmt.Sprintf("%s%d", strings.ToLower(string(result.Spec.Ports[0].Protocol)), result.Spec.Ports[0].Port)
			}
		}

		servicePort := corev1.ServicePort{
			Name:       "acme",
			Protocol:   corev1.ProtocolTCP,
			Port:       443,
			TargetPort: utilv1.IntOrString{Type: utilv1.Int, IntVal: int32(443)},
		}

		result.Spec.Ports = append(result.Spec.Ports, servicePort)

		// result.Spec.Replicas = int32Ptr(1)                           // reduce replica count
		// result.Spec.Template.Spec.Containers[0].Image = "nginx:1.13" // change nginx version
		_, updateErr := servicesClient.Update(context.TODO(), result, metav1.UpdateOptions{})
		return updateErr
	})
	if retryErr != nil {
		return fmt.Errorf("Update failed: %v", retryErr)
	}
	return nil
}
