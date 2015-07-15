// kubegateway is a bridge between Kubernetes services and the outside world.
package main // import "github.com/joelpm/kubegateway"

import (
	"flag"
	"fmt"
	"hash/fnv"
	"net/url"
	"os"
	"sync"
	"time"

	kapi "github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	kclient "github.com/GoogleCloudPlatform/kubernetes/pkg/client"
	kcache "github.com/GoogleCloudPlatform/kubernetes/pkg/client/cache"
	kclientcmd "github.com/GoogleCloudPlatform/kubernetes/pkg/client/clientcmd"
	kframework "github.com/GoogleCloudPlatform/kubernetes/pkg/controller/framework"
	kSelector "github.com/GoogleCloudPlatform/kubernetes/pkg/fields"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/util"
	"github.com/golang/glog"
)

var (
	// TODO: switch to pflag and make - and _ equivalent.
	argKubecfgFile   = flag.String("kubecfg_file", "", "Location of kubecfg file for access to kubernetes master service; --kube_master_url overrides the URL part of this; if neither this nor --kube_master_url are provided, defaults to service account tokens")
	argKubeMasterURL = flag.String("kube_master_url", "", "URL to reach kubernetes master. Env variables in this flag will be expanded.")
)

const (
	// Resync period for the kube controller loop.
	resyncPeriod = 30 * time.Minute
	// A subdomain added to the user specified domain for all services.
	serviceSubdomain = "svc"
)

type kubeGateway struct {
	// Cache for all the services in the system.
	servicesStore kcache.Store
	mlock         sync.Mutex
}

type serviceRule struct {
	service string
	ip      string
	port    int
	host    string
	path    string
}

// Returns a cache.ListWatch that gets all changes to services.
func createServiceLW(kubeClient *kclient.Client) *kcache.ListWatch {
	return kcache.NewListWatchFromClient(kubeClient, "services", kapi.NamespaceAll, kSelector.Everything())
}

func (kgw *kubeGateway) newService(obj interface{}) {
	if s, ok := obj.(*kapi.Service); ok {
		glog.V(1).Infof("new service: %+v\n", s)
	}
}

func (kgw *kubeGateway) removeService(obj interface{}) {
	if s, ok := obj.(*kapi.Service); ok {
		glog.V(1).Infof("removed service: %+v\n", s)
	}
}

func (kgw *kubeGateway) updateService(oldObj, newObj interface{}) {
	if s, ok := newObj.(*kapi.Service); ok {
		glog.V(1).Infof("updated service: %+v\n", s)
	}
}

func expandKubeMasterURL() (string, error) {
	parsedURL, err := url.Parse(os.ExpandEnv(*argKubeMasterURL))
	if err != nil {
		return "", fmt.Errorf("failed to parse --kube_master_url %s - %v", *argKubeMasterURL, err)
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" || parsedURL.Host == ":" {
		return "", fmt.Errorf("invalid --kube_master_url specified %s", *argKubeMasterURL)
	}
	return parsedURL.String(), nil
}

// TODO: evaluate using pkg/client/clientcmd
func newKubeClient() (*kclient.Client, error) {
	var (
		config    *kclient.Config
		err       error
		masterURL string
	)
	// If the user specified --kube_master_url, expand env vars and verify it.
	if *argKubeMasterURL != "" {
		masterURL, err = expandKubeMasterURL()
		if err != nil {
			return nil, err
		}
	}
	if masterURL != "" && *argKubecfgFile == "" {
		// Only --kube_master_url was provided.
		config = &kclient.Config{
			Host:    masterURL,
			Version: "v1",
		}
	} else {
		// We either have:
		//  1) --kube_master_url and --kubecfg_file
		//  2) just --kubecfg_file
		//  3) neither flag
		// In any case, the logic is the same.  If (3), this will automatically
		// fall back on the service account token.
		overrides := &kclientcmd.ConfigOverrides{}
		overrides.ClusterInfo.Server = masterURL                                     // might be "", but that is OK
		rules := &kclientcmd.ClientConfigLoadingRules{ExplicitPath: *argKubecfgFile} // might be "", but that is OK
		if config, err = kclientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides).ClientConfig(); err != nil {
			return nil, err
		}
	}

	glog.Infof("Using %s for kubernetes master", config.Host)
	glog.Infof("Using kubernetes API %s", config.Version)
	return kclient.New(config)
}

func watchForServices(kubeClient *kclient.Client, kgw *kubeGateway) kcache.Store {
	serviceStore, serviceController := kframework.NewInformer(
		createServiceLW(kubeClient),
		&kapi.Service{},
		resyncPeriod,
		kframework.ResourceEventHandlerFuncs{
			AddFunc:    kgw.newService,
			DeleteFunc: kgw.removeService,
			UpdateFunc: kgw.updateService,
		},
	)
	go serviceController.Run(util.NeverStop)
	return serviceStore
}

func getHash(text string) string {
	h := fnv.New32a()
	h.Write([]byte(text))
	return fmt.Sprintf("%x", h.Sum32())
}

func main() {
	flag.Parse()
	var err error

	// TODO: Validate input flags.

	kubeClient, err := newKubeClient()
	if err != nil {
		glog.Fatalf("Failed to create a kubernetes client: %v", err)
	}

	kgw := kubeGateway{}

	kgw.servicesStore = watchForServices(kubeClient, &kgw)

	select {}
}
