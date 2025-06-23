package dynamicclienthandler

import (
	"fmt"
	"os"

	"github.com/NorskHelsenett/ror/pkg/rlog"

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
	rlog.Info("Starting dynamic watchers")

	schemas := dynamichandler.GetSchemas()
	rlog.Info("Found schemas to watch", rlog.Int("count", len(schemas)))

	for _, schema := range schemas {
		rlog.Info("Checking resource availability",
			rlog.String("group", schema.Group),
			rlog.String("version", schema.Version),
			rlog.String("resource", schema.Resource))

		check, err := discovery.IsResourceEnabled(discoveryClient, schema)
		if err != nil {
			rlog.Error("Could not query resources from cluster", err)
		}
		if check {
			rlog.Info("Resource is available, creating watcher", rlog.String("resource", schema.Resource))
			controller := newDynamicWatcher(dynamichandler, dynamicClient, schema)
			go func(res string) {
				rlog.Info("Starting watcher for resource", rlog.String("resource", res))
				controller.Run(stop)
			}(schema.Resource)
		} else {
			errmsg := fmt.Sprintf("Could not register resource %s", schema.Resource)
			rlog.Info(errmsg)
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
		rlog.Error("Failed to sync cache", nil)
		return
	}

	rlog.Info("Cache synced successfully, starting to process events")

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
		rlog.Error("Error adding event handler", err)
	}

	return dynWatcher
}
