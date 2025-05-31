
import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import { RefreshCw, Blocks, Hash } from "lucide-react";
import blockchainService, { RawBlockData } from '@/services/blockchainService';

interface BlockMonitorProps {
  limit?: number;
}

interface ProcessedBlock {
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

export const BlockMonitor: React.FC<BlockMonitorProps> = ({ limit }) => {
  const [blocks, setBlocks] = useState<ProcessedBlock[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedBlock, setSelectedBlock] = useState<ProcessedBlock | null>(null);

  useEffect(() => {
    fetchBlocks();
    const interval = setInterval(fetchBlocks, 10000);
    return () => clearInterval(interval);
  }, [limit]);

  const fetchBlocks = async () => {
    setLoading(true);
    try {
      const latestBlock = await blockchainService.getLatestBlock();
      if (latestBlock) {
        const latestBlockNumber = parseInt(latestBlock.number, 16);
        const blocksToFetch = limit || 20;
        const startBlock = Math.max(0, latestBlockNumber - blocksToFetch + 1);
        
        const blockData = await blockchainService.getBlocks(startBlock, latestBlockNumber);
        const processedBlocks = blockData.map(block => processBlock(block)).reverse();
        setBlocks(processedBlocks);
      }
    } catch (error) {
      console.error('Failed to fetch blocks:', error);
    }
    setLoading(false);
  };

  const processBlock = (block: RawBlockData): ProcessedBlock => {
    return {
      number: parseInt(block.number, 16),
      hash: block.hash,
      parentHash: block.parentHash,
      timestamp: parseInt(block.timestamp, 16),
      transactionCount: Array.isArray(block.transactions) ? block.transactions.length : 0,
      gasUsed: parseInt(block.gasUsed, 16).toLocaleString(),
      gasLimit: parseInt(block.gasLimit, 16).toLocaleString(),
      difficulty: parseInt(block.difficulty, 16).toLocaleString(),
      miner: block.miner || '0x0000000000000000000000000000000000000000',
      size: parseInt(block.size, 16),
    };
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
                <div className="text-sm font-semibold">Hash</div>
                <p className="font-mono text-sm break-all">{selectedBlock.hash}</p>
              </div>
              <div>
                <div className="text-sm font-semibold">Parent Hash</div>
                <p className="font-mono text-sm break-all">{selectedBlock.parentHash}</p>
              </div>
              <div>
                <div className="text-sm font-semibold">Miner</div>
                <p className="font-mono text-sm">{selectedBlock.miner}</p>
              </div>
              <div>
                <div className="text-sm font-semibold">Difficulty</div>
                <p className="text-sm">{selectedBlock.difficulty}</p>
              </div>
              <div>
                <div className="text-sm font-semibold">Gas Limit</div>
                <p className="text-sm">{selectedBlock.gasLimit}</p>
              </div>
              <div>
                <div className="text-sm font-semibold">Gas Used</div>
                <p className="text-sm">{selectedBlock.gasUsed}</p>
              </div>
              <div>
                <div className="text-sm font-semibold">Size</div>
                <p className="text-sm">{selectedBlock.size.toLocaleString()} bytes</p>
              </div>
              <div>
                <div className="text-sm font-semibold">Timestamp</div>
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
