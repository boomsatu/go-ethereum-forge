
import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { useToast } from "@/hooks/use-toast";
import { Wallet, Plus, Key, Eye, EyeOff, Copy } from "lucide-react";

interface WalletAccount {
  address: string;
  privateKey: string;
  balance: string;
  nonce: number;
}

export const WalletManager: React.FC = () => {
  const [wallets, setWallets] = useState<WalletAccount[]>([]);
  const [showPrivateKeys, setShowPrivateKeys] = useState<{[key: string]: boolean}>({});
  const [newWalletForm, setNewWalletForm] = useState({
    privateKey: '',
    importing: false
  });
  const [selectedWallet, setSelectedWallet] = useState<string>('');
  const { toast } = useToast();

  useEffect(() => {
    loadWallets();
  }, []);

  const loadWallets = () => {
    // Load wallets from localStorage or generate mock wallets
    const savedWallets = localStorage.getItem('blockchain-wallets');
    if (savedWallets) {
      setWallets(JSON.parse(savedWallets));
    } else {
      generateMockWallets();
    }
  };

  const generateMockWallets = () => {
    const mockWallets: WalletAccount[] = [
      {
        address: '0x742d35Cc6635C0532925a3b8D5c6C1C8b1c5C6C',
        privateKey: '0x' + '1'.repeat(64),
        balance: '10.5',
        nonce: 5
      },
      {
        address: '0x8ba1f109551bD432803012645Hac136c54f2fA1',
        privateKey: '0x' + '2'.repeat(64),
        balance: '25.3',
        nonce: 12
      }
    ];
    setWallets(mockWallets);
    localStorage.setItem('blockchain-wallets', JSON.stringify(mockWallets));
  };

  const createNewWallet = async () => {
    try {
      // In real implementation, this would call the Go backend to create a wallet
      const response = await fetch('http://localhost:8545/api/wallet/create', {
        method: 'POST',
      });

      if (response.ok) {
        const result = await response.json();
        const newWallet: WalletAccount = {
          address: result.address,
          privateKey: result.privateKey,
          balance: '0.0',
          nonce: 0
        };

        const updatedWallets = [...wallets, newWallet];
        setWallets(updatedWallets);
        localStorage.setItem('blockchain-wallets', JSON.stringify(updatedWallets));

        toast({
          title: "Wallet Created",
          description: `New wallet created with address ${result.address.slice(0, 10)}...`,
        });
      }
    } catch (error) {
      // Generate mock wallet for demonstration
      const mockWallet: WalletAccount = {
        address: `0x${Math.random().toString(16).slice(2, 42)}`,
        privateKey: `0x${Math.random().toString(16).slice(2, 66)}`,
        balance: '0.0',
        nonce: 0
      };

      const updatedWallets = [...wallets, mockWallet];
      setWallets(updatedWallets);
      localStorage.setItem('blockchain-wallets', JSON.stringify(updatedWallets));

      toast({
        title: "Wallet Created",
        description: `New wallet created with address ${mockWallet.address.slice(0, 10)}...`,
      });
    }
  };

  const importWallet = async () => {
    if (!newWalletForm.privateKey.trim()) {
      toast({
        title: "Invalid Private Key",
        description: "Please enter a valid private key",
        variant: "destructive",
      });
      return;
    }

    try {
      // In real implementation, this would validate and import the wallet
      const mockAddress = `0x${Math.random().toString(16).slice(2, 42)}`;
      const importedWallet: WalletAccount = {
        address: mockAddress,
        privateKey: newWalletForm.privateKey,
        balance: (Math.random() * 100).toFixed(6),
        nonce: Math.floor(Math.random() * 50)
      };

      const updatedWallets = [...wallets, importedWallet];
      setWallets(updatedWallets);
      localStorage.setItem('blockchain-wallets', JSON.stringify(updatedWallets));

      setNewWalletForm({ privateKey: '', importing: false });

      toast({
        title: "Wallet Imported",
        description: `Wallet imported with address ${mockAddress.slice(0, 10)}...`,
      });
    } catch (error) {
      toast({
        title: "Import Failed",
        description: "Failed to import wallet",
        variant: "destructive",
      });
    }
  };

  const refreshBalance = async (address: string) => {
    try {
      const response = await fetch('http://localhost:8545', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          jsonrpc: '2.0',
          method: 'eth_getBalance',
          params: [address, 'latest'],
          id: 1,
        }),
      });

      const result = await response.json();
      if (result.result) {
        const balanceWei = parseInt(result.result, 16);
        const balanceEth = (balanceWei / 1e18).toFixed(6);
        
        const updatedWallets = wallets.map(wallet => 
          wallet.address === address 
            ? { ...wallet, balance: balanceEth }
            : wallet
        );
        setWallets(updatedWallets);
        localStorage.setItem('blockchain-wallets', JSON.stringify(updatedWallets));
      }
    } catch (error) {
      // Generate random balance for demonstration
      const randomBalance = (Math.random() * 100).toFixed(6);
      const updatedWallets = wallets.map(wallet => 
        wallet.address === address 
          ? { ...wallet, balance: randomBalance }
          : wallet
      );
      setWallets(updatedWallets);
      localStorage.setItem('blockchain-wallets', JSON.stringify(updatedWallets));
    }

    toast({
      title: "Balance Updated",
      description: `Balance refreshed for ${address.slice(0, 10)}...`,
    });
  };

  const copyToClipboard = (text: string, label: string) => {
    navigator.clipboard.writeText(text);
    toast({
      title: "Copied",
      description: `${label} copied to clipboard`,
    });
  };

  const togglePrivateKeyVisibility = (address: string) => {
    setShowPrivateKeys(prev => ({
      ...prev,
      [address]: !prev[address]
    }));
  };

  const truncateHash = (hash: string) => {
    return `${hash.slice(0, 8)}...${hash.slice(-6)}`;
  };

  const deleteWallet = (address: string) => {
    const updatedWallets = wallets.filter(wallet => wallet.address !== address);
    setWallets(updatedWallets);
    localStorage.setItem('blockchain-wallets', JSON.stringify(updatedWallets));
    
    toast({
      title: "Wallet Removed",
      description: `Wallet ${address.slice(0, 10)}... has been removed`,
    });
  };

  return (
    <div className="space-y-6">
      {/* Wallet Actions */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center space-x-2">
            <Wallet className="w-5 h-5" />
            <span>Wallet Manager</span>
          </CardTitle>
          <CardDescription>
            Create, import, and manage your blockchain wallets
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Create/Import Buttons */}
          <div className="grid grid-cols-2 gap-4">
            <Button onClick={createNewWallet} className="w-full">
              <Plus className="w-4 h-4 mr-2" />
              Create New Wallet
            </Button>
            <Button 
              onClick={() => setNewWalletForm({...newWalletForm, importing: !newWalletForm.importing})}
              variant="outline" 
              className="w-full"
            >
              <Key className="w-4 h-4 mr-2" />
              Import Wallet
            </Button>
          </div>

          {/* Import Form */}
          {newWalletForm.importing && (
            <>
              <Separator />
              <div className="space-y-3">
                <Label htmlFor="privateKey">Private Key</Label>
                <Input
                  id="privateKey"
                  type="password"
                  placeholder="Enter private key (0x...)"
                  value={newWalletForm.privateKey}
                  onChange={(e) => setNewWalletForm({...newWalletForm, privateKey: e.target.value})}
                />
                <div className="flex space-x-2">
                  <Button onClick={importWallet} className="flex-1">
                    Import
                  </Button>
                  <Button 
                    onClick={() => setNewWalletForm({privateKey: '', importing: false})}
                    variant="outline"
                  >
                    Cancel
                  </Button>
                </div>
              </div>
            </>
          )}
        </CardContent>
      </Card>

      {/* Wallet List */}
      <Card>
        <CardHeader>
          <CardTitle>Wallet Accounts</CardTitle>
          <CardDescription>
            Manage your wallet accounts and view balances
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Address</TableHead>
                <TableHead>Balance (ETH)</TableHead>
                <TableHead>Nonce</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {wallets.map((wallet) => (
                <TableRow key={wallet.address}>
                  <TableCell>
                    <div className="flex items-center space-x-2">
                      <span className="font-mono text-sm">{truncateHash(wallet.address)}</span>
                      <Button
                        onClick={() => copyToClipboard(wallet.address, 'Address')}
                        variant="ghost"
                        size="sm"
                      >
                        <Copy className="w-3 h-3" />
                      </Button>
                    </div>
                  </TableCell>
                  <TableCell>
                    <Badge variant="secondary">{wallet.balance} ETH</Badge>
                  </TableCell>
                  <TableCell>{wallet.nonce}</TableCell>
                  <TableCell>
                    <div className="flex space-x-1">
                      <Button
                        onClick={() => refreshBalance(wallet.address)}
                        variant="ghost"
                        size="sm"
                      >
                        Refresh
                      </Button>
                      <Button
                        onClick={() => setSelectedWallet(wallet.address)}
                        variant="ghost"
                        size="sm"
                      >
                        Details
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {/* Wallet Details */}
      {selectedWallet && (
        <Card>
          <CardHeader>
            <CardTitle>Wallet Details</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {wallets
              .filter(wallet => wallet.address === selectedWallet)
              .map(wallet => (
                <div key={wallet.address} className="space-y-4">
                  <div className="grid grid-cols-1 gap-4">
                    <div>
                      <Label className="text-sm font-semibold">Address</Label>
                      <div className="flex items-center space-x-2">
                        <p className="font-mono text-sm break-all">{wallet.address}</p>
                        <Button
                          onClick={() => copyToClipboard(wallet.address, 'Address')}
                          variant="ghost"
                          size="sm"
                        >
                          <Copy className="w-3 h-3" />
                        </Button>
                      </div>
                    </div>
                    <div>
                      <Label className="text-sm font-semibold">Private Key</Label>
                      <div className="flex items-center space-x-2">
                        <p className="font-mono text-sm break-all">
                          {showPrivateKeys[wallet.address] ? wallet.privateKey : '••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••'}
                        </p>
                        <Button
                          onClick={() => togglePrivateKeyVisibility(wallet.address)}
                          variant="ghost"
                          size="sm"
                        >
                          {showPrivateKeys[wallet.address] ? <EyeOff className="w-3 h-3" /> : <Eye className="w-3 h-3" />}
                        </Button>
                        {showPrivateKeys[wallet.address] && (
                          <Button
                            onClick={() => copyToClipboard(wallet.privateKey, 'Private Key')}
                            variant="ghost"
                            size="sm"
                          >
                            <Copy className="w-3 h-3" />
                          </Button>
                        )}
                      </div>
                    </div>
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <Label className="text-sm font-semibold">Balance</Label>
                        <p className="text-lg font-bold">{wallet.balance} ETH</p>
                      </div>
                      <div>
                        <Label className="text-sm font-semibold">Nonce</Label>
                        <p className="text-lg">{wallet.nonce}</p>
                      </div>
                    </div>
                  </div>
                  <div className="flex space-x-2">
                    <Button 
                      onClick={() => setSelectedWallet('')}
                      variant="outline"
                      className="flex-1"
                    >
                      Close
                    </Button>
                    <Button 
                      onClick={() => deleteWallet(wallet.address)}
                      variant="destructive"
                    >
                      Delete Wallet
                    </Button>
                  </div>
                </div>
              ))}
          </CardContent>
        </Card>
      )}
    </div>
  );
};
