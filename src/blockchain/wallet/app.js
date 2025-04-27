// app.js - Main wallet application file

// Check if ethers is available before proceeding
if (typeof window.ethers === 'undefined') {
    console.error('Ethers library is not available. Please check your internet connection and reload the page.');
    throw new Error('Ethers library is not defined');
}

// Log ethers version for debugging
console.log('Ethers version in app.js:', window.ethers.version);

// Network configurations
const SEPOLIA_CHAIN_ID = '0xaa36a7'; // 11155111 in hex
const SEPOLIA_CONFIG = {
    chainId: SEPOLIA_CHAIN_ID,
    chainName: 'Sepolia',
    nativeCurrency: {
        name: 'Sepolia ETH',
        symbol: 'ETH',
        decimals: 18
    },
    rpcUrls: ['https://sepolia.infura.io/v3/'],
    blockExplorerUrls: ['https://sepolia.etherscan.io/']
};

// Global variables and contract instances
let contractAddresses = {};
let contractABIs = {};
let contractInstances = {};
let provider = null;
let signer = null;
let selectedContract = null;
let connectedAccount = null;
let isConnecting = false;
const eventLog = [];
let transactions = [];
const transactionTypes = {
    ALLOCATION: 0,
    DISBURSEMENT: 1,
    EXPENDITURE: 2,
    RETURNS: 3
};

const transactionStatus = {
    PENDING: 0,
    APPROVED: 1,
    COMPLETED: 2,
    FLAGGED: 3
};

// Initialize the application when the DOM is loaded
document.addEventListener('DOMContentLoaded', async () => {
  console.log('DOM loaded in app.js');
  
  try {
    // UI elements
    const connectButton = document.getElementById('connect-button');
    const disconnectButton = document.getElementById('disconnect-button');
    const connectionIndicator = document.getElementById('connection-indicator');
    const addressDisplay = document.getElementById('account-address');
    const balanceDisplay = document.getElementById('account-balance');
    const networkDisplay = document.getElementById('network-name');
    const contractSelect = document.getElementById('contract-select');
    const functionSelect = document.getElementById('function-select');
    const functionParams = document.getElementById('function-params');
    const executeButton = document.getElementById('execute-function');
    const resultOutput = document.getElementById('result-output');
    const eventLogContainer = document.getElementById('event-log');
    const loadingOverlay = document.getElementById('loading-overlay');
    const loadingMessage = document.getElementById('loading-message');
    const clearLogButton = document.getElementById('clear-log');
    const transactionsBody = document.getElementById('transactions-body');
    const filterType = document.getElementById('filter-type');
    const filterStatus = document.getElementById('filter-status');

    // Initialize provider and check for MetaMask
    if (typeof window.ethereum !== 'undefined') {
      try {
        // Set up ethers provider
        provider = new ethers.providers.Web3Provider(window.ethereum, 'any');
        
        // Setup MetaMask event listeners
        window.ethereum.on('accountsChanged', handleAccountsChanged);
        window.ethereum.on('chainChanged', () => window.location.reload());
        
        // Load contract data
        await loadContractAddresses();
        await loadContractABIs();
        populateContractSelect();

        // Check for previous connection
        const accounts = await window.ethereum.request({ method: 'eth_accounts' });
        if (accounts.length > 0) {
          console.log('Found existing connected accounts:', accounts);
          await initializeWithAccount(accounts[0]);
        }
        
        // Set up connect button's event listener to handle connection after MetaMask popup
        window.ethereum.on('connect', async () => {
          const accounts = await window.ethereum.request({ method: 'eth_accounts' });
          if (accounts.length > 0) {
            await initializeWithAccount(accounts[0]);
          }
        });
      } catch (error) {
        console.error('Error initializing wallet:', error);
        showError('Failed to initialize wallet: ' + error.message);
      }
      
      // Set up remaining event listeners
      disconnectButton.addEventListener('click', disconnectWallet);
      contractSelect.addEventListener('change', handleContractChange);
      functionSelect.addEventListener('change', handleFunctionChange);
      executeButton.addEventListener('click', executeFunction);
      clearLogButton.addEventListener('click', clearEventLog);
      filterType.addEventListener('change', updateTransactionsUI);
      filterStatus.addEventListener('change', updateTransactionsUI);
      
    } else {
      connectButton.disabled = true;
      connectButton.textContent = 'MetaMask Not Installed';
      showMetaMaskError();
    }

    // Function to initialize the app with a connected account
    async function initializeWithAccount(account) {
      console.log('Initializing with account:', account);
      provider = new ethers.providers.Web3Provider(window.ethereum);
      signer = provider.getSigner();
      connectedAccount = account;
      
      await updateUIWithWalletInfo(account);
      await loadContractData();
      setupEventListeners();
      await loadTransactions();
    }

    // Function to handle account changes
    function handleAccountsChanged(accounts) {
      console.log('Accounts changed:', accounts);
      if (accounts.length === 0) {
        disconnectWallet();
      } else {
        initializeWithAccount(accounts[0]);
      }
    }

    // Function to disconnect wallet
    function disconnectWallet() {
      isConnecting = false;
      connectedAccount = null;
      signer = null;
      
      // Clear any listeners
      if (window.ethereum) {
          window.ethereum.removeAllListeners('accountsChanged');
          window.ethereum.removeAllListeners('chainChanged');
      }
      
      // Update UI
      connectionIndicator.textContent = 'Disconnected';
      connectionIndicator.classList.remove('connected');
      addressDisplay.textContent = 'Not connected';
      balanceDisplay.textContent = '0 ETH';
      networkDisplay.textContent = 'Not connected';
      
      // Update button states
      connectButton.style.display = '';
      disconnectButton.style.display = 'none';
      
      // Disable contract interactions
      contractSelect.disabled = true;
      functionSelect.disabled = true;
      executeButton.disabled = true;
      
      // Clear function parameters
      functionParams.innerHTML = '';
      resultOutput.textContent = '';
      
      logEvent('info', 'Wallet Disconnected', 'Wallet has been disconnected');
    }

    // Update UI with wallet info
    async function updateUIWithWalletInfo(address) {
      try {
        // Update connection status
        connectionIndicator.textContent = 'Connected';
        connectionIndicator.classList.add('connected');
        
        // Format and display address
        const shortAddress = `${address.substring(0, 6)}...${address.substring(address.length - 4)}`;
        addressDisplay.textContent = shortAddress;
        addressDisplay.title = address;
        
        // Get and display balance
        const balance = await provider.getBalance(address);
        const formattedBalance = ethers.utils.formatEther(balance);
        balanceDisplay.textContent = `${parseFloat(formattedBalance).toFixed(4)} ETH`;
        
        // Get and display network
        const network = await provider.getNetwork();
        let networkName = network.name;
        if (network.chainId === 11155111) {
          networkName = 'Sepolia';
        }
        networkDisplay.textContent = networkName;
        
        // Enable UI elements
        connectButton.style.display = 'none';
        disconnectButton.style.display = '';
        contractSelect.disabled = false;
        
        logEvent('success', 'Wallet Connected', `Connected to address: ${shortAddress}`);
      } catch (error) {
        console.error('Error updating wallet info:', error);
        logEvent('error', 'UI Update Failed', error.message);
      }
    }

    // Populate contract select dropdown
    function populateContractSelect() {
      contractSelect.innerHTML = '<option value="">-- Select Contract --</option>';
      
      for (const [name, address] of Object.entries(contractAddresses)) {
        if (contractABIs[name]) {
          const option = document.createElement('option');
          option.value = name;
          option.textContent = name;
          contractSelect.appendChild(option);
        }
      }
    }

    // Handle contract selection change
    function handleContractChange() {
      const contractName = contractSelect.value;
      
      // Clear previous function options and params
      functionSelect.innerHTML = '<option value="">-- Select Function --</option>';
      functionParams.innerHTML = '';
      resultOutput.textContent = '';
      
      if (contractName && contractAddresses[contractName] && contractABIs[contractName]) {
        // Create contract instance
        selectedContract = new ethers.Contract(
          contractAddresses[contractName],
          contractABIs[contractName],
          signer
        );
        
        // Populate function select
        populateFunctionSelect(contractABIs[contractName]);
        
        // Enable function select
        functionSelect.disabled = false;
      } else {
        selectedContract = null;
        functionSelect.disabled = true;
        executeButton.disabled = true;
      }
    }

    // Populate function select dropdown
    function populateFunctionSelect(abi) {
      // Add option group for read functions
      const readOptGroup = document.createElement('optgroup');
      readOptGroup.label = 'Read Functions';
      
      // Add option group for write functions
      const writeOptGroup = document.createElement('optgroup');
      writeOptGroup.label = 'Write Functions';
      
      // Filter and sort functions
      abi.forEach(item => {
        if (item.type === 'function') {
          const option = document.createElement('option');
          option.value = item.name;
          
          // Create function signature for display
          let signature = `${item.name}(`;
          if (item.inputs && item.inputs.length > 0) {
            signature += item.inputs.map(input => `${input.type} ${input.name || ''}`).join(', ');
          }
          signature += ')';
          
          if (item.outputs && item.outputs.length > 0) {
            signature += ' returns (';
            signature += item.outputs.map(output => output.type).join(', ');
            signature += ')';
          }
          
          option.textContent = signature;
          
          // Add to appropriate group
          if (item.stateMutability === 'view' || item.stateMutability === 'pure') {
            readOptGroup.appendChild(option);
          } else {
            writeOptGroup.appendChild(option);
          }
        }
      });
      
      // Add option groups to select if they have children
      if (readOptGroup.children.length > 0) {
        functionSelect.appendChild(readOptGroup);
      }
      
      if (writeOptGroup.children.length > 0) {
        functionSelect.appendChild(writeOptGroup);
      }
    }

    // Handle function selection change
    function handleFunctionChange() {
      const functionName = functionSelect.value;
      functionParams.innerHTML = '';
      resultOutput.textContent = '';
      
      if (functionName && selectedContract) {
        // Find function definition
        const functionDef = contractABIs[contractSelect.value].find(
          item => item.type === 'function' && item.name === functionName
        );
        
        if (functionDef && functionDef.inputs) {
          // Create input fields for function parameters
          functionDef.inputs.forEach((input, index) => {
            const paramDiv = document.createElement('div');
            paramDiv.className = 'param-input';
            
            const label = document.createElement('label');
            label.textContent = `${input.name || `param${index}`} (${input.type})`;
            label.htmlFor = `param-${index}`;
            
            // For enum types, create a select instead of text input
            if (input.type.includes('uint8') && 
               (input.name && (input.name.toLowerCase().includes('type') || 
                              input.name.toLowerCase().includes('status')))) {
              createEnumSelect(paramDiv, input, index);
            } else {
              const inputField = document.createElement('input');
              inputField.type = 'text';
              inputField.id = `param-${index}`;
              inputField.name = input.name || `param${index}`;
              inputField.dataset.type = input.type;
              
              paramDiv.appendChild(label);
              paramDiv.appendChild(inputField);
            }
            
            functionParams.appendChild(paramDiv);
          });
          
          // Enable execute button
          executeButton.disabled = false;
        } else {
          // No parameters, still enable button
          executeButton.disabled = false;
        }
      } else {
        executeButton.disabled = true;
      }
    }

    // Create select for enum types
    function createEnumSelect(container, input, index) {
      const label = document.createElement('label');
      label.textContent = `${input.name || `param${index}`} (${input.type})`;
      label.htmlFor = `param-${index}`;
      
      const select = document.createElement('select');
      select.id = `param-${index}`;
      select.name = input.name || `param${index}`;
      select.dataset.type = input.type;
      
      // Add enum options based on input name
      if (input.name && input.name.toLowerCase().includes('status')) {
        // TransactionStatus enum
        const statuses = ['PENDING', 'APPROVED', 'REJECTED', 'COMPLETED', 'FLAGGED'];
        statuses.forEach((status, i) => {
          const option = document.createElement('option');
          option.value = i;
          option.textContent = status;
          select.appendChild(option);
        });
      } else if (input.name && input.name.toLowerCase().includes('type')) {
        // TransactionType enum
        const types = ['DISBURSEMENT', 'ALLOCATION', 'REFUND', 'TRANSFER'];
        types.forEach((type, i) => {
          const option = document.createElement('option');
          option.value = i;
          option.textContent = type;
          select.appendChild(option);
        });
      }
      
      container.appendChild(label);
      container.appendChild(select);
    }

    // Execute the selected function
    async function executeFunction() {
      if (!connectedAccount) {
        showError('Please connect your wallet first');
        return;
      }

      const contractName = document.getElementById('contract-select').value;
      const functionName = document.getElementById('function-select').value;

      if (!contractName || !functionName) {
        showError('Please select a contract and function');
        return;
      }

      try {
        showLoading(`Executing ${functionName}...`);

        // Get contract instance
        const contract = contractInstances[contractName];
        if (!contract) {
          throw new Error(`Contract ${contractName} is not initialized`);
        }

        // Get function parameters
        const params = [];
        const paramInputs = document.querySelectorAll('#function-params .param-input');
        
        paramInputs.forEach(input => {
          let value = input.value;
          const type = input.getAttribute('data-type');
          
          if (type.includes('uint')) {
            if (value === '') {
              value = '0';
            }
            value = ethers.BigNumber.from(value);
          } else if (type === 'bool') {
            value = value === 'true';
          } else if (type === 'address') {
            // Validate address format
            if (value && !ethers.utils.isAddress(value)) {
              throw new Error(`Invalid address format: ${value}`);
            }
          }
          
          params.push(value);
        });

        // Check if function is read-only or requires a transaction
        const functionFragment = contract.interface.getFunction(functionName);
        const isReadOnly = functionFragment.constant || functionFragment.stateMutability === 'view' || functionFragment.stateMutability === 'pure';
        
        let result;
        if (isReadOnly) {
          // Call read-only function
          result = await contract[functionName](...params);
        } else {
          // Send transaction
          const tx = await contract[functionName](...params);
          logEvent('info', 'Transaction Sent', `Transaction hash: ${tx.hash}`);
          
          // Wait for transaction to be mined
          showLoading('Waiting for transaction to be mined...');
          const receipt = await tx.wait();
          
          // Create a result object for display
          result = {
            transactionHash: receipt.transactionHash,
            blockNumber: receipt.blockNumber,
            gasUsed: receipt.gasUsed.toString(),
            events: receipt.events ? receipt.events.length : 0,
            status: receipt.status === 1 ? 'Success' : 'Failed'
          };
          
          // Log events
          if (receipt.events && receipt.events.length > 0) {
            receipt.events.forEach(event => {
              try {
                const eventName = event.event;
                const args = {};
                
                // Convert event arguments for display
                for (const key in event.args) {
                  if (isNaN(parseInt(key))) continue; // Skip numeric keys
                  
                  if (ethers.BigNumber.isBigNumber(event.args[key])) {
                    // Check if it might be an ether amount
                    if (key.toLowerCase().includes('amount')) {
                      args[key] = ethers.utils.formatEther(event.args[key]) + ' ETH';
                    } else if (key.toLowerCase().includes('time') || key.toLowerCase().includes('date')) {
                      // Could be a timestamp
                      args[key] = new Date(event.args[key].toNumber() * 1000).toISOString();
                    } else {
                      args[key] = event.args[key].toString();
                    }
                  } else if (typeof event.args[key] === 'object') {
                    args[key] = JSON.stringify(event.args[key]);
                  } else {
                    args[key] = event.args[key];
                  }
                }
                
                logEvent('success', `Event: ${eventName}`, JSON.stringify(args, null, 2));
              } catch (e) {
                console.error('Error parsing event:', e);
              }
            });
          }
        }

        // Display the result
        displayResult(result);
        
        hideLoading();
      } catch (error) {
        console.error('Error executing function:', error);
        hideLoading();
        showError(`Error: ${error.message}`);
      }
    }

    // Display function result
    function displayResult(result) {
      if (result === null || result === undefined) {
        resultOutput.textContent = 'No result (null or undefined)';
        return;
      }
      
      if (ethers.BigNumber.isBigNumber(result)) {
        resultOutput.textContent = result.toString();
      } else if (Array.isArray(result)) {
        // Handle array results (including tuples)
        let formattedArray = result.map(item => {
          if (ethers.BigNumber.isBigNumber(item)) {
            return item.toString();
          } else if (typeof item === 'object') {
            return JSON.stringify(item, null, 2);
          } else {
            return item;
          }
        });
        resultOutput.textContent = JSON.stringify(formattedArray, null, 2);
      } else if (typeof result === 'object') {
        // Handle object results (including structs)
        let formattedObject = {};
        
        // Convert BigNumber to strings for display
        for (const [key, value] of Object.entries(result)) {
          if (ethers.BigNumber.isBigNumber(value)) {
            formattedObject[key] = value.toString();
          } else {
            formattedObject[key] = value;
          }
        }
        
        resultOutput.textContent = JSON.stringify(formattedObject, null, 2);
      } else {
        resultOutput.textContent = result.toString();
      }
    }

    // Set up event listeners for contracts
    function setupEventListeners() {
      if (!signer) return;
      
      // Listen for TransactionEventLogger events
      if (contractAddresses['TransactionEventLogger']) {
        const logger = new ethers.Contract(
          contractAddresses['TransactionEventLogger'],
          contractABIs['TransactionEventLogger'],
          provider
        );
        
        // Remove existing listeners
        logger.removeAllListeners();
        
        // Listen for TransactionEvent events
        logger.on('TransactionEvent', (transactionId, fundId, txType, status, amount, recipient, timestamp, event) => {
          const eventObj = {
            name: 'TransactionEvent',
            timestamp: new Date().toISOString(),
            data: {
              transactionId: transactionId.toString(),
              fundId: fundId.toString(),
              txType: txType,
              status: status,
              amount: ethers.utils.formatEther(amount),
              recipient,
              timestamp: new Date(timestamp.toNumber() * 1000).toISOString(),
              blockNumber: event.blockNumber,
              transactionHash: event.transactionHash
            }
          };
          
          logEventToUI(eventObj);
        });
        
        console.log('Set up event listeners for TransactionEventLogger');
      }
      
      // Listen for FundManager events
      if (contractAddresses['FundManager']) {
        const manager = new ethers.Contract(
          contractAddresses['FundManager'],
          contractABIs['FundManager'],
          provider
        );
        
        // Remove existing listeners
        manager.removeAllListeners();
        
        // Listen for FundCreated events
        manager.on('FundCreated', (fundId, name, amount, createdBy, event) => {
          const eventObj = {
            name: 'FundCreated',
            timestamp: new Date().toISOString(),
            data: {
              fundId: fundId.toString(),
              name,
              amount: ethers.utils.formatEther(amount),
              createdBy,
              blockNumber: event.blockNumber,
              transactionHash: event.transactionHash
            }
          };
          
          logEventToUI(eventObj);
        });
        
        console.log('Set up event listeners for FundManager');
      }
    }

    // Log event to UI
    function logEventToUI(eventObj) {
      // Add to event log array
      eventLog.unshift(eventObj);
      updateEventLogUI();
    }

    // Log event from transaction
    function logEvent(event) {
      try {
        // Get event name and contract
        const eventName = event.event;
        const eventData = {};
        
        // Format event args based on event name and parameters
        for (const [key, value] of Object.entries(event.args)) {
          if (!isNaN(parseInt(key))) continue; // Skip numeric keys
          
          if (ethers.BigNumber.isBigNumber(value)) {
            // Check if it might be an ether amount
            if (key.toLowerCase().includes('amount')) {
              eventData[key] = ethers.utils.formatEther(value) + ' ETH';
            } else if (key.toLowerCase().includes('time') || key.toLowerCase().includes('date')) {
              // Could be a timestamp
              eventData[key] = new Date(value.toNumber() * 1000).toISOString();
            } else {
              eventData[key] = value.toString();
            }
          } else if (typeof value === 'object') {
            eventData[key] = JSON.stringify(value);
          } else {
            eventData[key] = value;
          }
        }
        
        // Add block and transaction info
        eventData.blockNumber = event.blockNumber;
        eventData.transactionHash = event.transactionHash;
        
        // Create event object
        const eventObj = {
          name: eventName,
          timestamp: new Date().toISOString(),
          data: eventData
        };
        
        // Log to UI
        logEventToUI(eventObj);
      } catch (error) {
        console.error('Error logging event:', error);
      }
    }

    // Update event log UI
    function updateEventLogUI() {
      eventLogContainer.innerHTML = '';
      
      if (eventLog.length === 0) {
        eventLogContainer.textContent = 'No events logged yet.';
        return;
      }
      
      // Display the latest events (limit to 10)
      const eventsToShow = eventLog.slice(0, 10);
      
      eventsToShow.forEach(event => {
        const eventItem = document.createElement('div');
        eventItem.className = 'event-item';
        
        const eventHeader = document.createElement('div');
        eventHeader.className = 'event-header';
        
        const eventName = document.createElement('span');
        eventName.className = 'event-name';
        eventName.textContent = event.name;
        
        const eventTimestamp = document.createElement('span');
        eventTimestamp.className = 'event-timestamp';
        eventTimestamp.textContent = new Date(event.timestamp).toLocaleString();
        
        eventHeader.appendChild(eventName);
        eventHeader.appendChild(eventTimestamp);
        
        const eventData = document.createElement('div');
        eventData.className = 'event-data';
        
        // Display event data
        for (const [key, value] of Object.entries(event.data)) {
          const dataItem = document.createElement('div');
          dataItem.className = 'event-data-item';
          
          const dataLabel = document.createElement('span');
          dataLabel.className = 'event-data-label';
          dataLabel.textContent = key + ':';
          
          const dataValue = document.createElement('span');
          dataValue.className = 'event-data-value';
          
          // For transaction hash, add a truncated view with full hover
          if (key === 'transactionHash') {
            const fullHash = value;
            const shortHash = `${fullHash.substring(0, 10)}...${fullHash.substring(fullHash.length - 8)}`;
            dataValue.textContent = shortHash;
            dataValue.title = fullHash;
          } else {
            dataValue.textContent = value;
          }
          
          dataItem.appendChild(dataLabel);
          dataItem.appendChild(dataValue);
          eventData.appendChild(dataItem);
        }
        
        eventItem.appendChild(eventHeader);
        eventItem.appendChild(eventData);
        eventLogContainer.appendChild(eventItem);
      });
    }

    // Clear event log
    function clearEventLog() {
      eventLog.length = 0; // Clear array
      updateEventLogUI();
    }

    // Show loading overlay
    function showLoading(message = 'Loading...') {
      loadingMessage.textContent = message;
      loadingOverlay.classList.add('active');
    }

    // Hide loading overlay
    function hideLoading() {
      loadingOverlay.classList.remove('active');
    }

    // Show MetaMask error
    function showMetaMaskError() {
      const message = document.createElement('div');
      message.className = 'error-message';
      message.innerHTML = `
          <h3>MetaMask Required</h3>
          <p>Please install MetaMask to use this application.</p>
          <a href="https://metamask.io" target="_blank" rel="noopener noreferrer">
              Install MetaMask
          </a>
      `;
      document.body.appendChild(message);
    }

    // Show error message
    function showError(message) {
      const errorDiv = document.createElement('div');
      errorDiv.className = 'error-message';
      errorDiv.textContent = message;
      document.body.appendChild(errorDiv);
      setTimeout(() => errorDiv.remove(), 5000);
    }

    // Transaction management
    async function loadTransactions() {
      showLoading('Loading transactions...');
      try {
          const fundManager = new ethers.Contract(
              contractAddresses['FundManager'],
              contractABIs['FundManager'],
              signer
          );
          
          // Get all transactions
          const txs = await fundManager.getAllTransactions();
          transactions = txs.map(tx => ({
              id: tx.id.toString(),
              fundId: tx.fundId.toString(),
              type: parseInt(tx.transactionType),
              amount: ethers.utils.formatEther(tx.amount),
              status: parseInt(tx.status),
              created: new Date(tx.timestamp.toNumber() * 1000).toLocaleString(),
              recipient: tx.recipient,
              description: tx.description
          }));
          
          updateTransactionsUI();
          hideLoading();
      } catch (error) {
          console.error('Error loading transactions:', error);
          hideLoading();
      }
    }

    function updateTransactionsUI() {
      const filterTypeValue = filterType.value;
      const filterStatusValue = filterStatus.value;
      
      // Filter transactions
      let filtered = [...transactions];
      if (filterTypeValue !== 'all') {
          filtered = filtered.filter(tx => tx.type === transactionTypes[filterTypeValue]);
      }
      if (filterStatusValue !== 'all') {
          filtered = filtered.filter(tx => tx.status === transactionStatus[filterStatusValue]);
      }
      
      // Sort by most recent first
      filtered.sort((a, b) => new Date(b.created) - new Date(a.created));
      
      // Generate table rows
      transactionsBody.innerHTML = filtered.map(tx => `
          <tr>
              <td>${tx.id}</td>
              <td>${getTransactionTypeName(tx.type)}</td>
              <td>${tx.amount} ETH</td>
              <td><span class="status-badge ${getStatusClass(tx.status)}">${getStatusName(tx.status)}</span></td>
              <td>${tx.created}</td>
              <td>
                  <button onclick="viewTransaction('${tx.id}')" class="action-button">View</button>
                  ${tx.status === transactionStatus.PENDING ? `
                      <button onclick="approveTransaction('${tx.id}')" class="action-button approve">Approve</button>
                      <button onclick="rejectTransaction('${tx.id}')" class="action-button reject">Reject</button>
                  ` : ''}
              </td>
          </tr>
      `).join('');
    }

    function getTransactionTypeName(type) {
      return Object.keys(transactionTypes).find(key => transactionTypes[key] === type) || 'Unknown';
    }

    function getStatusName(status) {
      return Object.keys(transactionStatus).find(key => transactionStatus[key] === status) || 'Unknown';
    }

    function getStatusClass(status) {
      switch(status) {
          case transactionStatus.PENDING: return 'status-pending';
          case transactionStatus.APPROVED: return 'status-approved';
          case transactionStatus.COMPLETED: return 'status-completed';
          case transactionStatus.FLAGGED: return 'status-flagged';
          default: return '';
      }
    }

    async function viewTransaction(id) {
      const tx = transactions.find(t => t.id === id);
      if (!tx) return;
      
      // Get transaction events
      const events = await getTransactionEvents(id);
      
      // Show transaction details modal
      showTransactionModal(tx, events);
    }

    async function getTransactionEvents(id) {
      try {
          const logger = new ethers.Contract(
              contractAddresses['TransactionEventLogger'],
              contractABIs['TransactionEventLogger'],
              signer
          );
          
          const events = await logger.getTransactionEvents(id);
          return events;
      } catch (error) {
          console.error('Error getting transaction events:', error);
          return [];
      }
    }

    function showTransactionModal(transaction, events) {
      const modal = document.createElement('div');
      modal.className = 'modal';
      modal.innerHTML = `
          <div class="modal-content">
              <h2>Transaction Details</h2>
              <div class="transaction-details">
                  <p><strong>ID:</strong> ${transaction.id}</p>
                  <p><strong>Type:</strong> ${getTransactionTypeName(transaction.type)}</p>
                  <p><strong>Amount:</strong> ${transaction.amount} ETH</p>
                  <p><strong>Status:</strong> ${getStatusName(transaction.status)}</p>
                  <p><strong>Recipient:</strong> ${transaction.recipient}</p>
                  <p><strong>Description:</strong> ${transaction.description}</p>
                  <p><strong>Created:</strong> ${transaction.created}</p>
              </div>
              <h3>Event History</h3>
              <div class="event-timeline">
                  ${events.map(event => `
                      <div class="timeline-item">
                          <div class="event-type">${getEventTypeName(event.eventType)}</div>
                          <div class="event-time">${new Date(event.timestamp.toNumber() * 1000).toLocaleString()}</div>
                          ${event.reason ? `<div class="event-reason">${event.reason}</div>` : ''}
                      </div>
                  `).join('')}
              </div>
              <button onclick="this.parentElement.parentElement.remove()" class="close-button">Close</button>
          </div>
      `;
      document.body.appendChild(modal);
    }

    function getEventTypeName(eventType) {
      const types = {
          0: 'Created',
          1: 'Approved',
          2: 'Rejected',
          3: 'Completed',
          4: 'Flagged'
      };
      return types[eventType] || 'Unknown';
    }

    async function approveTransaction(id) {
      showLoading('Approving transaction...');
      try {
          const fundManager = new ethers.Contract(
              contractAddresses['FundManager'],
              contractABIs['FundManager'],
              signer
          );
          
          const tx = await fundManager.approveTransaction(id);
          await tx.wait();
          
          // Refresh transactions
          await loadTransactions();
          hideLoading();
          
          // Show success message
          const message = document.createElement('div');
          message.className = 'success-message';
          message.textContent = 'Transaction approved successfully';
          document.body.appendChild(message);
          setTimeout(() => message.remove(), 3000);
      } catch (error) {
          console.error('Error approving transaction:', error);
          hideLoading();
          alert('Error approving transaction: ' + error.message);
      }
    }

    async function rejectTransaction(id) {
      if (!confirm('Are you sure you want to reject this transaction?')) return;
      
      showLoading('Rejecting transaction...');
      try {
          const fundManager = new ethers.Contract(
              contractAddresses['FundManager'],
              contractABIs['FundManager'],
              signer
          );
          
          const tx = await fundManager.rejectTransaction(id);
          await tx.wait();
          
          // Refresh transactions
          await loadTransactions();
          hideLoading();
          
          // Show success message
          const message = document.createElement('div');
          message.className = 'error-message';
          message.textContent = 'Transaction rejected';
          document.body.appendChild(message);
          setTimeout(() => message.remove(), 3000);
      } catch (error) {
          console.error('Error rejecting transaction:', error);
          hideLoading();
          alert('Error rejecting transaction: ' + error.message);
      }
    }

    // Add refresh transactions handler
    document.getElementById('refresh-transactions').addEventListener('click', loadTransactions);
  } catch (error) {
    console.error('Fatal error initializing application:', error);
    document.body.innerHTML = `<div class="error-container">
      <h1>Fatal Error</h1>
      <p>Failed to initialize application: ${error.message}</p>
      <p>Please check your internet connection and refresh the page.</p>
    </div>`;
  }
});