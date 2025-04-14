#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== Digital Watchdog Initiative - Setup Script ===${NC}"
echo -e "This script will install all dependencies and prepare your environment"
echo ""

# Check for npm
if ! command -v npm &> /dev/null; then
    echo -e "${RED}Error: npm is not installed!${NC}"
    echo "Please install Node.js and npm first."
    exit 1
fi

# Clean installation
echo -e "${YELLOW}Cleaning previous installation...${NC}"
rm -rf node_modules
rm -rf .cache
rm -f package-lock.json

# Install dependencies with exact versions
echo -e "${YELLOW}Installing dependencies...${NC}"
npm install --save-exact hardhat@2.16.1
npm install --save-exact @nomiclabs/hardhat-waffle@2.0.5 @nomiclabs/hardhat-ethers@2.2.2 @nomiclabs/hardhat-etherscan@3.1.7
npm install --save-exact @openzeppelin/contracts@4.9.3 dotenv@16.1.4
npm install --save-exact ethers@5.7.2 chai@4.3.7

if [ $? -ne 0 ]; then
    echo -e "${RED}Error installing dependencies!${NC}"
    exit 1
fi

echo -e "${GREEN}Dependencies installed successfully!${NC}"

# Initialize hardhat if not already initialized
if [ ! -f hardhat.config.js ]; then
    echo -e "${YELLOW}Initializing Hardhat project...${NC}"
    npx hardhat
fi

# Clean Hardhat cache
echo -e "${YELLOW}Cleaning Hardhat cache...${NC}"
npx hardhat clean

# Compile contracts
echo -e "${YELLOW}Compiling contracts...${NC}"
npx hardhat compile

if [ $? -ne 0 ]; then
    echo -e "${RED}Compilation failed!${NC}"
    exit 1
fi

echo -e "${GREEN}Compilation successful!${NC}"

# Create wallet directory if it doesn't exist
if [ ! -d wallet ]; then
    echo -e "${YELLOW}Creating wallet directory...${NC}"
    mkdir -p wallet
fi

# Copy sample contracts file to wallet directory
echo -e "${YELLOW}Setting up wallet interface...${NC}"
if [ -f wallet/sample-deployed-contracts.json ]; then
    cp wallet/sample-deployed-contracts.json wallet/deployed-contracts.json
    echo -e "${GREEN}Sample contract addresses copied to wallet directory${NC}"
fi

echo ""
echo -e "${GREEN}Setup completed successfully!${NC}"
echo ""
echo -e "${YELLOW}=== NEXT STEPS ===${NC}"
echo "1. Start Ganache (local blockchain):"
echo "   npx ganache-cli"
echo ""
echo "2. Deploy contracts to Ganache:"
echo "   ./deploy-to-ganache.sh"
echo ""
echo "3. Or deploy to Sepolia testnet:"
echo "   ./deploy-to-sepolia.sh"
echo ""
echo "4. Start the wallet interface:"
echo "   cd wallet && python -m http.server 8080"

exit 0 