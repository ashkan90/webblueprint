package blueprint

// BlueprintNode represents a node in a blueprint
type BlueprintNode struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Position   Position               `json:"position"`
	Properties []NodeProperty         `json:"properties"`
	Data       map[string]interface{} `json:"data,omitempty"`
}

type BlueprintNodeType struct {
	Inputs     []NodePin      `json:"inputs"`
	Outputs    []NodePin      `json:"outputs"`
	Properties []NodeProperty `json:"properties"`
}

type NodePin struct {
	ID          string       `json:"id,omitempty"`          // Unique identifier
	Name        string       `json:"name,omitempty"`        // Human-readable name
	Description string       `json:"description,omitempty"` // Description of what the pin does
	Type        *NodePinType `json:"type,omitempty"`        // Type of data for this pin
	Optional    bool         `json:"optional,omitempty"`    // Whether this pin is required
	Default     interface{}  `json:"default,omitempty"`     // Default value if not connected
}

type NodePinType struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// Position represents the node position on the canvas
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// NodeProperty represents a property of a node
type NodeProperty struct {
	Name        string       `json:"name"`
	DisplayName string       `json:"displayName"`
	Description string       `json:"description"`
	Value       interface{}  `json:"value"`
	Type        *NodePinType `json:"type"`
}

// Connection represents a connection between nodes
type Connection struct {
	ID             string         `json:"id"`
	SourceNodeID   string         `json:"sourceNodeId"`
	SourcePinID    string         `json:"sourcePinId"`
	TargetNodeID   string         `json:"targetNodeId"`
	TargetPinID    string         `json:"targetPinId"`
	ConnectionType string         `json:"connectionType"` // "execution" or "data"
	Data           map[string]any `json:"data"`
}

// Variable represents a blueprint variable
type Variable struct {
	Name  string      `json:"name"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

type Function struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	NodeType    BlueprintNodeType `json:"nodeType"`
	Nodes       []BlueprintNode   `json:"nodes"`
	Connections []Connection      `json:"connections"`
	Variables   []Variable        `json:"variables,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// Blueprint represents a complete blueprint definition
type Blueprint struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Version     string            `json:"version"`
	Nodes       []BlueprintNode   `json:"nodes"`
	Functions   []Function        `json:"functions"`
	Connections []Connection      `json:"connections"`
	Variables   []Variable        `json:"variables,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// NewBlueprint creates a new empty blueprint
func NewBlueprint(id, name, version string) *Blueprint {
	return &Blueprint{
		ID:          id,
		Name:        name,
		Version:     version,
		Nodes:       make([]BlueprintNode, 0),
		Connections: make([]Connection, 0),
		Variables:   make([]Variable, 0),
		Metadata:    make(map[string]string),
	}
}

// AddNode adds a node to the blueprint
func (b *Blueprint) AddNode(node BlueprintNode) {
	b.Nodes = append(b.Nodes, node)
}

// AddConnection adds a connection to the blueprint
func (b *Blueprint) AddConnection(conn Connection) {
	b.Connections = append(b.Connections, conn)
}

// AddVariable adds a variable to the blueprint
func (b *Blueprint) AddVariable(variable Variable) {
	b.Variables = append(b.Variables, variable)
}

// FindNode finds a node by ID
func (b *Blueprint) FindNode(id string) *BlueprintNode {
	for i := range b.Nodes {
		if b.Nodes[i].ID == id {
			return &b.Nodes[i]
		}
	}
	return nil
}

// FindConnection finds a connection by ID
func (b *Blueprint) FindConnection(id string) *Connection {
	for i := range b.Connections {
		if b.Connections[i].ID == id {
			return &b.Connections[i]
		}
	}
	return nil
}

// FindVariable finds a variable by name
func (b *Blueprint) FindVariable(name string) *Variable {
	for i := range b.Variables {
		if b.Variables[i].Name == name {
			return &b.Variables[i]
		}
	}
	return nil
}

// GetNodeConnections finds all connections for a node
func (b *Blueprint) GetNodeConnections(nodeID string) []Connection {
	connections := make([]Connection, 0)

	for _, conn := range b.Connections {
		if conn.SourceNodeID == nodeID || conn.TargetNodeID == nodeID {
			connections = append(connections, conn)
		}
	}

	return connections
}

// GetNodeInputConnections finds all input connections for a node
func (b *Blueprint) GetNodeInputConnections(nodeID string) []Connection {
	connections := make([]Connection, 0)

	for _, conn := range b.Connections {
		if conn.TargetNodeID == nodeID {
			connections = append(connections, conn)
		}
	}

	return connections
}

// GetNodeOutputConnections finds all output connections for a node
func (b *Blueprint) GetNodeOutputConnections(nodeID string) []Connection {
	connections := make([]Connection, 0)

	for _, conn := range b.Connections {
		if conn.SourceNodeID == nodeID {
			connections = append(connections, conn)
		}
	}

	return connections
}

// RemoveNode removes a node and all its connections
func (b *Blueprint) RemoveNode(nodeID string) {
	// Remove connections first
	newConnections := make([]Connection, 0)
	for _, conn := range b.Connections {
		if conn.SourceNodeID != nodeID && conn.TargetNodeID != nodeID {
			newConnections = append(newConnections, conn)
		}
	}
	b.Connections = newConnections

	// Remove node
	newNodes := make([]BlueprintNode, 0)
	for _, node := range b.Nodes {
		if node.ID != nodeID {
			newNodes = append(newNodes, node)
		}
	}
	b.Nodes = newNodes
}

// RemoveConnection removes a connection
func (b *Blueprint) RemoveConnection(connID string) {
	newConnections := make([]Connection, 0)
	for _, conn := range b.Connections {
		if conn.ID != connID {
			newConnections = append(newConnections, conn)
		}
	}
	b.Connections = newConnections
}

// FindEntryPoints finds nodes that should be triggered first
// (nodes with execution outputs but no execution inputs)
func (b *Blueprint) FindEntryPoints() []string {
	entryPoints := make([]string, 0)

	// Create maps to track nodes with execution inputs and outputs
	execInputs := make(map[string]bool)
	execOutputs := make(map[string]bool)

	// Find all nodes with execution connections
	for _, conn := range b.Connections {
		if conn.ConnectionType == "execution" {
			execOutputs[conn.SourceNodeID] = true
			execInputs[conn.TargetNodeID] = true
		}
	}

	// Find nodes with execution outputs but no execution inputs
	for _, node := range b.Nodes {
		if execOutputs[node.ID] && !execInputs[node.ID] {
			entryPoints = append(entryPoints, node.ID)
		}
	}

	// Also include special entry point nodes like DOM events
	for _, node := range b.Nodes {
		if node.Type == "dom-event" && !contains(entryPoints, node.ID) {
			entryPoints = append(entryPoints, node.ID)
		}
	}

	return entryPoints
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
