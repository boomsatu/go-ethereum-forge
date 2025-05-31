
import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { Wallet, Plus, Key } from "lucide-react";
import { NewWalletForm } from './types';

interface WalletCreationFormProps {
  newWalletForm: NewWalletForm;
  setNewWalletForm: (form: NewWalletForm) => void;
  createNewWallet: () => Promise<void>;
  importWallet: () => Promise<void>;
  loading: boolean;
}

export const WalletCreationForm: React.FC<WalletCreationFormProps> = ({
  newWalletForm,
  setNewWalletForm,
  createNewWallet,
  importWallet,
  loading
}) => {
  return (
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
  );
};
