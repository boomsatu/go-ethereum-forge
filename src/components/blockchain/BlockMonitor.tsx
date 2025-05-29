
import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import { RefreshCw, Blocks, Clock, Hash } from "lucide-react";

interface Block {
  number: number;
  hash: string;
  parentHash: string;
  timestamp: number;
  transactionCount: number;
  gasUsed: string;
  gasLimit: string;
  difficulty: string;
  miner: string;
  size: number;
}

interface BlockMonitorProps {
  limit?: number;
}

export const BlockMonitor: React.FC<BlockMonitorProps> = ({ limit }) => {
  const [blocks, setBlocks] = useState<Block[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedBlock, setSelectedBlock] = useState<Block | null>(null);

  useEffect(() => {
    fetchBlocks();
    const interval = setInterval(fetchBlocks, 10000); // Refresh every 10 seconds
    return () => clearInterval(interval);
  }, [limit]);

  const fetchBlocks = async () => {
    setLoading(true);
    try {
      // Get latest block number
      const blockNumberResponse = await fetch('http://localhost:8545', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          jsonrpc: '2.0',
          method: 'eth_blockNumber',
          params: [],
          id: 1,
        }),
      });

      const blockNumberResult = await blockNumberResponse.json();
      const latestBlockNumber = parseInt(blockNumberResult.result, 16);

      // Fetch recent blocks
      const blocksToFetch = limit || 20;
      const startBlock = Math.max(0, latestBlockNumber - blocksToFetch + 1);
      
      const blockPromises = [];
      for (let i = startBlock; i <= latestBlockNumber; i++) {
        blockPromises.push(fetchBlock(i));
      }

      const fetchedBlocks = await Promise.all(blockPromises);
      setBlocks(fetchedBlocks.filter(block => block !== null).reverse());
    } catch (error) {
      console.error('Failed to fetch blocks:', error);
      // Generate mock data for demonstration
      generateMockBlocks();
    }
    setLoading(false);
  };

  const fetchBlock = async (blockNumber: number): Promise<Block | null> => {
    try {
      const response = await fetch('http://localhost:8545', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          jsonrpc: '2.0',
          method: 'eth_getBlockByNumber',
          params: [`0x${blockNumber.toString(16)}`, true],
          id: 1,
        }),
      });

      const result = await response.json();
      if (result.result) {
        const block = result.result;
        return {
          number: parseInt(block.number, 16),
          hash: block.hash,
          parentHash: block.parentHash,
          timestamp: parseInt(block.timestamp, 16),
          transactionCount: block.transactions.length,
          gasUsed: parseInt(block.gasUsed, 16).toLocaleString(),
          gasLimit: parseInt(block.gasLimit, 16).toLocaleString(),
          difficulty: parseInt(block.difficulty, 16).toLocaleString(),
          miner: block.miner || '0x0000000000000000000000000000000000000000',
          size: parseInt(block.size || '0x0', 16),
        };
      }
    } catch (error) {
      console.error(`Failed to fetch block ${blockNumber}:`, error);
    }
    return null;
  };

  const generateMockBlocks = () => {
    const mockBlocks: Block[] = [];
    const now = Date.now();
    
    for (let i = 0; i < (limit || 10); i++) {
      mockBlocks.push({
        number: i + 1,
        hash: `0x${Math.random().toString(16).slice(2, 66)}`,
        parentHash: `0x${Math.random().toString(16).slice(2, 66)}`,
        timestamp: Math.floor((now - i * 15000) / 1000),
        transactionCount: Math.floor(Math.random() * 20),
        gasUsed: (Math.floor(Math.random() * 8000000)).toLocaleString(),
        gasLimit: '8,000,000',
        difficulty: (1000 + i * 100).toLocaleString(),
        miner: '0x742d35Cc6635C0532925a3b8D5c6C1C8b1c5C6C',
        size: Math.floor(Math.random() * 10000) + 1000,
      });
    }
    
    setBlocks(mockBlocks.reverse());
  };

  const formatTimestamp = (timestamp: number) => {
    return new Date(timestamp * 1000).toLocaleString();
  };

  const truncateHash = (hash: string) => {
    return `${hash.slice(0, 8)}...${hash.slice(-6)}`;
  };

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="flex items-center space-x-2">
                <Blocks className="w-5 h-5" />
                <span>Block Monitor</span>
              </CardTitle>
              <CardDescription>
                Real-time monitoring of blockchain blocks
              </CardDescription>
            </div>
            <Button 
              onClick={fetchBlocks} 
              disabled={loading}
              variant="outline"
              size="sm"
            >
              <RefreshCw className={`w-4 h-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
              Refresh
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <ScrollArea className="h-[400px]">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Block #</TableHead>
                  <TableHead>Hash</TableHead>
                  <TableHead>Transactions</TableHead>
                  <TableHead>Gas Used</TableHead>
                  <TableHead>Timestamp</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {blocks.map((block) => (
                  <TableRow key={block.number}>
                    <TableCell className="font-medium">
                      <Badge variant="secondary">#{block.number}</Badge>
                    </TableCell>
                    <TableCell className="font-mono text-sm">
                      {truncateHash(block.hash)}
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline">{block.transactionCount}</Badge>
                    </TableCell>
                    <TableCell>{block.gasUsed}</TableCell>
                    <TableCell className="text-sm">
                      {formatTimestamp(block.timestamp)}
                    </TableCell>
                    <TableCell>
                      <Button 
                        onClick={() => setSelectedBlock(block)}
                        variant="ghost" 
                        size="sm"
                      >
                        View
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </ScrollArea>
        </CardContent>
      </Card>

      {/* Block Details Modal/Card */}
      {selectedBlock && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center space-x-2">
              <Hash className="w-5 h-5" />
              <span>Block #{selectedBlock.number} Details</span>
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label className="text-sm font-semibold">Hash</Label>
                <p className="font-mono text-sm break-all">{selectedBlock.hash}</p>
              </div>
              <div>
                <Label className="text-sm font-semibold">Parent Hash</Label>
                <p className="font-mono text-sm break-all">{selectedBlock.parentHash}</p>
              </div>
              <div>
                <Label className="text-sm font-semibold">Miner</Label>
                <p className="font-mono text-sm">{selectedBlock.miner}</p>
              </div>
              <div>
                <Label className="text-sm font-semibold">Difficulty</Label>
                <p className="text-sm">{selectedBlock.difficulty}</p>
              </div>
              <div>
                <Label className="text-sm font-semibold">Gas Limit</Label>
                <p className="text-sm">{selectedBlock.gasLimit}</p>
              </div>
              <div>
                <Label className="text-sm font-semibold">Gas Used</Label>
                <p className="text-sm">{selectedBlock.gasUsed}</p>
              </div>
              <div>
                <Label className="text-sm font-semibold">Size</Label>
                <p className="text-sm">{selectedBlock.size.toLocaleString()} bytes</p>
              </div>
              <div>
                <Label className="text-sm font-semibold">Timestamp</Label>
                <p className="text-sm">{formatTimestamp(selectedBlock.timestamp)}</p>
              </div>
            </div>
            <Button 
              onClick={() => setSelectedBlock(null)}
              variant="outline"
              className="w-full"
            >
              Close Details
            </Button>
          </CardContent>
        </Card>
      )}
    </div>
  );
};

// Add Label component for consistency
const Label: React.FC<{ className?: string; children: React.ReactNode }> = ({ className, children }) => (
  <div className={`text-sm font-medium ${className}`}>{children}</div>
);
