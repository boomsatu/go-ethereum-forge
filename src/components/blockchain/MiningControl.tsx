import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { Progress } from "@/components/ui/progress";
import { useToast } from "@/hooks/use-toast";
import { Pickaxe, PlayCircle, StopCircle, Zap, Hash } from "lucide-react";
import blockchainService, { MiningStats } from '@/services/blockchainService';

interface MiningControlProps {
  connectionStatus: 'disconnected' | 'connecting' | 'connected';
}

export const MiningControl: React.FC<MiningControlProps> = ({ connectionStatus }) => {
  const [miningStats, setMiningStats] = useState<MiningStats>({
    isActive: false,
    hashRate: 0,
    blocksFound: 0,
    difficulty: '0'
  });
  const [minerAddress, setMinerAddress] = useState('0x742d35Cc6635C0532925a3b8D5c6C1C8b1c5C6C');
  const [threads, setThreads] = useState(1);
  const [isLoading, setIsLoading] = useState(false);
  const { toast } = useToast();

  useEffect(() => {
    if (connectionStatus === 'connected') {
      fetchMiningStats();
      const interval = setInterval(fetchMiningStats, 2000);
      return () => clearInterval(interval);
    }
  }, [connectionStatus]);

  const fetchMiningStats = async () => {
    try {
      const stats = await blockchainService.getMiningStats();
      if (stats) {
        setMiningStats(stats);
      }
    } catch (error) {
      console.error('Failed to fetch mining stats:', error);
    }
  };

  const startMining = async () => {
    if (!minerAddress.trim()) {
      toast({
        title: "Invalid Miner Address",
        description: "Please enter a valid miner address",
        variant: "destructive",
      });
      return;
    }

    setIsLoading(true);
    try {
      const success = await blockchainService.startMining(minerAddress, threads);
      if (success) {
        toast({
          title: "Mining Started",
          description: "Mining process has been started",
        });
        await fetchMiningStats();
      } else {
        toast({
          title: "Mining Start Failed",
          description: "Failed to start mining process",
          variant: "destructive",
        });
      }
    } catch (error) {
      toast({
        title: "Mining Start Failed",
        description: "Failed to start mining process",
        variant: "destructive",
      });
    }
    setIsLoading(false);
  };

  const stopMining = async () => {
    setIsLoading(true);
    try {
      const success = await blockchainService.stopMining();
      if (success) {
        toast({
          title: "Mining Stopped",
          description: "Mining process has been stopped",
        });
        await fetchMiningStats();
      } else {
        toast({
          title: "Mining Stop Failed",
          description: "Failed to stop mining process",
          variant: "destructive",
        });
      }
    } catch (error) {
      toast({
        title: "Mining Stop Failed",
        description: "Failed to stop mining process",
        variant: "destructive",
      });
    }
    setIsLoading(false);
  };

  const mineManualBlock = async () => {
    if (!minerAddress.trim()) {
      toast({
        title: "Invalid Miner Address",
        description: "Please enter a valid miner address",
        variant: "destructive",
      });
      return;
    }

    setIsLoading(true);
    try {
      const result = await blockchainService.mineBlock(minerAddress);
      if (result) {
        toast({
          title: "Manual Block Mined",
          description: `Block #${result.blockNumber} mined successfully`,
        });
        await fetchMiningStats();
      } else {
        toast({
          title: "Manual Mining Failed",
          description: "Failed to mine block manually",
          variant: "destructive",
        });
      }
    } catch (error) {
      toast({
        title: "Manual Mining Failed",
        description: "Failed to mine block manually",
        variant: "destructive",
      });
    }
    setIsLoading(false);
  };

  const formatHashRate = (hashRate: number) => {
    if (hashRate >= 1000000000) return `${(hashRate / 1000000000).toFixed(2)} GH/s`;
    if (hashRate >= 1000000) return `${(hashRate / 1000000).toFixed(2)} MH/s`;
    if (hashRate >= 1000) return `${(hashRate / 1000).toFixed(2)} KH/s`;
    return `${hashRate} H/s`;
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center space-x-2">
          <Pickaxe className="w-5 h-5" />
          <span>Mining Control</span>
        </CardTitle>
        <CardDescription>
          Control the mining process and mine new blocks
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="grid grid-cols-3 gap-4 text-center">
          <div className="p-3 bg-gray-50 rounded-lg">
            <div className="text-2xl font-bold text-blue-600">{miningStats.blocksFound}</div>
            <div className="text-sm text-gray-600">Blocks Found</div>
          </div>
          <div className="p-3 bg-gray-50 rounded-lg">
            <div className="text-2xl font-bold text-green-600">
              {formatHashRate(miningStats.hashRate)}
            </div>
            <div className="text-sm text-gray-600">Hash Rate</div>
          </div>
          <div className="p-3 bg-gray-50 rounded-lg">
            <div className="text-2xl font-bold text-purple-600">
              {miningStats.isActive ? 'Active' : 'Stopped'}
            </div>
            <div className="text-sm text-gray-600">Status</div>
          </div>
        </div>

        <div className="grid grid-cols-3 gap-3">
          <Button 
            onClick={startMining}
            disabled={miningStats.isActive || connectionStatus !== 'connected' || isLoading}
            className="w-full"
          >
            <PlayCircle className="w-4 h-4 mr-2" />
            Start Mining
          </Button>
          <Button 
            onClick={stopMining}
            disabled={!miningStats.isActive || isLoading}
            variant="destructive"
            className="w-full"
          >
            <StopCircle className="w-4 h-4 mr-2" />
            Stop Mining
          </Button>
          <Button 
            onClick={mineManualBlock}
            disabled={connectionStatus !== 'connected' || isLoading}
            variant="outline"
            className="w-full"
          >
            <Zap className="w-4 h-4 mr-2" />
            Mine Block
          </Button>
        </div>

        <Separator />

        <div className="space-y-4">
          <div className="flex items-center space-x-2">
            <Hash className="w-4 h-4" />
            <span className="font-semibold">Mining Configuration</span>
          </div>
          
          <div className="space-y-3">
            <div className="space-y-2">
              <Label htmlFor="minerAddress">Miner Address</Label>
              <Input
                id="minerAddress"
                value={minerAddress}
                onChange={(e) => setMinerAddress(e.target.value)}
                placeholder="0x..."
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="threads">Mining Threads</Label>
              <Input
                id="threads"
                value={threads}
                onChange={(e) => setThreads(parseInt(e.target.value) || 1)}
                placeholder="1"
                type="number"
                min="1"
                max="16"
              />
            </div>
            <div className="space-y-2">
              <Label>Current Difficulty</Label>
              <div className="text-sm text-gray-600">
                {parseInt(miningStats.difficulty).toLocaleString()}
              </div>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};
