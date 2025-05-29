
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "blockchain-node",
	Short: "A blockchain node with EVM support",
	Long: `A complete blockchain node implementation with Ethereum Virtual Machine support,
P2P networking, JSON-RPC API, and wallet functionality.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.blockchain-node.yaml)")
	rootCmd.PersistentFlags().String("datadir", "./data", "Data directory for blockchain data")
	rootCmd.PersistentFlags().Int("port", 8080, "P2P port")
	rootCmd.PersistentFlags().Int("rpcport", 8545, "JSON-RPC port")
	rootCmd.PersistentFlags().String("rpcaddr", "127.0.0.1", "JSON-RPC address")

	viper.BindPFlag("datadir", rootCmd.PersistentFlags().Lookup("datadir"))
	viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("rpcport", rootCmd.PersistentFlags().Lookup("rpcport"))
	viper.BindPFlag("rpcaddr", rootCmd.PersistentFlags().Lookup("rpcaddr"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".blockchain-node")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
