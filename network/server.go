
package network

import (
	"blockchain-node/core"
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

type Config struct {
	Port     int
	DataDir  string
	MaxPeers int
}

type Server struct {
	config     *Config
	blockchain *core.Blockchain
	peers      map[string]*Peer
	listener   net.Listener
	running    bool
	mu         sync.RWMutex
}

type Peer struct {
	conn     net.Conn
	address  string
	version  string
	services uint64
}

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func NewServer(config *Config, blockchain *core.Blockchain) *Server {
	return &Server{
		config:     config,
		blockchain: blockchain,
		peers:      make(map[string]*Peer),
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.Port))
	if err != nil {
		return fmt.Errorf("failed to start server: %v", err)
	}

	s.listener = listener
	s.running = true

	go s.acceptConnections()

	fmt.Printf("P2P server started on port %d\n", s.config.Port)
	return nil
}

func (s *Server) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.running = false
	s.listener.Close()

	// Close all peer connections
	for _, peer := range s.peers {
		peer.conn.Close()
	}

	fmt.Println("P2P server stopped")
}

func (s *Server) acceptConnections() {
	for s.running {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.running {
				fmt.Printf("Failed to accept connection: %v\n", err)
			}
			continue
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	peer := &Peer{
		conn:    conn,
		address: conn.RemoteAddr().String(),
	}

	s.mu.Lock()
	s.peers[peer.address] = peer
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.peers, peer.address)
		s.mu.Unlock()
	}()

	fmt.Printf("New peer connected: %s\n", peer.address)

	// Handle peer messages
	decoder := json.NewDecoder(conn)
	for {
		var msg Message
		if err := decoder.Decode(&msg); err != nil {
			fmt.Printf("Peer %s disconnected: %v\n", peer.address, err)
			break
		}

		s.handleMessage(peer, &msg)
	}
}

func (s *Server) handleMessage(peer *Peer, msg *Message) {
	switch msg.Type {
	case "version":
		s.handleVersion(peer, msg)
	case "getblocks":
		s.handleGetBlocks(peer, msg)
	case "inv":
		s.handleInv(peer, msg)
	case "getdata":
		s.handleGetData(peer, msg)
	case "block":
		s.handleBlock(peer, msg)
	case "tx":
		s.handleTransaction(peer, msg)
	default:
		fmt.Printf("Unknown message type from %s: %s\n", peer.address, msg.Type)
	}
}

func (s *Server) handleVersion(peer *Peer, msg *Message) {
	// Handle version message
	versionData, _ := msg.Data.(map[string]interface{})
	peer.version = versionData["version"].(string)

	// Send verack
	s.sendMessage(peer, &Message{
		Type: "verack",
		Data: map[string]interface{}{
			"success": true,
		},
	})
}

func (s *Server) handleGetBlocks(peer *Peer, msg *Message) {
	// Send inventory of available blocks
	currentBlock := s.blockchain.GetCurrentBlock()
	if currentBlock == nil {
		return
	}

	inv := make([]string, 0)
	for i := uint64(0); i <= currentBlock.Header.Number; i++ {
		if block := s.blockchain.GetBlockByNumber(i); block != nil {
			inv = append(inv, block.Header.Hash.Hex())
		}
	}

	s.sendMessage(peer, &Message{
		Type: "inv",
		Data: map[string]interface{}{
			"type":  "block",
			"items": inv,
		},
	})
}

func (s *Server) handleInv(peer *Peer, msg *Message) {
	// Handle inventory message
	invData, _ := msg.Data.(map[string]interface{})
	items, _ := invData["items"].([]interface{})

	// Request data for items we don't have
	needed := make([]string, 0)
	for _, item := range items {
		hash := common.HexToHash(item.(string))
		if s.blockchain.GetBlockByHash(hash) == nil {
			needed = append(needed, item.(string))
		}
	}

	if len(needed) > 0 {
		s.sendMessage(peer, &Message{
			Type: "getdata",
			Data: map[string]interface{}{
				"items": needed,
			},
		})
	}
}

func (s *Server) handleGetData(peer *Peer, msg *Message) {
	// Send requested data
	getData, _ := msg.Data.(map[string]interface{})
	items, _ := getData["items"].([]interface{})

	for _, item := range items {
		hash := common.HexToHash(item.(string))
		if block := s.blockchain.GetBlockByHash(hash); block != nil {
			s.sendMessage(peer, &Message{
				Type: "block",
				Data: block,
			})
		}
	}
}

func (s *Server) handleBlock(peer *Peer, msg *Message) {
	// Handle incoming block
	blockData, _ := json.Marshal(msg.Data)
	var block core.Block
	if err := json.Unmarshal(blockData, &block); err != nil {
		fmt.Printf("Failed to decode block from %s: %v\n", peer.address, err)
		return
	}

	// Add block to blockchain
	if err := s.blockchain.AddBlock(&block); err != nil {
		fmt.Printf("Failed to add block from %s: %v\n", peer.address, err)
		return
	}

	fmt.Printf("Added block %d from peer %s\n", block.Header.Number, peer.address)
}

func (s *Server) handleTransaction(peer *Peer, msg *Message) {
	// Handle incoming transaction
	txData, _ := json.Marshal(msg.Data)
	var tx core.Transaction
	if err := json.Unmarshal(txData, &tx); err != nil {
		fmt.Printf("Failed to decode transaction from %s: %v\n", peer.address, err)
		return
	}

	// Add transaction to mempool
	if err := s.blockchain.AddTransaction(&tx); err != nil {
		fmt.Printf("Failed to add transaction from %s: %v\n", peer.address, err)
		return
	}

	fmt.Printf("Added transaction %s from peer %s\n", tx.Hash.Hex(), peer.address)
}

func (s *Server) sendMessage(peer *Peer, msg *Message) {
	encoder := json.NewEncoder(peer.conn)
	if err := encoder.Encode(msg); err != nil {
		fmt.Printf("Failed to send message to %s: %v\n", peer.address, err)
	}
}

func (s *Server) BroadcastTransaction(tx *core.Transaction) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	msg := &Message{
		Type: "tx",
		Data: tx,
	}

	for _, peer := range s.peers {
		s.sendMessage(peer, msg)
	}
}

func (s *Server) BroadcastBlock(block *core.Block) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	msg := &Message{
		Type: "block",
		Data: block,
	}

	for _, peer := range s.peers {
		s.sendMessage(peer, msg)
	}
}

func (s *Server) GetPeerCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.peers)
}
