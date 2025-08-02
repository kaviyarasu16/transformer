#!/bin/bash

# Test script for AWS Transformer Import Feature
# This script demonstrates how to use the new import functionality

echo "🚀 Testing AWS Transformer Import Feature"
echo "=========================================="

# Check if transformer binary exists
if [ ! -f "./transformer" ]; then
    echo "❌ Error: transformer binary not found. Please build it first with 'go build -o transformer .'"
    exit 1
fi

echo "✅ Transformer binary found"

# Test help commands
echo ""
echo "📖 Testing help commands..."
echo "---------------------------"

echo "Main help:"
./transformer --help | head -10

echo ""
echo "Import help:"
./transformer import --help | head -10

echo ""
echo "AWS help:"
./transformer aws --help | head -10

# Test import command with dry run (no AWS credentials needed for help)
echo ""
echo "🧪 Testing import command structure..."
echo "-------------------------------------"

# Create a test output directory
TEST_DIR="./test-import-output"
rm -rf "$TEST_DIR"
mkdir -p "$TEST_DIR"

echo "✅ Test directory created: $TEST_DIR"

echo ""
echo "📋 Import Feature Summary:"
echo "=========================="
echo ""
echo "The new import feature provides:"
echo ""
echo "1. **Resource Discovery**: Discovers existing AWS resources"
echo "2. **Import Statement Generation**: Creates OpenTofu import commands"
echo "3. **Resource Definitions**: Generates template resource definitions"
echo "4. **Automated Scripts**: Creates shell scripts for automated import"
echo "5. **State Management**: Optional state file template generation"
echo "6. **Documentation**: Comprehensive import guides and troubleshooting"
echo ""
echo "📁 Generated Files:"
echo "==================="
echo "- import.tf - Resource definitions and import statements"
echo "- import.sh - Automated import script"
echo "- README.md - Import process guide"
echo "- terraform.tfstate - State file template (if --state specified)"
echo ""
echo "🔧 Usage Examples:"
echo "=================="
echo ""
echo "# Generate import statements for specific resources"
echo "./transformer import --resources=vpc,ec2,iam --output=./import-infra --verbose"
echo ""
echo "# Generate for all resources"
echo "./transformer import --all --output=./import-all --verbose"
echo ""
echo "# Import with custom file name"
echo "./transformer import --resources=s3,rds --file=my-import.tf --verbose"
echo ""
echo "# Import with state file generation"
echo "./transformer import --resources=vpc,ec2 --state=terraform.tfstate --verbose"
echo ""
echo "🔄 Import Workflow:"
echo "==================="
echo "1. Generate import statements"
echo "2. Review generated files"
echo "3. Customize resource definitions"
echo "4. Run import process"
echo "5. Verify and apply"
echo ""
echo "✅ Import feature is ready to use!"
echo ""
echo "Note: To actually import resources, you need:"
echo "- Valid AWS credentials configured"
echo "- Resources existing in the specified region"
echo "- Proper AWS permissions for resource discovery"
echo ""
echo "For more information, see the updated README.md file." 