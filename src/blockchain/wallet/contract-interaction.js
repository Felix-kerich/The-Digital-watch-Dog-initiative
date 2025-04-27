/**
 * Contract Interaction Utility
 * Provides functions to interact with deployed smart contracts using ethers.js
 */

// Use the global ethers object instead of importing it
const ethers = window.ethers;

// Check for ethers availability
if (typeof ethers === 'undefined') {
  console.error('Ethers library is not available in contract-interaction.js');
  throw new Error('Ethers library is not defined');
}

// Paths to artifacts
const TRANSACTION_LOGGER_ARTIFACT_PATH = './TransactionEventLogger.json';
const FUND_MANAGER_ARTIFACT_PATH = './FundManager.json';
const DEPLOYED_CONTRACTS_PATH = '../deployed-contracts.json';

// Contract ABI variables - will be populated during initialization
let TransactionEventLoggerABI = null;
let FundManagerABI = null;

// Default provider (for reading from blockchain without signing)
export const getProvider = (rpcUrl) => {
  return new ethers.providers.JsonRpcProvider(rpcUrl);
};

// Get signer from connected wallet provider
export const getSigner = (provider) => {
  return new ethers.providers.Web3Provider(provider).getSigner();
};

// Utility functions
const formatAmount = (amount) => ethers.utils.formatEther(amount);
const parseAmount = (amount) => ethers.utils.parseEther(amount.toString());

/**
 * Load deployed contract addresses from the JSON file
 * @returns {Object|null} Contract addresses or null if not found
 */
export async function loadDeployedAddresses() {
  try {
    const response = await fetch('./deployed-contracts.json');
    if (!response.ok) {
      throw new Error('Failed to load deployed contracts');
    }
    return await response.json();
  } catch (error) {
    console.error('Error loading deployed addresses:', error);
    return null;
  }
}

/**
 * Load contract ABIs from JSON files
 */
export async function loadContractABIs() {
  try {
    // Load TransactionEventLogger ABI
    let response = await fetch(TRANSACTION_LOGGER_ARTIFACT_PATH);
    if (response.ok) {
      const data = await response.json();
      TransactionEventLoggerABI = data.abi;
    } else {
      // Try alternative path
      response = await fetch('../artifacts/contracts/TransactionEventLogger.sol/TransactionEventLogger.json');
      if (response.ok) {
        const data = await response.json();
        TransactionEventLoggerABI = data.abi;
      } else {
        console.warn('Failed to load TransactionEventLogger ABI');
      }
    }

    // Load FundManager ABI
    response = await fetch(FUND_MANAGER_ARTIFACT_PATH);
    if (response.ok) {
      const data = await response.json();
      FundManagerABI = data.abi;
    } else {
      // Try alternative path
      response = await fetch('../artifacts/contracts/FundManager.sol/FundManager.json');
      if (response.ok) {
        const data = await response.json();
        FundManagerABI = data.abi;
      } else {
        console.warn('Failed to load FundManager ABI');
      }
    }
    
    return {
      TransactionEventLogger: TransactionEventLoggerABI,
      FundManager: FundManagerABI
    };
  } catch (error) {
    console.error('Error loading contract ABIs:', error);
    return null;
  }
}

/**
 * Get a TransactionEventLogger contract instance
 * @param {Object} params - Parameters
 * @param {ethers.providers.JsonRpcSigner} params.signer - Ethers signer
 * @param {string} params.address - Contract address (optional)
 * @returns {Promise<ethers.Contract>} Contract instance
 */
export async function getTransactionEventLoggerContract({ signer, address }) {
  // Use ABI if available, otherwise use minimal ABI
  const abi = TransactionEventLoggerABI || [
    "function logTransactionEvent(uint256 transactionId, uint8 eventType, string memory actorHash, string memory detailsHash, string memory reason) external",
    "function getTransactionEvents(uint256 transactionId) external view returns (tuple(uint8 eventType, uint256 timestamp, string actorHash, string detailsHash, string reason)[])",
    "function hasRole(bytes32 role, address account) external view returns (bool)",
    "function getRoleMember(bytes32 role, uint256 index) external view returns (address)",
    "function getRoleMemberCount(bytes32 role) external view returns (uint256)",
    "function grantRole(bytes32 role, address account) external",
    "function API_ROLE() external view returns (bytes32)"
  ];

  // Use provided address or load from deployed contracts
  let contractAddress = address;
  if (!contractAddress) {
    const deployedAddresses = await loadDeployedAddresses();
    if (deployedAddresses && deployedAddresses.TransactionEventLogger) {
      contractAddress = deployedAddresses.TransactionEventLogger;
    } else {
      throw new Error('TransactionEventLogger address not found');
    }
  }

  return new ethers.Contract(contractAddress, abi, signer);
}

/**
 * Get a FundManager contract instance
 * @param {Object} params - Parameters
 * @param {ethers.providers.JsonRpcSigner} params.signer - Ethers signer
 * @param {string} params.address - Contract address (optional)
 * @returns {Promise<ethers.Contract>} Contract instance
 */
export async function getFundManagerContract({ signer, address }) {
  // ABI for the FundManager contract
  const abi = [
    "function createFund(string memory name, string memory description, uint256 amount, string memory metadataHash) external returns (uint256)",
    "function createTransaction(uint256 fundId, uint8 transactionType, uint256 amount, string memory description, string memory recipientId, string memory metadataHash) external returns (uint256)",
    "function approveTransaction(uint256 transactionId) external",
    "function rejectTransaction(uint256 transactionId, string memory reason) external",
    "function completeTransaction(uint256 transactionId) external",
    "function flagTransaction(uint256 transactionId, string memory reason) external",
    "function getFundDetails(uint256 fundId) external view returns (string memory name, string memory description, uint256 totalAmount, uint256 allocatedAmount, uint256 spentAmount, uint8 status, string memory metadataHash)",
    "function getTransactionDetails(uint256 transactionId) external view returns (uint256 fundId, uint8 transactionType, uint256 amount, string memory description, string memory recipientId, uint8 status, string memory metadataHash)",
    "function addAuthorizedUpdater(address updater) external",
    "function isAuthorizedUpdater(address account) external view returns (bool)"
  ];

  // Use provided address or load from deployed contracts
  let contractAddress = address;
  if (!contractAddress) {
    const deployedAddresses = await loadDeployedAddresses();
    if (deployedAddresses && deployedAddresses.FundManager) {
      contractAddress = deployedAddresses.FundManager;
    } else {
      throw new Error('FundManager address not found');
    }
  }

  return new ethers.Contract(contractAddress, abi, signer);
}

// Helper function to convert string to bytes32
export function stringToBytes32(str) {
  // Right-pad the string with null bytes to 32 bytes
  return ethers.utils.formatBytes32String(str);
}

// Helper function to convert bytes32 to string
export function bytes32ToString(bytes32) {
  return ethers.utils.parseBytes32String(bytes32);
}

// Helper function to format amount to wei (for amounts stored in contracts)
export function formatAmountToWei(amount) {
  return ethers.utils.parseEther(amount.toString());
}

// Helper function to format wei to amount (for UI display)
export function formatWeiToAmount(wei) {
  return ethers.utils.formatEther(wei);
}

/**
 * Execute a function on the TransactionEventLogger contract
 * @param {Object} params - Parameters
 * @param {ethers.providers.JsonRpcSigner} params.signer - Ethers signer
 * @param {string} params.functionName - Function name to execute
 * @param {Object} params.args - Function arguments
 * @param {string} params.address - Contract address (optional)
 * @returns {Promise<any>} - Function result
 */
export async function executeEventLoggerFunction({ signer, functionName, args, address }) {
  const contract = await getTransactionEventLoggerContract({ signer, address });
  
  switch (functionName) {
    case 'logTransactionEvent': {
      const { transactionId, eventType, actorHash, detailsHash, reason } = args;
      
      // Convert eventType to number if it's a string
      const eventTypeNum = typeof eventType === 'string' ? parseInt(eventType) : eventType;
      
      const tx = await contract.logTransactionEvent(
        transactionId,
        eventTypeNum,
        actorHash || '',
        detailsHash || '',
        reason || ''
      );
      
      return tx;
    }
    
    case 'getTransactionEvents': {
      const { transactionId } = args;
      const events = await contract.getTransactionEvents(transactionId);
      
      return {
        events: events.map(event => ({
          eventType: event.eventType,
          timestamp: event.timestamp.toNumber(),
          actorHash: event.actorHash,
          detailsHash: event.detailsHash,
          reason: event.reason
        }))
      };
    }
    
    case 'hasRole': {
      const { role, account } = args;
      return await contract.hasRole(role, account);
    }
    
    case 'getRoleMember': {
      const { role, index } = args;
      return await contract.getRoleMember(role, parseInt(index));
    }
    
    case 'getRoleMemberCount': {
      const { role } = args;
      const count = await contract.getRoleMemberCount(role);
      return count.toNumber();
    }
    
    case 'grantRole': {
      const { role, account } = args;
      const tx = await contract.grantRole(role, account);
      return tx;
    }
    
    case 'API_ROLE': {
      return await contract.API_ROLE();
    }
    
    default:
      throw new Error(`Unknown function: ${functionName}`);
  }
}

/**
 * Execute a function on the FundManager contract
 * @param {Object} params - Parameters
 * @param {ethers.providers.JsonRpcSigner} params.signer - Ethers signer
 * @param {string} params.functionName - Function name to execute
 * @param {Object} params.args - Function arguments
 * @param {string} params.address - Contract address (optional)
 * @returns {Promise<any>} - Function result
 */
export async function executeFundManagerFunction({ signer, functionName, args, address }) {
  const contract = await getFundManagerContract({ signer, address });
  
  switch (functionName) {
    case 'createFund': {
      const { name, description, amount, metadataHash } = args;
      const amountWei = parseAmount(amount);
      
      const tx = await contract.createFund(
        name,
        description,
        amountWei,
        metadataHash || ''
      );
      
      return tx;
    }
    
    case 'createTransaction': {
      const { fundId, transactionType, amount, description, recipientId, metadataHash } = args;
      
      // Convert transactionType to number if it's a string
      const typeNum = typeof transactionType === 'string' ? parseInt(transactionType) : transactionType;
      const amountWei = parseAmount(amount);
      
      const tx = await contract.createTransaction(
        fundId,
        typeNum,
        amountWei,
        description,
        recipientId,
        metadataHash || ''
      );
      
      return tx;
    }
    
    case 'approveTransaction': {
      const { transactionId } = args;
      const tx = await contract.approveTransaction(transactionId);
      return tx;
    }
    
    case 'rejectTransaction': {
      const { transactionId, reason } = args;
      const tx = await contract.rejectTransaction(transactionId, reason || '');
      return tx;
    }
    
    case 'completeTransaction': {
      const { transactionId } = args;
      const tx = await contract.completeTransaction(transactionId);
      return tx;
    }
    
    case 'flagTransaction': {
      const { transactionId, reason } = args;
      const tx = await contract.flagTransaction(transactionId, reason || '');
      return tx;
    }
    
    case 'getFundDetails': {
      const { fundId } = args;
      const fundDetails = await contract.getFundDetails(fundId);
      
      return {
        name: fundDetails.name,
        description: fundDetails.description,
        totalAmount: formatAmount(fundDetails.totalAmount),
        allocatedAmount: formatAmount(fundDetails.allocatedAmount),
        spentAmount: formatAmount(fundDetails.spentAmount),
        status: fundDetails.status,
        metadataHash: fundDetails.metadataHash
      };
    }
    
    case 'getTransactionDetails': {
      const { transactionId } = args;
      const txDetails = await contract.getTransactionDetails(transactionId);
      
      return {
        fundId: txDetails.fundId.toNumber(),
        transactionType: txDetails.transactionType,
        amount: formatAmount(txDetails.amount),
        description: txDetails.description,
        recipientId: txDetails.recipientId,
        status: txDetails.status,
        metadataHash: txDetails.metadataHash
      };
    }
    
    case 'addAuthorizedUpdater': {
      const { updater } = args;
      const tx = await contract.addAuthorizedUpdater(updater);
      return tx;
    }
    
    case 'isAuthorizedUpdater': {
      const { account } = args;
      return await contract.isAuthorizedUpdater(account);
    }
    
    default:
      throw new Error(`Unknown function: ${functionName}`);
  }
}

/**
 * Record a transaction creation event
 * @param {Object} params - Parameters
 * @param {ethers.Signer} params.signer - Ethers signer
 * @param {string} params.transactionId - Transaction ID
 * @param {string} params.actorHash - Actor hash (user identifier)
 * @param {string} params.detailsHash - Transaction details hash
 * @returns {Promise<Object>} Transaction result
 */
export async function recordTransactionCreation({ signer, transactionId, actorHash, detailsHash }) {
  try {
    return await executeEventLoggerFunction({
      signer,
      functionName: 'recordTransactionCreation',
      args: { transactionId, actorHash, detailsHash }
    });
  } catch (error) {
    console.error('Error recording transaction creation:', error);
    throw error;
  }
}

/**
 * Record a transaction approval event
 * @param {Object} params - Parameters
 * @param {ethers.Signer} params.signer - Ethers signer
 * @param {string} params.transactionId - Transaction ID
 * @param {string} params.actorHash - Actor hash (user identifier)
 * @returns {Promise<Object>} Transaction result
 */
export async function recordApproval({ signer, transactionId, actorHash }) {
  try {
    return await executeEventLoggerFunction({
      signer,
      functionName: 'recordApproval',
      args: { transactionId, actorHash }
    });
  } catch (error) {
    console.error('Error recording approval:', error);
    throw error;
  }
}

/**
 * Record a transaction rejection event
 * @param {Object} params - Parameters
 * @param {ethers.Signer} params.signer - Ethers signer
 * @param {string} params.transactionId - Transaction ID
 * @param {string} params.actorHash - Actor hash (user identifier)
 * @returns {Promise<Object>} Transaction result
 */
export async function recordRejection({ signer, transactionId, actorHash }) {
  try {
    return await executeEventLoggerFunction({
      signer,
      functionName: 'recordRejection',
      args: { transactionId, actorHash }
    });
  } catch (error) {
    console.error('Error recording rejection:', error);
    throw error;
  }
}

/**
 * Record a transaction completion event
 * @param {Object} params - Parameters
 * @param {ethers.Signer} params.signer - Ethers signer
 * @param {string} params.transactionId - Transaction ID
 * @param {string} params.actorHash - Actor hash (user identifier)
 * @returns {Promise<Object>} Transaction result
 */
export async function recordCompletion({ signer, transactionId, actorHash }) {
  try {
    return await executeEventLoggerFunction({
      signer,
      functionName: 'recordCompletion',
      args: { transactionId, actorHash }
    });
  } catch (error) {
    console.error('Error recording completion:', error);
    throw error;
  }
}

/**
 * Record a transaction flagging event
 * @param {Object} params - Parameters
 * @param {ethers.Signer} params.signer - Ethers signer
 * @param {string} params.transactionId - Transaction ID
 * @param {string} params.reason - Reason for flagging
 * @returns {Promise<Object>} Transaction result
 */
export async function recordFlagging({ signer, transactionId, reason }) {
  try {
    return await executeEventLoggerFunction({
      signer,
      functionName: 'recordFlagging',
      args: { transactionId, reason }
    });
  } catch (error) {
    console.error('Error recording flagging:', error);
    throw error;
  }
}

/**
 * Get all events for a transaction
 * @param {Object} params - Parameters
 * @param {ethers.providers.Provider|ethers.Signer} params.providerOrSigner - Ethers provider or signer
 * @param {string} params.transactionId - Transaction ID
 * @returns {Promise<Array>} Array of transaction events
 */
export async function getTransactionEvents({ providerOrSigner, transactionId }) {
  try {
    const contract = await getTransactionEventLoggerContract({ providerOrSigner });
    const events = await contract.getTransactionEvents(transactionId);
    return events;
  } catch (error) {
    console.error('Error getting transaction events:', error);
    throw error;
  }
}

/**
 * Initialize the contract interaction module
 * Loads deployed addresses and sets up global references
 */
export async function initContractInteraction() {
  const addresses = await loadDeployedAddresses();
  if (addresses) {
    window.TRANSACTION_LOGGER_ADDRESS = addresses.transactionEventLogger;
    window.FUND_MANAGER_ADDRESS = addresses.fundManager;
    return addresses;
  }
  return null;
}

// Export additional functions and helpers
export {
  executeEventLoggerFunction,
  executeFundManagerFunction
}; 