/**
 * MetaMask Wallet Connector
 * Provides functions to connect to MetaMask and interact with the Ethereum blockchain
 */

// Function to check if MetaMask is installed
export const isMetaMaskInstalled = () => {
  return typeof window !== 'undefined' && window.ethereum && window.ethereum.isMetaMask;
};

// Function to connect to MetaMask
export const connectMetaMask = async () => {
  if (!isMetaMaskInstalled()) {
    throw new Error("MetaMask is not installed. Please install MetaMask to use this feature.");
  }
  
  try {
    const accounts = await window.ethereum.request({ method: 'eth_requestAccounts' });
    return {
      address: accounts[0],
      chainId: await getChainId(),
      provider: window.ethereum
    };
  } catch (error) {
    console.error("Failed to connect to MetaMask:", error);
    throw error;
  }
};

// Get current chain ID
export const getChainId = async () => {
  if (!isMetaMaskInstalled()) return null;
  try {
    return await window.ethereum.request({ method: 'eth_chainId' });
  } catch (error) {
    console.error("Error getting chain ID:", error);
    return null;
  }
};

// Switch networks
export const switchNetwork = async (chainId) => {
  if (!isMetaMaskInstalled()) return false;
  
  try {
    await window.ethereum.request({
      method: 'wallet_switchEthereumChain',
      params: [{ chainId }],
    });
    return true;
  } catch (error) {
    // This error code indicates that the chain has not been added to MetaMask
    if (error.code === 4902) {
      return false;
    }
    console.error("Error switching network:", error);
    return false;
  }
};

// Add Sepolia network if it doesn't exist
export const addSepoliaNetwork = async () => {
  if (!isMetaMaskInstalled()) return false;
  
  try {
    await window.ethereum.request({
      method: 'wallet_addEthereumChain',
      params: [
        {
          chainId: '0xaa36a7', // 11155111 in hex
          chainName: 'Sepolia Testnet',
          nativeCurrency: {
            name: 'Sepolia ETH',
            symbol: 'ETH',
            decimals: 18,
          },
          rpcUrls: ['https://sepolia.infura.io/v3/'],
          blockExplorerUrls: ['https://sepolia.etherscan.io/'],
        },
      ],
    });
    return true;
  } catch (error) {
    console.error("Error adding Sepolia network:", error);
    return false;
  }
};

// MetaMask network configuration
export const GANACHE_NETWORK = {
    chainId: '0x539', // 1337 in hex
    chainName: 'Ganache',
    nativeCurrency: {
        name: 'ETH',
        symbol: 'ETH',
        decimals: 18
    },
    rpcUrls: ['http://127.0.0.1:8545'],
    blockExplorerUrls: []
};

// Function to setup Ganache network in MetaMask
export async function setupGanacheNetwork() {
    try {
        if (!window.ethereum) {
            throw new Error('MetaMask is not installed');
        }

        // First try to switch to the network if it exists
        try {
            await window.ethereum.request({
                method: 'wallet_switchEthereumChain',
                params: [{ chainId: GANACHE_NETWORK.chainId }]
            });
        } catch (switchError) {
            // Network doesn't exist, add it
            if (switchError.code === 4902) {
                await window.ethereum.request({
                    method: 'wallet_addEthereumChain',
                    params: [GANACHE_NETWORK]
                });
            } else {
                throw switchError;
            }
        }

        return true;
    } catch (error) {
        console.error('Error setting up Ganache network:', error);
        throw error;
    }
}

// Function to check if currently on Ganache network
export async function isGanacheNetwork() {
    try {
        const chainId = await window.ethereum.request({ method: 'eth_chainId' });
        return chainId === GANACHE_NETWORK.chainId;
    } catch (error) {
        console.error('Error checking network:', error);
        return false;
    }
}

// Sign a message using MetaMask
export const signMessage = async (message) => {
  if (!isMetaMaskInstalled()) return null;
  
  try {
    const accounts = await window.ethereum.request({ method: 'eth_accounts' });
    if (!accounts || accounts.length === 0) {
      throw new Error("No accounts found. Please connect to MetaMask first.");
    }
    
    const signature = await window.ethereum.request({
      method: 'personal_sign',
      params: [message, accounts[0]],
    });
    
    return signature;
  } catch (error) {
    console.error("Error signing message:", error);
    return null;
  }
};

// Send a transaction
export const sendTransaction = async (to, value, data = '') => {
  if (!isMetaMaskInstalled()) return null;
  
  try {
    const accounts = await window.ethereum.request({ method: 'eth_accounts' });
    if (!accounts || accounts.length === 0) {
      throw new Error("No accounts found. Please connect to MetaMask first.");
    }
    
    const transactionParameters = {
      from: accounts[0],
      to,
      value: `0x${parseInt(value).toString(16)}`,
      data,
    };
    
    const txHash = await window.ethereum.request({
      method: 'eth_sendTransaction',
      params: [transactionParameters],
    });
    
    return txHash;
  } catch (error) {
    console.error("Error sending transaction:", error);
    return null;
  }
};

// Listen for account changes
export const listenForAccountChanges = (callback) => {
  if (!isMetaMaskInstalled()) return () => {};
  
  window.ethereum.on('accountsChanged', callback);
  
  // Return function to remove listener
  return () => {
    window.ethereum.removeListener('accountsChanged', callback);
  };
};

// Listen for chain changes
export const listenForChainChanges = (callback) => {
  if (!isMetaMaskInstalled()) return () => {};
  
  window.ethereum.on('chainChanged', callback);
  
  // Return function to remove listener
  return () => {
    window.ethereum.removeListener('chainChanged', callback);
  };
};

// Get ETH balance for an address
export const getBalance = async (address) => {
  if (!isMetaMaskInstalled()) return null;
  
  try {
    const balance = await window.ethereum.request({
      method: 'eth_getBalance',
      params: [address, 'latest'],
    });
    
    // Convert balance from wei to ETH
    return parseInt(balance, 16) / 1e18;
  } catch (error) {
    console.error("Error getting balance:", error);
    return null;
  }
};