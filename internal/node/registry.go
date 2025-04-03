package node

// NodeTypeRegistry provides an interface for registering node types
type NodeTypeRegistry interface {
	// RegisterNodeType registers a node factory with the registry
	RegisterNodeType(typeID string, factory NodeFactory)
}
