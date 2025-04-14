/**
 * EIP-6963 Wallet Connector
 * Modern wallet detection that supports multiple installed wallets
 */

// Store for detected wallets
let walletProviders = [];

// Initialize the EIP-6963 wallet detection
export const initWalletDetection = () => {
  walletProviders = [];
  
  // Function to handle announce events
  const handleAnnounce = (event) => {
    // Check if provider is already registered
    if (!walletProviders.some(p => p.info.uuid === event.detail.info.uuid)) {
      walletProviders.push(event.detail);
    }
  };

  // Add event listener for wallet announcements
  if (typeof window !== 'undefined') {
    window.addEventListener('eip6963:announceProvider', handleAnnounce);
    
    // Request providers to announce themselves
    window.dispatchEvent(new Event('eip6963:requestProvider'));
  }
  
  // Return cleanup function
  return () => {
    if (typeof window !== 'undefined') {
      window.removeEventListener('eip6963:announceProvider', handleAnnounce);
    }
  };
};

// Get all detected wallet providers
export const getWalletProviders = () => {
  return walletProviders;
};

// Connect to a specific wallet provider
export const connectWallet = async (provider) => {
  if (!provider || !provider.provider) {
    throw new Error('Invalid wallet provider');
  }
  
  try {
    const accounts = await provider.provider.request({
      method: 'eth_requestAccounts'
    });
    
    const chainId = await provider.provider.request({
      method: 'eth_chainId'
    });
    
    return {
      address: accounts[0],
      chainId,
      provider: provider.provider,
      walletInfo: provider.info
    };
  } catch (error) {
    console.error(`Failed to connect to ${provider.info.name}:`, error);
    throw error;
  }
};

// Utilities for working with connected wallets

// Get the balance of an account
export const getAccountBalance = async (provider, address) => {
  try {
    const balance = await provider.request({
      method: 'eth_getBalance',
      params: [address, 'latest']
    });
    
    return parseInt(balance, 16) / 1e18; // Convert wei to ETH
  } catch (error) {
    console.error('Error getting account balance:', error);
    throw error;
  }
};

// Sign a message
export const signMessage = async (provider, address, message) => {
  try {
    return await provider.request({
      method: 'personal_sign',
      params: [message, address]
    });
  } catch (error) {
    console.error('Error signing message:', error);
    throw error;
  }
};

// Send a transaction
export const sendTransaction = async (provider, params) => {
  try {
    return await provider.request({
      method: 'eth_sendTransaction',
      params: [params]
    });
  } catch (error) {
    console.error('Error sending transaction:', error);
    throw error;
  }
};

// Switch to a specific chain
export const switchChain = async (provider, chainId) => {
  try {
    await provider.request({
      method: 'wallet_switchEthereumChain',
      params: [{ chainId }]
    });
    return true;
  } catch (error) {
    if (error.code === 4902) {
      // Chain not added
      return false;
    }
    console.error('Error switching chain:', error);
    throw error;
  }
};

// Add Sepolia testnet to wallet
export const addSepoliaNetwork = async (provider) => {
  try {
    await provider.request({
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
    console.error('Error adding Sepolia network:', error);
    throw error;
  }
}; 