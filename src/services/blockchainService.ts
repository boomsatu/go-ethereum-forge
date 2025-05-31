
export interface BlockchainConfig {
  host: string;
  port: number;
  chainId: string;
  gasLimit: string;
  dataDir: string;
}

export interface BlockData {
  number: string;
  hash: string;
  parentHash: string;
  timestamp: string;
  transactionCount: number;
  gasUsed: string;
  gasLimit: string;
  difficulty: string;
  miner: string;
  size: string;
  transactions: string[] | TransactionData[];
}

export interface TransactionData {
  hash: string;
  from: string;
  to: string | null;
  value: string;
  gasPrice: string;
  gas: string;
  nonce: string;
  blockHash: string | null;
  blockNumber: string | null;
  transactionIndex: string | null;
  input: string;
}

export interface WalletData {
  address: string;
  privateKey: string;
  publicKey?: string;
  balance: string;
  balanceEth?: string;
  nonce?: string;
  valid?: boolean;
}

export interface NetworkStats {
  chainId: string;
  networkId: string;
  blockHeight: number;
  peerCount: number;
  difficulty: string;
  hashRate: string;
}

export interface MiningStats {
  isActive: boolean;
  hashRate: number;
  blocksFound: number;
  difficulty: string;
}

export interface NodeMetrics {
  uptime: number;
  memoryUsage: number;
  diskUsage: number;
  cpuUsage: number;
  blockCount: number;
  transactionCount: number;
  peersConnected: number;
}

export interface GenesisAllocation {
  address: string;
  balance: string;
  balanceEth: string;
}

class BlockchainService {
  private baseUrl: string;

  constructor(baseUrl: string = 'http://localhost:8545') {
    this.baseUrl = baseUrl;
  }

  // Node Control
  async startNode(config: Partial<BlockchainConfig>): Promise<boolean> {
    try {
      const response = await fetch(`${this.baseUrl}/api/admin/start`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(config),
      });
      return response.ok;
    } catch (error) {
      console.error('Failed to start node:', error);
      return false;
    }
  }

  async stopNode(): Promise<boolean> {
    try {
      const response = await fetch(`${this.baseUrl}/api/admin/stop`, {
        method: 'POST',
      });
      return response.ok;
    } catch (error) {
      console.error('Failed to stop node:', error);
      return false;
    }
  }

  async getNodeStatus(): Promise<{ status: string; config: BlockchainConfig } | null> {
    try {
      const response = await fetch(`${this.baseUrl}/api/admin/status`);
      if (response.ok) {
        return await response.json();
      }
      return null;
    } catch (error) {
      console.error('Failed to get node status:', error);
      return null;
    }
  }

  // Mining Control
  async startMining(minerAddress: string, threads: number = 1): Promise<boolean> {
    try {
      const response = await fetch(`${this.baseUrl}/api/mining/start`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ minerAddress, threads }),
      });
      return response.ok;
    } catch (error) {
      console.error('Failed to start mining:', error);
      return false;
    }
  }

  async stopMining(): Promise<boolean> {
    try {
      const response = await fetch(`${this.baseUrl}/api/mining/stop`, {
        method: 'POST',
      });
      return response.ok;
    } catch (error) {
      console.error('Failed to stop mining:', error);
      return false;
    }
  }

  async getMiningStats(): Promise<MiningStats | null> {
    try {
      const response = await fetch(`${this.baseUrl}/api/mining/stats`);
      if (response.ok) {
        return await response.json();
      }
      return null;
    } catch (error) {
      console.error('Failed to get mining stats:', error);
      return null;
    }
  }

  async mineBlock(minerAddress: string): Promise<{ blockNumber: number; hash: string } | null> {
    try {
      const response = await fetch(`${this.baseUrl}/api/mining/mine-block`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ minerAddress }),
      });
      if (response.ok) {
        return await response.json();
      }
      return null;
    } catch (error) {
      console.error('Failed to mine block:', error);
      return null;
    }
  }

  // Blockchain Data
  async getLatestBlock(): Promise<BlockData | null> {
    try {
      const response = await this.rpcCall('eth_getBlockByNumber', ['latest', true]);
      return response.result;
    } catch (error) {
      console.error('Failed to get latest block:', error);
      return null;
    }
  }

  async getBlockByNumber(blockNumber: number, fullTx: boolean = false): Promise<BlockData | null> {
    try {
      const response = await this.rpcCall('eth_getBlockByNumber', [`0x${blockNumber.toString(16)}`, fullTx]);
      return response.result;
    } catch (error) {
      console.error('Failed to get block by number:', error);
      return null;
    }
  }

  async getBlockByHash(hash: string, fullTx: boolean = false): Promise<BlockData | null> {
    try {
      const response = await this.rpcCall('eth_getBlockByHash', [hash, fullTx]);
      return response.result;
    } catch (error) {
      console.error('Failed to get block by hash:', error);
      return null;
    }
  }

  async getTransaction(hash: string): Promise<TransactionData | null> {
    try {
      const response = await this.rpcCall('eth_getTransactionByHash', [hash]);
      return response.result;
    } catch (error) {
      console.error('Failed to get transaction:', error);
      return null;
    }
  }

  async getTransactionReceipt(hash: string): Promise<any> {
    try {
      const response = await this.rpcCall('eth_getTransactionReceipt', [hash]);
      return response.result;
    } catch (error) {
      console.error('Failed to get transaction receipt:', error);
      return null;
    }
  }

  // Wallet Management dengan validasi yang lebih baik
  async createWallet(): Promise<WalletData | null> {
    try {
      const response = await fetch(`${this.baseUrl}/api/wallet/create`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
      });
      if (response.ok) {
        const wallet = await response.json();
        console.log('Created wallet:', {
          address: wallet.address,
          privateKeyLength: wallet.privateKey?.length,
          hasPublicKey: !!wallet.publicKey
        });
        return wallet;
      }
      return null;
    } catch (error) {
      console.error('Failed to create wallet:', error);
      return null;
    }
  }

  async importWallet(privateKey: string): Promise<WalletData | null> {
    try {
      // Validate private key format
      const cleanPrivateKey = privateKey.startsWith('0x') ? privateKey.slice(2) : privateKey;
      if (cleanPrivateKey.length !== 64) {
        throw new Error('Private key must be 64 hex characters');
      }

      const response = await fetch(`${this.baseUrl}/api/wallet/import`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ privateKey: cleanPrivateKey }),
      });
      
      if (response.ok) {
        const wallet = await response.json();
        console.log('Imported wallet:', {
          address: wallet.address,
          privateKeyLength: wallet.privateKey?.length,
          valid: wallet.valid
        });
        return wallet;
      } else {
        const error = await response.text();
        console.error('Import failed:', error);
        return null;
      }
    } catch (error) {
      console.error('Failed to import wallet:', error);
      return null;
    }
  }

  async checkBalance(address: string): Promise<WalletData | null> {
    try {
      const response = await fetch(`${this.baseUrl}/api/wallet/balance?address=${encodeURIComponent(address)}`);
      if (response.ok) {
        return await response.json();
      }
      return null;
    } catch (error) {
      console.error('Failed to check balance:', error);
      return null;
    }
  }

  async sendTransaction(transaction: {
    from: string;
    to?: string;
    value?: string;
    gas?: string;
    gasPrice?: string;
    data?: string;
    privateKey: string;
  }): Promise<{ hash: string; success: boolean } | null> {
    try {
      const response = await fetch(`${this.baseUrl}/api/wallet/send`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          from: transaction.from,
          to: transaction.to || '',
          value: transaction.value || '0x0',
          gasLimit: transaction.gas || '0x5208',
          gasPrice: transaction.gasPrice || '0x4A817C800',
          data: transaction.data || '0x',
          privateKey: transaction.privateKey,
        }),
      });
      
      if (response.ok) {
        return await response.json();
      } else {
        const error = await response.text();
        console.error('Send transaction failed:', error);
        return null;
      }
    } catch (error) {
      console.error('Failed to send transaction:', error);
      return null;
    }
  }

  async getBalance(address: string): Promise<string> {
    try {
      const response = await this.rpcCall('eth_getBalance', [address, 'latest']);
      if (response.result) {
        const wei = parseInt(response.result, 16);
        return (wei / 1e18).toFixed(6); // Convert wei to ETH
      }
      return '0';
    } catch (error) {
      console.error('Failed to get balance:', error);
      return '0';
    }
  }

  async getNonce(address: string): Promise<number> {
    try {
      const response = await this.rpcCall('eth_getTransactionCount', [address, 'latest']);
      return parseInt(response.result, 16);
    } catch (error) {
      console.error('Failed to get nonce:', error);
      return 0;
    }
  }

  // Network Information
  async getNetworkStats(): Promise<NetworkStats | null> {
    try {
      const [chainId, networkId, blockNumber] = await Promise.all([
        this.rpcCall('eth_chainId', []),
        this.rpcCall('net_version', []),
        this.rpcCall('eth_blockNumber', []),
      ]);

      const statsResponse = await fetch(`${this.baseUrl}/api/network/stats`);
      const stats = statsResponse.ok ? await statsResponse.json() : {};

      return {
        chainId: chainId.result ? parseInt(chainId.result, 16).toString() : '1337',
        networkId: networkId.result || '1337',
        blockHeight: blockNumber.result ? parseInt(blockNumber.result, 16) : 0,
        peerCount: stats.peerCount || 0,
        difficulty: stats.difficulty || '1024',
        hashRate: stats.hashRate || '0',
      };
    } catch (error) {
      console.error('Failed to get network stats:', error);
      return null;
    }
  }

  async getPeers(): Promise<any[]> {
    try {
      const response = await fetch(`${this.baseUrl}/api/network/peers`);
      if (response.ok) {
        return await response.json();
      }
      return [];
    } catch (error) {
      console.error('Failed to get peers:', error);
      return [];
    }
  }

  // Genesis Configuration
  async getGenesisAllocations(): Promise<GenesisAllocation[]> {
    try {
      // Hardcoded genesis allocations for now
      const allocations = [
        { address: '0x742d35Cc6635C0532925a3b8D5c6C1C8b1c5C6C7', balance: '1000000000000000000000' },
        { address: '0x8ba1f109551bD432803012645Hac136c46C01C23', balance: '5000000000000000000000' },
        { address: '0x1234567890123456789012345678901234567890', balance: '10000000000000000000000' },
        { address: '0xabcdefabcdefabcdefabcdefabcdefabcdefabcd', balance: '500000000000000000000' },
      ];

      return allocations.map(alloc => ({
        address: alloc.address,
        balance: alloc.balance,
        balanceEth: (parseInt(alloc.balance) / 1e18).toString()
      }));
    } catch (error) {
      console.error('Failed to get genesis allocations:', error);
      return [];
    }
  }

  // Metrics
  async getMetrics(): Promise<NodeMetrics | null> {
    try {
      const response = await fetch(`${this.baseUrl}/api/metrics`);
      if (response.ok) {
        return await response.json();
      }
      return null;
    } catch (error) {
      console.error('Failed to get metrics:', error);
      return null;
    }
  }

  async getHealthCheck(): Promise<{ status: string; timestamp: number } | null> {
    try {
      const response = await fetch(`${this.baseUrl}/health`);
      if (response.ok) {
        const data = await response.json();
        return { ...data, timestamp: Date.now() };
      }
      return null;
    } catch (error) {
      console.error('Failed to get health check:', error);
      return null;
    }
  }

  // Utility functions
  validateAddress(address: string): boolean {
    if (!address) return false;
    const cleanAddress = address.startsWith('0x') ? address.slice(2) : address;
    return cleanAddress.length === 40 && /^[0-9a-fA-F]+$/.test(cleanAddress);
  }

  validatePrivateKey(privateKey: string): boolean {
    if (!privateKey) return false;
    const cleanKey = privateKey.startsWith('0x') ? privateKey.slice(2) : privateKey;
    return cleanKey.length === 64 && /^[0-9a-fA-F]+$/.test(cleanKey);
  }

  formatWei(wei: string | number): string {
    const weiValue = typeof wei === 'string' ? parseInt(wei, 16) : wei;
    return (weiValue / 1e18).toFixed(6);
  }

  parseEther(eth: string): string {
    const weiValue = parseFloat(eth) * 1e18;
    return '0x' + Math.floor(weiValue).toString(16);
  }

  // Private RPC helper
  private async rpcCall(method: string, params: any[]): Promise<any> {
    const response = await fetch(this.baseUrl, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        jsonrpc: '2.0',
        method,
        params,
        id: Date.now(),
      }),
    });

    if (!response.ok) {
      throw new Error(`RPC call failed: ${response.status}`);
    }

    const data = await response.json();
    if (data.error) {
      throw new Error(`RPC error: ${data.error.message}`);
    }

    return data;
  }
}

export const blockchainService = new BlockchainService();
