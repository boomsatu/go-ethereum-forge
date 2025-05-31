
import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Eye, EyeOff, Copy, RefreshCw } from "lucide-react";
import { WalletAccount } from './types';

interface WalletDetailsProps {
  selectedWallet: string;
  wallets: WalletAccount[];
  showPrivateKeys: {[key: string]: boolean};
  togglePrivateKeyVisibility: (address: string) => void;
  copyToClipboard: (text: string, label: string) => void;
  refreshBalance: (address: string) => Promise<void>;
  setSelectedWallet: (address: string) => void;
  deleteWallet: (address: string) => void;
  refreshing: string;
}

export const WalletDetails: React.FC<WalletDetailsProps> = ({
  selectedWallet,
  wallets,
  showPrivateKeys,
  togglePrivateKeyVisibility,
  copyToClipboard,
  refreshBalance,
  setSelectedWallet,
  deleteWallet,
  refreshing
}) => {
  if (!selectedWallet) return null;

  return (
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
  );
};
