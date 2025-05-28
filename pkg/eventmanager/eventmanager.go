package eventmanager

import (
	"sync"

	"github.com/NorskHelsenett/ror/pkg/rlog"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// EventType defines the type of resource event
type EventType string

// Event types
const (
	EventAdd    EventType = "ADD"
	EventUpdate EventType = "UPDATE"
	EventDelete EventType = "DELETE"
)

// ResourceEvent represents a resource change event
type ResourceEvent struct {
	Type     EventType
	Resource *unstructured.Unstructured
}

// EventHandler is a function that handles resource events
type EventHandler func(event ResourceEvent)

// EventManager manages event subscriptions and notifications
type EventManager struct {
	handlers       map[string][]EventHandler
	mutex          sync.RWMutex
	globalHandlers []EventHandler
}

// NewEventManager creates a new event manager
func NewEventManager() *EventManager {
	return &EventManager{
		handlers:       make(map[string][]EventHandler),
		globalHandlers: make([]EventHandler, 0),
	}
}

// Subscribe registers a handler for events of a specific resource kind
func (em *EventManager) Subscribe(resourceKind string, eventHandler EventHandler) {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	if _, exists := em.handlers[resourceKind]; !exists {
		em.handlers[resourceKind] = make([]EventHandler, 0)
	}
	em.handlers[resourceKind] = append(em.handlers[resourceKind], eventHandler)
	rlog.Info("Subscribed handler for resource kind: " + resourceKind)
}

// SubscribeAll registers a handler for all resource events
func (em *EventManager) SubscribeAll(eventHandler EventHandler) {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	em.globalHandlers = append(em.globalHandlers, eventHandler)
	rlog.Info("Subscribed handler for all resource events")
}

// Publish notifies all registered handlers of a resource event
func (em *EventManager) Publish(event ResourceEvent) {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	resourceKind := event.Resource.GetKind()

	// Notify specific handlers for this resource kind
	if handlers, exists := em.handlers[resourceKind]; exists {
		for _, handler := range handlers {
			go func(h EventHandler) {
				defer func() {
					if r := recover(); r != nil {
						rlog.Error("Panic in event handler", nil)
					}
				}()
				h(event)
			}(handler)
		}
	}

	// Notify global handlers
	for _, handler := range em.globalHandlers {
		go func(h EventHandler) {
			defer func() {
				if r := recover(); r != nil {
					rlog.Error("Panic in global event handler", nil)
				}
			}()
			h(event)
		}(handler)
	}
}

// Global event manager instance
var EventBus *EventManager

func init() {
	EventBus = NewEventManager()
}
