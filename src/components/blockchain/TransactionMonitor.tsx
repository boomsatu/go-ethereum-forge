import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { RefreshCw, Send, Search, Hash } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import blockchainService, { TransactionData } from '@/services/blockchainService';

interface ProcessedTransaction {
  hash: string;
  from: string;
  to: string | null;
  value: string;
  gasPrice: string;
  gasLimit: string;
  nonce: number;
  blockNumber: number | null;
  status: 'pending' | 'confirmed' | 'failed';
  timestamp: number;
}

interface TransactionMonitorProps {
  limit?: number;
}

export const TransactionMonitor: React.FC<TransactionMonitorProps> = ({ limit }) => {
  const [transactions, setTransactions] = useState<ProcessedTransaction[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedTx, setSelectedTx] = useState<ProcessedTransaction | null>(null);
  const [searchHash, setSearchHash] = useState('');
  const [sendTxForm, setSendTxForm] = useState({
    from: '',
    to: '',
    value: '',
    gasPrice: '20000000000',
    gasLimit: '21000',
    data: ''
  });
  const { toast } = useToast();

  useEffect(() => {
    fetchTransactions();
    const interval = setInterval(fetchTransactions, 15000);
    return () => clearInterval(interval);
  }, [limit]);

  const fetchTransactions = async () => {
    setLoading(true);
    try {
      const latestBlock = await blockchainService.getLatestBlock();
      if (latestBlock) {
        const latestBlockNumber = parseInt(latestBlock.number, 16);
        const blocksToCheck = Math.min(10, latestBlockNumber + 1);
        const allTransactions: ProcessedTransaction[] = [];

        for (let i = Math.max(0, latestBlockNumber - blocksToCheck + 1); i <= latestBlockNumber; i++) {
          const block = await blockchainService.getBlockByNumber(i, true);
          if (block && Array.isArray(block.transactions)) {
            for (const tx of block.transactions) {
              if (typeof tx === 'object') {
                const processedTx = processTransaction(tx as TransactionData, i);
                allTransactions.push(processedTx);
              }
            }
          }
        }

        allTransactions.sort((a, b) => b.timestamp - a.timestamp);
        setTransactions(allTransactions.slice(0, limit || 50));
      }
    } catch (error) {
      console.error('Failed to fetch transactions:', error);
    }
    setLoading(false);
  };

  const processTransaction = (tx: TransactionData, blockNumber: number): ProcessedTransaction => {
    return {
      hash: tx.hash,
      from: tx.from,
      to: tx.to,
      value: (parseInt(tx.value, 16) / 1e18).toFixed(6),
      gasPrice: (parseInt(tx.gasPrice, 16) / 1e9).toFixed(0),
      gasLimit: parseInt(tx.gas, 16).toString(),
      nonce: parseInt(tx.nonce, 16),
      blockNumber: tx.blockNumber ? parseInt(tx.blockNumber, 16) : blockNumber,
      status: tx.blockHash ? 'confirmed' : 'pending',
      timestamp: Date.now() / 1000,
    };
  };

  const searchTransaction = async () => {
    if (!searchHash.trim()) {
      toast({
        title: "Invalid Hash",
        description: "Please enter a valid transaction hash",
        variant: "destructive",
      });
      return;
    }
    
    setLoading(true);
    try {
      const tx = await blockchainService.getTransaction(searchHash);
      if (tx) {
        const processedTx = processTransaction(tx, 0);
        setSelectedTx(processedTx);
        toast({
          title: "Transaction Found",
          description: `Transaction ${searchHash.slice(0, 10)}... found`,
        });
      } else {
        toast({
          title: "Transaction Not Found",
          description: "No transaction found with this hash",
          variant: "destructive",
        });
      }
    } catch (error) {
      toast({
        title: "Search Failed",
        description: "Failed to search for transaction",
        variant: "destructive",
      });
    }
    setLoading(false);
  };

  const sendTransaction = async () => {
    if (!sendTxForm.from || !sendTxForm.to || !sendTxForm.value) {
      toast({
        title: "Invalid Transaction",
        description: "Please fill in all required fields",
        variant: "destructive",
      });
      return;
    }

    try {
      const txHash = await blockchainService.sendTransaction({
        from: sendTxForm.from,
        to: sendTxForm.to,
        value: `0x${(parseFloat(sendTxForm.value) * 1e18).toString(16)}`,
        gas: `0x${parseInt(sendTxForm.gasLimit).toString(16)}`,
        gasPrice: `0x${parseInt(sendTxForm.gasPrice).toString(16)}`,
        data: sendTxForm.data || '0x',
      });

      if (txHash) {
        toast({
          title: "Transaction Sent",
          description: `Transaction ${txHash.slice(0, 10)}... submitted`,
        });
        
        setSendTxForm({
          from: '',
          to: '',
          value: '',
          gasPrice: '20000000000',
          gasLimit: '21000',
          data: ''
        });
        
        setTimeout(fetchTransactions, 2000);
      } else {
        toast({
          title: "Transaction Failed",
          description: "Failed to send transaction",
          variant: "destructive",
        });
      }
    } catch (error) {
      toast({
        title: "Transaction Failed",
        description: "Failed to send transaction",
        variant: "destructive",
      });
    }
  };

  const formatTimestamp = (timestamp: number) => {
    return new Date(timestamp * 1000).toLocaleString();
  };

  const truncateHash = (hash: string) => {
    return `${hash.slice(0, 8)}...${hash.slice(-6)}`;
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'confirmed': return 'bg-green-100 text-green-800';
      case 'pending': return 'bg-yellow-100 text-yellow-800';
      case 'failed': return 'bg-red-100 text-red-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="flex items-center space-x-2">
                <Send className="w-5 h-5" />
                <span>Transaction Monitor</span>
              </CardTitle>
              <CardDescription>
                Monitor transactions and blockchain activity
              </CardDescription>
            </div>
            <Button 
              onClick={fetchTransactions} 
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
          <div className="flex space-x-2 mb-4">
            <Input
              placeholder="Search by transaction hash..."
              value={searchHash}
              onChange={(e) => setSearchHash(e.target.value)}
              className="flex-1"
            />
            <Button onClick={searchTransaction} disabled={loading}>
              <Search className="w-4 h-4 mr-2" />
              Search
            </Button>
          </div>

          <ScrollArea className="h-[300px]">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Hash</TableHead>
                  <TableHead>From</TableHead>
                  <TableHead>To</TableHead>
                  <TableHead>Value (ETH)</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {transactions.map((tx) => (
                  <TableRow key={tx.hash}>
                    <TableCell className="font-mono text-sm">
                      {truncateHash(tx.hash)}
                    </TableCell>
                    <TableCell className="font-mono text-sm">
                      {truncateHash(tx.from)}
                    </TableCell>
                    <TableCell className="font-mono text-sm">
                      {tx.to ? truncateHash(tx.to) : 'Contract Creation'}
                    </TableCell>
                    <TableCell>{tx.value} ETH</TableCell>
                    <TableCell>
                      <Badge className={getStatusColor(tx.status)}>
                        {tx.status}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Button 
                        onClick={() => setSelectedTx(tx)}
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

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center space-x-2">
            <Send className="w-5 h-5" />
            <span>Send Transaction</span>
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="from">From Address</Label>
              <Input
                id="from"
                placeholder="0x..."
                value={sendTxForm.from}
                onChange={(e) => setSendTxForm({...sendTxForm, from: e.target.value})}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="to">To Address</Label>
              <Input
                id="to"
                placeholder="0x..."
                value={sendTxForm.to}
                onChange={(e) => setSendTxForm({...sendTxForm, to: e.target.value})}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="value">Value (ETH)</Label>
              <Input
                id="value"
                placeholder="0.0"
                type="number"
                step="0.000001"
                value={sendTxForm.value}
                onChange={(e) => setSendTxForm({...sendTxForm, value: e.target.value})}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="gasPrice">Gas Price (Wei)</Label>
              <Input
                id="gasPrice"
                placeholder="20000000000"
                type="number"
                value={sendTxForm.gasPrice}
                onChange={(e) => setSendTxForm({...sendTxForm, gasPrice: e.target.value})}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="gasLimit">Gas Limit</Label>
              <Input
                id="gasLimit"
                placeholder="21000"
                type="number"
                value={sendTxForm.gasLimit}
                onChange={(e) => setSendTxForm({...sendTxForm, gasLimit: e.target.value})}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="data">Data (Hex)</Label>
              <Input
                id="data"
                placeholder="0x..."
                value={sendTxForm.data}
                onChange={(e) => setSendTxForm({...sendTxForm, data: e.target.value})}
              />
            </div>
          </div>
          <Button onClick={sendTransaction} className="w-full">
            Send Transaction
          </Button>
        </CardContent>
      </Card>

      {selectedTx && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center space-x-2">
              <Hash className="w-5 h-5" />
              <span>Transaction Details</span>
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <div className="text-sm font-semibold">Hash</div>
                <p className="font-mono text-sm break-all">{selectedTx.hash}</p>
              </div>
              <div>
                <div className="text-sm font-semibold">Status</div>
                <Badge className={getStatusColor(selectedTx.status)}>
                  {selectedTx.status}
                </Badge>
              </div>
              <div>
                <div className="text-sm font-semibold">From</div>
                <p className="font-mono text-sm">{selectedTx.from}</p>
              </div>
              <div>
                <div className="text-sm font-semibold">To</div>
                <p className="font-mono text-sm">{selectedTx.to || 'Contract Creation'}</p>
              </div>
              <div>
                <div className="text-sm font-semibold">Value</div>
                <p className="text-sm">{selectedTx.value} ETH</p>
              </div>
              <div>
                <div className="text-sm font-semibold">Gas Price</div>
                <p className="text-sm">{selectedTx.gasPrice} Gwei</p>
              </div>
              <div>
                <div className="text-sm font-semibold">Nonce</div>
                <p className="text-sm">{selectedTx.nonce}</p>
              </div>
              <div>
                <div className="text-sm font-semibold">Block Number</div>
                <p className="text-sm">{selectedTx.blockNumber || 'Pending'}</p>
              </div>
            </div>
            <Button 
              onClick={() => setSelectedTx(null)}
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
