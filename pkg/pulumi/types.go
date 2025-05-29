package pulumi

// Node represents a single node in the network
type Node struct {
	Host          string
	IP            string
	NetworkNodeID string
}

// Nodes represents a collection of validators and sentry nodes
type Nodes struct {
	Validators []Node
	Sentries   []Node
}

// NodeConfig represents the configuration for a node type
type NodeConfig struct {
	NodePoolName       string   `json:"nodePoolName"`
	NodePoolSize       int      `json:"nodePoolSize"`
	NodeMachineType    string   `json:"nodeMachineType"`
	NodeImage          string   `json:"nodeImage"`
	NodeStartupScript  string   `json:"nodeStartupScript"`
	NodeDiskSizeGB     int      `json:"nodeDiskSizeGB"`
	NodeDiskType       string   `json:"nodeDiskType"`
	NodesAutoDelete    bool     `json:"nodesAutoDelete"`
	NodePoolLabel      []string `json:"nodePoolLabel"`
	NodeRegions        []string `json:"nodeRegions"`
	NodeZonesPerRegion int      `json:"nodeZonesPerRegion"`
	NodePorts          []string `json:"nodePorts"`
}

// Peers represents a list of peer addresses
type Peers []string

type NodePoolConfig struct {
	NodeRegions        []string
	NodeZonesPerRegion int
	NodeMachineType    string
	NodeImage          string
	NodeDiskSizeGB     int
	NodeDiskType       string
	NodesAutoDelete    bool
	NodePoolSize       int
}

const (
	VALIDATORS_TYPE = "validators"
	SENTRIES_TYPE   = "sentry-nodes"
)
