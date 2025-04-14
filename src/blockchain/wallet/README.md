# Digital Watchdog Initiative - Blockchain Wallet Interface

A modern blockchain wallet interface for interacting with the Digital Watchdog Initiative's smart contracts.

## Overview

This wallet interface allows users to:

- Connect to MetaMask or other Web3 wallets
- Interact with deployed smart contracts
- Execute contract functions and view results
- Monitor blockchain events in real-time
- Track transaction status and history

## Getting Started

### Prerequisites

- MetaMask or another Web3 wallet browser extension installed
- Local Ganache instance running (for local development)
- Smart contracts deployed to the blockchain

### Setup

1. Deploy smart contracts to your desired network (local or testnet)
2. The deployment script will automatically copy contract ABIs and addresses to the wallet directory
3. Start the wallet interface HTTP server

```bash
# From the blockchain directory
npm run wallet
```

Alternatively, you can use any HTTP server:

```bash
# Using Python's built-in HTTP server
cd wallet
python -m http.server 8080
```

4. Open your browser to http://localhost:8080

### Connecting to Different Networks

#### Local Ganache

- Make sure Ganache is running: `npx ganache-cli --port 8545 --gasLimit 30000000`
- In MetaMask, configure a custom network with:
  - Network Name: Ganache Local
  - RPC URL: http://localhost:8545
  - Chain ID: 1337
  - Currency Symbol: ETH

#### Sepolia Testnet

- In MetaMask, select the Sepolia test network
- Make sure you have test ETH (from a faucet)
- See SEPOLIA_GUIDE.md for detailed instructions

## Usage

1. **Connect Wallet**: Click the "Connect Wallet" button to connect your MetaMask wallet
2. **Select Contract**: Choose from the deployed contracts dropdown
3. **Select Function**: Choose a function to execute from the dropdown
4. **Enter Parameters**: Fill in the required parameters for the selected function
5. **Execute Function**: Click "Execute Function" to call the contract function
6. **View Results**: See the function results and transaction details
7. **Event Log**: Monitor contract events and transaction status

## Troubleshooting

### Common Issues

1. **Wallet Not Connecting**
   - Make sure MetaMask is installed and unlocked
   - Refresh the page and try connecting again
   - Check if you're on the correct network

2. **Contract Functions Not Working**
   - Verify you have sufficient ETH for gas fees
   - Check if you have the necessary permissions for the function
   - Ensure you're on the same network where contracts are deployed

3. **Missing Contract ABIs or Addresses**
   - Run the deployment script again to regenerate contract artifacts
   - Check if `deployed-contracts.json` exists in the wallet directory

### Debugging

The Event Log section shows details about your interactions with the blockchain, including:
- Wallet connection status
- Contract function calls
- Transaction hashes and receipts
- Error messages and reasons for failed transactions

## Architecture

- **Frontend**: Vanilla JavaScript and HTML/CSS
- **Contracts Interface**: ethers.js library 
- **Wallet Connectivity**: Web3 provider standard

## License

This project is licensed under the MIT License.