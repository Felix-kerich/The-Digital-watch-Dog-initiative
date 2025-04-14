# Setting up MetaMask for Digital Watchdog Initiative

This guide will help you set up MetaMask to work with your local Ganache blockchain.

## Install MetaMask

1. Visit [MetaMask.io](https://metamask.io/) and install the browser extension for your browser (Chrome, Firefox, Brave, or Edge).
2. Create a new wallet or import an existing one following the on-screen instructions.
3. Make sure you securely store your recovery phrase.

## Add Ganache Network to MetaMask

1. Open MetaMask and click on the network dropdown at the top of the extension (likely says "Ethereum Mainnet").
2. Select "Add Network".
3. Click "Add a network manually".
4. Fill in the following details:
   - **Network Name**: `Ganache Local`
   - **New RPC URL**: `http://localhost:8545`
   - **Chain ID**: `1337`
   - **Currency Symbol**: `ETH`
   - **Block Explorer URL**: *(leave blank)*
5. Click "Save".

## Import a Ganache Account

To interact with your deployed contracts, you'll need to use the same account that deployed them. Here's how to import a Ganache account:

1. Run Ganache using our script or the Ganache CLI/UI app.
2. Copy one of the private keys from the Ganache output (preferably the first account).
3. In MetaMask, click on your account icon in the top-right corner.
4. Select "Import Account".
5. Paste the private key (including the "0x" prefix if needed) and click "Import".

The imported account should now appear in your account list and have 1000 ETH (or another amount, depending on your Ganache configuration).

## Test the Connection

1. Make sure Ganache is running and you're connected to the Ganache network in MetaMask.
2. Open the wallet interface in your browser.
3. Click "Connect Wallet" - MetaMask should prompt you to connect.
4. After connecting, you should see your account address and balance in the wallet interface.
5. Select a contract and try executing a view function (like `getTransactionCount` or `getFundCount`).

## Troubleshooting

### Cannot connect to Ganache

- Make sure Ganache is running (`npx ganache-cli --port 8545`)
- Verify the RPC URL in MetaMask matches the port Ganache is running on
- Try restarting your browser

### Transaction fails with "insufficient funds"

- Make sure you're using an account with ETH in Ganache
- Check the account balance in MetaMask
- Verify you're on the correct network

### Contract functions not available

- Make sure contracts are deployed to the Ganache network
- Check that the contract artifacts are correctly copied to the wallet directory
- Verify the `deployed-contracts.json` file has the correct addresses

### Gas estimation failed

- Try increasing the gas limit in hardhat.config.js
- Some complex functions may require manual gas estimation

## Advanced: Connect to Sepolia Testnet

See the [SEPOLIA_GUIDE.md](./SEPOLIA_GUIDE.md) file for instructions on connecting to the Sepolia testnet for more permanent deployments.