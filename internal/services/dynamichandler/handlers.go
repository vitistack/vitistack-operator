package dynamichandler

import (
	"context"
	"fmt"

	"github.com/vitistack/common/pkg/loggers/vlog"
	localcache "github.com/vitistack/vitistack-operator/internal/cache"
	"github.com/vitistack/vitistack-operator/internal/clients/dynamicclienthandler"
	"github.com/vitistack/vitistack-operator/pkg/eventmanager"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/tools/cache"
)

type handler struct {
}

func NewDynamicClientHandler() dynamicclienthandler.DynamicClientHandler {
	ret := handler{}
	return &ret
}

func (handler) AddResource(obj any) {
	if obj == nil {
		vlog.Error("AddResource called with nil object", nil)
		return
	}
	unstructuredObject, ok := obj.(*unstructured.Unstructured)
	if !ok || unstructuredObject == nil {
		vlog.Error("AddResource: failed to cast object to Unstructured", nil)
		return
	}
	vlog.Info(fmt.Sprintf("AddResource called - name: %s, kind: %s, namespace: %s",
		unstructuredObject.GetName(),
		unstructuredObject.GetKind(),
		unstructuredObject.GetNamespace()))

	// Determine cache key based on resource kind
	cacheKey := string(unstructuredObject.GetUID())
	if unstructuredObject.GetKind() == "ConfigMap" {
		// For ConfigMaps, use the name as the key
		namespace := unstructuredObject.GetNamespace()
		name := unstructuredObject.GetName()
		cacheKey = fmt.Sprintf("configmap-%s-%s", namespace, name)
		vlog.Info(fmt.Sprintf("Using ConfigMap name as cache key: %s", cacheKey))
	}

	err := localcache.Cache.Set(context.TODO(), cacheKey, unstructuredObject.Object)
	if err != nil {
		vlog.Error("Error setting cache:", err)
		return
	}
	vlog.Info("Cache set successfully")

	// Publish event to notify subscribers
	eventmanager.EventBus.Publish(eventmanager.ResourceEvent{
		Type:     eventmanager.EventAdd,
		Resource: unstructuredObject,
	})

	vlog.Info("Published add event for resource",
		"name: ", unstructuredObject.GetName(),
		"kind: ", unstructuredObject.GetKind())
}

func (handler) DeleteResource(obj any) {
	if obj == nil {
		vlog.Error("DeleteResource called with nil object", nil)
		return
	}

	// Handle tombstone objects (deleted final state unknown)
	var unstructuredObject *unstructured.Unstructured
	if tombstone, ok := obj.(cache.DeletedFinalStateUnknown); ok {
		vlog.Info("DeleteResource: processing tombstone object")
		unstructuredObject, ok = tombstone.Obj.(*unstructured.Unstructured)
		if !ok || unstructuredObject == nil {
			vlog.Error("DeleteResource: failed to cast tombstone object to Unstructured", nil)
			return
		}
	} else {
		var castOk bool
		unstructuredObject, castOk = obj.(*unstructured.Unstructured)
		if !castOk || unstructuredObject == nil {
			vlog.Error("DeleteResource: failed to cast object to Unstructured", nil)
			return
		}
	}

	vlog.Info(fmt.Sprintf("Delete Resource, name: %s, kind: %s", unstructuredObject.GetName(), unstructuredObject.GetKind()))

	// Determine cache key based on resource kind
	cacheKey := string(unstructuredObject.GetUID())
	if unstructuredObject.GetKind() == "ConfigMap" {
		// For ConfigMaps, use the name as the key
		namespace := unstructuredObject.GetNamespace()
		name := unstructuredObject.GetName()
		cacheKey = fmt.Sprintf("configmap-%s-%s", namespace, name)
		vlog.Info(fmt.Sprintf("Using ConfigMap name as cache key: %s", cacheKey))
	}

	err := localcache.Cache.Delete(context.TODO(), cacheKey)
	if err != nil {
		vlog.Error("Error deleting cache:", err)
		return
	}
	vlog.Info("Cache deleted successfully")

	// Publish event to notify subscribers
	eventmanager.EventBus.Publish(eventmanager.ResourceEvent{
		Type:     eventmanager.EventDelete,
		Resource: unstructuredObject,
	})
}

func (handler) UpdateResource(_ any, obj any) {
	if obj == nil {
		vlog.Error("UpdateResource called with nil object", nil)
		return
	}
	unstructuredObject, ok := obj.(*unstructured.Unstructured)
	if !ok || unstructuredObject == nil {
		vlog.Error("UpdateResource: failed to cast object to Unstructured", nil)
		return
	}
	vlog.Info(fmt.Sprintf("UpdateResource called - name: %s, kind: %s, namespace: %s",
		unstructuredObject.GetName(),
		unstructuredObject.GetKind(),
		unstructuredObject.GetNamespace()))

	// Determine cache key based on resource kind
	cacheKey := string(unstructuredObject.GetUID())
	if unstructuredObject.GetKind() == "ConfigMap" {
		// For ConfigMaps, use the name as the key
		namespace := unstructuredObject.GetNamespace()
		name := unstructuredObject.GetName()
		cacheKey = fmt.Sprintf("configmap-%s-%s", namespace, name)
		vlog.Info(fmt.Sprintf("Using ConfigMap name as cache key: %s", cacheKey))
	}

	err := localcache.Cache.Set(context.TODO(), cacheKey, unstructuredObject.Object)
	if err != nil {
		vlog.Error("Error updating cache:", err)
		return
	}
	vlog.Info("Cache updated successfully")

	// Additional logging for ConfigMap updates
	if unstructuredObject.GetKind() == "ConfigMap" {
		vlog.Info("ConfigMap updated in cache",
			"name: ", unstructuredObject.GetName(),
			"namespace: ", unstructuredObject.GetNamespace(),
			"cacheKey: ", cacheKey)
	}

	// Publish event to notify subscribers
	eventmanager.EventBus.Publish(eventmanager.ResourceEvent{
		Type:     eventmanager.EventUpdate,
		Resource: unstructuredObject,
	})

	vlog.Info("Published update event for resource",
		"name: ", unstructuredObject.GetName(),
		"kind: ", unstructuredObject.GetKind())
}
