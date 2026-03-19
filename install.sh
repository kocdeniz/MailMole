# MailMole Installation Script
#!/bin/bash

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}   MailMole Installation Script${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Detect OS
OS=""
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    OS="linux"
elif [[ "$OSTYPE" == "darwin"* ]]; then
    OS="darwin"
elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]]; then
    OS="windows"
else
    echo -e "${RED}❌ Unsupported operating system: $OSTYPE${NC}"
    exit 1
fi

echo -e "${BLUE}📋 Detected OS: $OS${NC}"
echo ""

# Check if Go is installed
echo -e "${YELLOW}🔍 Checking Go installation...${NC}"
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}')
    echo -e "${GREEN}✅ Go is installed: $GO_VERSION${NC}"
else
    echo -e "${RED}❌ Go is not installed${NC}"
    echo ""
    echo -e "${YELLOW}📥 Installing Go...${NC}"
    
    if [[ "$OS" == "linux" ]]; then
        # For Linux
        if command -v apt-get &> /dev/null; then
            # Debian/Ubuntu
            sudo apt-get update
            sudo apt-get install -y golang-go
        elif command -v yum &> /dev/null; then
            # RHEL/CentOS/Fedora
            sudo yum install -y golang
        elif command -v pacman &> /dev/null; then
            # Arch Linux
            sudo pacman -S go
        else
            echo -e "${RED}❌ Could not install Go automatically. Please install manually:${NC}"
            echo "   Visit: https://golang.org/dl/"
            exit 1
        fi
    elif [[ "$OS" == "darwin" ]]; then
        # For macOS
        if command -v brew &> /dev/null; then
            brew install go
        else
            echo -e "${YELLOW}⚠️  Homebrew not found. Installing...${NC}"
            /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
            brew install go
        fi
    fi
    
    echo -e "${GREEN}✅ Go installed successfully${NC}"
fi

# Verify Go version
echo ""
echo -e "${YELLOW}🔍 Checking Go version...${NC}"
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
REQUIRED_VERSION="1.20"

if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" = "$REQUIRED_VERSION" ]; then 
    echo -e "${GREEN}✅ Go version $GO_VERSION meets requirements (>= $REQUIRED_VERSION)${NC}"
else
    echo -e "${YELLOW}⚠️  Go version $GO_VERSION is outdated. Minimum required: $REQUIRED_VERSION${NC}"
    echo "   Please update Go: https://golang.org/dl/"
fi

echo ""

# Create directories
echo -e "${YELLOW}📁 Creating directories...${NC}"
mkdir -p logs
mkdir -p exports
mkdir -p backups
echo -e "${GREEN}✅ Directories created${NC}"

echo ""

# Build MailMole
echo -e "${YELLOW}🔨 Building MailMole...${NC}"
if CGO_ENABLED=0 go build -o mailmole .; then
    echo -e "${GREEN}✅ MailMole built successfully${NC}"
else
    echo -e "${RED}❌ Build failed${NC}"
    exit 1
fi

echo ""

# Make executable
echo -e "${YELLOW}🔧 Setting permissions...${NC}"
chmod +x mailmole
chmod +x install.sh 2>/dev/null || true
echo -e "${GREEN}✅ Permissions set${NC}"

echo ""

# Create example files
echo -e "${YELLOW}📝 Creating example files...${NC}"

cat > accounts-example.csv << 'EOF'
# MailMole Bulk Migration Example
# Format: src_host,src_port,src_user,src_pass,dst_host,dst_port,dst_user,dst_pass
#
# This is an example file. Replace with your actual accounts.

# Example 1: Gmail to Office 365
imap.gmail.com,993,user1@gmail.com,yourpassword1,outlook.office365.com,993,user1@company.com,yourpassword2

# Example 2: Yandex to Gmail
imap.yandex.com,993,user2@yandex.com,yourpassword3,imap.gmail.com,993,user2@newgmail.com,yourpassword4

# Add more accounts below...
EOF

cat > config.json << 'EOF'
{
  "default_source": {
    "host": "imap.gmail.com",
    "port": 993,
    "tls": true
  },
  "default_destination": {
    "host": "outlook.office365.com",
    "port": 993,
    "tls": true
  },
  "concurrent_workers": 3,
  "batch_size": 50,
  "retry_attempts": 3
}
EOF

echo -e "${GREEN}✅ Example files created:${NC}"
echo "   - accounts-example.csv"
echo "   - config.json"

echo ""

# Final message
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}   Installation Complete! ✅${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${BLUE}📖 Quick Start:${NC}"
echo ""
echo "   Terminal UI:        ./mailmole"
echo "   Web Dashboard:      ./mailmole -web :8080"
echo "   With config:        ./mailmole -config config.json"
echo ""
echo -e "${BLUE}📚 Documentation:${NC}"
echo "   README.md           # Main documentation"
echo "   Docs/               # Detailed guides"
echo "   accounts-example.csv # Example migration file"
echo ""
echo -e "${BLUE}💡 Tips:${NC}"
echo "   - Edit accounts-example.csv with your accounts"
echo "   - Run './mailmole -web :8080' for web interface"
echo "   - Check Docs/getting-started.md for detailed guide"
echo ""
echo -e "${YELLOW}⚠️  Important:${NC}"
echo "   - Keep your credentials secure"
echo "   - Use app passwords for Gmail/Outlook with 2FA"
echo "   - Review Docs/providers/ for server-specific settings"
echo ""
