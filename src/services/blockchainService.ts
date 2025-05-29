
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
  balance: string;
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

  async updateConfig(config: Partial<BlockchainConfig>): Promise<boolean> {
    try {
      const response = await fetch(`${this.baseUrl}/api/admin/config`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(config),
      });
      return response.ok;
    } catch (error) {
      console.error('Failed to update config:', error);
      return false;
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

  async getBlocks(fromBlock: number, toBlock: number): Promise<BlockData[]> {
    const blocks: BlockData[] = [];
    for (let i = fromBlock; i <= toBlock; i++) {
      const block = await this.getBlockByNumber(i, false);
      if (block) {
        blocks.push(block);
      }
    }
    return blocks;
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

  async sendTransaction(transaction: {
    from: string;
    to?: string;
    value?: string;
    gas?: string;
    gasPrice?: string;
    data?: string;
  }): Promise<string | null> {
    try {
      const response = await this.rpcCall('eth_sendTransaction', [transaction]);
      return response.result;
    } catch (error) {
      console.error('Failed to send transaction:', error);
      return null;
    }
  }

  // Wallet Management
  async createWallet(): Promise<WalletData | null> {
    try {
      const response = await fetch(`${this.baseUrl}/api/wallet/create`, {
        method: 'POST',
      });
      if (response.ok) {
        return await response.json();
      }
      return null;
    } catch (error) {
      console.error('Failed to create wallet:', error);
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
      if (response.ok) {
        return await response.json();
      }
      return null;
    } catch (error) {
      console.error('Failed to import wallet:', error);
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
        chainId: chainId.result ? parseInt(chainId.result, 16).toString() : '0',
        networkId: networkId.result || '0',
        blockHeight: blockNumber.result ? parseInt(blockNumber.result, 16) : 0,
        peerCount: stats.peerCount || 0,
        difficulty: stats.difficulty || '0',
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
