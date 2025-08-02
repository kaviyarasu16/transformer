# AWS to OpenTofu Transformer

A powerful CLI tool that automatically discovers your existing AWS infrastructure and generates corresponding OpenTofu (formerly Terraform) configuration files.

## 🚀 Features

- **Comprehensive AWS Resource Discovery**: Supports 58 AWS services with full resource discovery and OpenTofu code generation
- **Automated Import Generation**: Generate OpenTofu import statements for existing AWS resources
- **Modular Architecture**: Generates clean, modular OpenTofu configurations with resource-specific modules
- **Security Best Practices**: Properly handles sensitive data with variables and encryption
- **Production Ready**: Includes comprehensive error handling, deduplication, and validation
- **Cross-Platform**: Works on macOS, Linux, and Windows
- **Modern Go Codebase**: Built with AWS SDK v2, easy to extend and maintain

## 📋 Supported AWS Services

### ✅ Fully Implemented Services (58)

These services have complete resource discovery and OpenTofu code generation:

#### Core Infrastructure
- **VPC** - Virtual Private Clouds, Subnets, Route Tables, Internet Gateways
- **EC2** - Instances, Security Groups, Key Pairs, Launch Templates
- **IAM** - Roles, Policies, Users, Groups with JSON policy formatting
- **RDS** - Database instances with proper password variable handling
- **S3** - Buckets with versioning and encryption
- **Auto Scaling** - Auto Scaling Groups and Launch Configurations

#### Networking & Load Balancing
- **ALB/ELB** - Application and Classic Load Balancers
- **Route53** - DNS records and hosted zones
- **CloudFront** - Content delivery networks
- **API Gateway** - REST and HTTP APIs

#### Compute & Containers
- **Lambda** - Serverless functions and layers
- **ECS** - Container services and task definitions
- **EKS** - Kubernetes clusters and node groups
- **ECR** - Container registries
- **Elastic Beanstalk** - Application environments

#### Storage & Databases
- **DynamoDB** - NoSQL database tables
- **ElastiCache** - Redis and Memcached clusters
- **Redshift** - Data warehouse clusters
- **Neptune** - Graph database clusters
- **DocumentDB** - MongoDB-compatible clusters
- **FSx** - File systems
- **Storage Gateway** - Hybrid storage solutions
- **Transfer Family** - File transfer services
- **Glacier** - Long-term storage vaults

#### Messaging & Integration
- **SQS** - Message queues
- **SNS** - Notification topics and subscriptions
- **MQ** - Message brokers
- **Kinesis** - Real-time data streaming
- **Firehose** - Data delivery streams

#### Monitoring & Logging
- **CloudWatch** - Metrics, alarms, and dashboards
- **CloudTrail** - API logging and audit trails with event selectors and insight selectors
- **Config** - Configuration compliance

#### Security & Management
- **KMS** - Key management and encryption
- **Secrets Manager** - Secret storage
- **SSM** - Systems Manager parameters
- **GuardDuty** - Threat detection
- **Backup** - Centralized backup service

#### Analytics & Data Processing
- **Glue** - ETL and data catalog
- **Athena** - Query service
- **QuickSight** - Business intelligence dashboards

#### Media Services
- **MediaStore** - Media storage containers
- **MediaConvert** - File-based video transcoding
- **MediaLive** - Live video processing
- **MediaTailor** - Personalization service

#### IoT & Edge Computing
- **IoT Core** - Device connectivity and management
- **IoT Analytics** - Analytics for IoT data
- **IoT Events** - Event detection and response
- **IoT SiteWise** - Industrial data collection
- **IoT Things Graph** - Visual workflow modeling
- **IoT Wireless** - Wireless connectivity

#### Developer Tools
- **CodeBuild** - Build service
- **CodeDeploy** - Deployment service
- **CodeCommit** - Source control
- **CodePipeline** - CI/CD pipelines
- **CloudFormation** - Infrastructure as code

#### Enterprise Services
- **WorkSpaces** - Virtual desktops

### ✅ Complete Implementation Status

All 58 services are fully implemented with:
- **Specific Resource Types**: No more GenericResource generation
- **Complete Discovery**: Full AWS API integration for each service
- **Proper OpenTofu Generation**: Service-specific HCL configuration
- **Error Handling**: Comprehensive error handling and nil checks
- **Resource Relationships**: Proper dependency management

## 🛠️ Installation

### Prerequisites

- **Go 1.21+** - [Download here](https://golang.org/dl/)
- **AWS CLI** - [Installation guide](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html)
- **AWS Credentials** - Configure with `aws configure`

### Build from Source

```bash
# Clone the repository
git clone <your-repo-url>
cd transformer

# Build the binary
make build

# Or build manually
go build -o transformer .
```

### Using Makefile

```bash
# Build the application
make build

# Run tests
make test

# Clean build artifacts
make clean

# Build Docker image
make docker-build
```

## 🚀 Quick Start

### 1. Configure AWS Credentials

```bash
aws configure
# Enter your AWS Access Key ID, Secret Access Key, and default region
```

### 2. Discover and Generate Infrastructure

```bash
# Discover VPC and EC2 resources only
./transformer aws --region=us-east-1 --resources=vpc,ec2 --output=./my-infrastructure --verbose
```

### 3. Discover Complete Infrastructure

```bash
# Discover all 58 AWS services in your account
./transformer aws --region=us-east-1 --all --output=./complete-infrastructure --verbose
```

### 4. Generate Import Statements

```bash
# Generate import statements for existing resources
./transformer import --resources=vpc,ec2,iam --output=./import-infrastructure --verbose
```

### 5. Use Generated Configuration

```bash
cd complete-infrastructure

# Initialize OpenTofu
tofu init

# Review the plan
tofu plan

# Apply the configuration (in non-production first!)
tofu apply
```

### 6. Import Existing Resources

```bash
cd import-infrastructure

# Review generated import statements
cat import.tf

# Run automated import script
chmod +x import.sh
./import.sh

# Verify import
tofu plan
```

## 📖 Usage Examples

### Basic Usage

```bash
# Help
./transformer --help
./transformer aws --help
./transformer import --help

# Discover specific resource types
./transformer aws --region=us-west-2 --resources=vpc,ec2,rds --output=./infra

# Discover all 58 AWS services
./transformer aws --region=us-east-1 --all --output=./complete-infra --verbose

# Generate import statements
./transformer import --resources=vpc,ec2,iam --output=./import-infra --verbose
```

### Advanced Usage

```bash
# Custom output directory
./transformer aws --region=eu-west-1 --resources=iam,s3 --output=./security-resources

# Verbose output for debugging
./transformer aws --region=us-east-1 --all --output=./debug-infra --verbose

# Specific resource discovery
./transformer aws --region=us-east-1 --resources=vpc,ec2,alb,rds,lambda --output=./web-app-infra

# Import with custom file name
./transformer import --resources=s3,rds --file=my-import.tf --output=./imports --verbose

# Import with state file generation
./transformer import --resources=vpc,ec2 --state=terraform.tfstate --output=./imports --verbose
```

## 📁 Generated Structure

The tool generates a complete OpenTofu project structure:

### Infrastructure Generation (`transformer aws`)

```
output-directory/
├── main.tf              # Main configuration with provider and module calls
├── variables.tf         # Variable definitions including sensitive values
├── outputs.tf          # Output definitions for all modules
├── versions.tf         # Provider version constraints
├── README.md           # Comprehensive documentation
└── modules/            # Resource-specific modules
    ├── vpc/
    │   ├── main.tf
    │   ├── variables.tf
    │   └── outputs.tf
    ├── ec2/
    ├── rds/
    ├── iam/
    └── ...
```

### Import Generation (`transformer import`)

```
output-directory/
├── import.tf            # Resource definitions and import statements
├── import.sh            # Automated import script
├── README.md            # Import process guide
└── terraform.tfstate    # State file template (if --state specified)
```

## 🔧 Key Features

### Infrastructure Generation

- **Comprehensive Resource Discovery**: Discovers and generates OpenTofu configurations for 58 AWS services
- **Modular Architecture**: Each AWS service gets its own module with clean dependencies
- **Security Best Practices**: Properly handles sensitive data with variables and encryption
- **Production Ready**: Includes comprehensive error handling, deduplication, and validation

### Import Generation

- **Automated Import Statements**: Generates OpenTofu import commands for existing resources
- **Resource Definitions**: Creates template resource definitions for customization
- **Automated Scripts**: Generates shell scripts to automate the import process
- **State Management**: Optional state file template generation
- **Comprehensive Documentation**: Step-by-step import guides and troubleshooting

### Security & Sensitive Data Handling

- **RDS Passwords**: Automatically converted to variables with `sensitive = true`
- **IAM Policies**: Properly formatted JSON without URL encoding
- **KMS Keys**: Secure key references
- **Secrets**: Proper variable handling for sensitive data

### Resource Deduplication

- **Automatic Deduplication**: Prevents duplicate resources in generated output
- **Unique Identifiers**: Uses resource IDs for proper tracking
- **Clean Output**: Ensures no resource conflicts

### Error Handling

- **Graceful Failures**: Continues processing even if some resources fail
- **Detailed Logging**: Verbose mode for debugging
- **Resource Validation**: Validates discovered resources before generation

## 🔒 Security Best Practices

### Before Applying Generated Configuration

1. **Review All Resources**: Ensure generated configuration matches expectations
2. **Update Sensitive Values**: Replace `CHANGE_ME` placeholders for passwords
3. **Test in Non-Production**: Always test in a safe environment first
4. **Backup Existing Infrastructure**: Create backups before making changes
5. **Check Dependencies**: Ensure all referenced resources exist

### Before Importing Existing Resources

1. **Backup Current State**: Always backup your existing OpenTofu state
2. **Review Resource Definitions**: Customize generated resource definitions to match your requirements
3. **Test Import Process**: Test the import process in a non-production environment
4. **Verify Resource IDs**: Ensure resource IDs are correct and resources exist
5. **Check Permissions**: Verify AWS permissions for import operations

### Security Checklist

- [ ] Review and update all password variables
- [ ] Ensure proper IAM permissions
- [ ] Validate security group configurations
- [ ] Check encryption settings
- [ ] Review public access settings
- [ ] Verify KMS key permissions
- [ ] Test with subset of resources first

## 🐛 Troubleshooting

### Common Issues

1. **Permission Errors**
   ```bash
   # Ensure AWS credentials are configured
   aws sts get-caller-identity
   
   # Check required permissions
   # - EC2:DescribeInstances, DescribeVpcs, etc.
   # - IAM:ListRoles, GetRole, GetRolePolicy, etc.
   # - RDS:DescribeDBInstances, etc.
   ```

2. **Resource Discovery Failures**
   ```bash
   # Use verbose mode for detailed error messages
   ./transformer aws --region=us-east-1 --resources=vpc --output=./test --verbose
   ```

3. **Generated Configuration Issues**
   ```bash
   # Validate OpenTofu syntax
   tofu validate
   
   # Check for missing dependencies
   tofu plan
   ```

4. **Import Process Issues**
   ```bash
   # Verify resource exists
   aws ec2 describe-instances --instance-ids i-12345678
   
   # Check import command syntax
   tofu import --help
   
   # Remove conflicting resources from state
   tofu state rm aws_instance.example
   ```

### Debug Mode

```bash
# Enable verbose logging
./transformer aws --region=us-east-1 --all --output=./debug --verbose

# Check specific resource types
./transformer aws --region=us-east-1 --resources=vpc,ec2 --output=./test --verbose

# Debug import generation
./transformer import --resources=vpc,ec2 --output=./debug-import --verbose
```

## 📋 Import Workflow

### Step 1: Generate Import Statements

```bash
# Generate import statements for specific resources
./transformer import --resources=vpc,ec2,iam --output=./import-infra --verbose

# Generate for all resources
./transformer import --all --output=./import-all --verbose
```

### Step 2: Review Generated Files

```bash
cd import-infra

# Review resource definitions
cat import.tf

# Review import script
cat import.sh

# Review documentation
cat README.md
```

### Step 3: Customize Resource Definitions

Edit `import.tf` to customize resource configurations:

```hcl
resource "aws_vpc" "my_vpc" {
  # Uncomment and customize these fields
  # cidr_block = "10.0.0.0/16"
  # enable_dns_hostnames = true
  # enable_dns_support = true
  
  tags = {
    Name = "my-vpc"
    Environment = "production"
  }
}
```

### Step 4: Run Import Process

```bash
# Make script executable
chmod +x import.sh

# Run automated import
./import.sh

# Or run commands manually
tofu import aws_vpc.my_vpc vpc-12345678
tofu import aws_instance.my_instance i-87654321
```

### Step 5: Verify and Apply

```bash
# Verify import
tofu plan

# Apply if everything looks correct
tofu apply
```

## 🏗️ Architecture

### Project Structure

```
transformer/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command setup
│   ├── aws.go             # AWS discovery command
│   └── import.go          # Import generation command
├── internal/              # Core application logic
│   ├── aws/              # AWS service integration
│   │   ├── client.go     # AWS client management (58+ services)
│   │   ├── discovery.go  # Resource discovery logic
│   │   └── types.go      # 58+ resource type definitions
│   ├── generator/        # OpenTofu code generation
│   │   └── generator.go  # Configuration file generation
│   └── utils/            # Utility functions
├── main.go               # Application entry point
├── go.mod               # Go module definition
├── Makefile             # Build automation
├── Dockerfile           # Container configuration
└── README.md            # This documentation
```

### Key Components

- **AWS Client**: Manages 58+ AWS SDK connections and service clients
- **Discovery Engine**: Discovers resources across all 58 AWS services
- **Generator**: Converts discovered resources to OpenTofu HCL
- **CLI Interface**: User-friendly command-line interface

## 🤝 Contributing

### Development Setup

```bash
# Clone and setup
git clone <your-repo-url>
cd transformer

# Install dependencies
go mod download

# Run tests
make test

# Build
make build
```

### Adding New AWS Services

1. **Add Service Client** in `internal/aws/client.go`
2. **Implement Discovery** in `internal/aws/discovery.go`
3. **Define Resource Type** in `internal/aws/types.go`
4. **Add Generation Logic** in `internal/generator/generator.go`
5. **Update CLI** in `cmd/aws.go`

### Service Implementation Status

- **Fully Implemented**: 58 services with complete discovery and generation
- **Total Coverage**: 58 AWS services with specific resource types
- **No Generic Resources**: All services have proper type definitions

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

### Getting Help

1. **Check Documentation**: Review this README and generated documentation
2. **Open Issues**: Report bugs and feature requests
3. **Community**: Join discussions and share solutions

### Resources

- [OpenTofu Documentation](https://opentofu.org/docs)
- [AWS SDK for Go](https://aws.github.io/aws-sdk-go-v2/)
- [HCL Syntax](https://opentofu.org/docs/language/syntax)

---

**Made with ❤️ for the OpenTofu community**
