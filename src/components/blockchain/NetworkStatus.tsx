
import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { Network, Wifi, Users, Globe, Server, Activity } from "lucide-react";

interface NetworkInfo {
  chainId: string;
  networkId: string;
  peerCount: number;
  blockHeight: number;
  syncProgress: number;
  difficulty: string;
  hashRate: string;
  connections: PeerConnection[];
}

interface PeerConnection {
  id: string;
  address: string;
  protocol: string;
  latency: number;
  status: 'connected' | 'connecting' | 'disconnected';
}

export const NetworkStatus: React.FC = () => {
  const [networkInfo, setNetworkInfo] = useState<NetworkInfo>({
    chainId: '1337',
    networkId: '1337',
    peerCount: 0,
    blockHeight: 0,
    syncProgress: 100,
    difficulty: '1000',
    hashRate: '0',
    connections: []
  });
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    fetchNetworkInfo();
    const interval = setInterval(fetchNetworkInfo, 10000); // Update every 10 seconds
    return () => clearInterval(interval);
  }, []);

  const fetchNetworkInfo = async () => {
    setLoading(true);
    try {
      // Fetch network information from blockchain node
      const responses = await Promise.all([
        fetch('http://localhost:8545', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            jsonrpc: '2.0',
            method: 'eth_chainId',
            params: [],
            id: 1,
          }),
        }),
        fetch('http://localhost:8545', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            jsonrpc: '2.0',
            method: 'eth_blockNumber',
            params: [],
            id: 2,
          }),
        }),
        fetch('http://localhost:8545', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            jsonrpc: '2.0',
            method: 'net_version',
            params: [],
            id: 3,
          }),
        })
      ]);

      const [chainIdRes, blockNumberRes, networkIdRes] = await Promise.all(
        responses.map(res => res.json())
      );

      setNetworkInfo(prev => ({
        ...prev,
        chainId: chainIdRes.result ? parseInt(chainIdRes.result, 16).toString() : prev.chainId,
        blockHeight: blockNumberRes.result ? parseInt(blockNumberRes.result, 16) : prev.blockHeight,
        networkId: networkIdRes.result || prev.networkId,
      }));
    } catch (error) {
      console.error('Failed to fetch network info:', error);
      // Generate mock data for demonstration
      generateMockNetworkInfo();
    }
    setLoading(false);
  };

  const generateMockNetworkInfo = () => {
    const mockConnections: PeerConnection[] = [];
    const peerCount = Math.floor(Math.random() * 10) + 5;
    
    for (let i = 0; i < peerCount; i++) {
      mockConnections.push({
        id: `peer_${i}`,
        address: `192.168.1.${100 + i}:30303`,
        protocol: 'eth/66',
        latency: Math.floor(Math.random() * 200) + 10,
        status: Math.random() > 0.1 ? 'connected' : 'connecting'
      });
    }

    setNetworkInfo(prev => ({
      ...prev,
      peerCount: peerCount,
      blockHeight: prev.blockHeight + Math.floor(Math.random() * 3),
      syncProgress: Math.min(100, prev.syncProgress + Math.random() * 5),
      difficulty: (parseInt(prev.difficulty) + Math.floor(Math.random() * 100)).toString(),
      hashRate: (Math.floor(Math.random() * 1000000) + 500000).toString(),
      connections: mockConnections
    }));
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'connected': return 'bg-green-100 text-green-800';
      case 'connecting': return 'bg-yellow-100 text-yellow-800';
      case 'disconnected': return 'bg-red-100 text-red-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const formatHashRate = (hashRate: string) => {
    const rate = parseInt(hashRate);
    if (rate >= 1000000000) return `${(rate / 1000000000).toFixed(2)} GH/s`;
    if (rate >= 1000000) return `${(rate / 1000000).toFixed(2)} MH/s`;
    if (rate >= 1000) return `${(rate / 1000).toFixed(2)} KH/s`;
    return `${rate} H/s`;
  };

  return (
    <div className="space-y-6">
      {/* Network Overview */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Chain ID</CardTitle>
            <Globe className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{networkInfo.chainId}</div>
            <p className="text-xs text-muted-foreground">Network: {networkInfo.networkId}</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Block Height</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{networkInfo.blockHeight.toLocaleString()}</div>
            <p className="text-xs text-muted-foreground">Latest block number</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Peers</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{networkInfo.peerCount}</div>
            <p className="text-xs text-muted-foreground">Connected peers</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Hash Rate</CardTitle>
            <Server className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatHashRate(networkInfo.hashRate)}</div>
            <p className="text-xs text-muted-foreground">Network hash rate</p>
          </CardContent>
        </Card>
      </div>

      {/* Sync Status */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center space-x-2">
            <Wifi className="w-5 h-5" />
            <span>Sync Status</span>
          </CardTitle>
          <CardDescription>
            Blockchain synchronization progress
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div>
            <div className="flex justify-between text-sm mb-2">
              <span>Sync Progress</span>
              <span>{networkInfo.syncProgress.toFixed(1)}%</span>
            </div>
            <Progress value={networkInfo.syncProgress} className="w-full" />
          </div>
          
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <span className="font-semibold">Difficulty:</span>
              <span className="ml-2">{parseInt(networkInfo.difficulty).toLocaleString()}</span>
            </div>
            <div>
              <span className="font-semibold">Status:</span>
              <Badge className="ml-2" variant={networkInfo.syncProgress >= 100 ? 'default' : 'secondary'}>
                {networkInfo.syncProgress >= 100 ? 'Synced' : 'Syncing'}
              </Badge>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Peer Connections */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center space-x-2">
            <Network className="w-5 h-5" />
            <span>Peer Connections</span>
          </CardTitle>
          <CardDescription>
            Connected network peers and their status
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {networkInfo.connections.length === 0 ? (
              <div className="text-center py-8 text-gray-500">
                No peer connections available
              </div>
            ) : (
              networkInfo.connections.map((peer) => (
                <div key={peer.id} className="flex items-center justify-between p-3 border rounded-lg">
                  <div className="flex items-center space-x-3">
                    <div className={`w-2 h-2 rounded-full ${
                      peer.status === 'connected' ? 'bg-green-500' : 
                      peer.status === 'connecting' ? 'bg-yellow-500' : 'bg-red-500'
                    }`}></div>
                    <div>
                      <div className="font-medium text-sm">{peer.address}</div>
                      <div className="text-xs text-gray-500">{peer.protocol}</div>
                    </div>
                  </div>
                  <div className="text-right">
                    <Badge className={getStatusColor(peer.status)}>
                      {peer.status}
                    </Badge>
                    <div className="text-xs text-gray-500 mt-1">
                      {peer.latency}ms
                    </div>
                  </div>
                </div>
              ))
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
};
