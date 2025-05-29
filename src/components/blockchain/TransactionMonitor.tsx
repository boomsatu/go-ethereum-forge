
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

interface Transaction {
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
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedTx, setSelectedTx] = useState<Transaction | null>(null);
  const [searchHash, setSearchHash] = useState('');
  const [sendTxForm, setSendTxForm] = useState({
    to: '',
    value: '',
    gasPrice: '20',
    gasLimit: '21000',
    data: ''
  });
  const { toast } = useToast();

  useEffect(() => {
    fetchTransactions();
    const interval = setInterval(fetchTransactions, 15000); // Refresh every 15 seconds
    return () => clearInterval(interval);
  }, [limit]);

  const fetchTransactions = async () => {
    setLoading(true);
    try {
      // In a real implementation, this would fetch from the blockchain
      // For now, we'll generate mock data
      generateMockTransactions();
    } catch (error) {
      console.error('Failed to fetch transactions:', error);
      generateMockTransactions();
    }
    setLoading(false);
  };

  const generateMockTransactions = () => {
    const mockTxs: Transaction[] = [];
    const statuses: ('pending' | 'confirmed' | 'failed')[] = ['pending', 'confirmed', 'failed'];
    const now = Date.now();
    
    for (let i = 0; i < (limit || 15); i++) {
      const status = statuses[Math.floor(Math.random() * 3)];
      mockTxs.push({
        hash: `0x${Math.random().toString(16).slice(2, 66)}`,
        from: `0x${Math.random().toString(16).slice(2, 42)}`,
        to: Math.random() > 0.1 ? `0x${Math.random().toString(16).slice(2, 42)}` : null,
        value: (Math.random() * 10).toFixed(6),
        gasPrice: (20 + Math.random() * 80).toFixed(0),
        gasLimit: '21000',
        nonce: Math.floor(Math.random() * 100),
        blockNumber: status === 'confirmed' ? Math.floor(Math.random() * 1000) + 1 : null,
        status,
        timestamp: Math.floor((now - i * 30000) / 1000),
      });
    }
    
    setTransactions(mockTxs);
  };

  const searchTransaction = async () => {
    if (!searchHash.trim()) return;
    
    setLoading(true);
    try {
      const response = await fetch('http://localhost:8545', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          jsonrpc: '2.0',
          method: 'eth_getTransactionByHash',
          params: [searchHash],
          id: 1,
        }),
      });

      const result = await response.json();
      if (result.result) {
        // Process found transaction
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
    try {
      // In real implementation, this would create and send a transaction
      const mockTxHash = `0x${Math.random().toString(16).slice(2, 66)}`;
      
      const newTx: Transaction = {
        hash: mockTxHash,
        from: '0x742d35Cc6635C0532925a3b8D5c6C1C8b1c5C6C', // Mock sender
        to: sendTxForm.to || null,
        value: sendTxForm.value,
        gasPrice: sendTxForm.gasPrice,
        gasLimit: sendTxForm.gasLimit,
        nonce: Math.floor(Math.random() * 100),
        blockNumber: null,
        status: 'pending',
        timestamp: Math.floor(Date.now() / 1000),
      };

      setTransactions(prev => [newTx, ...prev]);
      
      toast({
        title: "Transaction Sent",
        description: `Transaction ${mockTxHash.slice(0, 10)}... submitted to mempool`,
      });

      // Reset form
      setSendTxForm({
        to: '',
        value: '',
        gasPrice: '20',
        gasLimit: '21000',
        data: ''
      });
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
                Monitor transactions and mempool activity
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
          {/* Search Transaction */}
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

      {/* Send Transaction Form */}
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
              <Label htmlFor="gasPrice">Gas Price (Gwei)</Label>
              <Input
                id="gasPrice"
                placeholder="20"
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
          <Button onClick={sendTransaction} className="w-full">
            Send Transaction
          </Button>
        </CardContent>
      </Card>

      {/* Transaction Details */}
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
                <Label className="text-sm font-semibold">Hash</Label>
                <p className="font-mono text-sm break-all">{selectedTx.hash}</p>
              </div>
              <div>
                <Label className="text-sm font-semibold">Status</Label>
                <Badge className={getStatusColor(selectedTx.status)}>
                  {selectedTx.status}
                </Badge>
              </div>
              <div>
                <Label className="text-sm font-semibold">From</Label>
                <p className="font-mono text-sm">{selectedTx.from}</p>
              </div>
              <div>
                <Label className="text-sm font-semibold">To</Label>
                <p className="font-mono text-sm">{selectedTx.to || 'Contract Creation'}</p>
              </div>
              <div>
                <Label className="text-sm font-semibold">Value</Label>
                <p className="text-sm">{selectedTx.value} ETH</p>
              </div>
              <div>
                <Label className="text-sm font-semibold">Gas Price</Label>
                <p className="text-sm">{selectedTx.gasPrice} Gwei</p>
              </div>
              <div>
                <Label className="text-sm font-semibold">Nonce</Label>
                <p className="text-sm">{selectedTx.nonce}</p>
              </div>
              <div>
                <Label className="text-sm font-semibold">Block Number</Label>
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
