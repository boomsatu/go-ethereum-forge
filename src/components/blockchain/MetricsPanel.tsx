import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, BarChart, Bar } from 'recharts';
import { TrendingUp, Activity, Clock, Database, RefreshCw } from "lucide-react";
import blockchainService from '@/services/blockchainService';

interface MetricData {
  timestamp: string;
  blockCount: number;
  transactionCount: number;
  hashRate: number;
  memoryUsage: number;
  peerCount: number;
  blockTime: number;
}

interface SystemMetrics {
  uptime: string;
  memoryUsage: number;
  diskUsage: number;
  cpuUsage: number;
  blocksPerSecond: number;
  transactionsPerSecond: number;
  averageBlockTime: number;
  totalTransactions: number;
  totalBlocks: number;
}

export const MetricsPanel: React.FC = () => {
  const [metricsData, setMetricsData] = useState<MetricData[]>([]);
  const [systemMetrics, setSystemMetrics] = useState<SystemMetrics>({
    uptime: '0h 0m',
    memoryUsage: 0,
    diskUsage: 0,
    cpuUsage: 0,
    blocksPerSecond: 0,
    transactionsPerSecond: 0,
    averageBlockTime: 0,
    totalTransactions: 0,
    totalBlocks: 0
  });
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    fetchMetrics();
    const interval = setInterval(fetchMetrics, 5000);
    return () => clearInterval(interval);
  }, []);

  const fetchMetrics = async () => {
    setLoading(true);
    try {
      // Fetch real metrics from blockchain
      const [nodeMetrics, networkStats, latestBlock, miningStats] = await Promise.all([
        blockchainService.getMetrics(),
        blockchainService.getNetworkStats(),
        blockchainService.getLatestBlock(),
        blockchainService.getMiningStats()
      ]);

      if (nodeMetrics && networkStats) {
        updateMetricsData({
          block_count: networkStats.blockHeight,
          transaction_count: nodeMetrics.transactionCount || 0,
          total_hash_rate: parseInt(networkStats.hashRate) || 0,
          memory_usage_mb: nodeMetrics.memoryUsage || 0,
          peer_count: networkStats.peerCount,
          uptime_seconds: nodeMetrics.uptime || 0,
          blocks_per_second: nodeMetrics.blockCount > 0 ? nodeMetrics.blockCount / (nodeMetrics.uptime || 1) : 0,
          transactions_per_second: nodeMetrics.transactionCount > 0 ? nodeMetrics.transactionCount / (nodeMetrics.uptime || 1) : 0,
          disk_usage_mb: nodeMetrics.diskUsage || 0
        });
      } else {
        // If no real data available, reset to zero values
        updateMetricsData({
          block_count: 0,
          transaction_count: 0,
          total_hash_rate: 0,
          memory_usage_mb: 0,
          peer_count: 0,
          uptime_seconds: 0,
          blocks_per_second: 0,
          transactions_per_second: 0,
          disk_usage_mb: 0
        });
      }
    } catch (error) {
      console.error('Failed to fetch metrics:', error);
      // Reset to zero if blockchain is not accessible
      updateMetricsData({
        block_count: 0,
        transaction_count: 0,
        total_hash_rate: 0,
        memory_usage_mb: 0,
        peer_count: 0,
        uptime_seconds: 0,
        blocks_per_second: 0,
        transactions_per_second: 0,
        disk_usage_mb: 0
      });
    }
    setLoading(false);
  };

  const updateMetricsData = (data: any) => {
    const now = new Date();
    const newMetric: MetricData = {
      timestamp: now.toLocaleTimeString(),
      blockCount: data.block_count || 0,
      transactionCount: data.transaction_count || 0,
      hashRate: data.total_hash_rate || 0,
      memoryUsage: data.memory_usage_mb || 0,
      peerCount: data.peer_count || 0,
      blockTime: data.blocks_per_second > 0 ? 1 / data.blocks_per_second : 0
    };

    setMetricsData(prev => [...prev.slice(-19), newMetric]);
    
    setSystemMetrics({
      uptime: formatUptime(data.uptime_seconds || 0),
      memoryUsage: data.memory_usage_mb || 0,
      diskUsage: data.disk_usage_mb || 0,
      cpuUsage: 0, // CPU usage not available from blockchain
      blocksPerSecond: data.blocks_per_second || 0,
      transactionsPerSecond: data.transactions_per_second || 0,
      averageBlockTime: newMetric.blockTime,
      totalTransactions: data.transaction_count || 0,
      totalBlocks: data.block_count || 0
    });
  };

  const formatUptime = (seconds: number) => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return `${hours}h ${minutes}m`;
  };

  const formatBytes = (bytes: number) => {
    if (bytes >= 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`;
    if (bytes >= 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(2)} MB`;
    if (bytes >= 1024) return `${(bytes / 1024).toFixed(2)} KB`;
    return `${bytes} B`;
  };

  const formatHashRate = (hashRate: number) => {
    if (hashRate >= 1000000000) return `${(hashRate / 1000000000).toFixed(2)} GH/s`;
    if (hashRate >= 1000000) return `${(hashRate / 1000000).toFixed(2)} MH/s`;
    if (hashRate >= 1000) return `${(hashRate / 1000).toFixed(2)} KH/s`;
    return `${hashRate} H/s`;
  };

  return (
    <div className="space-y-6">
      {/* System Overview */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Uptime</CardTitle>
            <Clock className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{systemMetrics.uptime}</div>
            <p className="text-xs text-muted-foreground">Node uptime</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Memory Usage</CardTitle>
            <Database className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatBytes(systemMetrics.memoryUsage * 1024 * 1024)}</div>
            <p className="text-xs text-muted-foreground">RAM used</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Blocks/sec</CardTitle>
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{systemMetrics.blocksPerSecond.toFixed(3)}</div>
            <p className="text-xs text-muted-foreground">Block production rate</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">TPS</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{systemMetrics.transactionsPerSecond.toFixed(2)}</div>
            <p className="text-xs text-muted-foreground">Transactions per second</p>
          </CardContent>
        </Card>
      </div>

      {/* Control Panel */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Real-time Metrics</CardTitle>
              <CardDescription>Live blockchain performance data</CardDescription>
            </div>
            <Button 
              onClick={fetchMetrics} 
              disabled={loading}
              variant="outline"
              size="sm"
            >
              <RefreshCw className={`w-4 h-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
              Refresh
            </Button>
          </div>
        </CardHeader>
      </Card>

      {/* Performance Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>Block Production</CardTitle>
            <CardDescription>Real-time block production metrics</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="h-[300px]">
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={metricsData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="timestamp" />
                  <YAxis />
                  <Tooltip />
                  <Line 
                    type="monotone" 
                    dataKey="blockCount" 
                    stroke="#8884d8" 
                    strokeWidth={2}
                    name="Blocks"
                  />
                </LineChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Transaction Volume</CardTitle>
            <CardDescription>Transaction throughput over time</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="h-[300px]">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={metricsData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="timestamp" />
                  <YAxis />
                  <Tooltip />
                  <Bar 
                    dataKey="transactionCount" 
                    fill="#82ca9d"
                    name="Transactions"
                  />
                </BarChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Hash Rate</CardTitle>
            <CardDescription>Network hash rate over time</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="h-[300px]">
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={metricsData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="timestamp" />
                  <YAxis />
                  <Tooltip formatter={(value: number) => [formatHashRate(value), 'Hash Rate']} />
                  <Line 
                    type="monotone" 
                    dataKey="hashRate" 
                    stroke="#ff7300" 
                    strokeWidth={2}
                    name="Hash Rate"
                  />
                </LineChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>System Resources</CardTitle>
            <CardDescription>Memory usage and peer connections</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="h-[300px]">
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={metricsData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="timestamp" />
                  <YAxis />
                  <Tooltip />
                  <Line 
                    type="monotone" 
                    dataKey="memoryUsage" 
                    stroke="#8884d8" 
                    strokeWidth={2}
                    name="Memory (MB)"
                  />
                  <Line 
                    type="monotone" 
                    dataKey="peerCount" 
                    stroke="#82ca9d" 
                    strokeWidth={2}
                    name="Peers"
                  />
                </LineChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Detailed Statistics */}
      <Card>
        <CardHeader>
          <CardTitle>Blockchain Statistics</CardTitle>
          <CardDescription>Comprehensive blockchain metrics from real data</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
            <div className="text-center">
              <div className="text-2xl font-bold text-blue-600">{systemMetrics.totalBlocks.toLocaleString()}</div>
              <div className="text-sm text-gray-600">Total Blocks</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-green-600">{systemMetrics.totalTransactions.toLocaleString()}</div>
              <div className="text-sm text-gray-600">Total Transactions</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-purple-600">
                {systemMetrics.averageBlockTime > 0 ? `${systemMetrics.averageBlockTime.toFixed(1)}s` : 'N/A'}
              </div>
              <div className="text-sm text-gray-600">Avg Block Time</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-orange-600">{formatBytes(systemMetrics.diskUsage * 1024 * 1024)}</div>
              <div className="text-sm text-gray-600">Disk Usage</div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
};
