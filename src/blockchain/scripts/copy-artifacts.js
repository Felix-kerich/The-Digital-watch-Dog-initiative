const fs = require('fs');
const path = require('path');

// Define paths
const artifactsDir = path.resolve(__dirname, '../artifacts/contracts');
const walletDir = path.resolve(__dirname, '../wallet');

// Create wallet directory if it doesn't exist
if (!fs.existsSync(walletDir)) {
  fs.mkdirSync(walletDir, { recursive: true });
  console.log('Created wallet directory');
}

// Copy TransactionEventLogger.json and FundManager.json
const contracts = ['TransactionEventLogger', 'FundManager'];

contracts.forEach(contractName => {
  const sourceFile = path.resolve(artifactsDir, `${contractName}.sol/${contractName}.json`);
  const destFile = path.resolve(walletDir, `${contractName}.json`);
  
  try {
    if (fs.existsSync(sourceFile)) {
      fs.copyFileSync(sourceFile, destFile);
      console.log(`Copied ${contractName}.json to wallet directory`);
    } else {
      console.error(`Source file not found: ${sourceFile}`);
    }
  } catch (error) {
    console.error(`Error copying ${contractName}.json:`, error);
  }
});

console.log('Artifacts copy complete'); 