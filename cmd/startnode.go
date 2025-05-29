
package cmd

import (
	"blockchain-node/config"
	"blockchain-node/core"
	"blockchain-node/health"
	"blockchain-node/logger"
	"blockchain-node/metrics"
	"blockchain-node/network"
	"blockchain-node/rpc"
	"blockchain-node/security"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var startNodeCmd = &cobra.Command{
	Use:   "startnode",
	Short: "Start the blockchain node",
	Long:  `Start the blockchain node with P2P networking, RPC server, and optional mining.`,
	RunE:  runStartNode,
}

func init() {
	rootCmd.AddCommand(startNodeCmd)
	
	startNodeCmd.Flags().Bool("mining", false, "Enable mining")
	startNodeCmd.Flags().String("miner", "", "Miner address for block rewards")
	startNodeCmd.Flags().Bool("enable-metrics", true, "Enable metrics collection")
	startNodeCmd.Flags().Bool("enable-health", true, "Enable health check endpoints")
}

func runStartNode(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.LoadConfig("")
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}
	
	// Set logging level based on config
	logger.SetLevel(logger.LogLevel(cfg.GetLogLevel()))
	
	logger.Info("Starting blockchain node...")
	logger.Infof("Configuration loaded: DataDir=%s, Port=%d, RPCPort=%d", cfg.DataDir, cfg.Port, cfg.RPCPort)
	
	// Initialize security manager
	securityManager := security.NewSecurityManager()
	
	// Initialize blockchain
	blockchainConfig := &core.Config{
		DataDir:       cfg.DataDir,
		ChainID:       cfg.ChainID,
		BlockGasLimit: cfg.BlockGasLimit,
	}
	
	blockchain, err := core.NewBlockchain(blockchainConfig)
	if err != nil {
		logger.Fatalf("Failed to initialize blockchain: %v", err)
		return err
	}
	defer func() {
		if err := blockchain.Close(); err != nil {
			logger.Errorf("Failed to close blockchain: %v", err)
		}
	}()
	
	// Initialize health checker
	var healthChecker *health.HealthChecker
	if cfg.EnableMetrics {
		healthChecker = health.NewHealthChecker(blockchain, blockchain.GetDatabase())
	}
	
	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	var wg sync.WaitGroup
	
	// Start P2P server
	p2pServer := network.NewServer(cfg.Port, blockchain)
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Infof("Starting P2P server on port %d", cfg.Port)
		if err := p2pServer.Start(ctx); err != nil {
			logger.Errorf("P2P server error: %v", err)
		}
	}()
	
	// Start RPC server
	rpcServer := rpc.NewServer(blockchain, securityManager)
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Infof("Starting RPC server on %s:%d", cfg.RPCAddr, cfg.RPCPort)
		if err := rpcServer.Start(fmt.Sprintf("%s:%d", cfg.RPCAddr, cfg.RPCPort)); err != nil {
			logger.Errorf("RPC server error: %v", err)
		}
	}()
	
	// Start health check server if enabled
	if cfg.EnableMetrics && healthChecker != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			healthPort := cfg.RPCPort + 1000 // Health port is RPC port + 1000
			
			mux := http.NewServeMux()
			mux.HandleFunc("/health", healthChecker.HealthHandler)
			mux.HandleFunc("/ready", healthChecker.ReadinessHandler)
			mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
				metricsData := metrics.GetMetrics().ToMap()
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				
				// Simple JSON response
				fmt.Fprintf(w, `{`)
				first := true
				for key, value := range metricsData {
					if !first {
						fmt.Fprintf(w, ",")
					}
					fmt.Fprintf(w, `"%s":%v`, key, value)
					first = false
				}
				fmt.Fprintf(w, `}`)
			})
			
			server := &http.Server{
				Addr:    fmt.Sprintf(":%d", healthPort),
				Handler: mux,
			}
			
			logger.Infof("Starting health check server on port %d", healthPort)
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Errorf("Health server error: %v", err)
			}
		}()
	}
	
	// Start miner if enabled
	mining, _ := cmd.Flags().GetBool("mining")
	minerAddr, _ := cmd.Flags().GetString("miner")
	
	if mining || cfg.Mining {
		if minerAddr == "" {
			minerAddr = cfg.Miner
		}
		if minerAddr == "" {
			logger.Warning("Mining enabled but no miner address specified")
		} else {
			miner := core.NewMiner(blockchain, minerAddr)
			wg.Add(1)
			go func() {
				defer wg.Done()
				logger.Infof("Starting miner with address: %s", minerAddr)
				miner.Start()
			}()
		}
	}
	
	// Start metrics collection goroutine
	if cfg.EnableMetrics {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ticker := time.NewTicker(cfg.HealthCheckInterval)
			defer ticker.Stop()
			
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					// Update system metrics
					memUsed, memSys := getMemoryUsage()
					metrics.GetMetrics().SetMemoryUsage(memUsed)
					
					// Update peer count (placeholder)
					metrics.GetMetrics().SetPeerCount(uint32(p2pServer.GetPeerCount()))
					
					// Update connection count
					metrics.GetMetrics().SetConnectionCount(uint32(p2pServer.GetConnectionCount()))
				}
			}
		}()
	}
	
	logger.Info("Blockchain node started successfully")
	logger.Info("Press Ctrl+C to stop the node")
	
	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	
	<-sigCh
	logger.Info("Received shutdown signal, stopping node...")
	
	// Cancel context to stop all goroutines
	cancel()
	
	// Wait for all goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		logger.Info("All services stopped gracefully")
	case <-time.After(30 * time.Second):
		logger.Warning("Timeout waiting for services to stop")
	}
	
	logger.Info("Blockchain node stopped")
	return nil
}

func getMemoryUsage() (uint64, uint64) {
	// Placeholder implementation
	return 100 * 1024 * 1024, 200 * 1024 * 1024 // 100MB used, 200MB system
}
