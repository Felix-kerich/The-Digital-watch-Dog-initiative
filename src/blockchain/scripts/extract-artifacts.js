const fs = require('fs');
const path = require('path');

// Paths
const artifactsDir = path.join(__dirname, '../artifacts/contracts');
const walletDir = path.join(__dirname, '../wallet');
const deployedContractsPath = path.join(__dirname, '../deployed-contracts.json');

// Ensure wallet directory exists
if (!fs.existsSync(walletDir)) {
  fs.mkdirSync(walletDir, { recursive: true });
}

// Contract names to extract
const contractsToExtract = [
  'TransactionEventLogger',
  'FundManager'
];

// Read deployed contracts file if it exists
let deployedContracts = {};
if (fs.existsSync(deployedContractsPath)) {
  try {
    deployedContracts = JSON.parse(fs.readFileSync(deployedContractsPath, 'utf8'));
    console.log('Loaded deployed contract addresses');
  } catch (error) {
    console.error('Error reading deployed contracts file:', error);
  }
}

// Copy deployed contracts file to wallet directory
fs.writeFileSync(
  path.join(walletDir, 'deployed-contracts.json'),
  JSON.stringify(deployedContracts, null, 2)
);
console.log('Copied deployed contracts to wallet directory');

// Extract artifacts for each contract
contractsToExtract.forEach(contractName => {
  // Find the artifact file
  const contractArtifactDir = path.join(artifactsDir, `${contractName}.sol`);
  if (!fs.existsSync(contractArtifactDir)) {
    console.error(`Artifact directory for ${contractName} not found`);
    return;
  }
  
  const artifactPath = path.join(contractArtifactDir, `${contractName}.json`);
  if (!fs.existsSync(artifactPath)) {
    console.error(`Artifact for ${contractName} not found`);
    return;
  }
  
  try {
    // Read the artifact
    const artifactData = JSON.parse(fs.readFileSync(artifactPath, 'utf8'));
    
    // Extract just what we need for the wallet interface
    const extractedData = {
      contractName: artifactData.contractName,
      abi: artifactData.abi,
      bytecode: artifactData.bytecode
    };
    
    // Write extracted data to wallet directory
    fs.writeFileSync(
      path.join(walletDir, `${contractName}.json`),
      JSON.stringify(extractedData, null, 2)
    );
    
    console.log(`Extracted ${contractName} artifact`);
  } catch (error) {
    console.error(`Error extracting ${contractName} artifact:`, error);
  }
});

console.log('Artifact extraction complete'); 