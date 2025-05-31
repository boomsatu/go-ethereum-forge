
import { WalletData } from '@/services/blockchainService';

export interface WalletAccount extends Omit<WalletData, 'nonce'> {
  nonce: number;
}

export interface NewWalletForm {
  privateKey: string;
  importing: boolean;
}
