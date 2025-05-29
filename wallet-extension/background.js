
class WalletBackground {
  constructor() {
    this.wallet = null;
    this.rpcUrl = 'http://localhost:8545';
    this.init();
  }

  init() {
    // Listen for extension icon clicks
    chrome.action.onClicked.addListener(() => {
      chrome.action.openPopup();
    });

    // Listen for messages from content scripts
    chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
      this.handleMessage(request, sender, sendResponse);
      return true; // Keep message channel open for async responses
    });

    // Load wallet data
    this.loadWallet();

    // Listen for web requests to inject provider
    chrome.webNavigation.onDOMContentLoaded.addListener((details) => {
      if (details.frameId === 0) { // Main frame only
        this.injectProvider(details.tabId);
      }
    });
  }

  async loadWallet() {
    const result = await chrome.storage.local.get(['wallet']);
    if (result.wallet) {
      this.wallet = result.wallet;
    }
  }

  async saveWallet(wallet) {
    this.wallet = wallet;
    await chrome.storage.local.set({ wallet: wallet });
  }

  async handleMessage(request, sender, sendResponse) {
    try {
      switch (request.type) {
        case 'GET_WALLET':
          sendResponse({ wallet: this.wallet });
          break;

        case 'SET_WALLET':
          await this.saveWallet(request.wallet);
          sendResponse({ success: true });
          break;

        case 'RPC_CALL':
          const result = await this.rpcCall(request.method, request.params);
          sendResponse({ result: result });
          break;

        case 'SIGN_TRANSACTION':
          const signedTx = await this.signTransaction(request.transaction);
          sendResponse({ signedTransaction: signedTx });
          break;

        case 'GET_ACCOUNTS':
          sendResponse({ accounts: this.wallet ? [this.wallet.address] : [] });
          break;

        case 'REQUEST_ACCOUNTS':
          if (this.wallet) {
            sendResponse({ accounts: [this.wallet.address] });
          } else {
            // Open popup to create/import wallet
            chrome.action.openPopup();
            sendResponse({ error: 'No wallet available' });
          }
          break;

        case 'SEND_TRANSACTION':
          try {
            const txHash = await this.sendTransaction(request.transaction);
            sendResponse({ transactionHash: txHash });
          } catch (error) {
            sendResponse({ error: error.message });
          }
          break;

        case 'PERSONAL_SIGN':
          try {
            const signature = await this.personalSign(request.message, request.address);
            sendResponse({ signature: signature });
          } catch (error) {
            sendResponse({ error: error.message });
          }
          break;

        default:
          sendResponse({ error: 'Unknown message type' });
      }
    } catch (error) {
      sendResponse({ error: error.message });
    }
  }

  async rpcCall(method, params = []) {
    try {
      const response = await fetch(this.rpcUrl, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          jsonrpc: '2.0',
          method: method,
          params: params,
          id: Date.now()
        })
      });

      const data = await response.json();
      if (data.error) {
        throw new Error(data.error.message);
      }
      return data.result;
    } catch (error) {
      console.error('RPC call failed:', error);
      throw error;
    }
  }

  async injectProvider(tabId) {
    try {
      await chrome.scripting.executeScript({
        target: { tabId: tabId },
        files: ['inject.js']
      });
    } catch (error) {
      console.error('Failed to inject provider:', error);
    }
  }

  async signTransaction(transaction) {
    if (!this.wallet) {
      throw new Error('No wallet available');
    }

    // Get transaction count for nonce
    const nonce = await this.rpcCall('eth_getTransactionCount', [this.wallet.address, 'latest']);
    
    // Build complete transaction
    const fullTransaction = {
      nonce: nonce,
      gasPrice: transaction.gasPrice || '0x4A817C800', // 20 Gwei
      gasLimit: transaction.gas || '0x5208', // 21000
      to: transaction.to,
      value: transaction.value || '0x0',
      data: transaction.data || '0x',
      from: this.wallet.address
    };

    // Sign transaction (simplified - in production use proper secp256k1)
    return this.createSignedTransaction(fullTransaction);
  }

  createSignedTransaction(transaction) {
    // This is a simplified implementation
    // In production, use proper RLP encoding and secp256k1 signing
    const serialized = this.serializeTransaction(transaction);
    return '0x' + serialized;
  }

  serializeTransaction(tx) {
    // Simplified serialization - in production use proper RLP
    const parts = [
      tx.nonce,
      tx.gasPrice,
      tx.gasLimit,
      tx.to || '',
      tx.value,
      tx.data,
      '0x1', // v (chain id)
      '0x', // r (signature)
      '0x'  // s (signature)
    ];
    
    return parts.join('').replace(/0x/g, '');
  }

  async sendTransaction(transaction) {
    const signedTx = await this.signTransaction(transaction);
    return this.rpcCall('eth_sendRawTransaction', [signedTx]);
  }

  async personalSign(message, address) {
    if (!this.wallet || this.wallet.address.toLowerCase() !== address.toLowerCase()) {
      throw new Error('Account not found or not authorized');
    }

    // Simplified message signing - in production use proper EIP-191/EIP-712
    const messageHash = await this.hashMessage(message);
    return this.signHash(messageHash);
  }

  async hashMessage(message) {
    // EIP-191 message hashing
    const prefix = '\x19Ethereum Signed Message:\n';
    const fullMessage = prefix + message.length + message;
    
    const encoder = new TextEncoder();
    const data = encoder.encode(fullMessage);
    const hashBuffer = await crypto.subtle.digest('SHA-256', data);
    
    return Array.from(new Uint8Array(hashBuffer), byte => 
      byte.toString(16).padStart(2, '0')).join('');
  }

  signHash(hash) {
    // Simplified signing - in production use secp256k1
    return '0x' + Array.from(crypto.getRandomValues(new Uint8Array(65)), 
      byte => byte.toString(16).padStart(2, '0')).join('');
  }

  // Network management
  async switchNetwork(chainId) {
    // For now, only support our local network
    if (chainId === '0x539' || chainId === 1337) {
      return true;
    }
    throw new Error('Unsupported network');
  }

  // Event system for DApp communication
  broadcastEvent(event, data) {
    chrome.tabs.query({}, (tabs) => {
      tabs.forEach(tab => {
        chrome.tabs.sendMessage(tab.id, {
          type: 'WALLET_EVENT',
          event: event,
          data: data
        }).catch(() => {
          // Ignore errors from tabs that don't have content script
        });
      });
    });
  }

  // Wallet lock/unlock functionality
  async lockWallet() {
    this.wallet = null;
    await chrome.storage.local.remove(['wallet']);
    this.broadcastEvent('accountsChanged', []);
  }

  async unlockWallet(password) {
    // In production, implement proper password-based encryption
    await this.loadWallet();
    if (this.wallet) {
      this.broadcastEvent('accountsChanged', [this.wallet.address]);
      return true;
    }
    return false;
  }
}

// Initialize background service
new WalletBackground();
