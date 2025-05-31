
import React, { useState, useEffect } from 'react';
import { useToast } from "@/hooks/use-toast";
import blockchainService from '@/services/blockchainService';
import { WalletCreationForm } from './wallet/WalletCreationForm';
import { WalletList } from './wallet/WalletList';
import { WalletDetails } from './wallet/WalletDetails';
import { WalletAccount, NewWalletForm } from './wallet/types';

export const WalletManager: React.FC = () => {
  const [wallets, setWallets] = useState<WalletAccount[]>([]);
  const [showPrivateKeys, setShowPrivateKeys] = useState<{[key: string]: boolean}>({});
  const [newWalletForm, setNewWalletForm] = useState<NewWalletForm>({
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
      <WalletCreationForm
        newWalletForm={newWalletForm}
        setNewWalletForm={setNewWalletForm}
        createNewWallet={createNewWallet}
        importWallet={importWallet}
        loading={loading}
      />

      <WalletList
        wallets={wallets}
        refreshBalance={refreshBalance}
        refreshAllBalances={refreshAllBalances}
        copyToClipboard={copyToClipboard}
        setSelectedWallet={setSelectedWallet}
        refreshing={refreshing}
        loading={loading}
      />

      <WalletDetails
        selectedWallet={selectedWallet}
        wallets={wallets}
        showPrivateKeys={showPrivateKeys}
        togglePrivateKeyVisibility={togglePrivateKeyVisibility}
        copyToClipboard={copyToClipboard}
        refreshBalance={refreshBalance}
        setSelectedWallet={setSelectedWallet}
        deleteWallet={deleteWallet}
        refreshing={refreshing}
      />
    </div>
  );
};
