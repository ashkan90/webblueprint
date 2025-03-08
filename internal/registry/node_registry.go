package registry

import (
	"sync"
	"webblueprint/internal/node"
)

// GlobalNodeRegistry provides a central point to access node factories
// This is useful for user-defined functions to access the same node types
// as the main execution engine
type GlobalNodeRegistry struct {
	factories map[string]node.NodeFactory
	mutex     sync.RWMutex
}

var (
	// instance is the singleton instance of GlobalNodeRegistry
	instance *GlobalNodeRegistry
	once     sync.Once
)

// GetInstance returns the singleton instance of GlobalNodeRegistry
func GetInstance() *GlobalNodeRegistry {
	once.Do(func() {
		instance = &GlobalNodeRegistry{
			factories: make(map[string]node.NodeFactory),
		}
	})
	return instance
}

// RegisterNodeType registers a node type with the registry
func (r *GlobalNodeRegistry) RegisterNodeType(typeID string, factory node.NodeFactory) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.factories[typeID] = factory
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
