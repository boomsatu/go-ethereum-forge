
import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Badge } from "@/components/ui/badge";
import { useToast } from "@/hooks/use-toast";
import { BlockchainControl } from "@/components/blockchain/BlockchainControl";
import { MiningControl } from "@/components/blockchain/MiningControl";
import { BlockMonitor } from "@/components/blockchain/BlockMonitor";
import { TransactionMonitor } from "@/components/blockchain/TransactionMonitor";
import { WalletManager } from "@/components/blockchain/WalletManager";
import { NetworkStatus } from "@/components/blockchain/NetworkStatus";
import { MetricsPanel } from "@/components/blockchain/MetricsPanel";
import { Activity, Blocks, Wallet, Zap, Settings, TrendingUp } from "lucide-react";
import { blockchainService } from '@/services/blockchainService';

const Index = () => {
  const [nodeStatus, setNodeStatus] = useState<'stopped' | 'starting' | 'running' | 'error'>('stopped');
  const [connectionStatus, setConnectionStatus] = useState<'disconnected' | 'connecting' | 'connected'>('disconnected');
  const { toast } = useToast();

  useEffect(() => {
    checkNodeConnection();
    const interval = setInterval(checkNodeConnection, 5000);
    return () => clearInterval(interval);
  }, []);

  const checkNodeConnection = async () => {
    try {
      const health = await blockchainService.getHealthCheck();
      if (health) {
        setConnectionStatus('connected');
        if (nodeStatus === 'starting') {
          setNodeStatus('running');
          toast({
            title: "Node Connected",
            description: "Successfully connected to blockchain node",
          });
        } else if (nodeStatus === 'stopped') {
          setNodeStatus('running');
        }
      } else {
        setConnectionStatus('disconnected');
        if (nodeStatus === 'running') {
          setNodeStatus('stopped');
        }
      }
    } catch (error) {
      setConnectionStatus('disconnected');
      if (nodeStatus === 'running') {
        setNodeStatus('error');
        toast({
          title: "Connection Lost",
          description: "Lost connection to blockchain node",
          variant: "destructive",
        });
      }
    }
  };

  const getStatusColor = () => {
    switch (connectionStatus) {
      case 'connected': return 'bg-green-500';
      case 'connecting': return 'bg-yellow-500';
      default: return 'bg-red-500';
    }
  };

  const getStatusText = () => {
    switch (connectionStatus) {
      case 'connected': return 'Connected';
      case 'connecting': return 'Connecting';
      default: return 'Disconnected';
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100 p-6">
      <div className="max-w-7xl mx-auto space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Blockchain Node Dashboard</h1>
            <p className="text-gray-600 mt-1">Monitor and control your blockchain node</p>
          </div>
          <div className="flex items-center space-x-3">
            <div className="flex items-center space-x-2">
              <div className={`w-3 h-3 rounded-full ${getStatusColor()}`}></div>
              <span className="text-sm font-medium">{getStatusText()}</span>
            </div>
            <Badge variant={connectionStatus === 'connected' ? 'default' : 'destructive'}>
              Node Status
            </Badge>
          </div>
        </div>

        <Tabs defaultValue="overview" className="space-y-6">
          <TabsList className="grid w-full grid-cols-6">
            <TabsTrigger value="overview" className="flex items-center space-x-2">
              <Activity className="w-4 h-4" />
              <span>Overview</span>
            </TabsTrigger>
            <TabsTrigger value="blocks" className="flex items-center space-x-2">
              <Blocks className="w-4 h-4" />
              <span>Blocks</span>
            </TabsTrigger>
            <TabsTrigger value="transactions" className="flex items-center space-x-2">
              <Activity className="w-4 h-4" />
              <span>Transactions</span>
            </TabsTrigger>
            <TabsTrigger value="wallet" className="flex items-center space-x-2">
              <Wallet className="w-4 h-4" />
              <span>Wallet</span>
            </TabsTrigger>
            <TabsTrigger value="network" className="flex items-center space-x-2">
              <Zap className="w-4 h-4" />
              <span>Network</span>
            </TabsTrigger>
            <TabsTrigger value="metrics" className="flex items-center space-x-2">
              <TrendingUp className="w-4 h-4" />
              <span>Metrics</span>
            </TabsTrigger>
          </TabsList>

          <TabsContent value="overview" className="space-y-6">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <BlockchainControl 
                nodeStatus={nodeStatus} 
                setNodeStatus={setNodeStatus}
                connectionStatus={connectionStatus}
              />
              <MiningControl connectionStatus={connectionStatus} />
            </div>
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <BlockMonitor limit={5} />
              <TransactionMonitor limit={5} />
            </div>
          </TabsContent>

          <TabsContent value="blocks">
            <BlockMonitor />
          </TabsContent>

          <TabsContent value="transactions">
            <TransactionMonitor />
          </TabsContent>

          <TabsContent value="wallet">
            <WalletManager />
          </TabsContent>

          <TabsContent value="network">
            <NetworkStatus />
          </TabsContent>

          <TabsContent value="metrics">
            <MetricsPanel />
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
};

export default Index;
