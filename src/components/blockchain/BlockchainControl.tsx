
import React, { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { useToast } from "@/hooks/use-toast";
import { PlayCircle, StopCircle, RotateCcw, Settings } from "lucide-react";

interface BlockchainControlProps {
  nodeStatus: 'stopped' | 'starting' | 'running' | 'error';
  setNodeStatus: (status: 'stopped' | 'starting' | 'running' | 'error') => void;
  connectionStatus: 'disconnected' | 'connecting' | 'connected';
}

export const BlockchainControl: React.FC<BlockchainControlProps> = ({
  nodeStatus,
  setNodeStatus,
  connectionStatus
}) => {
  const [config, setConfig] = useState({
    port: '8545',
    chainId: '1337',
    gasLimit: '8000000',
    dataDir: './data'
  });
  const { toast } = useToast();

  const startNode = async () => {
    setNodeStatus('starting');
    toast({
      title: "Starting Node",
      description: "Blockchain node is starting up...",
    });

    try {
      // Simulate node startup - in real implementation, this would call the Go backend
      setTimeout(() => {
        setNodeStatus('running');
        toast({
          title: "Node Started",
          description: "Blockchain node is now running",
        });
      }, 3000);
    } catch (error) {
      setNodeStatus('error');
      toast({
        title: "Start Failed",
        description: "Failed to start blockchain node",
        variant: "destructive",
      });
    }
  };

  const stopNode = async () => {
    try {
      const response = await fetch('http://localhost:8545/api/admin/stop', {
        method: 'POST',
      });
      
      setNodeStatus('stopped');
      toast({
        title: "Node Stopped",
        description: "Blockchain node has been stopped",
      });
    } catch (error) {
      toast({
        title: "Stop Failed",
        description: "Failed to stop blockchain node",
        variant: "destructive",
      });
    }
  };

  const restartNode = async () => {
    await stopNode();
    setTimeout(() => {
      startNode();
    }, 2000);
  };

  const updateConfig = async () => {
    try {
      const response = await fetch('http://localhost:8545/api/admin/config', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(config),
      });

      if (response.ok) {
        toast({
          title: "Configuration Updated",
          description: "Node configuration has been updated",
        });
      }
    } catch (error) {
      toast({
        title: "Update Failed",
        description: "Failed to update configuration",
        variant: "destructive",
      });
    }
  };

  const getStatusIcon = () => {
    switch (nodeStatus) {
      case 'running': return <PlayCircle className="w-5 h-5 text-green-500" />;
      case 'starting': return <RotateCcw className="w-5 h-5 text-yellow-500 animate-spin" />;
      case 'error': return <StopCircle className="w-5 h-5 text-red-500" />;
      default: return <StopCircle className="w-5 h-5 text-gray-500" />;
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center space-x-2">
          {getStatusIcon()}
          <span>Blockchain Control</span>
        </CardTitle>
        <CardDescription>
          Start, stop, and configure your blockchain node
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Node Control Buttons */}
        <div className="grid grid-cols-3 gap-3">
          <Button 
            onClick={startNode}
            disabled={nodeStatus === 'running' || nodeStatus === 'starting'}
            className="w-full"
            variant={nodeStatus === 'running' ? 'secondary' : 'default'}
          >
            <PlayCircle className="w-4 h-4 mr-2" />
            Start
          </Button>
          <Button 
            onClick={stopNode}
            disabled={nodeStatus === 'stopped' || nodeStatus === 'starting'}
            variant="destructive"
            className="w-full"
          >
            <StopCircle className="w-4 h-4 mr-2" />
            Stop
          </Button>
          <Button 
            onClick={restartNode}
            disabled={nodeStatus === 'starting'}
            variant="outline"
            className="w-full"
          >
            <RotateCcw className="w-4 h-4 mr-2" />
            Restart
          </Button>
        </div>

        <Separator />

        {/* Configuration */}
        <div className="space-y-4">
          <div className="flex items-center space-x-2">
            <Settings className="w-4 h-4" />
            <span className="font-semibold">Configuration</span>
          </div>
          
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="port">RPC Port</Label>
              <Input
                id="port"
                value={config.port}
                onChange={(e) => setConfig({...config, port: e.target.value})}
                placeholder="8545"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="chainId">Chain ID</Label>
              <Input
                id="chainId"
                value={config.chainId}
                onChange={(e) => setConfig({...config, chainId: e.target.value})}
                placeholder="1337"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="gasLimit">Gas Limit</Label>
              <Input
                id="gasLimit"
                value={config.gasLimit}
                onChange={(e) => setConfig({...config, gasLimit: e.target.value})}
                placeholder="8000000"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="dataDir">Data Directory</Label>
              <Input
                id="dataDir"
                value={config.dataDir}
                onChange={(e) => setConfig({...config, dataDir: e.target.value})}
                placeholder="./data"
              />
            </div>
          </div>

          <Button 
            onClick={updateConfig}
            variant="outline" 
            className="w-full"
            disabled={connectionStatus !== 'connected'}
          >
            Update Configuration
          </Button>
        </div>
      </CardContent>
    </Card>
  );
};
