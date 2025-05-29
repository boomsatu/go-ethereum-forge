
# Blockchain Node API Documentation

## Overview
This blockchain node provides a complete Ethereum-compatible JSON-RPC API for interacting with the blockchain.

## Base URL
```
http://localhost:8545
```

## Supported Methods

### Block Information

#### eth_blockNumber
Returns the number of most recent block.

**Parameters:** None

**Returns:** `QUANTITY` - integer of the current block number the client is on.

**Example:**
```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  -H "Content-Type: application/json" http://localhost:8545
```

#### eth_getBlockByNumber
Returns information about a block by block number.

**Parameters:**
1. `QUANTITY|TAG` - integer of a block number, or the string "latest"
2. `Boolean` - If true it returns the full transaction objects, if false only the hashes

**Returns:** `Object` - A block object

**Example:**
```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest", true],"id":1}' \
  -H "Content-Type: application/json" http://localhost:8545
```

#### eth_getBlockByHash
Returns information about a block by hash.

**Parameters:**
1. `DATA` - Hash of a block
2. `Boolean` - If true it returns the full transaction objects, if false only the hashes

### Account Information

#### eth_getBalance
Returns the balance of the account of given address.

**Parameters:**
1. `DATA` - 20 Bytes - address to check for balance
2. `QUANTITY|TAG` - integer block number, or the string "latest"

**Returns:** `QUANTITY` - integer of the current balance in wei

**Example:**
```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getBalance","params":["0x742d35Cc6635C0532925a3b8D5c6C1C8b1c5C6C", "latest"],"id":1}' \
  -H "Content-Type: application/json" http://localhost:8545
```

#### eth_getTransactionCount
Returns the number of transactions sent from an address.

**Parameters:**
1. `DATA` - 20 Bytes - address
2. `QUANTITY|TAG` - integer block number, or the string "latest"

**Returns:** `QUANTITY` - integer of the number of transactions send from this address

### Transaction Information

#### eth_getTransactionByHash
Returns the information about a transaction requested by transaction hash.

**Parameters:**
1. `DATA` - 32 Bytes - hash of a transaction

**Returns:** `Object` - A transaction object

#### eth_getTransactionReceipt
Returns the receipt of a transaction by transaction hash.

**Parameters:**
1. `DATA` - 32 Bytes - hash of a transaction

**Returns:** `Object` - A transaction receipt object

#### eth_sendRawTransaction
Creates new message call transaction or a contract creation for signed transactions.

**Parameters:**
1. `DATA` - The signed transaction data

**Returns:** `DATA` - 32 Bytes - the transaction hash

### Contract Interaction

#### eth_call
Executes a new message call immediately without creating a transaction on the block chain.

**Parameters:**
1. `Object` - The transaction call object
2. `QUANTITY|TAG` - integer block number, or the string "latest"

**Returns:** `DATA` - the return value of executed contract

#### eth_estimateGas
Generates and returns an estimate of how much gas is necessary to allow the transaction to complete.

**Parameters:**
1. `Object` - The transaction call object

**Returns:** `QUANTITY` - the amount of gas used

#### eth_getCode
Returns code at a given address.

**Parameters:**
1. `DATA` - 20 Bytes - address
2. `QUANTITY|TAG` - integer block number, or the string "latest"

**Returns:** `DATA` - the code from the given address

### Network Information

#### eth_chainId
Returns the chain ID of the current network.

**Returns:** `QUANTITY` - integer of the current chain id

#### eth_gasPrice
Returns the current price per gas in wei.

**Returns:** `QUANTITY` - integer of the current gas price in wei

#### net_version
Returns the current network id.

**Returns:** `String` - The current network id

#### web3_clientVersion
Returns the current client version.

**Returns:** `String` - The current client version

## Error Codes

- `-32700`: Parse error
- `-32600`: Invalid Request
- `-32601`: Method not found
- `-32602`: Invalid params
- `-32603`: Internal error

## Health Check

The node provides a health check endpoint:

```bash
curl http://localhost:8545/health
```

Returns:
```json
{"status": "ok"}
```
