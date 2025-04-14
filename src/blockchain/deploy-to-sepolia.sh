#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== Digital Watchdog Initiative - Deployment Script ===${NC}"
echo -e "This script will deploy the smart contracts to Sepolia testnet"
echo ""

# Check if .env file exists
if [ ! -f .env ]; then
    echo -e "${RED}Error: .env file not found!${NC}"
    echo "Please create a .env file with your deployment settings"
    exit 1
fi

# Clean Hardhat cache and artifacts
echo -e "${YELLOW}Cleaning Hardhat cache and artifacts...${NC}"
npx hardhat clean

# Compile contracts
echo -e "${YELLOW}Compiling contracts...${NC}"
npx hardhat compile
if [ $? -ne 0 ]; then
    echo -e "${RED}Compilation failed! Please fix the errors before deploying.${NC}"
    exit 1
fi
echo -e "${GREEN}Compilation successful!${NC}"

# Check if we have an Infura API key
source .env
if [ -z "$INFURA_API_KEY" ]; then
    echo -e "${RED}Error: INFURA_API_KEY not found in .env file!${NC}"
    exit 1
fi

if [ -z "$PRIVATE_KEY" ]; then
    echo -e "${RED}Error: PRIVATE_KEY not found in .env file!${NC}"
    exit 1
fi

# Set max retries
MAX_RETRIES=3
retry_count=0

echo -e "${YELLOW}Deploying to Sepolia testnet...${NC}"
echo -e "This may take several minutes. Please be patient."

while [ $retry_count -lt $MAX_RETRIES ]; do
    echo -e "${YELLOW}Deployment attempt $(($retry_count + 1))/${MAX_RETRIES}${NC}"
    
    # Run the deployment with a 5-minute timeout
    timeout 300 npx hardhat run scripts/deploy.js --network sepolia
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Deployment successful!${NC}"
        
        # Check if deployed-contracts.json was created
        if [ -f deployed-contracts.json ]; then
            echo -e "${GREEN}Contract addresses saved to deployed-contracts.json${NC}"
            
            # Copy to wallet directory for frontend
            if [ -d wallet ]; then
                cp deployed-contracts.json wallet/
                echo -e "${GREEN}Contract addresses copied to wallet/deployed-contracts.json${NC}"
            fi
            
            # Display next steps
            echo ""
            echo -e "${YELLOW}=== NEXT STEPS ===${NC}"
            echo "1. Verify contracts on Etherscan:"
            echo "   npx hardhat verify --network sepolia CONTRACT_ADDRESS"
            echo "2. Start the wallet interface to interact with your contracts"
            
            exit 0
        else
            echo -e "${RED}Warning: deployed-contracts.json not found!${NC}"
            echo "Deployment may have partially completed."
        fi
        
        break
    elif [ $? -eq 124 ]; then
        echo -e "${RED}Deployment timed out after 5 minutes!${NC}"
    else
        echo -e "${RED}Deployment failed!${NC}"
    fi
    
    # Increment retry counter
    retry_count=$(($retry_count + 1))
    
    if [ $retry_count -lt $MAX_RETRIES ]; then
        echo -e "${YELLOW}Waiting 30 seconds before retrying...${NC}"
        sleep 30
    else
        echo -e "${RED}Maximum retry attempts reached. Deployment failed.${NC}"
        echo ""
        echo -e "${YELLOW}Troubleshooting:${NC}"
        echo "1. Check your internet connection"
        echo "2. Verify your Infura API key is correct"
        echo "3. Make sure your wallet has enough Sepolia ETH"
        echo "4. Try increasing gas price and limit in hardhat.config.js"
        exit 1
    fi
done 