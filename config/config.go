
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	// Node configuration
	DataDir    string `mapstructure:"datadir"`
	Port       int    `mapstructure:"port"`
	RPCPort    int    `mapstructure:"rpcport"`
	RPCAddr    string `mapstructure:"rpcaddr"`
	
	// Mining configuration
	Mining   bool   `mapstructure:"mining"`
	Miner    string `mapstructure:"miner"`
	
	// Network configuration
	MaxPeers  int      `mapstructure:"maxpeers"`
	BootNodes []string `mapstructure:"bootnode"`
	
	// Chain configuration
	ChainID        uint64 `mapstructure:"chainid"`
	BlockGasLimit  uint64 `mapstructure:"blockgaslimit"`
	
	// Database configuration
	Cache   int `mapstructure:"cache"`
	Handles int `mapstructure:"handles"`
	
	// Logging configuration
	Verbosity int `mapstructure:"verbosity"`
	
	// Security configuration
	EnableRateLimit bool          `mapstructure:"enable_rate_limit"`
	RateLimit       int           `mapstructure:"rate_limit"`
	RateLimitWindow time.Duration `mapstructure:"rate_limit_window"`
	
	// Performance configuration
	EnableCache       bool          `mapstructure:"enable_cache"`
	CacheSize         int           `mapstructure:"cache_size"`
	ConnectionTimeout time.Duration `mapstructure:"connection_timeout"`
	
	// Health check configuration
	HealthCheckInterval time.Duration `mapstructure:"health_check_interval"`
	EnableMetrics       bool          `mapstructure:"enable_metrics"`
}

var defaultConfig = Config{
	DataDir:             "./data",
	Port:                8080,
	RPCPort:             8545,
	RPCAddr:             "127.0.0.1",
	Mining:              false,
	Miner:               "",
	MaxPeers:            50,
	BootNodes:           []string{},
	ChainID:             1337,
	BlockGasLimit:       8000000,
	Cache:               256,
	Handles:             256,
	Verbosity:           3,
	EnableRateLimit:     true,
	RateLimit:           100,
	RateLimitWindow:     time.Minute,
	EnableCache:         true,
	CacheSize:           1000,
	ConnectionTimeout:   30 * time.Second,
	HealthCheckInterval: 30 * time.Second,
	EnableMetrics:       true,
}

func LoadConfig(configPath string) (*Config, error) {
	config := defaultConfig
	
	if configPath != "" {
		// Set config file path
		viper.SetConfigFile(configPath)
	} else {
		// Search for config in working directory and home directory
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.blockchain-node")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}
	
	// Set environment variable prefix
	viper.SetEnvPrefix("BLOCKCHAIN")
	viper.AutomaticEnv()
	
	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %v", err)
		}
		// Config file not found, use defaults
	}
	
	// Unmarshal config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}
	
	// Validate and create directories
	if err := validateAndCreateDirs(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %v", err)
	}
	
	return &config, nil
}

func validateAndCreateDirs(config *Config) error {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(config.DataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %v", err)
	}
	
	// Create chaindata subdirectory
	chaindataDir := filepath.Join(config.DataDir, "chaindata")
	if err := os.MkdirAll(chaindataDir, 0755); err != nil {
		return fmt.Errorf("failed to create chaindata directory: %v", err)
	}
	
	// Create wallet directory
	walletDir := filepath.Join(config.DataDir, "wallet")
	if err := os.MkdirAll(walletDir, 0755); err != nil {
		return fmt.Errorf("failed to create wallet directory: %v", err)
	}
	
	// Validate ports
	if config.Port <= 0 || config.Port > 65535 {
		return fmt.Errorf("invalid port: %d", config.Port)
	}
	
	if config.RPCPort <= 0 || config.RPCPort > 65535 {
		return fmt.Errorf("invalid RPC port: %d", config.RPCPort)
	}
	
	if config.Port == config.RPCPort {
		return fmt.Errorf("port and RPC port cannot be the same")
	}
	
	// Validate other parameters
	if config.MaxPeers <= 0 {
		config.MaxPeers = 50
	}
	
	if config.BlockGasLimit == 0 {
		config.BlockGasLimit = 8000000
	}
	
	if config.Cache <= 0 {
		config.Cache = 256
	}
	
	if config.Handles <= 0 {
		config.Handles = 256
	}
	
	return nil
}

func (c *Config) GetLogLevel() int {
	switch c.Verbosity {
	case 0:
		return 5 // Fatal
	case 1:
		return 4 // Error
	case 2:
		return 3 // Warning
	case 3:
		return 2 // Info
	case 4:
		return 1 // Debug
	default:
		return 2 // Info
	}
}

func (c *Config) IsMainnet() bool {
	return c.ChainID == 1
}

func (c *Config) IsTestnet() bool {
	return c.ChainID == 3 || c.ChainID == 4 || c.ChainID == 5
}

func (c *Config) GetDataSubDir(subdir string) string {
	return filepath.Join(c.DataDir, subdir)
}
