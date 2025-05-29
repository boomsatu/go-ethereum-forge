
// Content script for injecting Web3 provider into web pages

(function() {
  'use strict';

  // Check if provider is already injected
  if (window.ethereum) {
    return;
  }

  // Inject the provider script
  const script = document.createElement('script');
  script.src = chrome.runtime.getURL('inject.js');
  script.onload = function() {
    this.remove();
  };
  (document.head || document.documentElement).appendChild(script);

  // Listen for messages from the injected script
  window.addEventListener('message', async (event) => {
    if (event.source !== window || !event.data.type) {
      return;
    }

    if (event.data.type === 'BLOCKCHAIN_WALLET_REQUEST') {
      try {
        const response = await chrome.runtime.sendMessage({
          type: 'RPC_CALL',
          method: event.data.method,
          params: event.data.params
        });

        window.postMessage({
          type: 'BLOCKCHAIN_WALLET_RESPONSE',
          id: event.data.id,
          result: response.result,
          error: response.error
        }, '*');
      } catch (error) {
        window.postMessage({
          type: 'BLOCKCHAIN_WALLET_RESPONSE',
          id: event.data.id,
          error: error.message
        }, '*');
      }
    }

    if (event.data.type === 'BLOCKCHAIN_WALLET_ACCOUNTS_REQUEST') {
      try {
        const response = await chrome.runtime.sendMessage({
          type: 'GET_ACCOUNTS'
        });

        window.postMessage({
          type: 'BLOCKCHAIN_WALLET_ACCOUNTS_RESPONSE',
          id: event.data.id,
          accounts: response.accounts || []
        }, '*');
      } catch (error) {
        window.postMessage({
          type: 'BLOCKCHAIN_WALLET_ACCOUNTS_RESPONSE',
          id: event.data.id,
          accounts: []
        }, '*');
      }
    }

    if (event.data.type === 'BLOCKCHAIN_WALLET_SIGN_REQUEST') {
      try {
        const response = await chrome.runtime.sendMessage({
          type: 'PERSONAL_SIGN',
          message: event.data.message,
          address: event.data.address
        });

        window.postMessage({
          type: 'BLOCKCHAIN_WALLET_SIGN_RESPONSE',
          id: event.data.id,
          signature: response.signature,
          error: response.error
        }, '*');
      } catch (error) {
        window.postMessage({
          type: 'BLOCKCHAIN_WALLET_SIGN_RESPONSE',
          id: event.data.id,
          error: error.message
        }, '*');
      }
    }

    if (event.data.type === 'BLOCKCHAIN_WALLET_SEND_REQUEST') {
      try {
        const response = await chrome.runtime.sendMessage({
          type: 'SEND_TRANSACTION',
          transaction: event.data.transaction
        });

        window.postMessage({
          type: 'BLOCKCHAIN_WALLET_SEND_RESPONSE',
          id: event.data.id,
          transactionHash: response.transactionHash,
          error: response.error
        }, '*');
      } catch (error) {
        window.postMessage({
          type: 'BLOCKCHAIN_WALLET_SEND_RESPONSE',
          id: event.data.id,
          error: error.message
        }, '*');
      }
    }
  });

  // Listen for wallet events from background script
  chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
    if (message.type === 'WALLET_EVENT') {
      window.postMessage({
        type: 'BLOCKCHAIN_WALLET_EVENT',
        event: message.event,
        data: message.data
      }, '*');
    }
  });
})();
