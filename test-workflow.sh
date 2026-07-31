#!/bin/bash

echo "Testing CI workflow locally for retui packages..."

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Pin versions to match .github/workflows/ci.yml
GOLANGCI_LINT_VERSION="v2.12.2"
GOSEC_VERSION="latest" # gosec Action uses @master; go install below tracks latest tagged release

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Track overall failure without exiting immediately, so we get a full report
FAILED=0

# Test 1: Build
echo "📦 Testing build..."
if go build ./retui/...; then
    echo -e "${GREEN}✅ Build passed${NC}"
else
    echo -e "${RED}❌ Build failed${NC}"
    exit 1
fi

# Test 2: Vet
echo "🔍 Testing vet..."
if go vet ./retui/...; then
    echo -e "${GREEN}✅ Vet passed${NC}"
else
    echo -e "${RED}❌ Vet failed${NC}"
    exit 1
fi

# Test 3: Gofmt
echo "📝 Testing gofmt..."
fmt_out="$(gofmt -l ./retui)"
if [ -z "$fmt_out" ]; then
    echo -e "${GREEN}✅ Gofmt passed${NC}"
else
    echo -e "${RED}❌ Gofmt failed${NC}"
    echo "The following files are not gofmt'd:"
    echo "$fmt_out"
    exit 1
fi

# Test 4: Tests with coverage (matches CI: race + atomic covermode + coverpkg)
echo "🧪 Testing tests..."
if go test ./retui/... -v -race -covermode=atomic -coverprofile=coverage.out -coverpkg=./retui/...; then
    echo -e "${GREEN}✅ Tests passed${NC}"
    go tool cover -func=coverage.out | grep total
else
    echo -e "${RED}❌ Tests failed${NC}"
    exit 1
fi

# # Test 5: Lint with golangci-lint (pinned to v2.12.2 to match CI — v1.x lacks go1.26 support)
# echo "🔎 Testing lint (golangci-lint ${GOLANGCI_LINT_VERSION})..."

# installed_version=""
# if command_exists golangci-lint; then
#     installed_version="$(golangci-lint version 2>/dev/null | grep -oE 'version [^ ]+' | awk '{print $2}')"
# fi

# if [ "$installed_version" != "${GOLANGCI_LINT_VERSION#v}" ]; then
#     echo -e "${YELLOW}⚠️  golangci-lint ${GOLANGCI_LINT_VERSION} not found (have: ${installed_version:-none}). Installing...${NC}"
#     go install "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}"
# fi

# if golangci-lint run ./retui/... --timeout=5m; then
#     echo -e "${GREEN}✅ golangci-lint passed${NC}"
# else
#     echo -e "${RED}❌ golangci-lint failed${NC}"
#     exit 1
# fi

# Test 6: gosec (matches CI's securego/gosec action)
echo "🛡️  Testing gosec..."

if ! command_exists gosec; then
    echo -e "${YELLOW}⚠️  gosec not found. Installing...${NC}"
    go install github.com/securego/gosec/v2/cmd/gosec@latest
fi

if gosec ./retui/...; then
    echo -e "${GREEN}✅ gosec passed${NC}"
else
    echo -e "${RED}❌ gosec failed${NC}"
    exit 1
fi

# Test 7: govulncheck (optional / non-failing, matches CI)
echo "🔒 Testing govulncheck..."

if ! command_exists govulncheck; then
    echo -e "${YELLOW}⚠️  govulncheck not found. Installing...${NC}"
    go install golang.org/x/vuln/cmd/govulncheck@latest
fi

if govulncheck ./retui/...; then
    echo -e "${GREEN}✅ govulncheck passed${NC}"
else
    echo -e "${YELLOW}⚠️  govulncheck found issues (not failing)${NC}"
fi

# Test 8: Dependency check
echo "📦 Checking dependencies..."
go mod tidy
if git diff --exit-code go.mod go.sum; then
    echo -e "${GREEN}✅ Dependencies are tidy${NC}"
else
    echo -e "${RED}❌ Dependencies are not tidy. Run 'go mod tidy' and commit the changes${NC}"
    exit 1
fi

echo -e "${GREEN}✅ All tests passed!${NC}"
echo -e "${GREEN}📊 Coverage report saved to coverage.out${NC}"