
class BlockchainService {
  private baseUrl: string;

  constructor(baseUrl: string = 'http://localhost:8545') {
    this.baseUrl = baseUrl;
  }

  // Health check
  async getHealthCheck(): Promise<boolean> {
    try {
      const response = await fetch(`${this.baseUrl}/api/health`);
      return response.ok;
    } catch (error) {
      return false;
    }
  }

  // Node operations
  async startNode(config: BlockchainConfig): Promise<boolean> {
    try {
      const response = await fetch(`${this.baseUrl}/api/admin/start`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(config),
      });
      return response.ok;
    } catch (error) {
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
      return false;
    }
  }

  async getNodeStatus(): Promise<NodeStatus | null> {
    try {
      const response = await fetch(`${this.baseUrl}/api/admin/status`);
      if (!response.ok) return null;
      return response.json();
    } catch (error) {
      return null;
    }
  }

  // Wallet operations
  async createWallet(): Promise<WalletData | null> {
    try {
      const response = await fetch(`${this.baseUrl}/api/wallet/create`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
      });
      if (!response.ok) return null;
      return response.json();
    } catch (error) {
      return null;
    }
  }

  async importWallet(privateKey: string): Promise<WalletData | null> {
    try {
      const response = await fetch(`${this.baseUrl}/api/wallet/import`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ privateKey }),
      });
      if (!response.ok) return null;
      return response.json();
    } catch (error) {
      return null;
    }
  }

  async getBalance(address: string): Promise<string> {
    try {
      const response = await fetch(`${this.baseUrl}/api/wallet/balance?address=${address}`);
      if (!response.ok) return '0';
      const data = await response.json();
      return data.balanceEth || '0';
    } catch (error) {
      return '0';
    }
  }

  async getNonce(address: string): Promise<number> {
    try {
      const response = await fetch(`${this.baseUrl}/api/wallet/nonce?address=${address}`);
      if (!response.ok) return 0;
      const data = await response.json();
      return parseInt(data.nonce) || 0;
    } catch (error) {
      return 0;
    }
  }

  // Block operations
  async getLatestBlock(): Promise<RawBlockData | null> {
    try {
      const response = await fetch(`${this.baseUrl}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          jsonrpc: '2.0',
          method: 'eth_getBlockByNumber',
          params: ['latest', true],
          id: 1,
        }),
      });
      if (!response.ok) return null;
      const data = await response.json();
      return data.result;
    } catch (error) {
      return null;
    }
  }

  async getBlockByNumber(blockNumber: number, includeTx: boolean = false): Promise<RawBlockData | null> {
    try {
      const response = await fetch(`${this.baseUrl}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          jsonrpc: '2.0',
          method: 'eth_getBlockByNumber',
          params: [`0x${blockNumber.toString(16)}`, includeTx],
          id: 1,
        }),
      });
      if (!response.ok) return null;
      const data = await response.json();
      return data.result;
    } catch (error) {
      return null;
    }
  }

  async getBlocks(startBlock: number, endBlock: number): Promise<RawBlockData[]> {
    try {
      const blocks: RawBlockData[] = [];
      for (let i = startBlock; i <= endBlock; i++) {
        const block = await this.getBlockByNumber(i, true);
        if (block) blocks.push(block);
      }
      return blocks;
    } catch (error) {
      return [];
    }
  }

  // Transaction operations
  async getTransaction(hash: string): Promise<TransactionData | null> {
    try {
      const response = await fetch(`${this.baseUrl}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          jsonrpc: '2.0',
          method: 'eth_getTransactionByHash',
          params: [hash],
          id: 1,
        }),
      });
      if (!response.ok) return null;
      const data = await response.json();
      return data.result;
    } catch (error) {
      return null;
    }
  }

  async sendTransaction(txData: {
    from: string;
    to?: string;
    value?: string;
    gas?: string;
    gasPrice?: string;
    data?: string;
  }): Promise<string | null> {
    try {
      const response = await fetch(`${this.baseUrl}/api/wallet/send`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(txData),
      });
      if (!response.ok) return null;
      const data = await response.json();
      return data.hash;
    } catch (error) {
      return null;
    }
  }

  // Mining operations
  async startMining(minerAddress: string, threads: number = 1): Promise<boolean> {
    try {
      const response = await fetch(`${this.baseUrl}/api/mining/start`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ miner: minerAddress, threads }),
      });
      return response.ok;
    } catch (error) {
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
      return false;
    }
  }

  async mineBlock(minerAddress: string): Promise<{ blockNumber: number } | null> {
    try {
      const response = await fetch(`${this.baseUrl}/api/mining/mine-block`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ miner: minerAddress }),
      });
      if (!response.ok) return null;
      return response.json();
    } catch (error) {
      return null;
    }
  }

  async getMiningStats(): Promise<MiningStats | null> {
    try {
      const response = await fetch(`${this.baseUrl}/api/mining/stats`);
      if (!response.ok) return null;
      return response.json();
    } catch (error) {
      return null;
    }
  }

  // Network operations
  async getNetworkStats(): Promise<NetworkStats | null> {
    try {
      const response = await fetch(`${this.baseUrl}/api/network/stats`);
      if (!response.ok) return null;
      return response.json();
    } catch (error) {
      return null;
    }
  }

  async getPeers(): Promise<any[]> {
    try {
      const response = await fetch(`${this.baseUrl}/api/network/peers`);
      if (!response.ok) return [];
      const data = await response.json();
      return data.peers || [];
    } catch (error) {
      return [];
    }
  }

  async getMetrics(): Promise<SystemMetrics | null> {
    try {
      const response = await fetch(`${this.baseUrl}/api/metrics`);
      if (!response.ok) return null;
      return response.json();
    } catch (error) {
      return null;
    }
  }

  async updateConfig(config: BlockchainConfig): Promise<boolean> {
    try {
      const response = await fetch(`${this.baseUrl}/api/admin/config`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(config),
      });
      return response.ok;
    } catch (error) {
      return false;
    }
  }
}

// Types
export interface BlockchainConfig {
  host: string;
  port: number;
  chainId: string;
  gasLimit: string;
  dataDir: string;
}

export interface WalletData {
  address: string;
  privateKey: string;
  publicKey: string;
  balance: string;
  balanceEth: string;
  nonce?: string;
}

export interface RawBlockData {
  number: string;
  hash: string;
  parentHash: string;
  timestamp: string;
  transactions: any[];
  gasUsed: string;
  gasLimit: string;
  difficulty: string;
  miner: string;
  size: string;
  nonce: string;
}

export interface BlockData {
  number: number;
  hash: string;
  timestamp: number;
  transactionCount: number;
  gasUsed: number;
  gasLimit: number;
}

export interface TransactionData {
  hash: string;
  from: string;
  to: string | null;
  value: string;
  gasPrice: string;
  gas: string;
  nonce: string;
  blockNumber?: string;
  blockHash?: string;
}

export interface NodeStatus {
  status: string;
  config: BlockchainConfig;
}

export interface MiningStats {
  isActive: boolean;
  hashRate: number;
  blocksFound: number;
  difficulty: string;
}

export interface NetworkStats {
  chainId: string;
  networkId: string;
  peerCount: number;
  blockHeight: number;
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

// Export default instance
const blockchainService = new BlockchainService();
export default blockchainService;
