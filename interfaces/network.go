
package interfaces

// Peer represents a network peer
type Peer interface {
	GetID() string
	GetAddress() string
	Send(message interface{}) error
	Close() error
}

// NetworkMessage represents a network message
type NetworkMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// BlockchainNetwork represents the blockchain network interface
type BlockchainNetwork interface {
	BroadcastBlock(block interface{}) error
	BroadcastTransaction(tx interface{}) error
	GetPeers() []Peer
	GetPeerCount() int
}
