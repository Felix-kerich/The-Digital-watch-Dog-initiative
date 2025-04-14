// Contract ABIs and addresses
let contractAddresses = {};
let contractABIs = {};
let provider, signer, selectedContract, connectedAddress;
const eventLog = [];

// Initialize the application when the DOM is loaded
document.addEventListener('DOMContentLoaded', async () => {
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

  // Try to load contracts data
  try {
    await loadContractAddresses();
    await loadContractABIs();
    populateContractSelect();
  } catch (error) {
    console.error('Error initializing wallet:', error);
  }

  // Check if MetaMask is installed
  if (window.ethereum) {
    provider = new ethers.providers.Web3Provider(window.ethereum);
    
    // Check for previous connection
    try {
      const accounts = await provider.listAccounts();
      if (accounts.length > 0) {
        await connectWallet(false);
      }
    } catch (error) {
      console.error('Error checking previous connection:', error);
    }
    
    // Setup MetaMask events
    window.ethereum.on('accountsChanged', (accounts) => {
      if (accounts.length === 0) {
        disconnectWallet();
      } else {
        updateUIWithWalletInfo(accounts[0]);
      }
    });
    
    window.ethereum.on('chainChanged', () => {
      window.location.reload();
    });
  } else {
    connectButton.disabled = true;
    connectButton.textContent = 'MetaMask Not Installed';
    alert('Please install MetaMask to use this application.');
  }

  // Event Listeners
  connectButton.addEventListener('click', () => connectWallet(true));
  disconnectButton.addEventListener('click', disconnectWallet);
  contractSelect.addEventListener('change', handleContractChange);
  functionSelect.addEventListener('change', handleFunctionChange);
  executeButton.addEventListener('click', executeFunction);
  clearLogButton.addEventListener('click', clearEventLog);

  // Load contract addresses
  async function loadContractAddresses() {
    try {
      const response = await fetch('./deployed-contracts.json');
      if (response.ok) {
        contractAddresses = await response.json();
        console.log('Loaded contract addresses:', contractAddresses);
        return true;
      } else {
        throw new Error('Failed to load contracts data');
      }
    } catch (error) {
      console.error('Error loading contract addresses:', error);
      
      // Try loading sample contract addresses
      try {
        const sampleResponse = await fetch('./sample-deployed-contracts.json');
        if (sampleResponse.ok) {
          contractAddresses = await sampleResponse.json();
          console.log('Loaded sample contract addresses:', contractAddresses);
          return true;
        } 
      } catch (sampleError) {
        console.error('Error loading sample contract addresses:', sampleError);
      }
      
      return false;
    }
  }

  // Load contract ABIs from artifacts
  async function loadContractABIs() {
    try {
      // Try to fetch ABIs from artifacts
      const abiPaths = {
        'TransactionEventLogger': '../artifacts/contracts/TransactionEventLogger.sol/TransactionEventLogger.json',
        'FundManager': '../artifacts/contracts/FundManager.sol/FundManager.json'
      };
      
      for (const [name, path] of Object.entries(abiPaths)) {
        try {
          const response = await fetch(path);
          if (response.ok) {
            const data = await response.json();
            contractABIs[name] = data.abi;
            console.log(`Loaded ABI for ${name}`);
          }
        } catch (error) {
          console.error(`Failed to load ABI for ${name}:`, error);
        }
      }
      
      // If no ABIs were loaded, use hardcoded minimal ABIs
      if (Object.keys(contractABIs).length === 0) {
        console.log('Using minimal hardcoded ABIs');
        contractABIs = {
          'TransactionEventLogger': [
            "function logTransactionEvent(uint256 transactionId, uint256 fundId, uint8 txType, uint8 status, uint256 amount, string recipient, uint256 timestamp) external",
            "function grantRole(bytes32 role, address account) external",
            "function revokeRole(bytes32 role, address account) external",
            "function hasRole(bytes32 role, address account) external view returns (bool)",
            "event TransactionEvent(uint256 indexed transactionId, uint256 indexed fundId, uint8 txType, uint8 status, uint256 amount, string recipient, uint256 timestamp)"
          ],
          'FundManager': [
            "function createFund(string name, string description, uint256 amount, address owner) external returns (uint256)",
            "function createTransaction(uint256 fundId, uint8 transactionType, uint256 amount, string description, string recipient) external returns (uint256)",
            "function approveTransaction(uint256 transactionId) external",
            "function rejectTransaction(uint256 transactionId) external",
            "function completeTransaction(uint256 transactionId) external",
            "function flagTransaction(uint256 transactionId) external",
            "function addAuthorizedUpdater(address updater) external",
            "function isAuthorizedUpdater(address account) external view returns (bool)",
            "event FundCreated(uint256 indexed fundId, string name, uint256 amount, address indexed createdBy)"
          ]
        };
      }
      
      return true;
    } catch (error) {
      console.error('Error loading contract ABIs:', error);
      return false;
    }
  }

  // Function to connect wallet
  async function connectWallet(showPrompt = true) {
    showLoading('Connecting wallet...');
    try {
      let accounts;
      if (showPrompt) {
        accounts = await window.ethereum.request({ method: 'eth_requestAccounts' });
      } else {
        accounts = await window.ethereum.request({ method: 'eth_accounts' });
        if (accounts.length === 0) throw new Error('No connected accounts');
      }
      
      signer = provider.getSigner();
      connectedAddress = accounts[0];
      
      // Update UI
      await updateUIWithWalletInfo(connectedAddress);
      
      // Start listening for events
      setupEventListeners();
      
      hideLoading();
      return true;
    } catch (error) {
      console.error('Error connecting wallet:', error);
      hideLoading();
      return false;
    }
  }

  // Function to disconnect wallet
  function disconnectWallet() {
    connectedAddress = null;
    signer = null;
    
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
  }

  // Update UI with wallet info
  async function updateUIWithWalletInfo(address) {
    try {
      // Update connection status
      connectionIndicator.textContent = 'Connected';
      connectionIndicator.classList.add('connected');
      
      // Format and display address
      addressDisplay.textContent = `${address.substring(0, 6)}...${address.substring(address.length - 4)}`;
      addressDisplay.title = address;
      
      // Get and display balance
      const balance = await provider.getBalance(address);
      balanceDisplay.textContent = `${ethers.utils.formatEther(balance).substring(0, 8)} ETH`;
      
      // Get and display network
      const network = await provider.getNetwork();
      networkDisplay.textContent = network.name === 'homestead' ? 'Mainnet' : network.name;
      
      // Enable UI elements
      connectButton.style.display = 'none';
      disconnectButton.style.display = '';
      contractSelect.disabled = false;
    } catch (error) {
      console.error('Error updating wallet info:', error);
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
    const functionName = functionSelect.value;
    
    if (!functionName || !selectedContract) {
      alert('Please select a contract and function first.');
      return;
    }
    
    // Find function definition
    const functionDef = contractABIs[contractSelect.value].find(
      item => item.type === 'function' && item.name === functionName
    );
    
    // Determine if function is read-only
    const isReadFunction = functionDef.stateMutability === 'view' || 
                           functionDef.stateMutability === 'pure';
    
    showLoading(`Executing ${isReadFunction ? 'read' : 'write'} function...`);
    
    try {
      // Collect parameters
      const params = [];
      const paramInputs = functionParams.querySelectorAll('input, select');
      
      paramInputs.forEach(input => {
        const type = input.dataset.type;
        let value = input.value.trim();
        
        // Convert value based on type
        if (type.includes('uint')) {
          if (value === '') {
            value = 0;
          } else if (type.includes('uint256')) {
            // Check if value looks like ETH amount
            if (value.includes('.')) {
              value = ethers.utils.parseEther(value);
            } else {
              value = ethers.BigNumber.from(value);
            }
          } else {
            value = parseInt(value);
          }
        } else if (type === 'bool') {
          value = value.toLowerCase() === 'true';
        } else if (type === 'address' && value === '') {
          value = ethers.constants.AddressZero;
        }
        
        params.push(value);
      });
      
      let result;
      
      if (isReadFunction) {
        // Call read function
        result = await selectedContract[functionName](...params);
        displayResult(result);
      } else {
        // Send transaction for write function
        const tx = await selectedContract[functionName](...params);
        showLoading('Transaction submitted. Waiting for confirmation...');
        
        // Wait for transaction to be mined
        const receipt = await tx.wait();
        
        result = {
          transactionHash: receipt.transactionHash,
          blockNumber: receipt.blockNumber,
          gasUsed: receipt.gasUsed.toString(),
          status: receipt.status === 1 ? 'Success' : 'Failed'
        };
        
        // Process any events emitted
        if (receipt.events && receipt.events.length > 0) {
          receipt.events.forEach(event => {
            logEvent(event);
          });
        }
        
        displayResult(result);
      }
      
      hideLoading();
    } catch (error) {
      console.error('Error executing function:', error);
      resultOutput.textContent = `Error: ${error.message}`;
      hideLoading();
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
}); 