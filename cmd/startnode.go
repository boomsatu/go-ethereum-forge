
package cmd

import (
	"blockchain-node/core"
	"blockchain-node/network"
	"blockchain-node/rpc"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var startnodeCmd = &cobra.Command{
	Use:   "startnode",
	Short: "Start the blockchain node",
	Long:  `Start the blockchain node with P2P networking and JSON-RPC server`,
	Run: func(cmd *cobra.Command, args []string) {
		startNode()
	},
}

func init() {
	rootCmd.AddCommand(startnodeCmd)
	startnodeCmd.Flags().Bool("mining", false, "Enable mining")
	startnodeCmd.Flags().String("miner", "", "Miner address")
	viper.BindPFlag("mining", startnodeCmd.Flags().Lookup("mining"))
	viper.BindPFlag("miner", startnodeCmd.Flags().Lookup("miner"))
}

func startNode() {
	// Initialize blockchain
	config := &core.Config{
		DataDir:       viper.GetString("datadir"),
		ChainID:       1337,
		BlockGasLimit: 8000000,
	}

	blockchain, err := core.NewBlockchain(config)
	if err != nil {
		fmt.Printf("Failed to initialize blockchain: %v\n", err)
		return
	}
	defer blockchain.Close()

	// Start P2P server
	p2pConfig := &network.Config{
		Port:     viper.GetInt("port"),
		DataDir:  viper.GetString("datadir"),
		MaxPeers: 50,
	}

	p2pServer := network.NewServer(p2pConfig, blockchain)
	if err := p2pServer.Start(); err != nil {
		fmt.Printf("Failed to start P2P server: %v\n", err)
		return
	}
	defer p2pServer.Stop()

	// Start JSON-RPC server
	rpcConfig := &rpc.Config{
		Host: viper.GetString("rpcaddr"),
		Port: viper.GetInt("rpcport"),
	}

	rpcServer := rpc.NewServer(rpcConfig, blockchain)
	if err := rpcServer.Start(); err != nil {
		fmt.Printf("Failed to start RPC server: %v\n", err)
		return
	}
	defer rpcServer.Stop()

	// Start mining if enabled
	if viper.GetBool("mining") {
		minerAddr := viper.GetString("miner")
		if minerAddr == "" {
			fmt.Println("Mining enabled but no miner address specified")
			return
		}
		
		miner := core.NewMiner(blockchain, minerAddr)
		go miner.Start()
		defer miner.Stop()
		
		fmt.Printf("Mining started with address: %s\n", minerAddr)
	}

	fmt.Printf("Node started successfully!\n")
	fmt.Printf("P2P server listening on port: %d\n", viper.GetInt("port"))
	fmt.Printf("JSON-RPC server listening on %s:%d\n", viper.GetString("rpcaddr"), viper.GetInt("rpcport"))

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("\nShutting down node...")
}
