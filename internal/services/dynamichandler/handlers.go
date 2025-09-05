package dynamichandler

import (
	"context"
	"fmt"

	"github.com/vitistack/common/pkg/loggers/vlog"
	"github.com/vitistack/vitistack-operator/internal/cache"
	"github.com/vitistack/vitistack-operator/internal/clients/dynamicclienthandler"
	"github.com/vitistack/vitistack-operator/pkg/eventmanager"
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

	err := cache.Cache.Set(context.TODO(), cacheKey, unstructuredObject.Object)
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
	unstructuredObject := obj.(*unstructured.Unstructured)
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

	err := cache.Cache.Delete(context.TODO(), cacheKey)
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
	unstructuredObject := obj.(*unstructured.Unstructured)
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

	err := cache.Cache.Set(context.TODO(), cacheKey, unstructuredObject.Object)
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
