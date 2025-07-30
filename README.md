# AWS to OpenTofu Transformer

A powerful CLI tool that automatically discovers your existing AWS infrastructure and generates corresponding OpenTofu (formerly Terraform) configuration files.

## 🚀 Features

- **Comprehensive AWS Resource Discovery**: Supports 50+ AWS services including VPC, EC2, RDS, IAM, S3, ALB, Lambda, EKS, and more
- **Modular Architecture**: Generates clean, modular OpenTofu configurations with resource-specific modules
- **Security Best Practices**: Properly handles sensitive data with variables and encryption
- **Production Ready**: Includes comprehensive error handling, deduplication, and validation
- **Cross-Platform**: Works on macOS, Linux, and Windows

## 📋 Supported AWS Services

### Core Infrastructure
- **VPC** - Virtual Private Clouds, Subnets, Route Tables, Internet Gateways
- **EC2** - Instances, Security Groups, Key Pairs, Launch Templates
- **RDS** - Database instances with proper password variable handling
- **IAM** - Roles, Policies, Users, Groups with JSON policy formatting

### Networking & Load Balancing
- **ALB/ELB** - Application and Classic Load Balancers
- **Auto Scaling** - Auto Scaling Groups and Launch Configurations
- **Route53** - DNS records and hosted zones
- **CloudFront** - Content delivery networks

### Compute & Containers
- **Lambda** - Serverless functions and layers
- **ECS** - Container services and task definitions
- **EKS** - Kubernetes clusters and node groups
- **ECR** - Container registries

### Storage & Databases
- **S3** - Buckets with versioning and encryption
- **DynamoDB** - NoSQL database tables
- **ElastiCache** - Redis and Memcached clusters
- **Redshift** - Data warehouse clusters

### Messaging & Integration
- **SQS** - Message queues
- **SNS** - Notification topics and subscriptions
- **API Gateway** - REST and HTTP APIs

### Monitoring & Logging
- **CloudWatch** - Metrics, alarms, and dashboards
- **CloudTrail** - API logging and audit trails
- **X-Ray** - Distributed tracing

### Security & Management
- **KMS** - Key management and encryption
- **Secrets Manager** - Secret storage
- **SSM** - Systems Manager parameters
- **Config** - Configuration compliance

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

### 2. Discover Specific Resources

```bash
# Discover VPC and EC2 resources only
./transformer aws --region=us-east-1 --resources=vpc,ec2 --output=./my-infrastructure --verbose
```

### 3. Discover Complete Infrastructure

```bash
# Discover all resources in your AWS account
./transformer aws --region=us-east-1 --all --output=./complete-infrastructure --verbose
```

### 4. Use Generated Configuration

```bash
cd complete-infrastructure

# Initialize OpenTofu
tofu init

# Review the plan
tofu plan

# Apply the configuration (in non-production first!)
tofu apply
```

## 📖 Usage Examples

### Basic Usage

```bash
# Help
./transformer --help
./transformer aws --help

# Discover specific resource types
./transformer aws --region=us-west-2 --resources=vpc,ec2,rds --output=./infra

# Discover all resources
./transformer aws --region=us-east-1 --all --output=./complete-infra --verbose
```

### Advanced Usage

```bash
# Custom output directory
./transformer aws --region=eu-west-1 --resources=iam,s3 --output=./security-resources

# Verbose output for debugging
./transformer aws --region=us-east-1 --all --output=./debug-infra --verbose

# Specific resource discovery
./transformer aws --region=us-east-1 --resources=vpc,ec2,alb,rds,lambda --output=./web-app-infra
```

## 📁 Generated Structure

The tool generates a complete OpenTofu project structure:

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

## 🔧 Key Features

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

### Modular Architecture

- **Resource Separation**: Each AWS service gets its own module
- **Clean Dependencies**: Proper module references and data sources
- **Maintainable Code**: Easy to modify and extend

## 🔒 Security Best Practices

### Before Applying Generated Configuration

1. **Review All Resources**: Ensure generated configuration matches expectations
2. **Update Sensitive Values**: Replace `CHANGE_ME` placeholders for passwords
3. **Test in Non-Production**: Always test in a safe environment first
4. **Backup Existing Infrastructure**: Create backups before making changes
5. **Check Dependencies**: Ensure all referenced resources exist

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

### Debug Mode

```bash
# Enable verbose logging
./transformer aws --region=us-east-1 --all --output=./debug --verbose

# Check specific resource types
./transformer aws --region=us-east-1 --resources=vpc,ec2 --output=./test --verbose
```

## 🏗️ Architecture

### Project Structure

```
transformer/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command setup
│   └── aws.go             # AWS discovery command
├── internal/              # Core application logic
│   ├── aws/              # AWS service integration
│   │   ├── client.go     # AWS client management
│   │   ├── discovery.go  # Resource discovery logic
│   │   └── types.go      # Resource type definitions
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

- **AWS Client**: Manages AWS SDK connections and service clients
- **Discovery Engine**: Discovers resources across multiple AWS services
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
