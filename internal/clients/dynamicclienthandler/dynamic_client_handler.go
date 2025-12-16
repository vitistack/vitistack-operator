package dynamicclienthandler

import (
	"fmt"
	"os"

	"github.com/vitistack/common/pkg/loggers/vlog"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

type DynamicClientHandler interface {
	AddResource(obj any)
	DeleteResource(obj any)
	UpdateResource(_ any, obj any)
	GetSchemas() []schema.GroupVersionResource
}

func Start(discoveryClient *discovery.DiscoveryClient, dynamicClient dynamic.Interface, dynamichandler DynamicClientHandler, stop chan struct{}, sigs chan os.Signal) error {
	vlog.Info("Starting dynamic watchers")

	schemas := dynamichandler.GetSchemas()
	vlog.Info(fmt.Sprintf("Found schemas to watch count=%d", len(schemas)))

	for _, schema := range schemas {
		vlog.Info(fmt.Sprintf(
			"Checking resource availability group=%s version=%s resource=%s",
			schema.Group,
			schema.Version,
			schema.Resource,
		))

		check, err := discovery.IsResourceEnabled(discoveryClient, schema)
		if err != nil {
			vlog.Error("Could not query resources from cluster", err)
		}
		if check {
			vlog.Info(fmt.Sprintf("Resource is available, creating watcher resource=%s", schema.Resource))
			controller := newDynamicWatcher(dynamichandler, dynamicClient, schema)
			go func(res string) {
				vlog.Info(fmt.Sprintf("Starting watcher for resource %s", res))
				controller.Run(stop)
			}(schema.Resource)
		} else {
			errmsg := fmt.Sprintf("Could not register resource %s", schema.Resource)
			vlog.Info(errmsg)
		}
	}
	return nil
}

type DynamicWatcher struct {
	dynInformer cache.SharedIndexInformer
	client      dynamic.Interface
	factory     dynamicinformer.DynamicSharedInformerFactory
}

func (c *DynamicWatcher) Run(stop <-chan struct{}) {
	// Start the informer factory
	c.factory.Start(stop)

	// Wait for cache to sync before processing events
	if !cache.WaitForCacheSync(stop, c.dynInformer.HasSynced) {
		vlog.Error("Failed to sync cache", nil)
		return
	}

	vlog.Info("Cache synced successfully, starting to process events")

	// Wait for stop signal
	<-stop
}

// Function creates a new dynamic controller to listen for api-changes in provided GroupVersionResource
func newDynamicWatcher(dynamichandler DynamicClientHandler, client dynamic.Interface, resource schema.GroupVersionResource) *DynamicWatcher {
	dynWatcher := &DynamicWatcher{}
	dynInformer := dynamicinformer.NewDynamicSharedInformerFactory(client, 0)
	informer := dynInformer.ForResource(resource).Informer()

	dynWatcher.client = client
	dynWatcher.dynInformer = informer
	dynWatcher.factory = dynInformer

	_, err := informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    dynamichandler.AddResource,
		UpdateFunc: dynamichandler.UpdateResource,
		DeleteFunc: dynamichandler.DeleteResource,
	})
	if err != nil {
		vlog.Error("Error adding event handler", err)
	}

	return dynWatcher
}
