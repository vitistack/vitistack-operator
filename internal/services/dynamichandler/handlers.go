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

	// Determine cache key based on resource kind
	cacheKey := string(unstructuredObject.GetUID())
	if unstructuredObject.GetKind() == "ConfigMap" {
		// For ConfigMaps, use the name as the key
		namespace := unstructuredObject.GetNamespace()
		name := unstructuredObject.GetName()
		cacheKey = fmt.Sprintf("configmap-%s-%s", namespace, name)
	}

	err := localcache.Cache.Set(context.TODO(), cacheKey, unstructuredObject.Object)
	if err != nil {
		vlog.Error("Error setting cache:", err)
		return
	}

	// Publish event to notify subscribers
	eventmanager.EventBus.Publish(eventmanager.ResourceEvent{
		Type:     eventmanager.EventAdd,
		Resource: unstructuredObject,
	})
}

func (handler) DeleteResource(obj any) {
	if obj == nil {
		vlog.Error("DeleteResource called with nil object", nil)
		return
	}

	// Handle tombstone objects (deleted final state unknown)
	var unstructuredObject *unstructured.Unstructured
	if tombstone, ok := obj.(cache.DeletedFinalStateUnknown); ok {
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

	// Determine cache key based on resource kind
	cacheKey := string(unstructuredObject.GetUID())
	if unstructuredObject.GetKind() == "ConfigMap" {
		// For ConfigMaps, use the name as the key
		namespace := unstructuredObject.GetNamespace()
		name := unstructuredObject.GetName()
		cacheKey = fmt.Sprintf("configmap-%s-%s", namespace, name)
	}

	err := localcache.Cache.Delete(context.TODO(), cacheKey)
	if err != nil {
		vlog.Error("Error deleting cache:", err)
		return
	}

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

	// Determine cache key based on resource kind
	cacheKey := string(unstructuredObject.GetUID())
	if unstructuredObject.GetKind() == "ConfigMap" {
		// For ConfigMaps, use the name as the key
		namespace := unstructuredObject.GetNamespace()
		name := unstructuredObject.GetName()
		cacheKey = fmt.Sprintf("configmap-%s-%s", namespace, name)
	}

	err := localcache.Cache.Set(context.TODO(), cacheKey, unstructuredObject.Object)
	if err != nil {
		vlog.Error("Error updating cache:", err)
		return
	}

	// Publish event to notify subscribers
	eventmanager.EventBus.Publish(eventmanager.ResourceEvent{
		Type:     eventmanager.EventUpdate,
		Resource: unstructuredObject,
	})
}
