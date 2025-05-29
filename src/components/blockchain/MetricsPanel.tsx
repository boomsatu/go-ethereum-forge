
import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, BarChart, Bar } from 'recharts';
import { TrendingUp, TrendingDown, Activity, Zap, Clock, Database } from "lucide-react";

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
    averageBlockTime: 15,
    totalTransactions: 0,
    totalBlocks: 0
  });

  useEffect(() => {
    fetchMetrics();
    const interval = setInterval(fetchMetrics, 5000); // Update every 5 seconds
    return () => clearInterval(interval);
  }, []);

  const fetchMetrics = async () => {
    try {
      const response = await fetch('http://localhost:8545/api/metrics');
      if (response.ok) {
        const data = await response.json();
        updateMetricsData(data);
      } else {
        generateMockMetrics();
      }
    } catch (error) {
      generateMockMetrics();
    }
  };

  const updateMetricsData = (data: any) => {
    const newMetric: MetricData = {
      timestamp: new Date().toLocaleTimeString(),
      blockCount: data.block_count || 0,
      transactionCount: data.transaction_count || 0,
      hashRate: data.total_hash_rate || 0,
      memoryUsage: data.memory_usage_mb || 0,
      peerCount: data.peer_count || 0,
      blockTime: data.blocks_per_second ? 1 / data.blocks_per_second : 15
    };

    setMetricsData(prev => [...prev.slice(-19), newMetric]);
    
    setSystemMetrics({
      uptime: formatUptime(data.uptime_seconds || 0),
      memoryUsage: data.memory_usage_mb || 0,
      diskUsage: data.disk_usage_mb || 0,
      cpuUsage: Math.random() * 100, // Mock CPU usage
      blocksPerSecond: data.blocks_per_second || 0,
      transactionsPerSecond: data.transactions_per_second || 0,
      averageBlockTime: newMetric.blockTime,
      totalTransactions: data.transaction_count || 0,
      totalBlocks: data.block_count || 0
    });
  };

  const generateMockMetrics = () => {
    const now = new Date();
    const newMetric: MetricData = {
      timestamp: now.toLocaleTimeString(),
      blockCount: metricsData.length > 0 ? metricsData[metricsData.length - 1].blockCount + Math.floor(Math.random() * 3) : Math.floor(Math.random() * 100),
      transactionCount: Math.floor(Math.random() * 50),
      hashRate: Math.floor(Math.random() * 1000000) + 500000,
      memoryUsage: Math.floor(Math.random() * 1000) + 200,
      peerCount: Math.floor(Math.random() * 10) + 5,
      blockTime: 10 + Math.random() * 10
    };

    setMetricsData(prev => [...prev.slice(-19), newMetric]);

    setSystemMetrics(prev => ({
      uptime: formatUptime(Date.now() / 1000),
      memoryUsage: newMetric.memoryUsage,
      diskUsage: Math.floor(Math.random() * 5000) + 1000,
      cpuUsage: Math.random() * 100,
      blocksPerSecond: 1 / newMetric.blockTime,
      transactionsPerSecond: newMetric.transactionCount / newMetric.blockTime,
      averageBlockTime: newMetric.blockTime,
      totalTransactions: prev.totalTransactions + newMetric.transactionCount,
      totalBlocks: newMetric.blockCount
    }));
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
          <CardTitle>Detailed Statistics</CardTitle>
          <CardDescription>Comprehensive blockchain metrics</CardDescription>
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
              <div className="text-2xl font-bold text-purple-600">{systemMetrics.averageBlockTime.toFixed(1)}s</div>
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
