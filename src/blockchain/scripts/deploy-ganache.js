const hre = require("hardhat");
const { ethers } = require("hardhat");
const fs = require("fs");
const path = require("path");

async function main() {
  console.log("Starting Ganache deployment process...");
  
  try {
    // Get accounts
    const accounts = await ethers.getSigners();
    const deployer = accounts[0];
    console.log(`Deploying with account: ${deployer.address}`);
    
    // Display account balance
    const balance = await deployer.getBalance();
    console.log(`Account balance: ${ethers.utils.formatEther(balance)} ETH`);

    // Make sure account has enough balance on Ganache
    if (balance.eq(ethers.BigNumber.from(0))) {
      console.log("Warning: Account balance is 0 ETH. This may be due to Ganache misconfiguration.");
      console.log("Using alternative account if available...");

      // Try to find an account with balance
      for (let i = 1; i < accounts.length; i++) {
        const altBalance = await accounts[i].getBalance();
        if (altBalance.gt(ethers.BigNumber.from(0))) {
          console.log(`Using alternative account ${accounts[i].address} with balance ${ethers.utils.formatEther(altBalance)} ETH`);
          deployer = accounts[i];
          break;
        }
      }
      
      // Get updated balance
      const updatedBalance = await deployer.getBalance();
      console.log(`Updated account balance: ${ethers.utils.formatEther(updatedBalance)} ETH`);
      
      if (updatedBalance.eq(ethers.BigNumber.from(0))) {
        throw new Error("No account with balance found. Please ensure Ganache is properly configured.");
      }
    }
    
    // Get network info
    const network = await ethers.provider.getNetwork();
    console.log(`Network: ${network.name}`);
    console.log(`Chain ID: ${network.chainId}`);
    
    // Deploy TransactionEventLogger with appropriate settings
    console.log("Deploying TransactionEventLogger...");
    const TransactionEventLogger = await ethers.getContractFactory("TransactionEventLogger");
    const logger = await TransactionEventLogger.deploy();
    
    console.log("Waiting for TransactionEventLogger transaction to be mined...");
    await logger.deployed();
    console.log("✅ TransactionEventLogger deployed to:", logger.address);

    // Deploy FundManager
    console.log("Deploying FundManager...");
    const FundManager = await ethers.getContractFactory("FundManager");
    const manager = await FundManager.deploy();
    
    console.log("Waiting for FundManager transaction to be mined...");
    await manager.deployed();
    console.log("✅ FundManager deployed to:", manager.address);

    // Grant API_ROLE to the deployer
    console.log("Granting API_ROLE to deployer...");
    const API_ROLE = ethers.utils.keccak256(ethers.utils.toUtf8Bytes("API_ROLE"));
    const grantRoleTx = await logger.grantRole(API_ROLE, deployer.address);
    
    console.log("Waiting for grantRole transaction to be mined...");
    await grantRoleTx.wait();
    console.log("✅ API_ROLE granted to:", deployer.address);

    // Add deployer as authorized updater in FundManager
    console.log("Adding deployer as authorized updater in FundManager...");
    const authUpdaterTx = await manager.addAuthorizedUpdater(deployer.address);
    
    console.log("Waiting for addAuthorizedUpdater transaction to be mined...");
    await authUpdaterTx.wait();
    console.log("✅ Authorized updater added to FundManager:", deployer.address);

    // Write contract addresses to files for the backend and wallet
    const contracts = {
      TransactionEventLogger: logger.address,
      FundManager: manager.address,
      network: "ganache",
      chainId: network.chainId,
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
    console.log(`Network: ganache (Chain ID: ${network.chainId})`);
    console.log(`TransactionEventLogger: ${logger.address}`);
    console.log(`FundManager: ${manager.address}`);
    console.log("Deployment successful!");
    
  } catch (error) {
    console.error("\n❌ DEPLOYMENT FAILED");
    console.error("Error message:", error.message);
    
    // Handle common Ganache issues
    if (error.message.includes("doesn't have enough funds")) {
      console.error("Error: Account has insufficient funds for deployment");
      console.error("Please configure Ganache to provide test ETH to accounts");
      console.error("Try running: npx ganache-cli --gasLimit 30000000 --accounts 10 --defaultBalanceEther 1000");
    } else if (error.message.includes("Exceeds block gas limit")) {
      console.error("Error: Exceeds block gas limit");
      console.error("Try increasing block gas limit in Ganache: --gasLimit 30000000");
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