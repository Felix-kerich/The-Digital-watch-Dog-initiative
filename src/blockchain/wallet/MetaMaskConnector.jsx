import React, { useState, useEffect } from 'react';
import * as metamask from './metamask';
import * as eip6963 from './eip6963';
import * as contractInteraction from './contract-interaction';

// Replace with your actual deployed contract addresses
const TRANSACTION_LOGGER_ADDRESS = process.env.TRANSACTION_LOGGER_ADDRESS || '';
const FUND_MANAGER_ADDRESS = process.env.FUND_MANAGER_ADDRESS || '';

const MetaMaskConnector = () => {
  const [walletProviders, setWalletProviders] = useState([]);
  const [walletConnection, setWalletConnection] = useState(null);
  const [error, setError] = useState('');
  const [transactionHash, setTransactionHash] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  // Initialize wallet detection on component mount
  useEffect(() => {
    const cleanup = eip6963.initWalletDetection();
    
    // Update wallet providers every 2 seconds
    const intervalId = setInterval(() => {
      setWalletProviders([...eip6963.getWalletProviders()]);
    }, 2000);
    
    return () => {
      cleanup();
      clearInterval(intervalId);
    };
  }, []);

  // Legacy MetaMask fallback connection
  const connectMetaMask = async () => {
    try {
      setIsLoading(true);
      setError('');
      
      const connection = await metamask.connectMetaMask();
      setWalletConnection(connection);
    } catch (err) {
      setError(err.message || 'Failed to connect to MetaMask');
      console.error(err);
    } finally {
      setIsLoading(false);
    }
  };

  // EIP-6963 connection
  const connectWallet = async (provider) => {
    try {
      setIsLoading(true);
      setError('');
      
      const connection = await eip6963.connectWallet(provider);
      setWalletConnection(connection);
    } catch (err) {
      setError(err.message || `Failed to connect to ${provider.info.name}`);
      console.error(err);
    } finally {
      setIsLoading(false);
    }
  };

  // Example of interacting with TransactionEventLogger contract
  const logTransactionEvent = async () => {
    if (!walletConnection) {
      setError('Please connect your wallet first');
      return;
    }
    
    try {
      setIsLoading(true);
      setError('');
      setTransactionHash('');
      
      // Create ethers provider and signer
      const signer = contractInteraction.getSigner(walletConnection.provider);
      
      // Get contract instance
      const contract = contractInteraction.getTransactionEventLoggerContract(
        TRANSACTION_LOGGER_ADDRESS,
        signer
      );
      
      // Example transaction ID and data
      const transactionId = contractInteraction.stringToBytes32('test-transaction-' + Date.now());
      const actorHash = contractInteraction.stringToBytes32('user-' + walletConnection.address.substring(2, 10));
      const detailsHash = contractInteraction.stringToBytes32('details-' + Date.now());
      
      // Record transaction creation
      const receipt = await contractInteraction.recordTransactionCreation(
        contract,
        transactionId,
        actorHash,
        detailsHash
      );
      
      setTransactionHash(receipt.transactionHash);
    } catch (err) {
      setError(err.message || 'Failed to log transaction event');
      console.error(err);
    } finally {
      setIsLoading(false);
    }
  };

  // Disconnect wallet
  const disconnectWallet = () => {
    setWalletConnection(null);
  };

  return (
    <div className="metamask-connector">
      <h2>Wallet Connection</h2>
      
      {error && (
        <div className="error-message">
          <p>{error}</p>
        </div>
      )}
      
      {!walletConnection ? (
        <div className="connection-section">
          <h3>Connect Your Wallet</h3>
          
          {walletProviders.length > 0 ? (
            <div className="wallet-providers">
              <p>Available Wallets:</p>
              <div className="wallet-buttons">
                {walletProviders.map((provider) => (
                  <button
                    key={provider.info.uuid}
                    onClick={() => connectWallet(provider)}
                    disabled={isLoading}
                  >
                    <img 
                      src={provider.info.icon} 
                      alt={provider.info.name} 
                      width="24" 
                      height="24" 
                    />
                    Connect {provider.info.name}
                  </button>
                ))}
              </div>
            </div>
          ) : (
            <button 
              onClick={connectMetaMask} 
              disabled={isLoading}
            >
              Connect MetaMask (Legacy)
            </button>
          )}
        </div>
      ) : (
        <div className="wallet-info">
          <h3>Wallet Connected</h3>
          <p>
            <strong>Address:</strong> {walletConnection.address}
          </p>
          <p>
            <strong>Chain ID:</strong> {walletConnection.chainId}
          </p>
          {walletConnection.walletInfo && (
            <p>
              <strong>Wallet:</strong> {walletConnection.walletInfo.name}
            </p>
          )}
          
          <div className="actions">
            <button 
              onClick={logTransactionEvent}
              disabled={isLoading || !TRANSACTION_LOGGER_ADDRESS}
            >
              Log Transaction Event
            </button>
            
            <button 
              onClick={disconnectWallet}
              disabled={isLoading}
            >
              Disconnect
            </button>
          </div>
          
          {transactionHash && (
            <div className="transaction-info">
              <p>
                <strong>Transaction Hash:</strong> {transactionHash}
              </p>
              <a 
                href={`https://sepolia.etherscan.io/tx/${transactionHash}`}
                target="_blank"
                rel="noopener noreferrer"
              >
                View on Etherscan
              </a>
            </div>
          )}
        </div>
      )}
      
      {isLoading && <p>Loading...</p>}
    </div>
  );
};

export default MetaMaskConnector; 