
# Multi-Node Blockchain Setup Guide

## Overview

Panduan ini menjelaskan cara menjalankan beberapa node blockchain dalam satu mesin atau jaringan lokal untuk menciptakan jaringan blockchain yang terdistribusi.

## Prerequisites

- Go 1.21 atau lebih tinggi
- Git
- Port yang tersedia untuk setiap node

## Struktur Direktori Multi-Node

```
blockchain-network/
├── node1/
│   ├── blockchain-node
│   ├── config.yaml
│   ├── genesis.json
│   └── data/
├── node2/
│   ├── blockchain-node
│   ├── config.yaml
│   ├── genesis.json
│   └── data/
├── node3/
│   ├── blockchain-node
│   ├── config.yaml
│   ├── genesis.json
│   └── data/
└── scripts/
    ├── start-all.sh
    ├── stop-all.sh
    └── reset-all.sh
```

## Setup Multi-Node

### 1. Persiapan Binary dan Genesis

Pertama, compile blockchain node:

```bash
go build -o blockchain-node
```

Buat direktori untuk setiap node:

```bash
mkdir -p blockchain-network/{node1,node2,node3,scripts}
```

Copy binary ke setiap direktori node:

```bash
cp blockchain-node blockchain-network/node1/
cp blockchain-node blockchain-network/node2/
cp blockchain-node blockchain-network/node3/
```

Copy file genesis yang sama ke semua node:

```bash
cp genesis.json blockchain-network/node1/
cp genesis.json blockchain-network/node2/
cp genesis.json blockchain-network/node3/
```

### 2. Konfigurasi Node

#### Node 1 (Bootstrap Node)
File: `blockchain-network/node1/config.yaml`

```yaml
# Node 1 - Bootstrap Node
datadir: "./data"
port: 8080
rpcport: 8545
rpcaddr: "127.0.0.1"

# Mining Configuration
mining: true
miner: "0x742d35Cc6635C0532925a3b8D5c6C1C8b1c5C6C7"

# Network Configuration
maxpeers: 50
bootnode: []

# Chain Configuration
chainid: 1337
blockgaslimit: 8000000

# Database Configuration
cache: 256
handles: 256

# Logging Configuration
verbosity: 3

# Security Configuration
enable_rate_limit: true
rate_limit: 1000
rate_limit_window: "1m"

# Performance Configuration
enable_cache: true
cache_size: 10000
connection_timeout: "30s"

# Health Check Configuration
health_check_interval: "30s"
enable_metrics: true
```

#### Node 2
File: `blockchain-network/node2/config.yaml`

```yaml
# Node 2
datadir: "./data"
port: 8081
rpcport: 8546
rpcaddr: "127.0.0.1"

# Mining Configuration
mining: true
miner: "0x8ba1f109551bD432803012645Hac136c46C01C23"

# Network Configuration
maxpeers: 50
bootnode: ["127.0.0.1:8080"]

# Chain Configuration
chainid: 1337
blockgaslimit: 8000000

# Database Configuration
cache: 256
handles: 256

# Logging Configuration
verbosity: 3

# Security Configuration
enable_rate_limit: true
rate_limit: 1000
rate_limit_window: "1m"

# Performance Configuration
enable_cache: true
cache_size: 10000
connection_timeout: "30s"

# Health Check Configuration
health_check_interval: "30s"
enable_metrics: true
```

#### Node 3
File: `blockchain-network/node3/config.yaml`

```yaml
# Node 3
datadir: "./data"
port: 8082
rpcport: 8547
rpcaddr: "127.0.0.1"

# Mining Configuration
mining: false
miner: ""

# Network Configuration
maxpeers: 50
bootnode: ["127.0.0.1:8080", "127.0.0.1:8081"]

# Chain Configuration
chainid: 1337
blockgaslimit: 8000000

# Database Configuration
cache: 256
handles: 256

# Logging Configuration
verbosity: 3

# Security Configuration
enable_rate_limit: true
rate_limit: 1000
rate_limit_window: "1m"

# Performance Configuration
enable_cache: true
cache_size: 10000
connection_timeout: "30s"

# Health Check Configuration
health_check_interval: "30s"
enable_metrics: true
```

### 3. Script Manajemen

#### Start All Nodes
File: `blockchain-network/scripts/start-all.sh`

```bash
#!/bin/bash

echo "Starting Blockchain Network..."

# Start Node 1 (Bootstrap)
echo "Starting Node 1 (Bootstrap Node)..."
cd ../node1
./blockchain-node startnode --config config.yaml &
NODE1_PID=$!
echo "Node 1 PID: $NODE1_PID"

# Wait for Node 1 to start
sleep 5

# Start Node 2
echo "Starting Node 2..."
cd ../node2
./blockchain-node startnode --config config.yaml &
NODE2_PID=$!
echo "Node 2 PID: $NODE2_PID"

# Wait for Node 2 to start
sleep 3

# Start Node 3
echo "Starting Node 3..."
cd ../node3
./blockchain-node startnode --config config.yaml &
NODE3_PID=$!
echo "Node 3 PID: $NODE3_PID"

# Save PIDs for stopping later
cd ../scripts
echo $NODE1_PID > node1.pid
echo $NODE2_PID > node2.pid
echo $NODE3_PID > node3.pid

echo "All nodes started successfully!"
echo "Node 1: RPC on 8545, P2P on 8080"
echo "Node 2: RPC on 8546, P2P on 8081"
echo "Node 3: RPC on 8547, P2P on 8082"
echo ""
echo "To stop all nodes, run: ./stop-all.sh"
```

#### Stop All Nodes
File: `blockchain-network/scripts/stop-all.sh`

```bash
#!/bin/bash

echo "Stopping Blockchain Network..."

# Stop Node 1
if [ -f node1.pid ]; then
    NODE1_PID=$(cat node1.pid)
    echo "Stopping Node 1 (PID: $NODE1_PID)..."
    kill $NODE1_PID 2>/dev/null
    rm node1.pid
fi

# Stop Node 2
if [ -f node2.pid ]; then
    NODE2_PID=$(cat node2.pid)
    echo "Stopping Node 2 (PID: $NODE2_PID)..."
    kill $NODE2_PID 2>/dev/null
    rm node2.pid
fi

# Stop Node 3
if [ -f node3.pid ]; then
    NODE3_PID=$(cat node3.pid)
    echo "Stopping Node 3 (PID: $NODE3_PID)..."
    kill $NODE3_PID 2>/dev/null
    rm node3.pid
fi

echo "All nodes stopped."
```

#### Reset All Nodes
File: `blockchain-network/scripts/reset-all.sh`

```bash
#!/bin/bash

echo "Resetting Blockchain Network..."

# Stop all nodes first
./stop-all.sh

# Remove data directories
echo "Removing blockchain data..."
rm -rf ../node1/data
rm -rf ../node2/data
rm -rf ../node3/data

echo "Blockchain network reset completed."
echo "Run ./start-all.sh to start fresh network."
```

### 4. Menjalankan Network

```bash
# Masuk ke direktori scripts
cd blockchain-network/scripts

# Berikan permission execute
chmod +x *.sh

# Start semua nodes
./start-all.sh

# Untuk monitor logs
tail -f ../node1/data/logs/*.log
tail -f ../node2/data/logs/*.log
tail -f ../node3/data/logs/*.log

# Stop semua nodes
./stop-all.sh

# Reset dan mulai ulang
./reset-all.sh
./start-all.sh
```

## Testing Konektivitas

### 1. Cek Status Node

```bash
# Node 1
curl http://localhost:8545/api/admin/status

# Node 2
curl http://localhost:8546/api/admin/status

# Node 3
curl http://localhost:8547/api/admin/status
```

### 2. Cek Network Stats

```bash
# Cek peer count Node 1
curl http://localhost:8545/api/network/stats

# Cek peer count Node 2
curl http://localhost:8546/api/network/stats

# Cek peer count Node 3
curl http://localhost:8547/api/network/stats
```

### 3. Test Transaction Sync

```bash
# Kirim transaction dari Node 1
curl -X POST http://localhost:8545/api/wallet/send \
  -H "Content-Type: application/json" \
  -d '{
    "from": "0x742d35Cc6635C0532925a3b8D5c6C1C8b1c5C6C7",
    "to": "0x8ba1f109551bD432803012645Hac136c46C01C23",
    "value": "1000000000000000000",
    "privateKey": "YOUR_PRIVATE_KEY"
  }'

# Cek balance di node lain
curl "http://localhost:8546/api/wallet/balance?address=0x8ba1f109551bD432803012645Hac136c46C01C23"
curl "http://localhost:8547/api/wallet/balance?address=0x8ba1f109551bD432803012645Hac136c46C01C23"
```

## Network Monitoring

### 1. Health Checks

```bash
# Health check endpoints
curl http://localhost:9545/health  # Node 1 health
curl http://localhost:9546/health  # Node 2 health
curl http://localhost:9547/health  # Node 3 health
```

### 2. Metrics

```bash
# Metrics endpoints
curl http://localhost:9545/metrics  # Node 1 metrics
curl http://localhost:9546/metrics  # Node 2 metrics
curl http://localhost:9547/metrics  # Node 3 metrics
```

## Troubleshooting

### 1. Node Tidak Connect

- Pastikan genesis.json sama di semua node
- Cek port tidak bentrok
- Periksa bootnode configuration
- Lihat logs untuk error handshake

### 2. Transaction Tidak Sync

- Pastikan chain ID sama
- Cek koneksi peer
- Verify mining node aktif
- Periksa mempool

### 3. Performance Issues

- Sesuaikan cache size
- Monitor memory usage
- Adjust connection timeout
- Optimize database settings

## Configuration Tuning

### Untuk Development

- Verbosity: 4 (Debug level)
- Cache: 128MB
- Connection timeout: 10s
- Block gas limit: 4M

### Untuk Production

- Verbosity: 2 (Info level)
- Cache: 512MB
- Connection timeout: 30s
- Block gas limit: 8M
- Enable rate limiting
- Enable metrics

## Advanced Setup

### Docker Compose

Untuk deployment yang lebih advanced, gunakan Docker Compose untuk orkestrasi multi-node.

### Load Balancer

Untuk high availability, setup load balancer di depan RPC endpoints.

### Monitoring Stack

Integrasikan dengan Prometheus + Grafana untuk monitoring production.
