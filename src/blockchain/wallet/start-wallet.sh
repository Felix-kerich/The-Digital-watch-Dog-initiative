#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== Digital Watchdog Initiative - Wallet Interface ===${NC}"

# Check if node is installed
if ! command -v node &> /dev/null; then
    echo -e "${RED}Error: Node.js is not installed!${NC}"
    echo "Please install Node.js and npm first."
    exit 1
fi

# Check if the current directory contains the wallet files
if [ ! -f "index.html" ] || [ ! -f "app.js" ] || [ ! -f "styles.css" ]; then
    echo -e "${RED}Error: Wallet files not found in the current directory!${NC}"
    echo "Please run this script from the wallet directory."
    exit 1
fi

# Check if ethers.js local file exists, download if needed
if [ ! -f "ethers-5.7.2.umd.min.js" ]; then
    echo -e "${YELLOW}Local ethers.js not found. Downloading...${NC}"
    if command -v curl &> /dev/null; then
        curl -s -o ethers-5.7.2.umd.min.js https://cdn.jsdelivr.net/npm/ethers@5.7.2/dist/ethers.umd.min.js
    elif command -v wget &> /dev/null; then
        wget -q -O ethers-5.7.2.umd.min.js https://cdn.jsdelivr.net/npm/ethers@5.7.2/dist/ethers.umd.min.js
    else
        echo -e "${RED}Warning: Could not download ethers.js - neither curl nor wget is available${NC}"
    fi
    
    if [ -f "ethers-5.7.2.umd.min.js" ]; then
        echo -e "${GREEN}ethers.js downloaded successfully${NC}"
    fi
fi

# Check if required files exist
if [ ! -f "deployed-contracts.json" ]; then
    echo -e "${YELLOW}Checking for deployed contracts in parent directory...${NC}"
    if [ -f "../deployed-contracts.json" ]; then
        cp ../deployed-contracts.json .
        echo -e "${GREEN}Copied deployed-contracts.json from parent directory${NC}"
    else
        echo -e "${RED}Warning: deployed-contracts.json not found!${NC}"
        echo "Make sure contracts are deployed first."
    fi
fi

# Copy contract artifacts if they exist
if [ ! -f "TransactionEventLogger.json" ] || [ ! -f "FundManager.json" ]; then
    echo -e "${YELLOW}Checking for contract artifacts...${NC}"
    if [ -d "../artifacts/contracts" ]; then
        cp ../artifacts/contracts/TransactionEventLogger.sol/TransactionEventLogger.json .
        cp ../artifacts/contracts/FundManager.sol/FundManager.json .
        echo -e "${GREEN}Copied contract artifacts${NC}"
    else
        echo -e "${RED}Warning: Contract artifacts not found!${NC}"
        echo "Please compile contracts first."
    fi
fi

# Find an available port starting with 8080
PORT=8080
while nc -z localhost $PORT &>/dev/null; do
    echo -e "${YELLOW}Port $PORT is already in use, trying next port...${NC}"
    PORT=$((PORT + 1))
done

# Start HTTP server
echo -e "${GREEN}Starting wallet interface on port $PORT...${NC}"
echo -e "${YELLOW}Use Ctrl+C to stop the server${NC}"
echo -e "Access the wallet interface at ${GREEN}http://localhost:$PORT${NC}"

# Try using http-server if available
if command -v npx &> /dev/null; then
    echo -e "${GREEN}Starting with npx http-server...${NC}"
    npx http-server -p $PORT --cors
elif command -v python3 &> /dev/null; then
    echo -e "${GREEN}Starting with Python 3 HTTP server...${NC}"
    python3 -m http.server $PORT
elif command -v python &> /dev/null; then
    echo -e "${GREEN}Starting with Python HTTP server...${NC}"
    python -m SimpleHTTPServer $PORT
else
    echo -e "${RED}Error: No HTTP server available.${NC}"
    echo "Please install Node.js/npx or Python to run the HTTP server."
    exit 1
fi