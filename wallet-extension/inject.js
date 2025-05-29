
// Injected script that creates the Web3 provider for DApps

(function() {
  'use strict';

  if (window.ethereum) {
    return; // Provider already exists
  }

  class BlockchainWalletProvider {
    constructor() {
      this.isBlockchainWallet = true;
      this.isConnected = true;
      this.chainId = '0x539'; // 1337 in hex
      this.networkVersion = '1337';
      this.selectedAddress = null;
      this.accounts = [];
      this.eventHandlers = {};
      
      this.init();
    }

    async init() {
      // Get initial accounts
      try {
        this.accounts = await this.getAccounts();
        this.selectedAddress = this.accounts[0] || null;
      } catch (error) {
        console.error('Failed to get initial accounts:', error);
      }

      // Listen for wallet events
      window.addEventListener('message', (event) => {
        if (event.source !== window || event.data.type !== 'BLOCKCHAIN_WALLET_EVENT') {
          return;
        }

        this.handleWalletEvent(event.data.event, event.data.data);
      });
    }

    handleWalletEvent(event, data) {
      switch (event) {
        case 'accountsChanged':
          this.accounts = data;
          this.selectedAddress = data[0] || null;
          this.emit('accountsChanged', data);
          break;
        case 'chainChanged':
          this.chainId = data;
          this.emit('chainChanged', data);
          break;
        case 'connect':
          this.emit('connect', { chainId: this.chainId });
          break;
        case 'disconnect':
          this.emit('disconnect');
          break;
      }
    }

    // Event management
    on(event, handler) {
      if (!this.eventHandlers[event]) {
        this.eventHandlers[event] = [];
      }
      this.eventHandlers[event].push(handler);
    }

    removeListener(event, handler) {
      if (this.eventHandlers[event]) {
        const index = this.eventHandlers[event].indexOf(handler);
        if (index > -1) {
          this.eventHandlers[event].splice(index, 1);
        }
      }
    }

    emit(event, data) {
      if (this.eventHandlers[event]) {
        this.eventHandlers[event].forEach(handler => {
          try {
            handler(data);
          } catch (error) {
            console.error('Error in event handler:', error);
          }
        });
      }
    }

    // Core wallet methods
    async request(args) {
      const { method, params = [] } = args;

      switch (method) {
        case 'eth_requestAccounts':
          return this.requestAccounts();
        
        case 'eth_accounts':
          return this.getAccounts();
        
        case 'eth_chainId':
          return this.chainId;
        
        case 'net_version':
          return this.networkVersion;
        
        case 'personal_sign':
          return this.personalSign(params[0], params[1]);
        
        case 'eth_sendTransaction':
          return this.sendTransaction(params[0]);
        
        case 'wallet_switchEthereumChain':
          return this.switchChain(params[0].chainId);
        
        case 'wallet_addEthereumChain':
          return this.addChain(params[0]);
        
        default:
          return this.rpcRequest(method, params);
      }
    }

    async requestAccounts() {
      return new Promise((resolve, reject) => {
        const id = Date.now();
        
        const handler = (event) => {
          if (event.source !== window || 
              event.data.type !== 'BLOCKCHAIN_WALLET_ACCOUNTS_RESPONSE' ||
              event.data.id !== id) {
            return;
          }
          
          window.removeEventListener('message', handler);
          
          if (event.data.accounts && event.data.accounts.length > 0) {
            this.accounts = event.data.accounts;
            this.selectedAddress = event.data.accounts[0];
            this.emit('accountsChanged', event.data.accounts);
            resolve(event.data.accounts);
          } else {
            reject(new Error('No accounts available. Please create or import a wallet.'));
          }
        };
        
        window.addEventListener('message', handler);
        
        window.postMessage({
          type: 'BLOCKCHAIN_WALLET_ACCOUNTS_REQUEST',
          id: id
        }, '*');
        
        // Timeout after 30 seconds
        setTimeout(() => {
          window.removeEventListener('message', handler);
          reject(new Error('Request timed out'));
        }, 30000);
      });
    }

    async getAccounts() {
      return this.accounts;
    }

    async personalSign(message, address) {
      return new Promise((resolve, reject) => {
        const id = Date.now();
        
        const handler = (event) => {
          if (event.source !== window || 
              event.data.type !== 'BLOCKCHAIN_WALLET_SIGN_RESPONSE' ||
              event.data.id !== id) {
            return;
          }
          
          window.removeEventListener('message', handler);
          
          if (event.data.error) {
            reject(new Error(event.data.error));
          } else {
            resolve(event.data.signature);
          }
        };
        
        window.addEventListener('message', handler);
        
        window.postMessage({
          type: 'BLOCKCHAIN_WALLET_SIGN_REQUEST',
          id: id,
          message: message,
          address: address
        }, '*');
        
        // Timeout after 60 seconds
        setTimeout(() => {
          window.removeEventListener('message', handler);
          reject(new Error('Sign request timed out'));
        }, 60000);
      });
    }

    async sendTransaction(transaction) {
      return new Promise((resolve, reject) => {
        const id = Date.now();
        
        const handler = (event) => {
          if (event.source !== window || 
              event.data.type !== 'BLOCKCHAIN_WALLET_SEND_RESPONSE' ||
              event.data.id !== id) {
            return;
          }
          
          window.removeEventListener('message', handler);
          
          if (event.data.error) {
            reject(new Error(event.data.error));
          } else {
            resolve(event.data.transactionHash);
          }
        };
        
        window.addEventListener('message', handler);
        
        window.postMessage({
          type: 'BLOCKCHAIN_WALLET_SEND_REQUEST',
          id: id,
          transaction: transaction
        }, '*');
        
        // Timeout after 60 seconds
        setTimeout(() => {
          window.removeEventListener('message', handler);
          reject(new Error('Transaction request timed out'));
        }, 60000);
      });
    }

    async rpcRequest(method, params) {
      return new Promise((resolve, reject) => {
        const id = Date.now();
        
        const handler = (event) => {
          if (event.source !== window || 
              event.data.type !== 'BLOCKCHAIN_WALLET_RESPONSE' ||
              event.data.id !== id) {
            return;
          }
          
          window.removeEventListener('message', handler);
          
          if (event.data.error) {
            reject(new Error(event.data.error));
          } else {
            resolve(event.data.result);
          }
        };
        
        window.addEventListener('message', handler);
        
        window.postMessage({
          type: 'BLOCKCHAIN_WALLET_REQUEST',
          id: id,
          method: method,
          params: params
        }, '*');
        
        // Timeout after 30 seconds
        setTimeout(() => {
          window.removeEventListener('message', handler);
          reject(new Error('RPC request timed out'));
        }, 30000);
      });
    }

    async switchChain(chainId) {
      if (chainId === '0x539' || chainId === 1337) {
        return null; // Success
      }
      throw new Error('Unsupported chain');
    }

    async addChain(chainParams) {
      throw new Error('Adding chains not supported');
    }

    // Legacy methods for compatibility
    enable() {
      return this.requestAccounts();
    }

    send(methodOrPayload, paramsOrCallback) {
      if (typeof methodOrPayload === 'string') {
        // Legacy format: send(method, params)
        return this.request({ method: methodOrPayload, params: paramsOrCallback });
      } else {
        // Legacy format: send(payload, callback)
        const payload = methodOrPayload;
        const callback = paramsOrCallback;
        
        this.request(payload)
          .then(result => callback(null, { result }))
          .catch(error => callback(error));
      }
    }

    sendAsync(payload, callback) {
      this.request(payload)
        .then(result => callback(null, { jsonrpc: '2.0', id: payload.id, result }))
        .catch(error => callback(error));
    }
  }

  // Create and expose the provider
  const provider = new BlockchainWalletProvider();
  
  // Make it available globally
  window.ethereum = provider;
  window.web3 = { currentProvider: provider };

  // Announce provider availability
  window.dispatchEvent(new Event('ethereum#initialized'));
  
  // For compatibility with some DApps
  document.addEventListener('DOMContentLoaded', () => {
    window.dispatchEvent(new Event('ethereum#initialized'));
  });

  console.log('Blockchain Wallet provider injected');
})();
