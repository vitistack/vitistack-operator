package dynamichandler

import (
	"context"
	"fmt"

	"github.com/NorskHelsenett/ror/pkg/rlog"
	"github.com/vitistack/datacenter-operator/internal/cache"
	"github.com/vitistack/datacenter-operator/internal/clients/dynamicclienthandler"
	"github.com/vitistack/datacenter-operator/pkg/eventmanager"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type handler struct {
}

func NewDynamicClientHandler() dynamicclienthandler.DynamicClientHandler {
	ret := handler{}
	return &ret
}

func (handler) AddResource(obj any) {
	unstructuredObject := obj.(*unstructured.Unstructured)
	rlog.Info(fmt.Sprintf("AddResource called - name: %s, kind: %s, namespace: %s",
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
		rlog.Info(fmt.Sprintf("Using ConfigMap name as cache key: %s", cacheKey))
	}

	err := cache.Cache.Set(context.TODO(), cacheKey, unstructuredObject.Object)
	if err != nil {
		rlog.Error("Error setting cache:", err)
		return
	}
	rlog.Info("Cache set successfully")

	// Publish event to notify subscribers
	eventmanager.EventBus.Publish(eventmanager.ResourceEvent{
		Type:     eventmanager.EventAdd,
		Resource: unstructuredObject,
	})

	rlog.Info("Published add event for resource",
		rlog.String("name", unstructuredObject.GetName()),
		rlog.String("kind", unstructuredObject.GetKind()))
}

func (handler) DeleteResource(obj any) {
	unstructuredObject := obj.(*unstructured.Unstructured)
	rlog.Info(fmt.Sprintf("Delete Resource, name: %s, kind: %s", unstructuredObject.GetName(), unstructuredObject.GetKind()))

	// Determine cache key based on resource kind
	cacheKey := string(unstructuredObject.GetUID())
	if unstructuredObject.GetKind() == "ConfigMap" {
		// For ConfigMaps, use the name as the key
		namespace := unstructuredObject.GetNamespace()
		name := unstructuredObject.GetName()
		cacheKey = fmt.Sprintf("configmap-%s-%s", namespace, name)
		rlog.Info(fmt.Sprintf("Using ConfigMap name as cache key: %s", cacheKey))
	}

	err := cache.Cache.Delete(context.TODO(), cacheKey)
	if err != nil {
		rlog.Error("Error deleting cache:", err)
		return
	}
	rlog.Info("Cache deleted successfully")

	// Publish event to notify subscribers
	eventmanager.EventBus.Publish(eventmanager.ResourceEvent{
		Type:     eventmanager.EventDelete,
		Resource: unstructuredObject,
	})
}

func (handler) UpdateResource(_ any, obj any) {
	unstructuredObject := obj.(*unstructured.Unstructured)
	rlog.Info(fmt.Sprintf("UpdateResource called - name: %s, kind: %s, namespace: %s",
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
		rlog.Info(fmt.Sprintf("Using ConfigMap name as cache key: %s", cacheKey))
	}

	err := cache.Cache.Set(context.TODO(), cacheKey, unstructuredObject.Object)
	if err != nil {
		rlog.Error("Error updating cache:", err)
		return
	}
	rlog.Info("Cache updated successfully")

	// Additional logging for ConfigMap updates
	if unstructuredObject.GetKind() == "ConfigMap" {
		rlog.Info("ConfigMap updated in cache",
			rlog.String("name", unstructuredObject.GetName()),
			rlog.String("namespace", unstructuredObject.GetNamespace()),
			rlog.String("cacheKey", cacheKey))
	}

	// Publish event to notify subscribers
	eventmanager.EventBus.Publish(eventmanager.ResourceEvent{
		Type:     eventmanager.EventUpdate,
		Resource: unstructuredObject,
	})

	rlog.Info("Published update event for resource",
		rlog.String("name", unstructuredObject.GetName()),
		rlog.String("kind", unstructuredObject.GetKind()))
}
