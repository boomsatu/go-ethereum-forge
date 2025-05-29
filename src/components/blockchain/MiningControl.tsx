
import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { Progress } from "@/components/ui/progress";
import { useToast } from "@/hooks/use-toast";
import { Pickaxe, Play, Pause, Zap, Hash } from "lucide-react";

interface MiningControlProps {
  connectionStatus: 'disconnected' | 'connecting' | 'connected';
}

export const MiningControl: React.FC<MiningControlProps> = ({ connectionStatus }) => {
  const [miningStatus, setMiningStatus] = useState<'stopped' | 'running'>('stopped');
  const [minerAddress, setMinerAddress] = useState('0x742d35Cc6635C0532925a3b8D5c6C1C8b1c5C6C');
  const [difficulty, setDifficulty] = useState('1000');
  const [hashRate, setHashRate] = useState(0);
  const [minedBlocks, setMinedBlocks] = useState(0);
  const [currentProgress, setCurrentProgress] = useState(0);
  const { toast } = useToast();

  useEffect(() => {
    let interval: NodeJS.Timeout;
    
    if (miningStatus === 'running') {
      interval = setInterval(() => {
        // Simulate mining progress
        setCurrentProgress((prev) => {
          const newProgress = prev + Math.random() * 10;
          if (newProgress >= 100) {
            setMinedBlocks(blocks => blocks + 1);
            toast({
              title: "Block Mined!",
              description: `Successfully mined block #${minedBlocks + 1}`,
            });
            return 0;
          }
          return newProgress;
        });
        
        // Update hash rate
        setHashRate(Math.floor(Math.random() * 1000000) + 500000);
      }, 1000);
    }

    return () => {
      if (interval) clearInterval(interval);
    };
  }, [miningStatus, minedBlocks, toast]);

  const startMining = async () => {
    try {
      const response = await fetch('http://localhost:8545/api/mining/start', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          minerAddress,
          difficulty: parseInt(difficulty)
        }),
      });

      if (response.ok) {
        setMiningStatus('running');
        toast({
          title: "Mining Started",
          description: "Mining process has been started",
        });
      }
    } catch (error) {
      toast({
        title: "Mining Start Failed",
        description: "Failed to start mining process",
        variant: "destructive",
      });
    }
  };

  const stopMining = async () => {
    try {
      const response = await fetch('http://localhost:8545/api/mining/stop', {
        method: 'POST',
      });

      setMiningStatus('stopped');
      setCurrentProgress(0);
      toast({
        title: "Mining Stopped",
        description: "Mining process has been stopped",
      });
    } catch (error) {
      toast({
        title: "Mining Stop Failed",
        description: "Failed to stop mining process",
        variant: "destructive",
      });
    }
  };

  const mineManualBlock = async () => {
    try {
      const response = await fetch('http://localhost:8545/api/mining/mine-block', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          minerAddress
        }),
      });

      if (response.ok) {
        const result = await response.json();
        setMinedBlocks(blocks => blocks + 1);
        toast({
          title: "Manual Block Mined",
          description: `Block #${result.blockNumber} mined successfully`,
        });
      }
    } catch (error) {
      toast({
        title: "Manual Mining Failed",
        description: "Failed to mine block manually",
        variant: "destructive",
      });
    }
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
        {/* Mining Status */}
        <div className="grid grid-cols-3 gap-4 text-center">
          <div className="p-3 bg-gray-50 rounded-lg">
            <div className="text-2xl font-bold text-blue-600">{minedBlocks}</div>
            <div className="text-sm text-gray-600">Blocks Mined</div>
          </div>
          <div className="p-3 bg-gray-50 rounded-lg">
            <div className="text-2xl font-bold text-green-600">
              {hashRate.toLocaleString()}
            </div>
            <div className="text-sm text-gray-600">H/s</div>
          </div>
          <div className="p-3 bg-gray-50 rounded-lg">
            <div className="text-2xl font-bold text-purple-600">
              {miningStatus === 'running' ? 'Active' : 'Stopped'}
            </div>
            <div className="text-sm text-gray-600">Status</div>
          </div>
        </div>

        {/* Mining Progress */}
        {miningStatus === 'running' && (
          <div className="space-y-2">
            <div className="flex justify-between text-sm">
              <span>Block Mining Progress</span>
              <span>{Math.round(currentProgress)}%</span>
            </div>
            <Progress value={currentProgress} className="w-full" />
          </div>
        )}

        {/* Control Buttons */}
        <div className="grid grid-cols-3 gap-3">
          <Button 
            onClick={startMining}
            disabled={miningStatus === 'running' || connectionStatus !== 'connected'}
            className="w-full"
          >
            <Play className="w-4 h-4 mr-2" />
            Start Mining
          </Button>
          <Button 
            onClick={stopMining}
            disabled={miningStatus === 'stopped'}
            variant="destructive"
            className="w-full"
          >
            <Pause className="w-4 h-4 mr-2" />
            Stop Mining
          </Button>
          <Button 
            onClick={mineManualBlock}
            disabled={connectionStatus !== 'connected'}
            variant="outline"
            className="w-full"
          >
            <Zap className="w-4 h-4 mr-2" />
            Mine Block
          </Button>
        </div>

        <Separator />

        {/* Mining Configuration */}
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
              <Label htmlFor="difficulty">Difficulty</Label>
              <Input
                id="difficulty"
                value={difficulty}
                onChange={(e) => setDifficulty(e.target.value)}
                placeholder="1000"
                type="number"
              />
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};
