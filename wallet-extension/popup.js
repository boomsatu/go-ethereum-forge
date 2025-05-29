
class BlockchainWallet {
  constructor() {
    this.rpcUrl = 'http://localhost:8545';
    this.wallet = null;
    this.init();
  }

  async init() {
    await this.loadWallet();
    this.setupEventListeners();
    this.updateUI();
    this.startBalanceUpdater();
  }

  async loadWallet() {
    const result = await chrome.storage.local.get(['wallet']);
    if (result.wallet) {
      this.wallet = result.wallet;
    }
  }

  async saveWallet() {
    await chrome.storage.local.set({ wallet: this.wallet });
  }

  setupEventListeners() {
    // Setup view buttons
    document.getElementById('createWalletBtn').addEventListener('click', () => this.createWallet());
    document.getElementById('importWalletBtn').addEventListener('click', () => this.showImportModal());
    document.getElementById('sendBtn').addEventListener('click', () => this.showSendModal());
    document.getElementById('receiveBtn').addEventListener('click', () => this.showReceiveModal());
    document.getElementById('importBtn').addEventListener('click', () => this.showImportModal());
    document.getElementById('exportBtn').addEventListener('click', () => this.exportWallet());

    // Setup modal forms
    document.getElementById('sendForm').addEventListener('submit', (e) => this.handleSendTransaction(e));
    document.getElementById('importForm').addEventListener('submit', (e) => this.handleImportWallet(e));

    // Setup modal close buttons
    document.getElementById('cancelSend').addEventListener('click', () => this.hideSendModal());
    document.getElementById('cancelImport').addEventListener('click', () => this.hideImportModal());
    document.getElementById('closeReceive').addEventListener('click', () => this.hideReceiveModal());
    document.getElementById('copyAddress').addEventListener('click', () => this.copyAddress());

    // Close modals when clicking outside
    document.querySelectorAll('.modal').forEach(modal => {
      modal.addEventListener('click', (e) => {
        if (e.target === modal) {
          modal.style.display = 'none';
        }
      });
    });
  }

  async createWallet() {
    try {
      // Generate new wallet using Web3 crypto
      const privateKey = this.generatePrivateKey();
      const address = await this.privateKeyToAddress(privateKey);
      
      this.wallet = {
        address: address,
        privateKey: privateKey
      };
      
      await this.saveWallet();
      this.updateUI();
      
      alert('Wallet created successfully! Please backup your private key.');
    } catch (error) {
      console.error('Error creating wallet:', error);
      alert('Failed to create wallet');
    }
  }

  generatePrivateKey() {
    const array = new Uint8Array(32);
    crypto.getRandomValues(array);
    return Array.from(array, byte => byte.toString(16).padStart(2, '0')).join('');
  }

  async privateKeyToAddress(privateKey) {
    // Simple address generation - in production use proper secp256k1
    const encoder = new TextEncoder();
    const data = encoder.encode(privateKey);
    const hashBuffer = await crypto.subtle.digest('SHA-256', data);
    const hashArray = new Uint8Array(hashBuffer);
    const address = '0x' + Array.from(hashArray.slice(12), byte => 
      byte.toString(16).padStart(2, '0')
    ).join('');
    return address;
  }

  updateUI() {
    if (this.wallet) {
      document.getElementById('setupView').style.display = 'none';
      document.getElementById('walletView').style.display = 'block';
      document.getElementById('walletAddress').textContent = this.wallet.address;
      this.updateBalance();
      this.updateTransactions();
    } else {
      document.getElementById('setupView').style.display = 'block';
      document.getElementById('walletView').style.display = 'none';
    }
  }

  async updateBalance() {
    if (!this.wallet) return;

    try {
      const balance = await this.rpcCall('eth_getBalance', [this.wallet.address, 'latest']);
      const balanceInEth = parseInt(balance, 16) / Math.pow(10, 18);
      document.getElementById('walletBalance').textContent = `${balanceInEth.toFixed(6)} ETH`;
      document.getElementById('balanceUsd').textContent = `$${(balanceInEth * 2000).toFixed(2)} USD`; // Mock price
    } catch (error) {
      console.error('Error updating balance:', error);
      document.getElementById('walletBalance').textContent = 'Error loading balance';
    }
  }

  async updateTransactions() {
    // This would fetch transaction history from the blockchain
    // For now, we'll show a placeholder
    const transactionList = document.getElementById('transactionList');
    transactionList.innerHTML = `
      <div style="text-align: center; opacity: 0.6; padding: 20px;">
        Transaction history not implemented yet
      </div>
    `;
  }

  async rpcCall(method, params = []) {
    const response = await fetch(this.rpcUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        jsonrpc: '2.0',
        method: method,
        params: params,
        id: 1
      })
    });

    const data = await response.json();
    if (data.error) {
      throw new Error(data.error.message);
    }
    return data.result;
  }

  showSendModal() {
    document.getElementById('sendModal').style.display = 'block';
  }

  hideSendModal() {
    document.getElementById('sendModal').style.display = 'none';
    document.getElementById('sendForm').reset();
    document.getElementById('sendError').textContent = '';
    document.getElementById('sendSuccess').textContent = '';
  }

  showImportModal() {
    document.getElementById('importModal').style.display = 'block';
  }

  hideImportModal() {
    document.getElementById('importModal').style.display = 'none';
    document.getElementById('importForm').reset();
    document.getElementById('importError').textContent = '';
  }

  showReceiveModal() {
    document.getElementById('receiveModal').style.display = 'block';
    document.getElementById('receiveAddress').textContent = this.wallet.address;
    
    // Generate QR code (placeholder)
    document.getElementById('qrCode').innerHTML = `
      <div style="width: 150px; height: 150px; background: #f0f0f0; border: 1px solid #ddd; 
                  display: flex; align-items: center; justify-content: center; margin: 0 auto;">
        QR Code<br>
        <small>(Not implemented)</small>
      </div>
    `;
  }

  hideReceiveModal() {
    document.getElementById('receiveModal').style.display = 'none';
  }

  async handleSendTransaction(e) {
    e.preventDefault();
    
    const toAddress = document.getElementById('sendToAddress').value;
    const amount = document.getElementById('sendAmount').value;
    const gasPrice = document.getElementById('sendGasPrice').value;
    
    const errorDiv = document.getElementById('sendError');
    const successDiv = document.getElementById('sendSuccess');
    
    errorDiv.textContent = '';
    successDiv.textContent = '';
    
    try {
      // Validate inputs
      if (!toAddress.match(/^0x[a-fA-F0-9]{40}$/)) {
        throw new Error('Invalid address format');
      }
      
      if (parseFloat(amount) <= 0) {
        throw new Error('Amount must be greater than 0');
      }
      
      // Get nonce
      const nonce = await this.rpcCall('eth_getTransactionCount', [this.wallet.address, 'latest']);
      
      // Create transaction object
      const transaction = {
        from: this.wallet.address,
        to: toAddress,
        value: '0x' + (parseFloat(amount) * Math.pow(10, 18)).toString(16),
        gas: '0x5208', // 21000
        gasPrice: '0x' + (parseFloat(gasPrice) * Math.pow(10, 9)).toString(16),
        nonce: nonce
      };
      
      // Sign transaction (simplified - in production use proper signing)
      const signedTx = await this.signTransaction(transaction);
      
      // Send transaction
      const txHash = await this.rpcCall('eth_sendRawTransaction', [signedTx]);
      
      successDiv.textContent = `Transaction sent! Hash: ${txHash}`;
      
      // Update balance after a delay
      setTimeout(() => this.updateBalance(), 2000);
      
    } catch (error) {
      errorDiv.textContent = error.message;
    }
  }

  async signTransaction(transaction) {
    // Simplified transaction signing - in production use proper secp256k1 signing
    return '0x' + Array.from(crypto.getRandomValues(new Uint8Array(32)), 
      byte => byte.toString(16).padStart(2, '0')).join('');
  }

  async handleImportWallet(e) {
    e.preventDefault();
    
    const privateKey = document.getElementById('importPrivateKey').value;
    const errorDiv = document.getElementById('importError');
    
    errorDiv.textContent = '';
    
    try {
      if (!privateKey.match(/^[a-fA-F0-9]{64}$/)) {
        throw new Error('Invalid private key format');
      }
      
      const address = await this.privateKeyToAddress(privateKey);
      
      this.wallet = {
        address: address,
        privateKey: privateKey
      };
      
      await this.saveWallet();
      this.hideImportModal();
      this.updateUI();
      
    } catch (error) {
      errorDiv.textContent = error.message;
    }
  }

  exportWallet() {
    if (!this.wallet) return;
    
    const walletData = {
      address: this.wallet.address,
      privateKey: this.wallet.privateKey
    };
    
    // Create downloadable file
    const blob = new Blob([JSON.stringify(walletData, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    
    const a = document.createElement('a');
    a.href = url;
    a.download = `wallet-${this.wallet.address.slice(0, 8)}.json`;
    a.click();
    
    URL.revokeObjectURL(url);
  }

  copyAddress() {
    navigator.clipboard.writeText(this.wallet.address).then(() => {
      const btn = document.getElementById('copyAddress');
      const originalText = btn.textContent;
      btn.textContent = 'Copied!';
      setTimeout(() => {
        btn.textContent = originalText;
      }, 2000);
    });
  }

  startBalanceUpdater() {
    // Update balance every 30 seconds
    setInterval(() => {
      if (this.wallet) {
        this.updateBalance();
        this.updateTransactions();
      }
    }, 30000);
  }

  // Web3 Provider methods for DApp interaction
  getProvider() {
    return {
      isConnected: () => !!this.wallet,
      getAccounts: () => this.wallet ? [this.wallet.address] : [],
      getChainId: () => '0x539', // 1337 in hex
      request: async ({ method, params }) => {
        switch (method) {
          case 'eth_accounts':
          case 'eth_requestAccounts':
            return this.wallet ? [this.wallet.address] : [];
          case 'eth_chainId':
            return '0x539';
          case 'personal_sign':
            return this.personalSign(params[0], params[1]);
          case 'eth_sendTransaction':
            return this.sendTransaction(params[0]);
          default:
            return this.rpcCall(method, params);
        }
      }
    };
  }

  async personalSign(message, address) {
    if (!this.wallet || this.wallet.address.toLowerCase() !== address.toLowerCase()) {
      throw new Error('Account not found');
    }
    
    // Simplified signing - in production use proper message signing
    return '0x' + Array.from(crypto.getRandomValues(new Uint8Array(65)), 
      byte => byte.toString(16).padStart(2, '0')).join('');
  }

  async sendTransaction(transaction) {
    // Use the same logic as handleSendTransaction but for DApp requests
    const nonce = await this.rpcCall('eth_getTransactionCount', [this.wallet.address, 'latest']);
    
    const fullTransaction = {
      ...transaction,
      from: this.wallet.address,
      nonce: nonce
    };
    
    const signedTx = await this.signTransaction(fullTransaction);
    return this.rpcCall('eth_sendRawTransaction', [signedTx]);
  }
}

// Initialize wallet when popup opens
document.addEventListener('DOMContentLoaded', () => {
  window.wallet = new BlockchainWallet();
});
