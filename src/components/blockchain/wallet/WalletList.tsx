
import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Copy, RefreshCw } from "lucide-react";
import { WalletAccount } from './types';

interface WalletListProps {
  wallets: WalletAccount[];
  refreshBalance: (address: string) => Promise<void>;
  refreshAllBalances: () => Promise<void>;
  copyToClipboard: (text: string, label: string) => void;
  setSelectedWallet: (address: string) => void;
  refreshing: string;
  loading: boolean;
}

export const WalletList: React.FC<WalletListProps> = ({
  wallets,
  refreshBalance,
  refreshAllBalances,
  copyToClipboard,
  setSelectedWallet,
  refreshing,
  loading
}) => {
  const truncateHash = (hash: string) => {
    return `${hash.slice(0, 8)}...${hash.slice(-6)}`;
  };

  return (
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
  );
};
