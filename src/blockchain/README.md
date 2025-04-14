# Digital Watchdog Initiative - Blockchain Component

This directory contains the blockchain component of the Digital Watchdog Initiative, a project focused on enhancing transparency and accountability in public fund management in Kenya using blockchain and AI.

## Components

The blockchain implementation consists of:

1. **Smart Contracts** (`contracts/`)
   - `TransactionEventLogger.sol`: Logs financial transaction events on the blockchain
   - `FundManager.sol`: Manages funds and transactions with role-based access control

2. **Deployment Scripts** (`scripts/`)
   - `deploy.js`: Deploys contracts to the blockchain network

3. **Tests** (`test/`)
   - `integration.test.js`: Tests the integration between contracts and simulates API interaction

4. **Wallet Interface** (`wallet/`)
   - Web interface for interacting with deployed contracts
   - Event monitoring and transaction management

## Getting Started

### Prerequisites

- Node.js (v14+)
- npm or yarn
- Ganache for local blockchain development

### Installation

1. Install dependencies:
   ```bash
   npm install
   ```

2. Compile smart contracts:
   ```bash
   npx hardhat compile
   ```

### Local Development

1. Start a local blockchain using Ganache:
   ```bash
   # Using Ganache CLI
   npx ganache-cli

   # Or using Docker
   docker-compose up -d blockchain
   ```

2. Deploy contracts to the local blockchain:
   ```bash
   # For Ganache running on port 8545 (default for Ganache CLI)
   npx hardhat run scripts/deploy.js --network ganache

   # For Ganache UI running on port 7545
   npx hardhat run scripts/deploy.js --network ganache_ui
   ```

3. Run tests:
   ```bash
   npx hardhat test
   ```

### Using the Wallet Interface

1. After deploying the contracts, navigate to the wallet directory:
   ```bash
   cd wallet
   ```

2. Serve the wallet interface:
   ```bash
   # Using Node.js http-server
   npx http-server -p 8080

   # Or using Python
   python -m http.server 8080
   ```

3. Open your browser and navigate to `http://localhost:8080`

4. Connect MetaMask to your local Ganache instance:
   - Network Name: Ganache
   - RPC URL: http://localhost:7545 (or 8545 depending on your setup)
   - Chain ID: 1337
   - Currency Symbol: ETH

## Integration with Go Backend

The Go backend interacts with the blockchain through the following components:

- `api/blockchain/connector.go`: Handles connection to the Ethereum network
- `api/blockchain/transaction_logger.go`: Interacts with the TransactionEventLogger contract
- `api/blockchain/fund_manager.go`: Interacts with the FundManager contract

Environment variables for configuration:
- `BLOCKCHAIN_SERVICE_URL`: URL of the blockchain node
- `TRANSACTION_LOGGER_ADDRESS`: Address of the deployed TransactionEventLogger contract
- `FUND_MANAGER_ADDRESS`: Address of the deployed FundManager contract
- `BLOCKCHAIN_PRIVATE_KEY`: Private key for signing transactions

## Deployment to Test Networks

To deploy to Ethereum test networks:

1. Configure environment variables:
   ```
   INFURA_API_KEY=your_infura_key
   PRIVATE_KEY=your_private_key_without_0x_prefix
   ETHERSCAN_API_KEY=your_etherscan_api_key
   ```

2. Deploy to Sepolia testnet:
   ```bash
   npx hardhat run scripts/deploy.js --network sepolia
   ```

3. Verify contract on Etherscan:
   ```bash
   npx hardhat verify --network sepolia DEPLOYED_CONTRACT_ADDRESS
   ```

## Security Considerations

- Role-based access control ensures only authorized users can update funds and transactions
- Events provide an immutable audit trail for all fund operations
- Pausable functionality allows for emergency response
- All transactions are logged with detailed metadata for transparency

## License

This project is part of the Digital Watchdog Initiative and is licensed under the MIT License. 