// wallet.js - Handles Web3 integration and smart contract interactions

// Global variables
let currentAccount = null;
let eventLoggerContract = null;
let fundManagerContract = null;
let web3Provider = null;
let deployedContracts = {};

// Contract ABIs will be loaded from JSON files
let eventLoggerABI;
let fundManagerABI;

// Wallet and Blockchain Interaction Logic
let provider, signer, connectedAccount;
let contracts = {};
let contractABIs = {};
let isConnected = false;

// Contract addresses - will be loaded from the deployed contracts
let contractAddresses = {};

// Initialize wallet and contract functionality
async function initContractInteraction() {
    try {
        // Load deployed contract addresses
        await loadDeployedContracts();
        
        // Load contract ABIs
        await loadContractABIs();
        
        if (window.ethereum && currentAccount) {
            // Initialize contract instances if wallet is connected
            await initializeContractInstances();
        }
    } catch (error) {
        console.error('Error initializing contracts:', error);
    }
}

async function loadDeployedContracts() {
    try {
        const response = await fetch('../deployed-contracts.json');
        if (response.ok) {
            deployedContracts = await response.json();
        } else {
            console.error('Failed to load deployed contracts');
        }
    } catch (error) {
        console.error('Error loading deployed contracts:', error);
    }
}

async function loadContractABIs() {
    try {
        // Load TransactionEventLogger ABI
        const eventLoggerResponse = await fetch('../artifacts/contracts/TransactionEventLogger.sol/TransactionEventLogger.json');
        if (eventLoggerResponse.ok) {
            const eventLoggerData = await eventLoggerResponse.json();
            eventLoggerABI = eventLoggerData.abi;
        } else {
            console.error('Failed to load TransactionEventLogger ABI');
        }
        
        // Load FundManager ABI
        const fundManagerResponse = await fetch('../artifacts/contracts/FundManager.sol/FundManager.json');
        if (fundManagerResponse.ok) {
            const fundManagerData = await fundManagerResponse.json();
            fundManagerABI = fundManagerData.abi;
        } else {
            console.error('Failed to load FundManager ABI');
        }
    } catch (error) {
        console.error('Error loading contract ABIs:', error);
    }
}

async function initializeContractInstances() {
    if (!window.ethereum || !currentAccount) {
        console.error('Ethereum provider not available or wallet not connected');
        return;
    }
    
    try {
        const provider = new ethers.providers.Web3Provider(window.ethereum);
        const signer = provider.getSigner();
        
        // Initialize TransactionEventLogger contract
        if (deployedContracts.transactionEventLogger && eventLoggerABI) {
            eventLoggerContract = new ethers.Contract(
                deployedContracts.transactionEventLogger,
                eventLoggerABI,
                signer
            );
        }
        
        // Initialize FundManager contract
        if (deployedContracts.fundManager && fundManagerABI) {
            fundManagerContract = new ethers.Contract(
                deployedContracts.fundManager,
                fundManagerABI,
                signer
            );
        }
    } catch (error) {
        console.error('Error initializing contract instances:', error);
    }
}

// Wallet Connection Functions
async function checkWalletConnection() {
    if (window.ethereum) {
        // Check if already connected
        const accounts = await window.ethereum.request({ method: 'eth_accounts' });
        if (accounts.length > 0) {
            currentAccount = accounts[0];
            return true;
        }
    }
    return false;
}

async function connectWallet() {
    if (!window.ethereum) {
        alert('Please install MetaMask to use this dApp!');
        return;
    }
    
    try {
        showLoading();
        const accounts = await window.ethereum.request({ 
            method: 'eth_requestAccounts' 
        });
        
        if (accounts.length > 0) {
            currentAccount = accounts[0];
            await initializeContractInstances();
            updateUIForConnectedWallet();
        }
        hideLoading();
    } catch (error) {
        hideLoading();
        console.error('Error connecting wallet:', error);
        alert('Failed to connect wallet');
    }
}

function disconnectWallet() {
    currentAccount = null;
    eventLoggerContract = null;
    fundManagerContract = null;
    updateUIForDisconnectedWallet();
}

function getCurrentAccount() {
    return currentAccount;
}

async function getAccountBalance() {
    if (!window.ethereum || !currentAccount) {
        return '0';
    }
    
    try {
        const provider = new ethers.providers.Web3Provider(window.ethereum);
        const balance = await provider.getBalance(currentAccount);
        return ethers.utils.formatEther(balance);
    } catch (error) {
        console.error('Error getting balance:', error);
        return '0';
    }
}

// Set up event listeners for account changes and chain changes
if (window.ethereum) {
    window.ethereum.on('accountsChanged', (accounts) => {
        if (accounts.length === 0) {
            // User disconnected wallet
            disconnectWallet();
        } else {
            // Account changed
            currentAccount = accounts[0];
            updateUIForConnectedWallet();
        }
    });
    
    window.ethereum.on('chainChanged', () => {
        // Handle chain change - reload the page
        window.location.reload();
    });
}

// Contract Interaction Functions
async function executeEventLoggerFunction(functionName, params = []) {
    if (!eventLoggerContract) {
        throw new Error('Transaction Event Logger contract not initialized');
    }
    
    try {
        const result = await eventLoggerContract[functionName](...params);
        return result;
    } catch (error) {
        console.error(`Error executing ${functionName}:`, error);
        throw error;
    }
}

async function executeFundManagerFunction(functionName, params = []) {
    if (!fundManagerContract) {
        throw new Error('Fund Manager contract not initialized');
    }
    
    try {
        const result = await fundManagerContract[functionName](...params);
        return result;
    } catch (error) {
        console.error(`Error executing ${functionName}:`, error);
        throw error;
    }
}

// Utility Functions
function parseEther(amount) {
    try {
        return ethers.utils.parseEther(amount);
    } catch (error) {
        console.error('Error parsing ether amount:', error);
        throw new Error('Invalid amount format');
    }
}

function formatEtherValue(value) {
    try {
        if (typeof value === 'bigint' || (typeof value === 'object' && value._isBigNumber)) {
            return ethers.utils.formatEther(value) + ' ETH';
        }
        return value.toString();
    } catch (error) {
        console.error('Error formatting ether value:', error);
        return value.toString();
    }
}

// Initialize the application
async function initApp() {
    try {
        // Load deployed contract addresses
        await loadContractAddresses();
        
        // Load contract ABIs
        await loadContractABIs();
        
        // Check if browser has ethereum provider
        if (window.ethereum) {
            provider = new ethers.providers.Web3Provider(window.ethereum);
            
            // Check if already connected
            const accounts = await provider.listAccounts();
            if (accounts.length > 0) {
                await connectWallet();
            }
            
            // Listen for account changes
            window.ethereum.on('accountsChanged', handleAccountChange);
            window.ethereum.on('chainChanged', () => window.location.reload());
        } else {
            addEventToLog('error', 'No Ethereum Provider', 'Please install MetaMask or use a Web3-enabled browser');
        }
        
        // Setup UI components
        setupUIComponents();
    } catch (error) {
        console.error('Initialization error:', error);
        addEventToLog('error', 'Initialization Failed', error.message);
    }
}

// Load contract addresses from deployed-contracts.json
async function loadContractAddresses() {
    try {
        const response = await fetch('../deployed-contracts.json');
        if (!response.ok) {
            throw new Error('Failed to load contract addresses');
        }
        contractAddresses = await response.json();
        console.log('Loaded contract addresses:', contractAddresses);
    } catch (error) {
        console.error('Error loading contract addresses:', error);
        addEventToLog('error', 'Failed to Load Addresses', error.message);
    }
}

// Load contract ABIs from artifacts
async function loadContractABIs() {
    try {
        // Load TransactionEventLogger ABI
        const loggerResponse = await fetch('../artifacts/contracts/TransactionEventLogger.sol/TransactionEventLogger.json');
        if (!loggerResponse.ok) {
            throw new Error('Failed to load TransactionEventLogger ABI');
        }
        const loggerData = await loggerResponse.json();
        contractABIs['TransactionEventLogger'] = loggerData.abi;
        
        // Load FundManager ABI
        const managerResponse = await fetch('../artifacts/contracts/FundManager.sol/FundManager.json');
        if (!managerResponse.ok) {
            throw new Error('Failed to load FundManager ABI');
        }
        const managerData = await managerResponse.json();
        contractABIs['FundManager'] = managerData.abi;
        
        console.log('Loaded contract ABIs');
    } catch (error) {
        console.error('Error loading contract ABIs:', error);
        addEventToLog('error', 'Failed to Load ABIs', error.message);
    }
}

// Connect wallet
async function connectWallet() {
    try {
        showLoading('Connecting wallet...');
        
        // Request account access
        const accounts = await window.ethereum.request({ method: 'eth_requestAccounts' });
        connectedAccount = accounts[0];
        
        // Get signer
        signer = provider.getSigner();
        
        // Initialize contract instances
        if (contractAddresses.TransactionEventLogger) {
            contracts.TransactionEventLogger = new ethers.Contract(
                contractAddresses.TransactionEventLogger,
                contractABIs['TransactionEventLogger'],
                signer
            );
        }
        
        if (contractAddresses.FundManager) {
            contracts.FundManager = new ethers.Contract(
                contractAddresses.FundManager,
                contractABIs['FundManager'],
                signer
            );
        }
        
        isConnected = true;
        updateUIOnConnect();
        populateContractSelector();
        
        // Update account info
        const balance = await provider.getBalance(connectedAccount);
        const formattedBalance = ethers.utils.formatEther(balance);
        
        document.getElementById('account-address').textContent = connectedAccount;
        document.getElementById('account-balance').textContent = `${formattedBalance} ETH`;
        
        addEventToLog('success', 'Wallet Connected', `Connected to account: ${connectedAccount}`);
        hideLoading();
    } catch (error) {
        console.error('Connection error:', error);
        addEventToLog('error', 'Connection Failed', error.message);
        hideLoading();
    }
}

// Disconnect wallet
function disconnectWallet() {
    connectedAccount = null;
    signer = null;
    isConnected = false;
    updateUIOnDisconnect();
    addEventToLog('info', 'Wallet Disconnected', 'Wallet has been disconnected');
}

// Handle account change
async function handleAccountChange(accounts) {
    if (accounts.length === 0) {
        // User disconnected their wallet
        disconnectWallet();
    } else if (accounts[0] !== connectedAccount) {
        // User switched accounts
        connectedAccount = accounts[0];
        signer = provider.getSigner();
        
        // Reinitialize contract instances with new signer
        if (contractAddresses.TransactionEventLogger) {
            contracts.TransactionEventLogger = new ethers.Contract(
                contractAddresses.TransactionEventLogger,
                contractABIs['TransactionEventLogger'],
                signer
            );
        }
        
        if (contractAddresses.FundManager) {
            contracts.FundManager = new ethers.Contract(
                contractAddresses.FundManager,
                contractABIs['FundManager'],
                signer
            );
        }
        
        // Update account info
        const balance = await provider.getBalance(connectedAccount);
        const formattedBalance = ethers.utils.formatEther(balance);
        
        document.getElementById('account-address').textContent = connectedAccount;
        document.getElementById('account-balance').textContent = `${formattedBalance} ETH`;
        
        addEventToLog('info', 'Account Changed', `Switched to account: ${connectedAccount}`);
    }
}

// Execute contract function
async function executeContractFunction() {
    if (!isConnected) {
        addEventToLog('error', 'Not Connected', 'Please connect your wallet first');
        return;
    }
    
    const contractName = document.getElementById('contract-select').value;
    const functionName = document.getElementById('function-select').value;
    
    if (!contractName || !functionName) {
        addEventToLog('error', 'Selection Required', 'Please select both contract and function');
        return;
    }
    
    try {
        showLoading(`Executing ${functionName}...`);
        
        // Get function parameters
        const params = [];
        const paramContainers = document.querySelectorAll('.param-container');
        
        paramContainers.forEach(container => {
            const input = container.querySelector('input, select');
            if (input) {
                // Convert value based on data type
                let value = input.value;
                const dataType = input.getAttribute('data-type');
                
                if (dataType === 'uint256' || dataType === 'uint') {
                    value = ethers.BigNumber.from(value);
                } else if (dataType === 'bool') {
                    value = value === 'true';
                } else if (dataType === 'bytes32') {
                    value = ethers.utils.formatBytes32String(value);
                }
                
                params.push(value);
            }
        });
        
        // Execute function
        const contract = contracts[contractName];
        const result = await contract[functionName](...params);
        
        // Check if result is a transaction
        if (result.hash) {
            addEventToLog('info', 'Transaction Sent', `Transaction hash: ${result.hash}`);
            
            // Wait for transaction to be mined
            const receipt = await result.wait();
            addEventToLog('success', 'Transaction Mined', `Transaction confirmed in block ${receipt.blockNumber}`);
            
            // Parse and display events
            if (receipt.events && receipt.events.length > 0) {
                receipt.events.forEach(event => {
                    try {
                        const eventName = event.event;
                        const eventData = JSON.stringify(event.args, (key, value) => {
                            if (typeof value === 'object' && value._isBigNumber) {
                                return value.toString();
                            }
                            return value;
                        }, 2);
                        
                        addEventToLog('success', `Event: ${eventName}`, eventData);
                    } catch (e) {
                        console.error('Error parsing event:', e);
                    }
                });
            }
            
            // Update result display
            displayResult(JSON.stringify({
                transactionHash: receipt.transactionHash,
                blockNumber: receipt.blockNumber,
                gasUsed: receipt.gasUsed.toString(),
                status: receipt.status === 1 ? 'Success' : 'Failed'
            }, null, 2));
        } else {
            // Call result
            const resultDisplay = typeof result === 'object' 
                ? JSON.stringify(result, (key, value) => {
                    if (typeof value === 'object' && value._isBigNumber) {
                        return value.toString();
                    }
                    return value;
                }, 2)
                : result.toString();
            
            displayResult(resultDisplay);
            addEventToLog('success', 'Call Successful', `Function ${functionName} executed successfully`);
        }
        
        hideLoading();
    } catch (error) {
        console.error('Execution error:', error);
        hideLoading();
        displayResult(`Error: ${error.message}`);
        addEventToLog('error', 'Execution Failed', error.message);
    }
}

// Listen for events
function listenForEvents(contractName) {
    if (!contracts[contractName]) return;
    
    const contract = contracts[contractName];
    
    // Get all event names from ABI
    const eventFragments = contractABIs[contractName]
        .filter(fragment => fragment.type === 'event')
        .map(fragment => fragment.name);
    
    // Listen for all events
    eventFragments.forEach(eventName => {
        contract.on(eventName, (...args) => {
            const eventObj = args[args.length - 1];
            const eventData = JSON.stringify(
                args.slice(0, args.length - 1).reduce((acc, arg, index) => {
                    // Try to get parameter name from ABI
                    const fragment = contract.interface.getEvent(eventName);
                    const paramName = fragment.inputs[index]?.name || `param${index}`;
                    acc[paramName] = arg.toString();
                    return acc;
                }, {}),
                null, 2
            );
            
            addEventToLog('success', `Event: ${eventName}`, eventData);
        });
    });
    
    addEventToLog('info', 'Event Listener', `Listening for events from ${contractName}`);
}

// Initialize app when DOM is ready
document.addEventListener('DOMContentLoaded', initApp); 