#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== Digital Watchdog Initiative - Local Deployment Script ===${NC}"
echo -e "This script will deploy the smart contracts to your local Ganache instance"
echo ""

# Check if Ganache is installed
if ! command -v npx &> /dev/null; then
    echo -e "${RED}Error: npx is not installed!${NC}"
    echo "Please install Node.js and npm first."
    exit 1
fi

# Stop any running Ganache instance
echo -e "${YELLOW}Stopping any running Ganache instances...${NC}"
pkill -f ganache-cli || true
sleep 2

# Create a temporary file to capture Ganache output
GANACHE_OUTPUT=$(mktemp)

# Start Ganache with proper settings
echo -e "${YELLOW}Starting Ganache with proper settings...${NC}"
npx ganache-cli --port 8545 --gasLimit 30000000 --accounts 10 --defaultBalanceEther 1000 > ${GANACHE_OUTPUT} &
GANACHE_PID=$!

# Wait for Ganache to start
echo -e "${YELLOW}Waiting for Ganache to start...${NC}"
sleep 5

# Display Ganache startup information
cat ${GANACHE_OUTPUT}

# Extract the first private key from Ganache output
PRIVATE_KEY=$(grep -A 11 "Private Keys" ${GANACHE_OUTPUT} | grep "(0)" | cut -d' ' -f2)
if [ -z "$PRIVATE_KEY" ]; then
    echo -e "${RED}Error: Could not extract private key from Ganache output${NC}"
    exit 1
fi

# Update the .env file with the first private key
echo -e "${YELLOW}Updating .env file with the first Ganache account private key...${NC}"
sed -i "s/^BLOCKCHAIN_PRIVATE_KEY=.*/BLOCKCHAIN_PRIVATE_KEY=${PRIVATE_KEY}/" .env
echo -e "${GREEN}Updated .env with private key: ${PRIVATE_KEY}${NC}"

# Verify Ganache is running
nc -z localhost 8545 > /dev/null 2>&1
if [ $? -ne 0 ]; then
    echo -e "${RED}Error: Failed to start Ganache on port 8545!${NC}"
    exit 1
fi

echo -e "${GREEN}Ganache started successfully on port 8545${NC}"

# Clean Hardhat cache and artifacts
echo -e "${YELLOW}Cleaning Hardhat cache and artifacts...${NC}"
npx hardhat clean

# Compile contracts
echo -e "${YELLOW}Compiling contracts...${NC}"
npx hardhat compile
if [ $? -ne 0 ]; then
    echo -e "${RED}Compilation failed! Please fix the errors before deploying.${NC}"
    pkill -P $GANACHE_PID
    exit 1
fi
echo -e "${GREEN}Compilation successful!${NC}"

# Update hardhat.config.js to use the extracted private key
echo -e "${YELLOW}Deploying contracts to local Ganache...${NC}"
npx hardhat run scripts/deploy.js --network ganache

if [ $? -ne 0 ]; then
    echo -e "${RED}Deployment failed!${NC}"
    echo -e "${YELLOW}Keeping Ganache running for debugging.${NC}"
    echo -e "Ganache is running on http://localhost:8545"
    echo -e "You can stop it with: pkill -f ganache-cli"
    exit 1
fi

# Check if deployed-contracts.json was created
if [ ! -f deployed-contracts.json ]; then
    echo -e "${RED}Warning: deployed-contracts.json not found!${NC}"
    echo "Deployment may have partially completed."
    exit 1
fi

# Copy the deployed contracts file to the wallet directory
echo -e "${YELLOW}Copying deployed contracts to wallet directory...${NC}"
cp deployed-contracts.json wallet/
echo -e "${GREEN}Contract addresses copied to wallet/deployed-contracts.json${NC}"

# Display next steps
echo ""
echo -e "${GREEN}Deployment successful! Ganache is running on http://localhost:8545${NC}"
echo ""
echo -e "${YELLOW}=== NEXT STEPS ===${NC}"
echo "1. Start the wallet interface:"
echo "   cd wallet && python -m http.server 8080"
echo "2. Connect MetaMask to your local Ganache:"
echo "   - Network Name: Ganache"
echo "   - RPC URL: http://localhost:8545"
echo "   - Chain ID: 1337"
echo "   - Currency Symbol: ETH"
echo ""
echo -e "${YELLOW}When you're done, you can stop Ganache with: pkill -f ganache-cli${NC}"

# Clean up
rm -f ${GANACHE_OUTPUT}

exit 0 