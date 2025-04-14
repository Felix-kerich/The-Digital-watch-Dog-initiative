require("@nomiclabs/hardhat-waffle");
require("@nomiclabs/hardhat-ethers");
require("@nomiclabs/hardhat-etherscan");
require("dotenv").config();

/**
 * @type import('hardhat/config').HardhatUserConfig
 */
module.exports = {
  solidity: {
    version: "0.8.19",
    settings: {
      viaIR: true,
      optimizer: {
        enabled: true,
        runs: 200
      }
    }
  },
  networks: {
    ganache: {
      url: process.env.BLOCKCHAIN_SERVICE_URL || "http://127.0.0.1:8545",
      accounts: [process.env.BLOCKCHAIN_PRIVATE_KEY].filter(Boolean),
      gas: 8000000,  // set this higher if needed
      gasPrice: 24000000000,
      chainId: 1337,
      timeout: 60000,
      blockGasLimit: 30000000
    },
    ganache_ui: {
      url: "http://127.0.0.1:7545",
      accounts: [process.env.BLOCKCHAIN_PRIVATE_KEY].filter(Boolean),
      gas: 8000000,  // set this higher if needed
      gasPrice: 24000000000,
      chainId: 1337,
      timeout: 60000,
      blockGasLimit: 30000000
    },
    hardhat: {
      chainId: 1337,
      blockGasLimit: 30000000
    },
    // Add other networks as needed (e.g., testnet, mainnet)
    sepolia: {
      url: `https://sepolia.infura.io/v3/${process.env.INFURA_API_KEY}`,
      accounts: [process.env.PRIVATE_KEY].filter(Boolean),
      chainId: 11155111,
      gasPrice: 80000000000, // 80 gwei
      gas: 12000000, // 12 million gas limit
      timeout: 300000 // 5 minutes timeout
    }
  },
  etherscan: {
    apiKey: process.env.ETHERSCAN_API_KEY
  },
  paths: {
    sources: "./contracts",
    tests: "./test",
    cache: "./cache",
    artifacts: "./artifacts"
  },
  mocha: {
    timeout: 100000 // Longer timeout for tests
  }
}; 