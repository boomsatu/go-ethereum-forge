
package network

import (
	"blockchain-node/core"
	"blockchain-node/interfaces"
	"blockchain-node/logger"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

type Config struct {
	Port     int
	DataDir  string
	MaxPeers int
}

type Server struct {
	port       int
	blockchain *core.Blockchain
	peers      map[string]*Peer
	listener   net.Listener
	running    bool
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

type Peer struct {
	conn        net.Conn
	address     string
	version     string
	services    uint64
	chainID     uint64
	genesisHash [32]byte
	bestHeight  uint64
	handshaked  bool
}

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type VersionMessage struct {
	Version     string   `json:"version"`
	ChainID     uint64   `json:"chainId"`
	GenesisHash [32]byte `json:"genesisHash"`
	BestHeight  uint64   `json:"bestHeight"`
	Services    uint64   `json:"services"`
}

type HandshakeData struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

func NewServer(port int, blockchain *core.Blockchain) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		port:       port,
		blockchain: blockchain,
		peers:      make(map[string]*Peer),
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (s *Server) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to start server: %v", err)
	}

	s.listener = listener
	s.running = true

	go s.acceptConnections()

	logger.Infof("P2P server started on port %d", s.port)
	logger.Infof("Genesis hash: %x", s.blockchain.GetGenesisHash())
	logger.Infof("Chain ID: %d", s.blockchain.GetChainID())
	
	// Wait for context cancellation
	<-ctx.Done()
	return s.Stop()
}

func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	s.running = false
	s.cancel()
	
	if s.listener != nil {
		s.listener.Close()
	}

	// Close all peer connections
	for _, peer := range s.peers {
		peer.conn.Close()
	}

	logger.Info("P2P server stopped")
	return nil
}

func (s *Server) acceptConnections() {
	for s.running {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.running {
				logger.Errorf("Failed to accept connection: %v", err)
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

	logger.Infof("New peer connected: %s", peer.address)

	// Set connection timeout for handshake
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	// Perform handshake
	if !s.performHandshake(peer) {
		logger.Errorf("Handshake failed with peer %s", peer.address)
		return
	}

	// Remove read deadline after successful handshake
	conn.SetReadDeadline(time.Time{})

	s.mu.Lock()
	s.peers[peer.address] = peer
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.peers, peer.address)
		s.mu.Unlock()
		logger.Infof("Peer disconnected: %s", peer.address)
	}()

	// Handle peer messages
	decoder := json.NewDecoder(conn)
	for {
		var msg Message
		if err := decoder.Decode(&msg); err != nil {
			logger.Debugf("Peer %s disconnected: %v", peer.address, err)
			break
		}

		s.handleMessage(peer, &msg)
	}
}

func (s *Server) performHandshake(peer *Peer) bool {
	// Send version message
	currentBlock := s.blockchain.GetCurrentBlock()
	bestHeight := uint64(0)
	if currentBlock != nil {
		bestHeight = currentBlock.Header.Number
	}

	versionMsg := VersionMessage{
		Version:     "1.0.0",
		ChainID:     s.blockchain.GetChainID(),
		GenesisHash: s.blockchain.GetGenesisHash(),
		BestHeight:  bestHeight,
		Services:    1, // Full node
	}

	if err := s.sendMessage(peer, &Message{
		Type: "version",
		Data: versionMsg,
	}); err != nil {
		logger.Errorf("Failed to send version to %s: %v", peer.address, err)
		return false
	}

	// Wait for version response
	decoder := json.NewDecoder(peer.conn)
	var response Message
	if err := decoder.Decode(&response); err != nil {
		logger.Errorf("Failed to receive version from %s: %v", peer.address, err)
		return false
	}

	if response.Type != "version" {
		logger.Errorf("Expected version message from %s, got %s", peer.address, response.Type)
		return false
	}

	// Parse peer version
	versionData, err := json.Marshal(response.Data)
	if err != nil {
		logger.Errorf("Failed to parse version data from %s: %v", peer.address, err)
		return false
	}

	var peerVersion VersionMessage
	if err := json.Unmarshal(versionData, &peerVersion); err != nil {
		logger.Errorf("Failed to unmarshal version from %s: %v", peer.address, err)
		return false
	}

	// Verify compatibility
	if peerVersion.ChainID != s.blockchain.GetChainID() {
		logger.Errorf("Chain ID mismatch with %s: expected %d, got %d", 
			peer.address, s.blockchain.GetChainID(), peerVersion.ChainID)
		
		s.sendMessage(peer, &Message{
			Type: "handshake_error",
			Data: HandshakeData{
				Success: false,
				Message: fmt.Sprintf("Chain ID mismatch: expected %d, got %d", 
					s.blockchain.GetChainID(), peerVersion.ChainID),
			},
		})
		return false
	}

	if peerVersion.GenesisHash != s.blockchain.GetGenesisHash() {
		logger.Errorf("Genesis hash mismatch with %s", peer.address)
		
		s.sendMessage(peer, &Message{
			Type: "handshake_error",
			Data: HandshakeData{
				Success: false,
				Message: "Genesis hash mismatch",
			},
		})
		return false
	}

	// Update peer info
	peer.version = peerVersion.Version
	peer.chainID = peerVersion.ChainID
	peer.genesisHash = peerVersion.GenesisHash
	peer.bestHeight = peerVersion.BestHeight
	peer.services = peerVersion.Services
	peer.handshaked = true

	// Send handshake success
	if err := s.sendMessage(peer, &Message{
		Type: "handshake_success",
		Data: HandshakeData{
			Success: true,
			Message: "Handshake completed successfully",
		},
	}); err != nil {
		logger.Errorf("Failed to send handshake success to %s: %v", peer.address, err)
		return false
	}

	logger.Infof("Handshake completed with peer %s (ChainID: %d, Height: %d)", 
		peer.address, peer.chainID, peer.bestHeight)

	// Request blocks if peer has higher height
	if peer.bestHeight > bestHeight {
		s.requestBlockSync(peer, bestHeight+1, peer.bestHeight)
	}

	return true
}

func (s *Server) requestBlockSync(peer *Peer, fromHeight, toHeight uint64) {
	logger.Infof("Requesting block sync from %s (blocks %d-%d)", peer.address, fromHeight, toHeight)
	
	syncRequest := map[string]interface{}{
		"from": fromHeight,
		"to":   toHeight,
	}

	s.sendMessage(peer, &Message{
		Type: "sync_request",
		Data: syncRequest,
	})
}

func (s *Server) handleMessage(peer *Peer, msg *Message) {
	if !peer.handshaked && msg.Type != "version" && msg.Type != "handshake_error" && msg.Type != "handshake_success" {
		logger.Errorf("Received %s message from non-handshaked peer %s", msg.Type, peer.address)
		return
	}

	switch msg.Type {
	case "version":
		// Version already handled in handshake
		break
	case "handshake_error":
		s.handleHandshakeError(peer, msg)
	case "handshake_success":
		s.handleHandshakeSuccess(peer, msg)
	case "sync_request":
		s.handleSyncRequest(peer, msg)
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
		logger.Debugf("Unknown message type from %s: %s", peer.address, msg.Type)
	}
}

func (s *Server) handleHandshakeError(peer *Peer, msg *Message) {
	handshakeData, _ := msg.Data.(map[string]interface{})
	message := handshakeData["message"].(string)
	logger.Errorf("Handshake error from %s: %s", peer.address, message)
}

func (s *Server) handleHandshakeSuccess(peer *Peer, msg *Message) {
	logger.Infof("Handshake success with %s", peer.address)
}

func (s *Server) handleSyncRequest(peer *Peer, msg *Message) {
	syncData, _ := msg.Data.(map[string]interface{})
	from := uint64(syncData["from"].(float64))
	to := uint64(syncData["to"].(float64))

	logger.Infof("Sync request from %s for blocks %d-%d", peer.address, from, to)

	// Send blocks
	for i := from; i <= to; i++ {
		if block := s.blockchain.GetBlockByNumber(i); block != nil {
			s.sendMessage(peer, &Message{
				Type: "block",
				Data: block,
			})
		}
	}
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
			inv = append(inv, fmt.Sprintf("%x", block.Header.Hash))
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
		hashStr := item.(string)
		var hash [32]byte
		// Convert hex string to hash
		for i := 0; i < 32 && i*2 < len(hashStr); i++ {
			fmt.Sscanf(hashStr[i*2:i*2+2], "%02x", &hash[i])
		}
		
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
		hashStr := item.(string)
		var hash [32]byte
		// Convert hex string to hash
		for i := 0; i < 32 && i*2 < len(hashStr); i++ {
			fmt.Sscanf(hashStr[i*2:i*2+2], "%02x", &hash[i])
		}
		
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
		logger.Errorf("Failed to decode block from %s: %v", peer.address, err)
		return
	}

	// Add block to blockchain
	if err := s.blockchain.AddBlock(&block); err != nil {
		logger.Debugf("Failed to add block from %s: %v", peer.address, err)
		return
	}

	logger.Infof("Added block %d from peer %s", block.Header.Number, peer.address)
}

func (s *Server) handleTransaction(peer *Peer, msg *Message) {
	// Handle incoming transaction
	txData, _ := json.Marshal(msg.Data)
	var tx core.Transaction
	if err := json.Unmarshal(txData, &tx); err != nil {
		logger.Errorf("Failed to decode transaction from %s: %v", peer.address, err)
		return
	}

	// Add transaction to mempool
	if err := s.blockchain.AddTransaction(&tx); err != nil {
		logger.Debugf("Failed to add transaction from %s: %v", peer.address, err)
		return
	}

	logger.Debugf("Added transaction %x from peer %s", tx.Hash, peer.address)
}

func (s *Server) sendMessage(peer *Peer, msg *Message) error {
	encoder := json.NewEncoder(peer.conn)
	if err := encoder.Encode(msg); err != nil {
		return fmt.Errorf("failed to send message to %s: %v", peer.address, err)
	}
	return nil
}

func (s *Server) BroadcastTransaction(tx *core.Transaction) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	msg := &Message{
		Type: "tx",
		Data: tx,
	}

	for _, peer := range s.peers {
		if peer.handshaked {
			s.sendMessage(peer, msg)
		}
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
		if peer.handshaked {
			s.sendMessage(peer, msg)
		}
	}
}

func (s *Server) GetPeerCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	count := 0
	for _, peer := range s.peers {
		if peer.handshaked {
			count++
		}
	}
	return count
}

func (s *Server) GetConnectionCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.peers)
}
