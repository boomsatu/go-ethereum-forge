
class BlockchainService {
  private baseUrl: string;

  constructor(baseUrl: string = 'http://localhost:8545') {
    this.baseUrl = baseUrl;
  }

  // Wallet operations
  async createWallet(): Promise<WalletData> {
    const response = await fetch(`${this.baseUrl}/api/wallet/create`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      throw new Error(`Failed to create wallet: ${response.statusText}`);
    }

    return response.json();
  }

  async importWallet(privateKey: string): Promise<WalletData> {
    const response = await fetch(`${this.baseUrl}/api/wallet/import`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ privateKey }),
    });

    if (!response.ok) {
      throw new Error(`Failed to import wallet: ${response.statusText}`);
    }

    return response.json();
  }

  async getBalance(address: string): Promise<BalanceData> {
    const response = await fetch(`${this.baseUrl}/api/wallet/balance?address=${address}`);

    if (!response.ok) {
      throw new Error(`Failed to get balance: ${response.statusText}`);
    }

    return response.json();
  }

  async sendTransaction(txData: {
    from: string;
    to?: string;
    value?: string;
    gas?: string;
    gasPrice?: string;
    data?: string;
    privateKey: string;
  }): Promise<TransactionResult> {
    const response = await fetch(`${this.baseUrl}/api/wallet/send`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        from: txData.from,
        to: txData.to || '',
        value: txData.value || '0',
        gasLimit: txData.gas || '21000',
        gasPrice: txData.gasPrice || '20000000000',
        data: txData.data || '',
        privateKey: txData.privateKey,
      }),
    });

    if (!response.ok) {
      throw new Error(`Failed to send transaction: ${response.statusText}`);
    }

    return response.json();
  }

  // Block operations
  async getBlocks(limit: number = 10): Promise<BlockData[]> {
    const response = await fetch(`${this.baseUrl}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        jsonrpc: '2.0',
        method: 'eth_blockNumber',
        params: [],
        id: 1,
      }),
    });

    if (!response.ok) {
      throw new Error('Failed to get block number');
    }

    const blockNumberResult = await response.json();
    const latestBlockNumber = parseInt(blockNumberResult.result, 16);

    const blocks: BlockData[] = [];
    for (let i = Math.max(0, latestBlockNumber - limit + 1); i <= latestBlockNumber; i++) {
      try {
        const blockResponse = await fetch(`${this.baseUrl}`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            jsonrpc: '2.0',
            method: 'eth_getBlockByNumber',
            params: [`0x${i.toString(16)}`, true],
            id: 1,
          }),
        });

        if (blockResponse.ok) {
          const blockResult = await blockResponse.json();
          if (blockResult.result) {
            blocks.push({
              number: parseInt(blockResult.result.number, 16),
              hash: blockResult.result.hash,
              timestamp: parseInt(blockResult.result.timestamp, 16),
              transactionCount: blockResult.result.transactionCount || 0,
              gasUsed: parseInt(blockResult.result.gasUsed, 16),
              gasLimit: parseInt(blockResult.result.gasLimit, 16),
            });
          }
        }
      } catch (error) {
        console.error(`Failed to fetch block ${i}:`, error);
      }
    }

    return blocks.reverse();
  }

  // Admin operations
  async getStatus(): Promise<NodeStatus> {
    const response = await fetch(`${this.baseUrl}/api/admin/status`);

    if (!response.ok) {
      throw new Error('Failed to get node status');
    }

    return response.json();
  }

  async updateConfig(config: any): Promise<{ success: boolean }> {
    // For now, return success since config updates require node restart
    console.log('Config update requested:', config);
    return { success: true };
  }

  // Mining operations
  async startMining(): Promise<{ success: boolean }> {
    const response = await fetch(`${this.baseUrl}/api/mining/start`, {
      method: 'POST',
    });

    if (!response.ok) {
      throw new Error('Failed to start mining');
    }

    return response.json();
  }

  async stopMining(): Promise<{ success: boolean }> {
    const response = await fetch(`${this.baseUrl}/api/mining/stop`, {
      method: 'POST',
    });

    if (!response.ok) {
      throw new Error('Failed to stop mining');
    }

    return response.json();
  }

  async getMiningStats(): Promise<MiningStats> {
    const response = await fetch(`${this.baseUrl}/api/mining/stats`);

    if (!response.ok) {
      throw new Error('Failed to get mining stats');
    }

    return response.json();
  }

  // Network operations
  async getNetworkStats(): Promise<NetworkStats> {
    const response = await fetch(`${this.baseUrl}/api/network/stats`);

    if (!response.ok) {
      throw new Error('Failed to get network stats');
    }

    return response.json();
  }

  async getMetrics(): Promise<SystemMetrics> {
    const response = await fetch(`${this.baseUrl}/api/metrics`);

    if (!response.ok) {
      throw new Error('Failed to get metrics');
    }

    return response.json();
  }
}

// Types
export interface WalletData {
  address: string;
  privateKey: string;
  publicKey: string;
  balance: string;
  balanceEth: string;
  nonce?: string;
}

export interface BalanceData {
  address: string;
  balance: string;
  balanceEth: string;
  nonce: string;
}

export interface TransactionResult {
  hash: string;
  success: boolean;
  nonce: string;
  gasUsed: string;
}

export interface BlockData {
  number: number;
  hash: string;
  timestamp: number;
  transactionCount: number;
  gasUsed: number;
  gasLimit: number;
}

export interface NodeStatus {
  status: string;
  config: {
    chainId: number;
    dataDir: string;
    gasLimit: number;
  };
}

export interface MiningStats {
  isActive: boolean;
  hashRate: number;
  blocksFound: number;
  difficulty: string;
}

export interface NetworkStats {
  peerCount: number;
  difficulty: string;
  hashRate: string;
}

export interface SystemMetrics {
  uptime: number;
  memoryUsage: number;
  diskUsage: number;
  cpuUsage: number;
  blockCount: number;
  transactionCount: number;
  peersConnected: number;
}

export default new BlockchainService();
