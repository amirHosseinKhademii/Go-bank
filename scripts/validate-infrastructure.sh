#!/bin/bash
# Validate infrastructure configurations before deployment

set -e

echo "🔍 Infrastructure Validation Script"
echo "===================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

check_tool() {
    if ! command -v "$1" &> /dev/null; then
        echo -e "${YELLOW}⚠️  $1 not found. Install it to enable full validation.${NC}"
        return 1
    fi
    echo -e "${GREEN}✓ $1 found${NC}"
    return 0
}

validate_yaml() {
    local file=$1
    echo ""
    echo "Checking YAML syntax: $file"
    if command -v yamllint &> /dev/null; then
        yamllint "$file" && echo -e "${GREEN}✓ $file is valid${NC}" || return 1
    else
        # Fallback to basic validation with python
        python3 -c "import yaml; yaml.safe_load(open('$file'))" && \
        echo -e "${GREEN}✓ $file is valid${NC}" || return 1
    fi
}

echo "Checking required tools..."
echo ""

has_terraform=0
has_kubectl=0
has_docker=0

check_tool "git" || true
check_tool "docker" && has_docker=1 || true
check_tool "terraform" && has_terraform=1 || true
check_tool "kubectl" && has_kubectl=1 || true

echo ""
echo "Validating Kubernetes manifests..."
echo "=================================="

for file in k8s/*.yaml; do
    validate_yaml "$file" || true
done

if [ $has_kubectl -eq 1 ]; then
    echo ""
    echo "Running kubectl dry-run validation..."
    for file in k8s/*.yaml; do
        kubectl apply -f "$file" --dry-run=client -o yaml > /dev/null 2>&1 && \
        echo -e "${GREEN}✓ $file passes kubectl validation${NC}" || \
        echo -e "${RED}✗ $file fails kubectl validation${NC}"
    done
fi

echo ""
echo "Validating Terraform configurations..."
echo "======================================"

if [ $has_terraform -eq 1 ]; then
    cd terraform

    echo "Checking Terraform format..."
    terraform fmt -check -recursive . && \
    echo -e "${GREEN}✓ Terraform files are properly formatted${NC}" || \
    echo -e "${YELLOW}⚠️  Run 'terraform fmt -recursive .' to fix formatting${NC}"

    echo ""
    echo "Validating Terraform syntax..."
    terraform init -backend=false > /dev/null 2>&1 && \
    terraform validate && \
    echo -e "${GREEN}✓ Terraform configuration is valid${NC}" || \
    echo -e "${RED}✗ Terraform validation failed${NC}"

    cd ..
else
    echo -e "${YELLOW}⚠️  Terraform not installed. Skipping Terraform validation.${NC}"
fi

echo ""
echo "Validating Docker configuration..."
echo "===================================="

if [ $has_docker -eq 1 ]; then
    echo "Checking Dockerfile syntax..."
    docker build --dry-run . &> /dev/null && \
    echo -e "${GREEN}✓ Dockerfile is valid${NC}" || \
    echo -e "${RED}✗ Dockerfile validation failed${NC}"
else
    echo -e "${YELLOW}⚠️  Docker not installed. Skipping Docker validation.${NC}"
fi

echo ""
echo "Checking GitHub Actions workflow..."
echo "===================================="
validate_yaml ".github/workflows/cd.yml" || true

echo ""
echo "Configuration Checklist..."
echo "=========================="
echo ""

checks=(
    ["terraform/terraform.tfvars"]="Terraform variables configured"
    [".env"]="Environment variables loaded"
    ["go.mod"]="Go module file present"
    ["Dockerfile"]="Dockerfile present"
)

for file in "${!checks[@]}"; do
    if [ -f "$file" ]; then
        echo -e "${GREEN}✓${NC} ${checks[$file]}"
    else
        echo -e "${YELLOW}⚠️ ${checks[$file]} - $file not found${NC}"
    fi
done

echo ""
echo "Next Steps:"
echo "==========="
echo "1. Copy terraform/terraform.tfvars.example to terraform/terraform.tfvars"
echo "2. Update terraform/terraform.tfvars with your AWS account ID and settings"
echo "3. Ensure AWS credentials are configured (aws configure or environment variables)"
echo "4. Run 'terraform plan' to preview infrastructure changes"
echo "5. Run 'terraform apply' to deploy infrastructure"
echo "6. Push code to main branch to trigger GitHub Actions CD pipeline"
echo ""
echo "Validation complete!"
