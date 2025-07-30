# AWS to OpenTofu Transformer

A powerful CLI tool to transform existing AWS infrastructure into OpenTofu (formerly Terraform) Infrastructure as Code (IaC) scripts. This tool discovers AWS resources and generates corresponding OpenTofu configuration files.

## 🚀 Quick Install

### One-liner Installation
```bash
curl -sSL https://raw.githubusercontent.com/kaviyarasu16/transformer/main/install.sh | bash
```

### Manual Installation
```bash
# Clone the repository
git clone https://github.com/kaviyarasu16/transformer.git
cd transformer

# Run the installer
./install.sh

# Or install directly with Go
go install github.com/kaviyarasu16/transformer@latest
```

### Prerequisites
- Go 1.21 or later
- AWS CLI configured with appropriate permissions
- OpenTofu (formerly Terraform) for applying generated configurations

## 📖 Usage

### Basic Commands
```bash
# Show help
transformer --help

# Show AWS command help
transformer aws --help

# Show version
transformer --version
```

### Discover Specific Resources
```bash
# Discover VPC and EC2 resources
transformer aws --region=us-east-1 --resources=vpc,ec2 --output=./infrastructure

# Discover IAM and RDS resources
transformer aws --region=us-east-1 --resources=iam,rds --output=./infrastructure
```

### Discover All Resources
```bash
# Discover all available resources
transformer aws --region=us-east-1 --all --output=./complete-infrastructure
```

### Verbose Output
```bash
# Enable verbose logging
transformer aws --region=us-east-1 --all --output=./infrastructure --verbose
```

## 🔧 Configuration

### AWS Credentials
Ensure your AWS credentials are configured:
```bash
aws configure
```

### Environment Variables
You can also use environment variables:
```bash
export AWS_REGION=us-east-1
export AWS_PROFILE=my-profile
transformer aws --all --output=./infrastructure
```

## 📁 Generated Output

The tool generates a complete OpenTofu project structure:

```
infrastructure/
├── main.tf              # Main configuration with provider setup
├── variables.tf         # Variable definitions
├── outputs.tf          # Output definitions
├── versions.tf         # Provider version constraints
├── README.md           # Documentation and usage guide
└── modules/            # Resource-specific modules
    ├── vpc/
    ├── ec2/
    ├── iam/
    ├── rds/
    └── ...
```

## 🛠️ Supported AWS Resources

- **VPC** - Virtual Private Clouds, Subnets, Route Tables
- **EC2** - Instances, Security Groups, Key Pairs
- **IAM** - Roles, Policies, Users, Groups
- **RDS** - Database Instances, Subnet Groups
- **S3** - Buckets, Bucket Policies
- **ALB/ELB** - Application and Classic Load Balancers
- **Lambda** - Functions, Layers, Event Sources
- **SQS** - Queues, Queue Policies
- **SNS** - Topics, Subscriptions
- **CloudWatch** - Log Groups, Dashboards, Alarms
- **CloudTrail** - Trails, Event Selectors
- **ECS** - Clusters, Services, Task Definitions
- **EKS** - Clusters, Node Groups
- **Auto Scaling** - Groups, Launch Configurations
- **And many more...**

## 🔒 Security Features

- **Sensitive Data Handling**: RDS passwords are stored as variables
- **IAM Policy Decoding**: Properly formatted JSON policies
- **Resource Deduplication**: Prevents duplicate resource generation
- **Nil Pointer Safety**: Robust error handling for missing data

## 🚀 Quick Start

1. **Install the tool**:
   ```bash
   curl -sSL https://raw.githubusercontent.com/kaviyarasu16/transformer/main/install.sh | bash
   ```

2. **Configure AWS credentials**:
   ```bash
   aws configure
   ```

3. **Discover your infrastructure**:
   ```bash
   transformer aws --region=us-east-1 --resources=vpc,ec2 --output=./my-infrastructure
   ```

4. **Review generated files**:
   ```bash
   cd my-infrastructure
   ls -la
   ```

5. **Initialize and plan with OpenTofu**:
   ```bash
   tofu init
   tofu plan
   ```

## 🏗️ Development

### Building from Source
```bash
# Clone the repository
git clone https://github.com/kaviyarasu16/transformer.git
cd transformer

# Build the binary
make build

# Or build manually
go build -o transformer .
```

### Running Tests
```bash
make test
```

### Local Installation
```bash
# Install from local source
./install.sh --local
```

## 📋 Requirements

- **Go**: 1.21 or later
- **AWS SDK**: v2 (automatically managed)
- **Cobra**: CLI framework (automatically managed)
- **Viper**: Configuration management (automatically managed)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **GitHub Issues**: [Report bugs or request features](https://github.com/kaviyarasu16/transformer/issues)
- **Documentation**: [Project Wiki](https://github.com/kaviyarasu16/transformer/wiki)
- **Discussions**: [Community discussions](https://github.com/kaviyarasu16/transformer/discussions)

## 🙏 Acknowledgments

- AWS SDK for Go v2 team
- HashiCorp for OpenTofu/Terraform
- Cobra CLI framework contributors
- The open-source community

---

**Made with ❤️ for the DevOps community**
