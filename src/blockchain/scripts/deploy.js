const hre = require("hardhat");
const { ethers } = require("hardhat");
const fs = require("fs");
const path = require("path");

async function main() {
  console.log("Starting deployment process...");
  console.log(`Network: ${hre.network.name}`);
  
  try {
    // Get deployer address
    const [deployer] = await ethers.getSigners();
    console.log(`Deploying with account: ${deployer.address}`);
    
    // Display account balance
    const balance = await deployer.getBalance();
    console.log(`Account balance: ${ethers.utils.formatEther(balance)} ETH`);
    
    // Get network info
    const { chainId } = await ethers.provider.getNetwork();
    console.log(`Chain ID: ${chainId}`);
    
    // Get current gas price and increase by 20%
    const gasPrice = await ethers.provider.getGasPrice();
    const adjustedGasPrice = gasPrice.mul(120).div(100);
    console.log(`Current gas price: ${ethers.utils.formatUnits(gasPrice, "gwei")} gwei`);
    console.log(`Using gas price: ${ethers.utils.formatUnits(adjustedGasPrice, "gwei")} gwei`);
    
    // Deploy TransactionEventLogger with specific gas settings
    console.log("Deploying TransactionEventLogger...");
    const TransactionEventLogger = await hre.ethers.getContractFactory("TransactionEventLogger");
    const logger = await TransactionEventLogger.deploy({
      gasPrice: adjustedGasPrice,
      gasLimit: 8000000 // Increased gas limit
    });
    
    console.log("Waiting for TransactionEventLogger transaction to be mined...");
    console.log("Transaction hash:", logger.deployTransaction.hash);
    
    await logger.deployed();
    console.log("✅ TransactionEventLogger deployed to:", logger.address);

    // Deploy FundManager
    console.log("Deploying FundManager...");
    const FundManager = await hre.ethers.getContractFactory("FundManager");
    const manager = await FundManager.deploy({
      gasPrice: adjustedGasPrice,
      gasLimit: 10000000 // Higher gas limit for potentially larger contract
    });
    
    console.log("Waiting for FundManager transaction to be mined...");
    console.log("Transaction hash:", manager.deployTransaction.hash);
    
    await manager.deployed();
    console.log("✅ FundManager deployed to:", manager.address);

    // Grant API_ROLE to the service account
    console.log("Granting API_ROLE to deployer...");
    const API_ROLE = ethers.utils.keccak256(ethers.utils.toUtf8Bytes("API_ROLE"));
    const grantRoleTx = await logger.grantRole(API_ROLE, deployer.address, {
      gasPrice: adjustedGasPrice,
      gasLimit: 300000
    });
    console.log("Transaction hash:", grantRoleTx.hash);
    
    console.log("Waiting for grantRole transaction to be mined...");
    await grantRoleTx.wait();
    console.log("✅ API_ROLE granted to:", deployer.address);

    // Add deployer as authorized updater in FundManager
    console.log("Adding deployer as authorized updater in FundManager...");
    const authUpdaterTx = await manager.addAuthorizedUpdater(deployer.address, {
      gasPrice: adjustedGasPrice,
      gasLimit: 300000
    });
    console.log("Transaction hash:", authUpdaterTx.hash);
    
    console.log("Waiting for addAuthorizedUpdater transaction to be mined...");
    await authUpdaterTx.wait();
    console.log("✅ Authorized updater added to FundManager:", deployer.address);

    // Write contract addresses to files for the backend and wallet
    const contracts = {
      TransactionEventLogger: logger.address,
      FundManager: manager.address,
      network: hre.network.name,
      chainId: chainId,
      deployedBy: deployer.address,
      deploymentTime: new Date().toISOString()
    };

    // Write to root directory
    fs.writeFileSync(
      "deployed-contracts.json",
      JSON.stringify(contracts, null, 2)
    );
    
    // Write to wallet directory for the frontend
    const walletDir = path.join(__dirname, "..", "wallet");
    if (fs.existsSync(walletDir)) {
      fs.writeFileSync(
        path.join(walletDir, "deployed-contracts.json"),
        JSON.stringify(contracts, null, 2)
      );
      console.log("✅ Contract addresses written to wallet/deployed-contracts.json");
    }
    
    console.log("✅ Contract addresses written to deployed-contracts.json");
    
    console.log("\n=== DEPLOYMENT SUMMARY ===");
    console.log(`Network: ${hre.network.name} (Chain ID: ${chainId})`);
    console.log(`TransactionEventLogger: ${logger.address}`);
    console.log(`FundManager: ${manager.address}`);
    console.log("Deployment successful!");
    
    // Suggest next steps
    console.log("\n=== NEXT STEPS ===");
    console.log("1. Update your backend .env file with the deployed contract addresses");
    console.log("2. Verify contracts on Etherscan:");
    console.log(`   npx hardhat verify --network ${hre.network.name} ${logger.address}`);
    console.log(`   npx hardhat verify --network ${hre.network.name} ${manager.address}`);
    
  } catch (error) {
    console.error("\n❌ DEPLOYMENT FAILED");
    console.error("Error message:", error.message);
    
    // Provide more helpful error information
    if (error.message.includes("insufficient funds")) {
      console.error("\nError: Insufficient funds for deployment");
      console.error("Make sure your account has enough ETH for deployment and gas fees");
      console.error("Get Sepolia testnet ETH from https://sepoliafaucet.com/");
    } else if (error.message.includes("timeout")) {
      console.error("\nError: Deployment timed out");
      console.error("Try again with higher gas price or check your network connection");
      console.error("If using Infura, their free tier might be rate limited. Try again in a few minutes.");
    } else if (error.message.includes("nonce")) {
      console.error("\nError: Nonce error - transaction ordering issue");
      console.error("Try running: npx hardhat clean");
    } else if (error.message.includes("network") || error.message.includes("connection")) {
      console.error("\nError: Network connection issue");
      console.error("Check your internet connection and the Infura API key");
      console.error("Make sure your URL is correct in hardhat.config.js");
    } else if (error.code === "NETWORK_ERROR") {
      console.error("\nError: Network error - could not connect to Infura/Ethereum node");
      console.error("The Infura free tier might be rate limited or your API key may be incorrect");
    } else if (error.message.includes("estimateGas") || error.message.includes("gas")) {
      console.error("\nError: Gas estimation failed");
      console.error("Try manually setting a higher gas limit in hardhat.config.js");
    }
    
    throw error;
  }
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  }); 