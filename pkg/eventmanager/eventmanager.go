package eventmanager

import (
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/vitistack/common/pkg/loggers/vlog"
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
	vlog.Info("Subscribed handler for resource kind: " + resourceKind)
}

// SubscribeAll registers a handler for all resource events
func (em *EventManager) SubscribeAll(eventHandler EventHandler) {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	em.globalHandlers = append(em.globalHandlers, eventHandler)
	vlog.Info("Subscribed handler for all resource events")
}

// Publish notifies all registered handlers of a resource event.
//
// Handlers are invoked synchronously. They must NOT be dispatched on a fresh
// goroutine per event: informers replay every existing object as an ADD event
// on startup, so a goroutine-per-event model spawns hundreds of concurrent
// handlers at once (e.g. one per Machine), each allocating its own working set.
// On large clusters that overruns the memory limit and the pod is OOMKilled.
// Synchronous dispatch bounds concurrency to one handler per informer goroutine.
func (em *EventManager) Publish(event ResourceEvent) {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	if event.Resource == nil {
		vlog.Error("Cannot publish event with nil resource", nil)
		return
	}

	resourceKind := event.Resource.GetKind()

	// Notify specific handlers for this resource kind
	if handlers, exists := em.handlers[resourceKind]; exists {
		for _, handler := range handlers {
			safeInvoke(handler, event, resourceKind, "event handler")
		}
	}

	// Notify global handlers
	for _, handler := range em.globalHandlers {
		safeInvoke(handler, event, resourceKind, "global event handler")
	}
}

// safeInvoke runs a single handler synchronously, recovering from panics so one
// faulty handler cannot crash the process or stop sibling handlers from running.
func safeInvoke(h EventHandler, event ResourceEvent, resourceKind, label string) {
	defer func() {
		if r := recover(); r != nil {
			vlog.Error("Panic in "+label,
				fmt.Errorf("%v", r),
				"kind", resourceKind,
				"stack", string(debug.Stack()))
		}
	}()
	h(event)
}

// Global event manager instance
var EventBus *EventManager

func init() {
	EventBus = NewEventManager()
}
