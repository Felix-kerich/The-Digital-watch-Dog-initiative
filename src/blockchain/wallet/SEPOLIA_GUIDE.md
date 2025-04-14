# Sepolia Testnet ETH Guide

## Getting Sepolia ETH For Deployment

Your current wallet address (from the deployment logs): `0x1976Ea2978538479BeE0C6713F6dD26B2f76415d`

The deployment failed because your wallet needs Sepolia testnet ETH. Here's how to get some:

### Option 1: Chainlink Faucet (Recommended)
1. Go to [Chainlink Faucet](https://faucets.chain.link/sepolia)
2. Connect your wallet or enter your wallet address
3. Request 0.5 ETH
4. Wait for the transaction to complete (usually takes less than a minute)

### Option 2: Automata Sepolia Faucet
1. Visit [Sepolia Faucet by Automata](https://www.sepoliafaucet.io/)
2. Enter your wallet address
3. Complete the attestation process
4. Receive testnet ETH

### Option 3: GetBlock Faucet
1. Visit [GetBlock Faucet](https://getblock.io/faucet/eth-sepolia/)
2. Register and connect your wallet
3. Request testnet ETH

## Deploying After Getting Sepolia ETH

Once you have Sepolia ETH in your wallet, run the deployment command:

```bash
cd ~/Documents/The\ Digital\ Watchdog\ Initiative/src/blockchain
npx hardhat run scripts/deploy.js --network sepolia
```

The improved deployment script will:
1. Show detailed logs of the deployment process
2. Use optimal gas settings to avoid timeouts
3. Provide a summary of all deployed contracts
4. Generate `deployed-contracts.json` with all contract addresses

## After Successful Deployment

After deploying your contracts, you'll need to:

1. Update the backend `.env` file with the contract addresses:
```
CONTRACT_ADDRESS=0x...  # TransactionEventLogger address
```

2. Verify your contracts on Etherscan:
```bash
npx hardhat verify --network sepolia TRANSACTION_LOGGER_ADDRESS
npx hardhat verify --network sepolia FUND_MANAGER_ADDRESS
```

3. Test the contracts by interacting with them using the wallet integration you've created.

## Troubleshooting

If the deployment keeps failing:

1. **Make sure you have enough ETH**: Check your balance on [Sepolia Etherscan](https://sepolia.etherscan.io/)
2. **Network congestion**: Try deploying again during a less busy time
3. **Clean the cache**: Run `npx hardhat clean` before deploying
4. **Check your private key**: Ensure your private key in `.env` is correct (no `0x` prefix needed)
5. **Increase gas limit**: Try adjusting the gas settings in the deployment script

## Creating a New Wallet (If Needed)

If you want to create a new wallet for testing:

1. Generate a new private key (you can use MetaMask and export the private key)
2. Update your `.env` file with the new private key
3. Get Sepolia ETH for the new wallet using the faucets above
4. Run the deployment with the new wallet 