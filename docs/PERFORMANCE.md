
# Blockchain Performance Guide

## Overview

Panduan ini menjelaskan cara mengoptimalkan performa blockchain node untuk berbagai skenario deployment.

## Key Performance Metrics

### 1. Transaction Throughput (TPS)
- **Target**: 100-1000 TPS untuk private network
- **Faktor**: Block gas limit, block time, network latency

### 2. Block Time
- **Target**: 1-15 detik per block
- **Faktor**: Mining difficulty, network size, consensus algorithm

### 3. Memory Usage
- **Target**: < 2GB untuk node standar
- **Faktor**: Cache size, state trie size, pending transactions

### 4. Network Latency
- **Target**: < 100ms untuk P2P communication
- **Faktor**: Geographic distribution, network quality

## Configuration Tuning

### 1. Cache Settings

```yaml
# High Performance
cache: 1024  # 1GB cache
cache_size: 100000  # Large application cache
enable_cache: true

# Memory Constrained
cache: 128   # 128MB cache
cache_size: 1000   # Small application cache
enable_cache: true

# Minimal Memory
cache: 64    # 64MB cache
cache_size: 100    # Tiny cache
enable_cache: false
```

### 2. Database Settings

```yaml
# High Throughput
handles: 1024
cache: 1024

# Balanced
handles: 512
cache: 512

# Low Resource
handles: 256
cache: 256
```

### 3. Network Settings

```yaml
# High Connectivity
maxpeers: 100
connection_timeout: "30s"

# Balanced
maxpeers: 50
connection_timeout: "20s"

# Low Bandwidth
maxpeers: 10
connection_timeout: "10s"
```

### 4. Block Gas Limit

```yaml
# High Throughput
blockgaslimit: 15000000  # 15M gas per block

# Standard
blockgaslimit: 8000000   # 8M gas per block

# Conservative
blockgaslimit: 4000000   # 4M gas per block
```

## Performance Monitoring

### 1. System Metrics

Monitor CPU, Memory, Disk I/O:

```bash
# CPU usage
top -p $(pgrep blockchain-node)

# Memory usage
ps aux | grep blockchain-node

# Disk I/O
iotop -p $(pgrep blockchain-node)
```

### 2. Application Metrics

```bash
# Get metrics via API
curl http://localhost:9545/metrics

# Key metrics to monitor:
# - blockCount: Current block height
# - transactionCount: Total transactions
# - memoryUsage: Memory consumption
# - peersConnected: Active peer connections
```

### 3. Network Performance

```bash
# Check peer connectivity
curl http://localhost:8545/api/network/stats

# Monitor transaction propagation
curl http://localhost:8545/api/network/peers
```

## Optimization Strategies

### 1. Hardware Recommendations

#### Minimum Requirements
- **CPU**: 2 cores
- **RAM**: 4GB
- **Storage**: 100GB SSD
- **Network**: 10 Mbps

#### Recommended
- **CPU**: 4+ cores
- **RAM**: 8GB+
- **Storage**: 500GB+ NVMe SSD
- **Network**: 100 Mbps+

#### High Performance
- **CPU**: 8+ cores
- **RAM**: 16GB+
- **Storage**: 1TB+ NVMe SSD
- **Network**: 1 Gbps+

### 2. Database Optimization

```yaml
# Use SSD storage
datadir: "/ssd/blockchain/data"

# Optimize cache
cache: 1024
handles: 1024

# Enable compression
enable_compression: true
```

### 3. Network Optimization

```yaml
# Increase peer connections
maxpeers: 100

# Optimize timeouts
connection_timeout: "30s"
handshake_timeout: "10s"

# Use dedicated network interface
bind_interface: "eth0"
```

### 4. Memory Management

```yaml
# Large cache for read performance
cache_size: 50000

# Enable memory optimization
enable_gc_optimization: true
gc_percent: 10

# Limit memory growth
max_memory_usage: "8GB"
```

## Load Testing

### 1. Transaction Load Test

```bash
# Generate multiple transactions
for i in {1..100}; do
  curl -X POST http://localhost:8545/api/wallet/send \
    -H "Content-Type: application/json" \
    -d '{
      "from": "0x742d35Cc6635C0532925a3b8D5c6C1C8b1c5C6C7",
      "to": "0x8ba1f109551bD432803012645Hac136c46C01C23",
      "value": "1000000000000000000",
      "privateKey": "YOUR_PRIVATE_KEY"
    }' &
done
```

### 2. Network Load Test

```bash
# Test with multiple clients
for port in 8545 8546 8547; do
  curl http://localhost:$port/api/admin/status &
done
```

### 3. Mining Performance Test

```bash
# Monitor mining stats
watch -n 1 'curl -s http://localhost:8545/api/mining/stats | jq'
```

## Troubleshooting Performance Issues

### 1. High Memory Usage

**Symptoms**: Node becomes slow, system swapping
**Solutions**:
- Reduce cache size
- Implement memory limits
- Restart node periodically
- Use memory profiling

### 2. Slow Transaction Processing

**Symptoms**: High transaction pending time
**Solutions**:
- Increase block gas limit
- Optimize database
- Check network connectivity
- Monitor disk I/O

### 3. Network Connectivity Issues

**Symptoms**: Low peer count, sync problems
**Solutions**:
- Check firewall settings
- Verify bootnode connectivity
- Increase connection timeout
- Monitor network quality

### 4. Database Performance

**Symptoms**: Slow block sync, high disk usage
**Solutions**:
- Use SSD storage
- Increase database cache
- Optimize disk I/O
- Consider database cleanup

## Benchmarking

### 1. Transaction Throughput

```bash
# Measure TPS
start_time=$(date +%s)
# Send 1000 transactions
end_time=$(date +%s)
tps=$((1000 / (end_time - start_time)))
echo "TPS: $tps"
```

### 2. Block Processing Time

Monitor block creation and processing times through metrics endpoint.

### 3. Network Latency

```bash
# Measure P2P message propagation
time curl -X POST http://localhost:8545/api/wallet/send -d '...'
```

## Production Optimization

### 1. Load Balancing

Use multiple RPC endpoints behind load balancer:

```yaml
# Load balancer config
upstream blockchain_rpc {
    server 127.0.0.1:8545;
    server 127.0.0.1:8546;
    server 127.0.0.1:8547;
}
```

### 2. Caching Layer

Implement Redis caching for frequent queries:

```yaml
# Redis config for RPC caching
redis:
  host: "localhost"
  port: 6379
  ttl: 60
```

### 3. Monitoring Setup

Use Prometheus + Grafana for production monitoring:

```yaml
# Prometheus config
prometheus:
  enabled: true
  endpoint: "/metrics"
  port: 9090
```

## Best Practices

1. **Monitor continuously**: Set up alerts for key metrics
2. **Regular maintenance**: Restart nodes, clean logs, backup data
3. **Gradual scaling**: Start small, scale based on actual usage
4. **Testing**: Load test before production deployment
5. **Documentation**: Keep configuration and performance docs updated
