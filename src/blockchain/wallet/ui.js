// UI.js - Handles user interface interactions for the blockchain wallet

// DOM Elements
let connectButton;
let disconnectButton;
let accountAddress;
let accountBalance;
let walletInfoSection;
let contractInteractionSection;
let contractTypeSelect;
let eventLoggerFunctions;
let fundManagerFunctions;
let eventLoggerFunctionSelect;
let fundManagerFunctionSelect;
let executeButton;
let functionResultSection;
let resultContent;
let txHashContainer;
let txHashLink;
let functionParamDivs = {};

// Initialize UI when DOM is loaded
document.addEventListener('DOMContentLoaded', async () => {
    initializeDOMElements();
    attachEventListeners();
    
    // Check if wallet is already connected
    const isConnected = await checkWalletConnection();
    if (isConnected) {
        updateUIForConnectedWallet();
    }
    
    // Initialize contract interaction
    await initContractInteraction();
    
    // Hide all function parameter divs initially
    document.querySelectorAll('.function-params').forEach(div => {
        div.style.display = 'none';
    });
    
    // Show the first function's parameters
    showFunctionParams('event-logger', 'logTransactionEvent');
});

function initializeDOMElements() {
    // Wallet elements
    connectButton = document.getElementById('connect-button');
    disconnectButton = document.getElementById('disconnect-button');
    accountAddress = document.getElementById('account-address');
    accountBalance = document.getElementById('account-balance');
    walletInfoSection = document.getElementById('wallet-info');
    
    // Contract interaction elements
    contractInteractionSection = document.getElementById('contract-interaction');
    contractTypeSelect = document.getElementById('contract-type');
    eventLoggerFunctions = document.getElementById('event-logger-functions');
    fundManagerFunctions = document.getElementById('fund-manager-functions');
    eventLoggerFunctionSelect = document.getElementById('event-logger-function');
    fundManagerFunctionSelect = document.getElementById('fund-manager-function');
    executeButton = document.getElementById('execute-function');
    functionResultSection = document.getElementById('function-result');
    resultContent = document.getElementById('result-content');
    txHashContainer = document.getElementById('tx-hash-container');
    txHashLink = document.getElementById('tx-hash-link');
    
    // Map function parameters divs
    document.querySelectorAll('.function-params').forEach(div => {
        functionParamDivs[div.getAttribute('data-function')] = div;
    });
}

function attachEventListeners() {
    // Wallet connection
    connectButton.addEventListener('click', connectWallet);
    disconnectButton.addEventListener('click', disconnectWallet);
    
    // Contract interaction
    contractTypeSelect.addEventListener('change', handleContractTypeChange);
    eventLoggerFunctionSelect.addEventListener('change', () => handleFunctionChange('event-logger'));
    fundManagerFunctionSelect.addEventListener('change', () => handleFunctionChange('fund-manager'));
    executeButton.addEventListener('click', executeSelectedFunction);
}

function handleContractTypeChange() {
    const contractType = contractTypeSelect.value;
    
    if (contractType === 'event-logger') {
        eventLoggerFunctions.style.display = 'block';
        fundManagerFunctions.style.display = 'none';
        handleFunctionChange('event-logger');
    } else {
        eventLoggerFunctions.style.display = 'none';
        fundManagerFunctions.style.display = 'block';
        handleFunctionChange('fund-manager');
    }
}

function handleFunctionChange(contractType) {
    // Hide all parameter divs
    document.querySelectorAll('.function-params').forEach(div => {
        div.style.display = 'none';
    });
    
    // Show the selected function's parameters
    let functionName;
    if (contractType === 'event-logger') {
        functionName = eventLoggerFunctionSelect.value;
    } else {
        functionName = fundManagerFunctionSelect.value;
    }
    
    showFunctionParams(contractType, functionName);
}

function showFunctionParams(contractType, functionName) {
    const paramDiv = document.querySelector(`.function-params[data-function="${functionName}"]`);
    if (paramDiv) {
        paramDiv.style.display = 'block';
    }
}

async function updateUIForConnectedWallet() {
    const currentAccount = getCurrentAccount();
    const balance = await getAccountBalance();
    
    if (currentAccount) {
        // Update wallet info
        accountAddress.textContent = `${currentAccount.substring(0, 6)}...${currentAccount.substring(currentAccount.length - 4)}`;
        accountBalance.textContent = balance;
        
        // Show wallet info and contract interaction sections
        walletInfoSection.classList.remove('hidden');
        contractInteractionSection.classList.remove('hidden');
        
        // Hide connect button
        connectButton.classList.add('hidden');
    }
}

function updateUIForDisconnectedWallet() {
    // Hide wallet info and contract interaction sections
    walletInfoSection.classList.add('hidden');
    contractInteractionSection.classList.add('hidden');
    
    // Show connect button
    connectButton.classList.remove('hidden');
    
    // Clear function results
    clearFunctionResult();
}

function showLoading() {
    document.querySelector('.loader').classList.remove('hidden');
}

function hideLoading() {
    document.querySelector('.loader').classList.add('hidden');
}

function clearFunctionResult() {
    resultContent.innerHTML = '';
    functionResultSection.classList.add('hidden');
    txHashContainer.classList.add('hidden');
}

function displayFunctionResult(result, isError = false) {
    functionResultSection.classList.remove('hidden');
    hideLoading();
    
    if (isError) {
        resultContent.innerHTML = `<div class="error">${result}</div>`;
        return;
    }
    
    if (typeof result === 'object') {
        if (result.hash) {
            // It's a transaction
            txHashContainer.classList.remove('hidden');
            txHashLink.href = `https://dashboard.ganache.org/transactions/${result.hash}`;
            txHashLink.textContent = result.hash;
            resultContent.innerHTML = '<div class="success">Transaction sent successfully!</div>';
        } else {
            // It's a complex object (e.g., fund details, events, etc.)
            displayObjectResult(result);
        }
    } else if (Array.isArray(result)) {
        // Display array results
        displayArrayResult(result);
    } else {
        // It's a simple result (boolean, number, string)
        resultContent.innerHTML = `<div class="success">${result.toString()}</div>`;
    }
}

function displayObjectResult(obj) {
    let html = '<div class="result-object">';
    for (const [key, value] of Object.entries(obj)) {
        if (key === 'hash') continue; // Skip hash as it's displayed separately
        
        html += `<div class="result-row">
            <span class="result-key">${formatKey(key)}:</span>
            <span class="result-value">${formatValue(value)}</span>
        </div>`;
    }
    html += '</div>';
    resultContent.innerHTML = html;
}

function displayArrayResult(arr) {
    let html = '<div class="result-array">';
    arr.forEach((item, index) => {
        html += `<div class="result-item">
            <div class="result-item-header">Item ${index + 1}</div>
            <div class="result-item-content">`;
        
        if (typeof item === 'object' && item !== null) {
            for (const [key, value] of Object.entries(item)) {
                html += `<div class="result-row">
                    <span class="result-key">${formatKey(key)}:</span>
                    <span class="result-value">${formatValue(value)}</span>
                </div>`;
            }
        } else {
            html += `<div class="result-value">${formatValue(item)}</div>`;
        }
        
        html += `</div></div>`;
    });
    html += '</div>';
    resultContent.innerHTML = html;
}

function formatKey(key) {
    return key.replace(/([A-Z])/g, ' $1')
        .replace(/^./, str => str.toUpperCase());
}

function formatValue(value) {
    if (typeof value === 'boolean') {
        return value ? 'Yes' : 'No';
    } else if (typeof value === 'bigint' || (typeof value === 'object' && value._isBigNumber)) {
        // Format big numbers (like wei values) to ETH
        return formatEtherValue(value);
    } else if (typeof value === 'object' && value !== null) {
        return JSON.stringify(value);
    }
    return value.toString();
}

async function executeSelectedFunction() {
    clearFunctionResult();
    showLoading();
    
    try {
        const contractType = contractTypeSelect.value;
        let functionName;
        
        if (contractType === 'event-logger') {
            functionName = eventLoggerFunctionSelect.value;
            await executeEventLoggerFunction(functionName);
        } else {
            functionName = fundManagerFunctionSelect.value;
            await executeFundManagerFunction(functionName);
        }
    } catch (error) {
        console.error('Error executing function:', error);
        displayFunctionResult(error.message || 'An error occurred while executing the function', true);
    }
}

async function executeEventLoggerFunction(functionName) {
    let result;
    
    switch (functionName) {
        case 'logTransactionEvent':
            const txId = parseInt(document.getElementById('log-transaction-id').value);
            const eventType = parseInt(document.getElementById('log-event-type').value);
            const actorHash = document.getElementById('log-actor-hash').value;
            const detailsHash = document.getElementById('log-details-hash').value;
            const reason = document.getElementById('log-reason').value;
            
            result = await executeEventLoggerFunction('logTransactionEvent', [
                txId, eventType, actorHash, detailsHash, reason
            ]);
            break;
            
        case 'getTransactionEvents':
            const transactionId = parseInt(document.getElementById('get-events-transaction-id').value);
            result = await executeEventLoggerFunction('getTransactionEvents', [transactionId]);
            break;
            
        case 'hasRole':
            const role = document.getElementById('has-role-role').value;
            const account = document.getElementById('has-role-account').value;
            result = await executeEventLoggerFunction('hasRole', [role, account]);
            break;
            
        case 'grantRole':
            const roleToGrant = document.getElementById('grant-role-role').value;
            const accountToGrant = document.getElementById('grant-role-account').value;
            result = await executeEventLoggerFunction('grantRole', [roleToGrant, accountToGrant]);
            break;
            
        case 'API_ROLE':
            result = await executeEventLoggerFunction('API_ROLE', []);
            break;
    }
    
    displayFunctionResult(result);
}

async function executeFundManagerFunction(functionName) {
    let result;
    
    switch (functionName) {
        case 'createFund':
            const fundName = document.getElementById('create-fund-name').value;
            const fundDescription = document.getElementById('create-fund-description').value;
            const fundAmount = document.getElementById('create-fund-amount').value;
            const fundMetadata = document.getElementById('create-fund-metadata').value || "";
            
            result = await executeFundManagerFunction('createFund', [
                fundName, fundDescription, parseEther(fundAmount), fundMetadata
            ]);
            break;
            
        case 'createTransaction':
            const fundId = parseInt(document.getElementById('create-tx-fund-id').value);
            const txType = parseInt(document.getElementById('create-tx-type').value);
            const txAmount = document.getElementById('create-tx-amount').value;
            const txDescription = document.getElementById('create-tx-description').value;
            const recipientId = document.getElementById('create-tx-recipient').value;
            const txMetadata = document.getElementById('create-tx-metadata').value || "";
            
            result = await executeFundManagerFunction('createTransaction', [
                fundId, txType, parseEther(txAmount), txDescription, recipientId, txMetadata
            ]);
            break;
            
        case 'approveTransaction':
            const approveId = parseInt(document.getElementById('approve-tx-id').value);
            result = await executeFundManagerFunction('approveTransaction', [approveId]);
            break;
            
        case 'rejectTransaction':
            const rejectId = parseInt(document.getElementById('reject-tx-id').value);
            const rejectReason = document.getElementById('reject-tx-reason').value;
            result = await executeFundManagerFunction('rejectTransaction', [rejectId, rejectReason]);
            break;
            
        case 'completeTransaction':
            const completeId = parseInt(document.getElementById('complete-tx-id').value);
            result = await executeFundManagerFunction('completeTransaction', [completeId]);
            break;
            
        case 'flagTransaction':
            const flagId = parseInt(document.getElementById('flag-tx-id').value);
            const flagReason = document.getElementById('flag-tx-reason').value;
            result = await executeFundManagerFunction('flagTransaction', [flagId, flagReason]);
            break;
            
        case 'getFundDetails':
            const fundDetailsId = parseInt(document.getElementById('fund-details-id').value);
            result = await executeFundManagerFunction('getFundDetails', [fundDetailsId]);
            break;
            
        case 'getTransactionDetails':
            const txDetailsId = parseInt(document.getElementById('tx-details-id').value);
            result = await executeFundManagerFunction('getTransactionDetails', [txDetailsId]);
            break;
            
        case 'addAuthorizedUpdater':
            const updaterAddress = document.getElementById('add-updater-address').value;
            result = await executeFundManagerFunction('addAuthorizedUpdater', [updaterAddress]);
            break;
            
        case 'isAuthorizedUpdater':
            const checkAddress = document.getElementById('check-updater-address').value;
            result = await executeFundManagerFunction('isAuthorizedUpdater', [checkAddress]);
            break;
    }
    
    displayFunctionResult(result);
} 