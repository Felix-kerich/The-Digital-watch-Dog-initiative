#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== Digital Watchdog Initiative - Wallet Demo ===${NC}"
echo "This script will deploy contracts to Ganache and start the wallet interface."

# Check dependencies
if ! command -v node &> /dev/null; then
    echo -e "${RED}Error: Node.js is not installed!${NC}"
    echo "Please install Node.js and npm first."
    exit 1
fi

if ! command -v npm &> /dev/null; then
    echo -e "${RED}Error: npm is not installed!${NC}"
    echo "Please install npm first."
    exit 1
fi

if ! command -v npx &> /dev/null; then
    echo -e "${RED}Error: npx is not installed!${NC}"
    echo "Please install npx first."
    exit 1
fi

# Step 1: Start Ganache if it's not already running
echo -e "\n${YELLOW}Step 1: Starting Ganache...${NC}"

# Check if Ganache is already running
if nc -z localhost 8545 2>/dev/null; then
    echo -e "${GREEN}Ganache is already running on port 8545.${NC}"
else
    echo "Starting Ganache with 1000 ETH per account..."
    # Start Ganache in the background
    npx ganache-cli --port 8545 --gasLimit 30000000 --accounts 10 --defaultBalanceEther 1000 > ganache.log 2>&1 &
    GANACHE_PID=$!
    
    # Wait for Ganache to start
    echo "Waiting for Ganache to start..."
    sleep 3
    
    # Check if Ganache started successfully
    if nc -z localhost 8545 2>/dev/null; then
        echo -e "${GREEN}Ganache started successfully on port 8545.${NC}"
        
        # Extract the first private key from Ganache output
        PRIVATE_KEY=$(grep -A 11 "Private Keys" ganache.log | grep "(0)" | cut -d' ' -f2)
        if [ -n "$PRIVATE_KEY" ]; then
            # Update .env file if it exists
            if [ -f ".env" ]; then
                sed -i "s/^BLOCKCHAIN_PRIVATE_KEY=.*/BLOCKCHAIN_PRIVATE_KEY=0x${PRIVATE_KEY}/" .env
                echo -e "${GREEN}Updated .env with private key from Ganache.${NC}"
            else
                echo "BLOCKCHAIN_PRIVATE_KEY=0x${PRIVATE_KEY}" > .env
                echo -e "${GREEN}Created .env with private key from Ganache.${NC}"
            fi
        else
            echo -e "${YELLOW}Warning: Could not extract private key from Ganache output.${NC}"
        fi
    else
        echo -e "${RED}Error: Failed to start Ganache on port 8545!${NC}"
        exit 1
    fi
fi

# Step 2: Compile smart contracts
echo -e "\n${YELLOW}Step 2: Compiling smart contracts...${NC}"
echo "Running: npx hardhat compile"
npx hardhat compile

if [ $? -ne 0 ]; then
    echo -e "${RED}Compilation failed! Please fix the errors before proceeding.${NC}"
    # Kill Ganache if we started it
    if [ -n "$GANACHE_PID" ]; then
        kill $GANACHE_PID
    fi
    exit 1
fi

echo -e "${GREEN}Compilation successful!${NC}"

# Step 3: Deploy contracts to Ganache
echo -e "\n${YELLOW}Step 3: Deploying contracts to Ganache...${NC}"
echo "Running: npx hardhat run scripts/deploy.js --network ganache"
npx hardhat run scripts/deploy.js --network ganache

if [ $? -ne 0 ]; then
    echo -e "${RED}Deployment failed!${NC}"
    echo -e "${YELLOW}Please check if Ganache is running and your account has sufficient ETH.${NC}"
    # Keep Ganache running for debugging
    exit 1
fi

echo -e "${GREEN}Deployment successful!${NC}"

# Step 4: Extract contract artifacts for the wallet
echo -e "\n${YELLOW}Step 4: Extracting contract artifacts for the wallet...${NC}"
node scripts/extract-artifacts.js

if [ $? -ne 0 ]; then
    echo -e "${RED}Failed to extract artifacts!${NC}"
    echo -e "${YELLOW}Trying fallback approach...${NC}"
    
    # Manually copy artifacts if the script fails
    if [ -d "artifacts/contracts/TransactionEventLogger.sol" ] && [ -d "artifacts/contracts/FundManager.sol" ]; then
        cp artifacts/contracts/TransactionEventLogger.sol/TransactionEventLogger.json wallet/
        cp artifacts/contracts/FundManager.sol/FundManager.json wallet/
        cp deployed-contracts.json wallet/
        echo -e "${GREEN}Copied contract artifacts to wallet directory${NC}"
    else
        echo -e "${RED}Could not find contract artifacts. Please compile contracts first.${NC}"
        exit 1
    fi
else
    echo -e "${GREEN}Successfully extracted contract artifacts.${NC}"
fi

# Step 5: Start the wallet interface
echo -e "\n${YELLOW}Step 5: Starting the wallet interface...${NC}"
echo -e "${GREEN}=== INSTRUCTIONS ===${NC}"
echo -e "1. Connect MetaMask to your local Ganache:"
echo -e "   - Network Name: Ganache Local"
echo -e "   - RPC URL: http://localhost:8545"
echo -e "   - Chain ID: 1337"
echo -e "   - Currency Symbol: ETH"
echo -e "2. Import an account using the private key from the Ganache output"
echo -e "3. The first account will have the necessary permissions to interact with contracts"
echo -e ""
echo -e "${YELLOW}Starting wallet interface...${NC}"
echo -e "Use Ctrl+C to stop both the wallet interface and Ganache when done."
echo -e ""

cd wallet
./start-wallet.sh

# Clean up when the wallet interface is closed
echo -e "\n${YELLOW}Shutting down...${NC}"

# Kill Ganache if we started it
if [ -n "$GANACHE_PID" ]; then
    echo "Stopping Ganache..."
    kill $GANACHE_PID
    echo -e "${GREEN}Ganache stopped.${NC}"
fi

echo -e "${GREEN}Demo completed. Thank you for using the Digital Watchdog Initiative!${NC}"
exit 0