package main

import (
	"flag"
	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	//"golang.org/x/tools/go/gcimporter15/testdata"
	//"golang.org/x/tools/go/gcimporter15/testdata"
	//"golang.org/x/tools/go/gcimporter15/testdata"
	//"golang.org/x/tools/go/gcimporter15/testdata"
	//"golang.org/x/tools/go/gcimporter15/testdata"
	//"golang.org/x/tools/go/gcimporter15/testdata"
)

//https://github.com/kubernetes/client-go/blob/master/examples/out-of-cluster-client-configuration/main.go
//https://github.com/kubernetes/client-go/tree/master/examples/in-cluster-client-configuration

// optional - local kubeconfig for testing
var kubeconfig = flag.String("kubeconfig", "", "Path to a kubeconfig file")

func main() {
	// send logs to stderr so we can use 'kubectl logs'
	flag.Set("logtostderr", "true")
	flag.Set("v", "3")
	flag.Parse()
	config, err := getConfig(*kubeconfig)
	if err != nil {
		glog.Errorf("Failed to load client config: %v", err)
		return
	}
	// build the Kubernetes client
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Errorf("Failed to create kubernetes client: %v", err)
		return
	}
	// list pods
	pods, err := client.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		glog.Errorf("Failed to retrieve pods: %v", err)
		return
	}
	for _, p := range pods.Items {
		glog.V(3).Infof("Found pods: %s/%s", p.Namespace, p.Name)
	}
}
func getConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}
