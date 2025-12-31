#!/bin/bash
# Quick setup script for Ubuntu 22.04 LTS
# Run this to set up Malice development environment

set -e

echo "🔧 Malice Development Environment Setup for Ubuntu 22.04"
echo "========================================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if running on Ubuntu
if [[ ! -f /etc/lsb-release ]]; then
    echo -e "${RED}✗ This script is designed for Ubuntu. Please ensure you're on Ubuntu 22.04${NC}"
    exit 1
fi

UBUNTU_VERSION=$(grep DISTRIB_RELEASE /etc/lsb-release | cut -d= -f2)
if [[ "$UBUNTU_VERSION" != "22.04" ]]; then
    echo -e "${YELLOW}⚠ Warning: This was tested on Ubuntu 22.04. You have $UBUNTU_VERSION${NC}"
fi

# 1. Update system
echo ""
echo -e "${YELLOW}→ Updating system packages...${NC}"
sudo apt-get update -qq
sudo apt-get upgrade -y -qq

# 2. Install Go 1.21+
echo ""
echo -e "${YELLOW}→ Checking Go installation...${NC}"
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}')
    echo -e "${GREEN}✓ Go already installed: $GO_VERSION${NC}"
else
    echo -e "${YELLOW}→ Installing Go 1.21...${NC}"
    cd /tmp
    wget -q https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
    
    # Add to PATH if not already there
    if ! grep -q "export PATH=\$PATH:/usr/local/go/bin" ~/.bashrc; then
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    fi
    
    export PATH=$PATH:/usr/local/go/bin
    echo -e "${GREEN}✓ Go 1.21 installed$(NC}"
fi

# 3. Install Docker
echo ""
echo -e "${YELLOW}→ Checking Docker installation...${NC}"
if command -v docker &> /dev/null; then
    DOCKER_VERSION=$(docker --version | awk '{print $3}' | tr -d ',')
    echo -e "${GREEN}✓ Docker already installed: $DOCKER_VERSION${NC}"
else
    echo -e "${YELLOW}→ Installing Docker...${NC}"
    curl -fsSL https://get.docker.com -o /tmp/get-docker.sh
    sudo sh /tmp/get-docker.sh
    sudo usermod -aG docker $USER
    echo -e "${GREEN}✓ Docker installed${NC}"
    echo -e "${YELLOW}  Note: Log out and back in or run 'newgrp docker' for changes to take effect${NC}"
fi

# 4. Install Docker Compose
echo ""
echo -e "${YELLOW}→ Checking Docker Compose installation...${NC}"
if command -v docker-compose &> /dev/null; then
    DC_VERSION=$(docker-compose --version | awk '{print $3}')
    echo -e "${GREEN}✓ Docker Compose already installed: $DC_VERSION${NC}"
else
    echo -e "${YELLOW}→ Installing Docker Compose...${NC}"
    COMPOSE_URL=$(curl -s https://api.github.com/repos/docker/compose/releases/latest | grep browser_download_url | grep "$(uname -s)-$(uname -m)" | cut -d'"' -f4)
    sudo curl -L "$COMPOSE_URL" -o /usr/local/bin/docker-compose
    sudo chmod +x /usr/local/bin/docker-compose
    echo -e "${GREEN}✓ Docker Compose installed${NC}"
fi

# 5. Install Git
echo ""
echo -e "${YELLOW}→ Checking Git installation...${NC}"
if command -v git &> /dev/null; then
    GIT_VERSION=$(git --version | awk '{print $3}')
    echo -e "${GREEN}✓ Git already installed: $GIT_VERSION${NC}"
else
    echo -e "${YELLOW}→ Installing Git...${NC}"
    sudo apt-get install -y -qq git
    echo -e "${GREEN}✓ Git installed${NC}"
fi

# 6. Install build tools
echo ""
echo -e "${YELLOW}→ Checking build tools...${NC}"
sudo apt-get install -y -qq build-essential

# 7. Verify installations
echo ""
echo -e "${YELLOW}→ Verifying installations...${NC}"
echo -e "${GREEN}✓ Go version:${NC}"
go version
echo -e "${GREEN}✓ Docker version:${NC}"
docker --version
echo -e "${GREEN}✓ Docker Compose version:${NC}"
docker-compose --version
echo -e "${GREEN}✓ Git version:${NC}"
git --version

# 8. Clone/Update repository
echo ""
echo -e "${YELLOW}→ Setting up Malice repository...${NC}"
if [ ! -d "malice" ]; then
    git clone https://github.com/maliceio/malice.git
    cd malice
else
    cd malice
    git pull origin master
fi

# 9. Setup Go project
echo ""
echo -e "${YELLOW}→ Setting up Go modules and dependencies...${NC}"
make setup

echo ""
echo -e "${GREEN}✓ Setup complete!${NC}"
echo ""
echo "Next steps:"
echo "==========="
echo "1. Build Malice:"
echo "   make build"
echo ""
echo "2. Run tests:"
echo "   make test"
echo ""
echo "3. Start Malice:"
echo "   ./build/malice -D"
echo ""
echo "4. Or use Docker Compose:"
echo "   docker-compose up -d"
echo "   # Access Kibana at http://localhost:5601"
echo ""
echo "For more information, see MODERNIZATION.md"
