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

# Extract artifacts from compiled contracts if they exist
cd ..
if [ -d "artifacts" ]; then
    echo -e "${YELLOW}Extracting contract artifacts...${NC}"
    
    # Run the extract-artifacts script if it exists
    if [ -f "scripts/extract-artifacts.js" ]; then
        node scripts/extract-artifacts.js
    else
        echo -e "${YELLOW}Artifact extraction script not found, checking for compiled contracts...${NC}"
        
        # Manually copy artifacts if the script doesn't exist
        if [ -d "artifacts/contracts/TransactionEventLogger.sol" ] && [ -d "artifacts/contracts/FundManager.sol" ]; then
            cp artifacts/contracts/TransactionEventLogger.sol/TransactionEventLogger.json wallet/
            cp artifacts/contracts/FundManager.sol/FundManager.json wallet/
            echo -e "${GREEN}Copied contract artifacts to wallet directory${NC}"
        else
            echo -e "${YELLOW}Warning: Contract artifacts not found. Please compile contracts first.${NC}"
        fi
    fi
    
    # Copy deployed contracts file if it exists
    if [ -f "deployed-contracts.json" ]; then
        cp deployed-contracts.json wallet/
        echo -e "${GREEN}Copied deployed contracts info to wallet directory${NC}"
    else
        echo -e "${YELLOW}Warning: No deployed-contracts.json found.${NC}"
        echo -e "${YELLOW}You may need to deploy contracts first or use sample data.${NC}"
    fi
fi

# Return to wallet directory
cd wallet

# Find an available port starting with 8080
PORT=8080
while nc -z localhost $PORT &>/dev/null; do
    echo -e "${YELLOW}Port $PORT is already in use, trying next port...${NC}"
    PORT=$((PORT + 1))
done

# Start HTTP server
echo -e "${GREEN}Starting wallet interface on port $PORT...${NC}"
echo -e "${YELLOW}Use Ctrl+C to stop the server${NC}"

# Try using http-server if available
if command -v npx &> /dev/null; then
    echo -e "${GREEN}Starting with npx http-server...${NC}"
    echo -e "${GREEN}Open your browser to http://localhost:$PORT${NC}"
    npx http-server -p $PORT -o
else
    # Fallback to Python's built-in HTTP server
    if command -v python3 &> /dev/null; then
        echo -e "${GREEN}Starting with Python 3 HTTP server...${NC}"
        echo -e "${GREEN}Open your browser to http://localhost:$PORT${NC}"
        python3 -m http.server $PORT
    elif command -v python &> /dev/null; then
        echo -e "${GREEN}Starting with Python HTTP server...${NC}"
        echo -e "${GREEN}Open your browser to http://localhost:$PORT${NC}"
        python -m http.server $PORT
    else
        echo -e "${RED}Error: No HTTP server available.${NC}"
        echo "Please install Node.js/npx or Python to run the HTTP server."
        exit 1
    fi
fi