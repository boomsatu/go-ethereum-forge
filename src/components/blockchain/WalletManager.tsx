
import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { useToast } from "@/hooks/use-toast";
import { Wallet, Plus, Key, Eye, EyeOff, Copy, RefreshCw } from "lucide-react";
import { blockchainService, WalletData } from '@/services/blockchainService';

interface WalletAccount extends WalletData {
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
  const [loading, setLoading] = useState(false);
  const [refreshing, setRefreshing] = useState<string>('');
  const { toast } = useToast();

  useEffect(() => {
    loadWallets();
  }, []);

  const loadWallets = () => {
    const savedWallets = localStorage.getItem('blockchain-wallets');
    if (savedWallets) {
      const parsedWallets = JSON.parse(savedWallets);
      setWallets(parsedWallets);
      // Refresh balances for all wallets
      parsedWallets.forEach((wallet: WalletAccount) => {
        refreshBalance(wallet.address);
      });
    }
  };

  const saveWallets = (walletsToSave: WalletAccount[]) => {
    localStorage.setItem('blockchain-wallets', JSON.stringify(walletsToSave));
  };

  const createNewWallet = async () => {
    setLoading(true);
    try {
      // Create wallet using real blockchain service
      const newWallet = await blockchainService.createWallet();
      if (newWallet) {
        const nonce = await blockchainService.getNonce(newWallet.address);
        const walletWithNonce: WalletAccount = {
          ...newWallet,
          nonce
        };

        const updatedWallets = [...wallets, walletWithNonce];
        setWallets(updatedWallets);
        saveWallets(updatedWallets);

        toast({
          title: "Wallet Created",
          description: `New wallet created with address ${newWallet.address.slice(0, 10)}...`,
        });

        await refreshBalance(newWallet.address);
      } else {
        toast({
          title: "Creation Failed",
          description: "Failed to create new wallet - blockchain node may not be running",
          variant: "destructive",
        });
      }
    } catch (error) {
      console.error('Wallet creation error:', error);
      toast({
        title: "Creation Failed",
        description: "Failed to create new wallet - check if blockchain node is running",
        variant: "destructive",
      });
    }
    setLoading(false);
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

    setLoading(true);
    try {
      // Import wallet using real blockchain service
      const importedWallet = await blockchainService.importWallet(newWalletForm.privateKey);
      if (importedWallet) {
        const nonce = await blockchainService.getNonce(importedWallet.address);
        const walletWithNonce: WalletAccount = {
          ...importedWallet,
          nonce
        };

        const updatedWallets = [...wallets, walletWithNonce];
        setWallets(updatedWallets);
        saveWallets(updatedWallets);

        setNewWalletForm({ privateKey: '', importing: false });

        toast({
          title: "Wallet Imported",
          description: `Wallet imported with address ${importedWallet.address.slice(0, 10)}...`,
        });

        await refreshBalance(importedWallet.address);
      } else {
        toast({
          title: "Import Failed",
          description: "Failed to import wallet - check private key format or blockchain connection",
          variant: "destructive",
        });
      }
    } catch (error) {
      console.error('Wallet import error:', error);
      toast({
        title: "Import Failed",
        description: "Failed to import wallet - check private key format or blockchain connection",
        variant: "destructive",
      });
    }
    setLoading(false);
  };

  const refreshBalance = async (address: string) => {
    setRefreshing(address);
    try {
      // Get real balance and nonce from blockchain
      const [balance, nonce] = await Promise.all([
        blockchainService.getBalance(address),
        blockchainService.getNonce(address)
      ]);
      
      const updatedWallets = wallets.map(wallet => 
        wallet.address === address 
          ? { ...wallet, balance, nonce }
          : wallet
      );
      setWallets(updatedWallets);
      saveWallets(updatedWallets);

      toast({
        title: "Balance Updated",
        description: `Wallet ${address.slice(0, 10)}... balance: ${balance} ETH`,
      });
    } catch (error) {
      console.error('Failed to refresh balance:', error);
      toast({
        title: "Refresh Failed",
        description: "Failed to refresh balance - check blockchain connection",
        variant: "destructive",
      });
    }
    setRefreshing('');
  };

  const refreshAllBalances = async () => {
    setLoading(true);
    try {
      const updatedWallets = await Promise.all(
        wallets.map(async (wallet) => {
          try {
            const [balance, nonce] = await Promise.all([
              blockchainService.getBalance(wallet.address),
              blockchainService.getNonce(wallet.address)
            ]);
            return { ...wallet, balance, nonce };
          } catch (error) {
            console.error(`Failed to refresh wallet ${wallet.address}:`, error);
            return wallet;
          }
        })
      );
      
      setWallets(updatedWallets);
      saveWallets(updatedWallets);

      toast({
        title: "Balances Refreshed",
        description: "All wallet balances have been updated",
      });
    } catch (error) {
      console.error('Failed to refresh all balances:', error);
      toast({
        title: "Refresh Failed",
        description: "Failed to refresh balances - check blockchain connection",
        variant: "destructive",
      });
    }
    setLoading(false);
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
    saveWallets(updatedWallets);
    
    if (selectedWallet === address) {
      setSelectedWallet('');
    }
    
    toast({
      title: "Wallet Removed",
      description: `Wallet ${address.slice(0, 10)}... has been removed`,
    });
  };

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center space-x-2">
            <Wallet className="w-5 h-5" />
            <span>Wallet Manager</span>
          </CardTitle>
          <CardDescription>
            Create, import, and manage your blockchain wallets with real data
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <Button onClick={createNewWallet} className="w-full" disabled={loading}>
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
                  <Button onClick={importWallet} className="flex-1" disabled={loading}>
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

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Wallet Accounts</CardTitle>
              <CardDescription>
                Manage your wallet accounts and view real balances from blockchain
              </CardDescription>
            </div>
            {wallets.length > 0 && (
              <Button 
                onClick={refreshAllBalances} 
                disabled={loading}
                variant="outline"
                size="sm"
              >
                <RefreshCw className={`w-4 h-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
                Refresh All
              </Button>
            )}
          </div>
        </CardHeader>
        <CardContent>
          {wallets.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              No wallets found. Create or import a wallet to get started.
            </div>
          ) : (
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
                          disabled={refreshing === wallet.address}
                        >
                          <RefreshCw className={`w-3 h-3 ${refreshing === wallet.address ? 'animate-spin' : ''}`} />
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
          )}
        </CardContent>
      </Card>

      {selectedWallet && (
        <Card>
          <CardHeader>
            <CardTitle>Wallet Details</CardTitle>
            <CardDescription>Real wallet data from blockchain</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {wallets
              .filter(wallet => wallet.address === selectedWallet)
              .map(wallet => (
                <div key={wallet.address} className="space-y-4">
                  <div className="grid grid-cols-1 gap-4">
                    <div>
                      <div className="text-sm font-semibold">Address</div>
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
                      <div className="text-sm font-semibold">Private Key</div>
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
                        <div className="text-sm font-semibold">Balance (Real)</div>
                        <p className="text-lg font-bold">{wallet.balance} ETH</p>
                      </div>
                      <div>
                        <div className="text-sm font-semibold">Nonce (Real)</div>
                        <p className="text-lg">{wallet.nonce}</p>
                      </div>
                    </div>
                  </div>
                  <div className="flex space-x-2">
                    <Button 
                      onClick={() => refreshBalance(wallet.address)}
                      variant="outline"
                      disabled={refreshing === wallet.address}
                    >
                      <RefreshCw className={`w-4 h-4 mr-2 ${refreshing === wallet.address ? 'animate-spin' : ''}`} />
                      Refresh Data
                    </Button>
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
