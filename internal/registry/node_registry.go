package registry

import (
	"fmt"
	"sync"
	"webblueprint/internal/node"
)

// GlobalNodeRegistry provides a central point to access node factories
// This is useful for user-defined functions to access the same node types
// as the main execution engine
type GlobalNodeRegistry struct {
	factories map[string]node.NodeFactory
	chanF     chan map[string]node.NodeFactory
	mutex     sync.RWMutex
	wg        sync.WaitGroup // Add WaitGroup to track registration goroutines
}

var (
	// instance is the singleton instance of GlobalNodeRegistry
	instance *GlobalNodeRegistry
	once     sync.Once
)

func Make(def ...map[string]node.NodeFactory) {
	once.Do(func() {
		var factories = make(map[string]node.NodeFactory)
		if len(def) == 1 {
			factories = def[0]
		}

		instance = &GlobalNodeRegistry{
			factories: factories,
			chanF:     make(chan map[string]node.NodeFactory),
			// wg is initialized with zero value
		}
	})
}

// GetInstance returns the singleton instance of GlobalNodeRegistry
func GetInstance() *GlobalNodeRegistry {
	return instance
}

// RegisterNodeType registers a node type with the registry
func (r *GlobalNodeRegistry) RegisterNodeType(typeID string, factory node.NodeFactory) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.factories[typeID] = factory
}

func (r *GlobalNodeRegistry) RegisterNodeTypeRuntime(typeID string, factory node.NodeFactory) {
	if factory == nil {
		return
	}

	r.mutex.Lock()
	r.factories[typeID] = factory
	r.mutex.Unlock()

	r.wg.Add(1) // Increment WaitGroup counter
	go func() {
		defer r.wg.Done() // Decrement counter when goroutine finishes
		// Use a select with a default case or timeout to prevent blocking indefinitely
		// if the channel is closed or never read.
		select {
		case r.chanF <- map[string]node.NodeFactory{typeID: factory}:
			// Sent successfully
		default:
			// Channel might be closed or blocked, log or handle appropriately
			// Depending on requirements, could try a timed send
			fmt.Printf("[WARN] Could not send runtime node factory update for %s to channel\n", typeID)
		}
	}()
}

// GetNodeFactory retrieves a node factory by type ID
func (r *GlobalNodeRegistry) GetNodeFactory(typeID string) (node.NodeFactory, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	factory, exists := r.factories[typeID]
	return factory, exists
}

// GetAllNodeFactories returns all registered node factories
func (r *GlobalNodeRegistry) GetAllNodeFactories() map[string]node.NodeFactory {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Create a copy to avoid concurrency issues
	factoriesCopy := make(map[string]node.NodeFactory)
	for k, v := range r.factories {
		factoriesCopy[k] = v
	}

	return factoriesCopy
}

func (r *GlobalNodeRegistry) FactoryChannel() chan map[string]node.NodeFactory {
	return r.chanF
}

func (r *GlobalNodeRegistry) Close() {
	// Wait for any pending registration goroutines to finish sending
	r.wg.Wait()

	// Now it's safe to close the channel
	close(r.chanF)

	// Drain any remaining items (optional, might not be necessary if readers stop on close)
	// for range r.chanF {
	// }
}
