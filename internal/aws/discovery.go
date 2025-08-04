package aws

import (
	"context"

	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"

	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/aws/aws-sdk-go-v2/service/codedeploy"
	"github.com/aws/aws-sdk-go-v2/service/codecommit"
	"github.com/aws/aws-sdk-go-v2/service/codepipeline"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticsearchservice"
	"github.com/aws/aws-sdk-go-v2/service/firehose"
	"github.com/aws/aws-sdk-go-v2/service/fsx"
	"github.com/aws/aws-sdk-go-v2/service/glacier"
	"github.com/aws/aws-sdk-go-v2/service/glue"

	"github.com/aws/aws-sdk-go-v2/service/iam"

	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/mediaconvert"
	"github.com/aws/aws-sdk-go-v2/service/medialive"
	"github.com/aws/aws-sdk-go-v2/service/mediastore"

	"github.com/aws/aws-sdk-go-v2/service/mq"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/sns"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/storagegateway"
	"github.com/aws/aws-sdk-go-v2/service/transfer"
	"github.com/aws/aws-sdk-go-v2/service/workspaces"
)

// Helper function to decode URL-encoded JSON
func decodeURLEncodedJSON(encoded string) string {
	decoded, err := url.QueryUnescape(encoded)
	if err != nil {
		// If decoding fails, return the original string
		return encoded
	}
	
	// Remove newlines and extra whitespace for OpenTofu compatibility
	decoded = strings.ReplaceAll(decoded, "\n", "")
	decoded = strings.ReplaceAll(decoded, "\r", "")
	decoded = strings.ReplaceAll(decoded, "\t", "")
	
	// Remove extra spaces between JSON elements
	re := regexp.MustCompile(`\s+`)
	decoded = re.ReplaceAllString(decoded, " ")
	
	return strings.TrimSpace(decoded)
}

// Helper function to parse string to int32
func parseStringToInt32(s string) int32 {
	if s == "" {
		return 0
	}
	val, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0
	}
	return int32(val)
}

// Helper function to check if a resource has already been seen
func isResourceSeen(seenResources map[string]bool, resourceID string) bool {
	if seenResources[resourceID] {
		return true
	}
	seenResources[resourceID] = true
	return false
}

// Helper function to convert EC2 tags to map
func convertTags(tags []ec2types.Tag) map[string]string {
	result := make(map[string]string)
	for _, tag := range tags {
		if tag.Key != nil && tag.Value != nil {
			result[*tag.Key] = *tag.Value
		}
	}
	return result
}

// extractSecurityGroupIDs extracts security group IDs from VPC security group memberships
func extractSecurityGroupIDs(vpcSecurityGroups []rdstypes.VpcSecurityGroupMembership) []string {
	var securityGroupIDs []string
	for _, sg := range vpcSecurityGroups {
		if sg.VpcSecurityGroupId != nil {
			securityGroupIDs = append(securityGroupIDs, *sg.VpcSecurityGroupId)
		}
	}
	return securityGroupIDs
}

// discoverVPCResources discovers VPC resources
func (c *Client) discoverVPCResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.ec2Client.DescribeVpcs(context.TODO(), &ec2.DescribeVpcsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe VPCs: %w", err)
	}

	for _, vpc := range result.Vpcs {
		// Get VPC name from tags
		name := *vpc.VpcId
		if vpc.Tags != nil {
			for _, tag := range vpc.Tags {
				if tag.Key != nil && *tag.Key == "Name" && tag.Value != nil {
					name = *tag.Value
					break
				}
			}
		}

		resource := &VPCResource{
			BaseResource: BaseResource{
				Type:   "vpc",
				ID:     *vpc.VpcId,
				Name:   name,
				Region: c.region,
				Tags:   convertTags(vpc.Tags),
			},
			CIDRBlock:          *vpc.CidrBlock,
			EnableDNSHostnames: true, // Default value, would need separate API call
			EnableDNSSupport:   true, // Default value, would need separate API call
			InstanceTenancy:    string(vpc.InstanceTenancy),
		}

		// Add IPv6 CIDR block if present
		if vpc.Ipv6CidrBlockAssociationSet != nil && len(vpc.Ipv6CidrBlockAssociationSet) > 0 {
			for _, ipv6Assoc := range vpc.Ipv6CidrBlockAssociationSet {
				if ipv6Assoc.Ipv6CidrBlock != nil {
					resource.IPv6CIDRBlock = *ipv6Assoc.Ipv6CidrBlock
					break
				}
			}
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverEC2Resources discovers EC2 instance resources
func (c *Client) discoverEC2Resources() ([]Resource, error) {
	var resources []Resource

	result, err := c.ec2Client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe EC2 instances: %w", err)
	}

	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			// Get instance name from tags
			name := *instance.InstanceId
			if instance.Tags != nil {
				for _, tag := range instance.Tags {
					if tag.Key != nil && *tag.Key == "Name" && tag.Value != nil {
						name = *tag.Value
						break
					}
				}
			}

			// Get security groups
			var securityGroups []string
			if instance.SecurityGroups != nil {
			for _, sg := range instance.SecurityGroups {
					if sg.GroupId != nil {
				securityGroups = append(securityGroups, *sg.GroupId)
					}
				}
			}

			// Get root block device
			var rootBlockDevice *BlockDeviceMapping
			if instance.BlockDeviceMappings != nil {
				for _, bdm := range instance.BlockDeviceMappings {
					if bdm.DeviceName != nil && *bdm.DeviceName == "/dev/xvda" {
						if bdm.Ebs != nil {
							rootBlockDevice = &BlockDeviceMapping{
								DeviceName: *bdm.DeviceName,
								VolumeSize: 20, // Default size, would need separate API call for actual size
								VolumeType: "gp2", // Default type
								Encrypted:  false, // Default value
							}
							// Note: EBS block device details would need separate API calls
						}
						break
					}
				}
			}

			resource := &EC2InstanceResource{
				BaseResource: BaseResource{
					Type:   "ec2",
					ID:     *instance.InstanceId,
					Name:   name,
					Region: c.region,
					Tags:   convertTags(instance.Tags),
				},
				InstanceType: string(instance.InstanceType),
				AMI:          *instance.ImageId,
				SubnetID:     *instance.SubnetId,
				SecurityGroups: securityGroups,
				RootBlockDevice: rootBlockDevice,
			}

			// Add optional fields
			if instance.KeyName != nil {
				resource.KeyName = *instance.KeyName
			}
			if instance.IamInstanceProfile != nil && instance.IamInstanceProfile.Arn != nil {
				resource.IAMInstanceProfile = *instance.IamInstanceProfile.Arn
			}
			// Note: UserData would need separate API call to retrieve

			resources = append(resources, resource)
		}
	}

	return resources, nil
}

// discoverIAMResources discovers IAM role resources
func (c *Client) discoverIAMResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.iamClient.ListRoles(context.TODO(), &iam.ListRolesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list IAM roles: %w", err)
	}

	for _, role := range result.Roles {
		// Skip roles with names that are too long for OpenTofu
		if role.RoleName != nil && len(*role.RoleName) > 64 {
			continue
		}
		
		// Get role details
		roleDetails, err := c.iamClient.GetRole(context.TODO(), &iam.GetRoleInput{
			RoleName: role.RoleName,
		})
		if err != nil {
			continue
		}

		// Get managed policies
		attachedPolicies, err := c.iamClient.ListAttachedRolePolicies(context.TODO(), &iam.ListAttachedRolePoliciesInput{
			RoleName: role.RoleName,
		})
		if err != nil {
			continue
		}

		var managedPolicyARNs []string
		for _, policy := range attachedPolicies.AttachedPolicies {
			managedPolicyARNs = append(managedPolicyARNs, *policy.PolicyArn)
		}

		// Get inline policies
		inlinePolicies, err := c.iamClient.ListRolePolicies(context.TODO(), &iam.ListRolePoliciesInput{
			RoleName: role.RoleName,
		})
		if err != nil {
			continue
		}

		inlinePolicyMap := make(map[string]string)
		for _, policyName := range inlinePolicies.PolicyNames {
			policyDoc, err := c.iamClient.GetRolePolicy(context.TODO(), &iam.GetRolePolicyInput{
				RoleName:   role.RoleName,
				PolicyName: &policyName,
			})
			if err == nil && policyDoc.PolicyDocument != nil {
				// Decode URL-encoded policy document
				decodedPolicy := decodeURLEncodedJSON(*policyDoc.PolicyDocument)
				inlinePolicyMap[policyName] = decodedPolicy
			}
		}

		// Convert tags
		tags := make(map[string]string)
		if roleDetails.Role.Tags != nil {
			for _, tag := range roleDetails.Role.Tags {
				if tag.Key != nil && tag.Value != nil {
					tags[*tag.Key] = *tag.Value
				}
			}
		}

		// Decode assume role policy
		var assumeRolePolicy string
		if roleDetails.Role.AssumeRolePolicyDocument != nil {
			assumeRolePolicy = decodeURLEncodedJSON(*roleDetails.Role.AssumeRolePolicyDocument)
		}

		resource := &IAMRoleResource{
			BaseResource: BaseResource{
				Type:   "iam",
				ID:     *role.RoleId,
				Name:   *role.RoleName,
				Region: c.region,
				Tags:   tags,
			},
			AssumeRolePolicy: assumeRolePolicy,
			ManagedPolicyARNs: managedPolicyARNs,
			InlinePolicies:    inlinePolicyMap,
			Path:              *role.Path,
			Description:       "",
		}

		if role.Description != nil {
			resource.Description = *role.Description
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverRDSResources discovers RDS instance resources
func (c *Client) discoverRDSResources() ([]Resource, error) {
	var resources []Resource
	seenResources := make(map[string]bool)

	result, err := c.rdsClient.DescribeDBInstances(context.TODO(), &rds.DescribeDBInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe RDS instances: %w", err)
	}

	for _, instance := range result.DBInstances {
		// Use DBInstanceIdentifier as the unique identifier to prevent duplicates
		resourceID := *instance.DBInstanceIdentifier
		
		// Skip if we've already seen this resource
		if isResourceSeen(seenResources, resourceID) {
			continue
		}

		// Convert tags
		tags := make(map[string]string)
		if instance.TagList != nil {
			for _, tag := range instance.TagList {
				if tag.Key != nil && tag.Value != nil {
					tags[*tag.Key] = *tag.Value
				}
			}
		}

		// Get VPC security groups
		var vpcSecurityGroupIDs []string
		if instance.VpcSecurityGroups != nil {
		for _, sg := range instance.VpcSecurityGroups {
				if sg.VpcSecurityGroupId != nil {
					vpcSecurityGroupIDs = append(vpcSecurityGroupIDs, *sg.VpcSecurityGroupId)
				}
			}
		}

		// Safely get values with nil checks
		engine := ""
		if instance.Engine != nil {
			engine = *instance.Engine
		}
		
		engineVersion := ""
		if instance.EngineVersion != nil {
			engineVersion = *instance.EngineVersion
		}
		
		instanceClass := ""
		if instance.DBInstanceClass != nil {
			instanceClass = *instance.DBInstanceClass
		}
		
		allocatedStorage := 0
		if instance.AllocatedStorage != nil {
			allocatedStorage = int(*instance.AllocatedStorage)
		}
		
		storageType := "gp2"
		if instance.StorageType != nil {
			storageType = *instance.StorageType
		}
		
		storageEncrypted := false
		if instance.StorageEncrypted != nil {
			storageEncrypted = *instance.StorageEncrypted
		}
		
		dbName := ""
		if instance.DBName != nil {
			dbName = *instance.DBName
		}

		username := ""
		if instance.MasterUsername != nil {
			username = *instance.MasterUsername
		}
		
		dbSubnetGroupName := ""
		if instance.DBSubnetGroup != nil && instance.DBSubnetGroup.DBSubnetGroupName != nil {
			dbSubnetGroupName = *instance.DBSubnetGroup.DBSubnetGroupName
		}
		
		backupRetentionPeriod := 0
		if instance.BackupRetentionPeriod != nil {
			backupRetentionPeriod = int(*instance.BackupRetentionPeriod)
		}
		
		multiAZ := false
		if instance.MultiAZ != nil {
			multiAZ = *instance.MultiAZ
		}
		
		publiclyAccessible := false
		if instance.PubliclyAccessible != nil {
			publiclyAccessible = *instance.PubliclyAccessible
		}

		resource := &RDSInstanceResource{
			BaseResource: BaseResource{
				Type:   "rds",
				ID:     *instance.DBInstanceIdentifier,
				Name:   *instance.DBInstanceIdentifier,
				Region: c.region,
				Tags:   tags,
			},
			Engine:               engine,
			EngineVersion:        engineVersion,
			InstanceClass:        instanceClass,
			AllocatedStorage:     allocatedStorage,
			StorageType:          storageType,
			StorageEncrypted:     storageEncrypted,
			DBName:               dbName,
			Username:             username,
			Password:             "CHANGE_ME", // Will be replaced with variable
			VpcSecurityGroupIDs:  vpcSecurityGroupIDs,
			DBSubnetGroupName:    dbSubnetGroupName,
			BackupRetentionPeriod: backupRetentionPeriod,
			MultiAZ:              multiAZ,
			PubliclyAccessible:   publiclyAccessible,
			SkipFinalSnapshot:    false,
		}

		// Add KMS key if present
		if instance.KmsKeyId != nil {
			resource.KMSKeyID = *instance.KmsKeyId
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverS3Resources discovers S3 bucket resources
func (c *Client) discoverS3Resources() ([]Resource, error) {
	var resources []Resource

	result, err := c.s3Client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list S3 buckets: %w", err)
	}

	for _, bucket := range result.Buckets {
		// Get bucket location
		location, err := c.s3Client.GetBucketLocation(context.TODO(), &s3.GetBucketLocationInput{
			Bucket: bucket.Name,
		})
		if err != nil {
			continue
		}

		// Only include buckets in the current region
		if location.LocationConstraint != "" && string(location.LocationConstraint) != c.region {
			continue
		}

		// Get bucket versioning
		versioning, err := c.s3Client.GetBucketVersioning(context.TODO(), &s3.GetBucketVersioningInput{
			Bucket: bucket.Name,
		})
		if err != nil {
			continue
		}

		// Get bucket encryption
		encryption, err := c.s3Client.GetBucketEncryption(context.TODO(), &s3.GetBucketEncryptionInput{
			Bucket: bucket.Name,
		})
		if err != nil {
			continue
		}

		// Get bucket tags
		tags, err := c.s3Client.GetBucketTagging(context.TODO(), &s3.GetBucketTaggingInput{
			Bucket: bucket.Name,
		})
		if err != nil {
			continue
		}

		tagMap := make(map[string]string)
		if tags.TagSet != nil {
			for _, tag := range tags.TagSet {
				if tag.Key != nil && tag.Value != nil {
					tagMap[*tag.Key] = *tag.Value
				}
			}
		}

		resource := &S3BucketResource{
			BaseResource: BaseResource{
				Type:   "s3",
				ID:     *bucket.Name,
				Name:   *bucket.Name,
				Region: c.region,
				Tags:   tagMap,
			},
			VersioningEnabled: versioning.Status == "Enabled",
			EncryptionEnabled: encryption.ServerSideEncryptionConfiguration != nil,
		}

		// Add KMS key if present
		if encryption.ServerSideEncryptionConfiguration != nil {
			for _, rule := range encryption.ServerSideEncryptionConfiguration.Rules {
				if rule.ApplyServerSideEncryptionByDefault != nil && rule.ApplyServerSideEncryptionByDefault.KMSMasterKeyID != nil {
					resource.KMSKeyID = *rule.ApplyServerSideEncryptionByDefault.KMSMasterKeyID
					break
				}
			}
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverLoadBalancerResources discovers Application Load Balancer resources
func (c *Client) discoverLoadBalancerResources() ([]Resource, error) {
	var resources []Resource
	seenResources := make(map[string]bool)

	result, err := c.elbv2Client.DescribeLoadBalancers(context.TODO(), &elasticloadbalancingv2.DescribeLoadBalancersInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe load balancers: %w", err)
	}

	for _, lb := range result.LoadBalancers {
		// Use LoadBalancerName as the unique identifier to prevent duplicates
		resourceID := *lb.LoadBalancerName
		
		// Skip if we've already seen this resource
		if isResourceSeen(seenResources, resourceID) {
			continue
		}

		// Convert tags - LoadBalancer doesn't have Tags field in this version
		tags := make(map[string]string)
		// Tags would need to be fetched separately if needed

		// Get security groups
		var securityGroups []string
		if lb.SecurityGroups != nil {
		for _, sg := range lb.SecurityGroups {
			securityGroups = append(securityGroups, sg)
			}
		}

		// Get subnets
		var subnets []string
		if lb.AvailabilityZones != nil {
		for _, az := range lb.AvailabilityZones {
				if az.SubnetId != nil {
			subnets = append(subnets, *az.SubnetId)
				}
			}
		}

		resource := &LoadBalancerResource{
			BaseResource: BaseResource{
				Type:   "alb",
				ID:     *lb.LoadBalancerArn,
				Name:   *lb.LoadBalancerName,
				Region: c.region,
				Tags:   tags,
			},
			Internal:           lb.Scheme == "internal",
			LoadBalancerType:   string(lb.Type),
			SecurityGroups:     securityGroups,
			Subnets:            subnets,
			EnableDeletionProtection: false, // Default value, would need separate API call
			IdleTimeout:        60, // Default value
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverAutoScalingResources discovers Auto Scaling Group resources
func (c *Client) discoverAutoScalingResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.asgClient.DescribeAutoScalingGroups(context.TODO(), &autoscaling.DescribeAutoScalingGroupsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe Auto Scaling Groups: %w", err)
	}

	for _, asg := range result.AutoScalingGroups {
		// Convert tags
		tags := make(map[string]string)
		if asg.Tags != nil {
			for _, tag := range asg.Tags {
				if tag.Key != nil && tag.Value != nil {
					tags[*tag.Key] = *tag.Value
				}
			}
		}

		// Convert VPC zone identifier
		var vpcZoneIdentifier []string
		if asg.VPCZoneIdentifier != nil {
			vpcZoneIdentifier = strings.Split(*asg.VPCZoneIdentifier, ",")
		}

		// Convert target group ARNs
		var targetGroupARNs []string
		if asg.TargetGroupARNs != nil {
			targetGroupARNs = asg.TargetGroupARNs
		}

		// Convert load balancer names
		var loadBalancerNames []string
		if asg.LoadBalancerNames != nil {
			loadBalancerNames = asg.LoadBalancerNames
		}

		// Create launch template specification if present
		var launchTemplate *LaunchTemplateSpecification
		if asg.LaunchTemplate != nil {
			launchTemplate = &LaunchTemplateSpecification{
				ID:      aws.ToString(asg.LaunchTemplate.LaunchTemplateId),
				Name:    aws.ToString(asg.LaunchTemplate.LaunchTemplateName),
				Version: aws.ToString(asg.LaunchTemplate.Version),
			}
		}

		// Create mixed instances policy if present
		var mixedInstancesPolicy *MixedInstancesPolicy
		if asg.MixedInstancesPolicy != nil {
			var instancesDistribution *InstancesDistribution
			if asg.MixedInstancesPolicy.InstancesDistribution != nil {
				instancesDistribution = &InstancesDistribution{
					OnDemandBaseCapacity:                aws.ToInt32(asg.MixedInstancesPolicy.InstancesDistribution.OnDemandBaseCapacity),
					OnDemandPercentageAboveBaseCapacity: aws.ToInt32(asg.MixedInstancesPolicy.InstancesDistribution.OnDemandPercentageAboveBaseCapacity),
					SpotAllocationStrategy:              aws.ToString(asg.MixedInstancesPolicy.InstancesDistribution.SpotAllocationStrategy),
					SpotInstancePools:                   aws.ToInt32(asg.MixedInstancesPolicy.InstancesDistribution.SpotInstancePools),
					SpotMaxPrice:                        aws.ToString(asg.MixedInstancesPolicy.InstancesDistribution.SpotMaxPrice),
				}
			}

			var policyLaunchTemplate *LaunchTemplateSpecification
			if asg.MixedInstancesPolicy.LaunchTemplate != nil && asg.MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification != nil {
				policyLaunchTemplate = &LaunchTemplateSpecification{
					ID:      aws.ToString(asg.MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification.LaunchTemplateId),
					Name:    aws.ToString(asg.MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification.LaunchTemplateName),
					Version: aws.ToString(asg.MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification.Version),
				}
			}

			mixedInstancesPolicy = &MixedInstancesPolicy{
				LaunchTemplate:       policyLaunchTemplate,
				InstancesDistribution: instancesDistribution,
			}
		}

		resource := &ASGResource{
			BaseResource: BaseResource{
				Type:   "asg",
				ID:     *asg.AutoScalingGroupName,
				Name:   *asg.AutoScalingGroupName,
				Region: c.region,
				Tags:   tags,
			},
			MaxSize:                aws.ToInt32(asg.MaxSize),
			MinSize:                aws.ToInt32(asg.MinSize),
			DesiredCapacity:        aws.ToInt32(asg.DesiredCapacity),
			HealthCheckType:        aws.ToString(asg.HealthCheckType),
			HealthCheckGracePeriod: aws.ToInt32(asg.HealthCheckGracePeriod),
			VPCZoneIdentifier:      vpcZoneIdentifier,
			LaunchTemplate:         launchTemplate,
			MixedInstancesPolicy:   mixedInstancesPolicy,
			TargetGroupARNs:        targetGroupARNs,
			LoadBalancerNames:      loadBalancerNames,
			ServiceLinkedRoleARN:   aws.ToString(asg.ServiceLinkedRoleARN),
			MaxInstanceLifetime:    aws.ToInt32(asg.MaxInstanceLifetime),
			CapacityRebalance:      aws.ToBool(asg.CapacityRebalance),
			ProtectFromScaleIn:     aws.ToBool(asg.NewInstancesProtectedFromScaleIn),
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverLambdaResources discovers Lambda function resources
func (c *Client) discoverLambdaResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.lambdaClient.ListFunctions(context.TODO(), &lambda.ListFunctionsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list Lambda functions: %w", err)
	}

	for _, function := range result.Functions {
		// Convert tags - Lambda tags would need separate API call
		tags := make(map[string]string)

		resource := &LambdaFunctionResource{
			BaseResource: BaseResource{
				Type:   "lambda",
				ID:     *function.FunctionName,
				Name:   *function.FunctionName,
				Region: c.region,
				Tags:   tags,
			},
			Runtime:    string(function.Runtime),
			Handler:    *function.Handler,
			Role:       *function.Role,
			Code:       "lambda_function.zip", // Placeholder
			Timeout:    int(*function.Timeout),
			MemorySize: int(*function.MemorySize),
		}

		// Add environment variables if present
		if function.Environment != nil && function.Environment.Variables != nil {
			resource.Environment = function.Environment.Variables
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverSQSResources discovers SQS queue resources
func (c *Client) discoverSQSResources() ([]Resource, error) {
	// TODO: Implement SQS resource discovery
	return []Resource{}, nil
}

// discoverSNSResources discovers SNS topic resources
func (c *Client) discoverSNSResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.snsClient.ListTopics(context.TODO(), &sns.ListTopicsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list SNS topics: %w", err)
	}

	for _, topic := range result.Topics {
		// Skip topics with invalid ARNs
		if topic.TopicArn == nil || *topic.TopicArn == "" {
			continue
		}

		// Validate ARN format
		if !isValidSNSArn(*topic.TopicArn) {
			continue
		}

		// Extract topic name from ARN
		parts := strings.Split(*topic.TopicArn, ":")
		if len(parts) < 6 {
			continue
		}
		topicName := parts[len(parts)-1]

		// Skip empty topic names
		if topicName == "" {
			continue
		}

		// Additional validation: ensure topic name is valid
		if !isValidSNSTopicName(topicName) {
			continue
		}

		// Get topic attributes
		attributes, err := c.snsClient.GetTopicAttributes(context.TODO(), &sns.GetTopicAttributesInput{
			TopicArn: topic.TopicArn,
		})
		if err != nil {
			continue
		}

		resource := &SNSResource{
			BaseResource: BaseResource{
				Type:   "sns",
				ID:     topicName,
				Name:   topicName,
				Region: c.region,
				Tags:   make(map[string]string), // SNS tags would need separate call
			},
			ARN:              *topic.TopicArn,
			KMSMasterKeyID:   attributes.Attributes["KmsMasterKeyId"],
			DeliveryPolicy:   attributes.Attributes["DeliveryPolicy"],
			Policy:           attributes.Attributes["Policy"],
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// isValidSNSArn validates SNS ARN format
func isValidSNSArn(arn string) bool {
	if arn == "" {
		return false
	}
	
	// Check if it starts with arn:aws:sns:
	if !strings.HasPrefix(arn, "arn:aws:sns:") {
		return false
	}
	
	// Split by : and check we have at least 6 parts
	parts := strings.Split(arn, ":")
	if len(parts) < 6 {
		return false
	}
	
	// Check format: arn:aws:sns:region:account:topic-name
	if parts[0] != "arn" || parts[1] != "aws" || parts[2] != "sns" {
		return false
	}
	
	// Check that topic name is not empty
	if parts[len(parts)-1] == "" {
		return false
	}
	
	return true
}

// isValidSNSTopicName validates SNS topic name format
func isValidSNSTopicName(name string) bool {
	if name == "" {
		return false
	}
	
	// Check length (1-256 characters)
	if len(name) < 1 || len(name) > 256 {
		return false
	}
	
	// Check pattern: [a-zA-Z0-9_-]+
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, name)
	return matched
}

// discoverCloudWatchResources discovers CloudWatch log group resources
func (c *Client) discoverCloudWatchResources() ([]Resource, error) {
	// TODO: Implement CloudWatch log group discovery
	return []Resource{}, nil
}

// discoverCloudTrailResources discovers CloudTrail resources
func (c *Client) discoverCloudTrailResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.cloudtrailClient.ListTrails(context.TODO(), &cloudtrail.ListTrailsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list CloudTrail trails: %w", err)
	}

	for _, trail := range result.Trails {
		// Get trail details
		trailDetails, err := c.cloudtrailClient.DescribeTrails(context.TODO(), &cloudtrail.DescribeTrailsInput{
			TrailNameList: []string{*trail.Name},
		})
		if err != nil || len(trailDetails.TrailList) == 0 {
			continue
		}

		trailInfo := trailDetails.TrailList[0]

		// Get event selectors
		var eventSelectors []*CloudTrailEventSelector
		if trailInfo.HasCustomEventSelectors != nil && *trailInfo.HasCustomEventSelectors {
			eventSelectorsResult, err := c.cloudtrailClient.GetEventSelectors(context.TODO(), &cloudtrail.GetEventSelectorsInput{
				TrailName: trail.Name,
			})
			if err == nil && len(eventSelectorsResult.EventSelectors) > 0 {
				for _, selector := range eventSelectorsResult.EventSelectors {
					eventSelector := &CloudTrailEventSelector{
						ReadWriteType:           string(selector.ReadWriteType),
						IncludeManagementEvents: aws.ToBool(selector.IncludeManagementEvents),
					}

					// Add data resources if present
					if len(selector.DataResources) > 0 {
						for _, dataResource := range selector.DataResources {
							eventSelector.DataResources = append(eventSelector.DataResources, &CloudTrailDataResource{
								Type:   aws.ToString(dataResource.Type),
								Values: dataResource.Values,
							})
						}
					}

					// Add exclude management event sources if present
					if len(selector.ExcludeManagementEventSources) > 0 {
						eventSelector.ExcludeManagementEventSources = selector.ExcludeManagementEventSources
					}

					eventSelectors = append(eventSelectors, eventSelector)
				}
			}
		}

		// Get insight selectors
		var insightSelectors []*CloudTrailInsightSelector
		insightSelectorsResult, err := c.cloudtrailClient.GetInsightSelectors(context.TODO(), &cloudtrail.GetInsightSelectorsInput{
			TrailName: trail.Name,
		})
		if err == nil && len(insightSelectorsResult.InsightSelectors) > 0 {
			for _, selector := range insightSelectorsResult.InsightSelectors {
				insightSelectors = append(insightSelectors, &CloudTrailInsightSelector{
					InsightType: string(selector.InsightType),
				})
			}
		}

		resource := &CloudTrailResource{
			BaseResource: BaseResource{
				Type:   "cloudtrail",
				ID:     *trail.Name,
				Name:   *trail.Name,
				Region: c.region,
				Tags:   make(map[string]string), // CloudTrail tags would need separate call
			},
			Name:                    *trail.Name,
			S3BucketName:            aws.ToString(trailInfo.S3BucketName),
			S3KeyPrefix:             aws.ToString(trailInfo.S3KeyPrefix),
			CloudWatchLogGroupARN:   aws.ToString(trailInfo.CloudWatchLogsLogGroupArn),
			CloudWatchLogsRoleARN:   aws.ToString(trailInfo.CloudWatchLogsRoleArn),
			IncludeGlobalServiceEvents: aws.ToBool(trailInfo.IncludeGlobalServiceEvents),
			IsMultiRegionTrail:      aws.ToBool(trailInfo.IsMultiRegionTrail),
			EnableLogFileValidation: aws.ToBool(trailInfo.LogFileValidationEnabled),
			KMSKeyID:                aws.ToString(trailInfo.KmsKeyId),
			EventSelectors:          eventSelectors,
			InsightSelectors:        insightSelectors,
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverECSResources discovers ECS resources
func (c *Client) discoverECSResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.ecsClient.ListServices(context.TODO(), &ecs.ListServicesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list ECS services: %w", err)
	}

	for _, serviceARN := range result.ServiceArns {
		// Get service details
		serviceDetails, err := c.ecsClient.DescribeServices(context.TODO(), &ecs.DescribeServicesInput{
			Services: []string{serviceARN},
		})
		if err != nil || len(serviceDetails.Services) == 0 {
			continue
		}

		service := serviceDetails.Services[0]

		// Build network configuration if present
		var networkConfig *NetworkConfiguration
		if service.NetworkConfiguration != nil {
			networkConfig = &NetworkConfiguration{
				Subnets:        service.NetworkConfiguration.AwsvpcConfiguration.Subnets,
				SecurityGroups: service.NetworkConfiguration.AwsvpcConfiguration.SecurityGroups,
				AssignPublicIP: service.NetworkConfiguration.AwsvpcConfiguration.AssignPublicIp == "ENABLED",
			}
		}

		// Build load balancers if present
		var loadBalancers []*LoadBalancer
		for _, lb := range service.LoadBalancers {
			loadBalancers = append(loadBalancers, &LoadBalancer{
				TargetGroupARN:   aws.ToString(lb.TargetGroupArn),
				ContainerName:    aws.ToString(lb.ContainerName),
				ContainerPort:    int32(aws.ToInt32(lb.ContainerPort)),
			})
		}

		// Build service registries if present
		var serviceRegistries []*ServiceRegistry
		for _, registry := range service.ServiceRegistries {
			serviceRegistries = append(serviceRegistries, &ServiceRegistry{
				RegistryARN:   aws.ToString(registry.RegistryArn),
				Port:          int32(aws.ToInt32(registry.Port)),
				ContainerName: aws.ToString(registry.ContainerName),
				ContainerPort: int32(aws.ToInt32(registry.ContainerPort)),
			})
		}

		// Build deployment configuration if present
		var deploymentConfig *DeploymentConfiguration
		if service.DeploymentConfiguration != nil {
			deploymentConfig = &DeploymentConfiguration{
				MaximumPercent:        int32(aws.ToInt32(service.DeploymentConfiguration.MaximumPercent)),
				MinimumHealthyPercent: int32(aws.ToInt32(service.DeploymentConfiguration.MinimumHealthyPercent)),
			}

			if service.DeploymentConfiguration.DeploymentCircuitBreaker != nil {
				deploymentConfig.DeploymentCircuitBreaker = &DeploymentCircuitBreaker{
					Enable:   service.DeploymentConfiguration.DeploymentCircuitBreaker.Enable,
					Rollback: service.DeploymentConfiguration.DeploymentCircuitBreaker.Rollback,
				}
			}
		}

		resource := &ECSResource{
			BaseResource: BaseResource{
				Type:   "ecs",
				ID:     *service.ServiceName,
				Name:   *service.ServiceName,
				Region: c.region,
				Tags:   make(map[string]string), // ECS tags would need separate call
			},
			ClusterARN:        aws.ToString(service.ClusterArn),
			TaskDefinitionARN: aws.ToString(service.TaskDefinition),
			DesiredCount:      service.DesiredCount,
			LaunchType:        string(service.LaunchType),
			PlatformVersion:   aws.ToString(service.PlatformVersion),
			NetworkConfiguration: networkConfig,
			LoadBalancers:      loadBalancers,
			ServiceRegistries:  serviceRegistries,
			DeploymentConfiguration: deploymentConfig,
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverEKSResources discovers EKS cluster resources
func (c *Client) discoverEKSResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.eksClient.ListClusters(context.TODO(), &eks.ListClustersInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list EKS clusters: %w", err)
	}

	for _, clusterName := range result.Clusters {
		// Get cluster details
		clusterDetails, err := c.eksClient.DescribeCluster(context.TODO(), &eks.DescribeClusterInput{
			Name: &clusterName,
		})
		if err != nil {
			continue
		}

		cluster := clusterDetails.Cluster

		// Extract VPC config
		var vpcConfig *VPCConfig
		if cluster.ResourcesVpcConfig != nil {
			vpcConfig = &VPCConfig{
				SubnetIDs:             cluster.ResourcesVpcConfig.SubnetIds,
				SecurityGroupIDs:      cluster.ResourcesVpcConfig.SecurityGroupIds,
				EndpointPrivateAccess: cluster.ResourcesVpcConfig.EndpointPrivateAccess,
				EndpointPublicAccess:  cluster.ResourcesVpcConfig.EndpointPublicAccess,
				PublicAccessCIDRs:     cluster.ResourcesVpcConfig.PublicAccessCidrs,
			}
		}

		// Build encryption config if present
		var encryptionConfig *EncryptionConfig
		if cluster.EncryptionConfig != nil && len(cluster.EncryptionConfig) > 0 {
			enc := cluster.EncryptionConfig[0]
			encryptionConfig = &EncryptionConfig{
				Provider: &EncryptionProvider{
					KeyARN: aws.ToString(enc.Provider.KeyArn),
				},
				Resources: enc.Resources,
			}
		}

		// Build kubernetes network config if present
		// TODO: Implement proper kubernetes network config handling

		resource := &EKSResource{
			BaseResource: BaseResource{
				Type:   "eks",
				ID:     *cluster.Name,
				Name:   *cluster.Name,
				Region: c.region,
				Tags:   cluster.Tags,
			},
			Version:          aws.ToString(cluster.Version),
			RoleARN:          aws.ToString(cluster.RoleArn),
			PlatformVersion:  aws.ToString(cluster.PlatformVersion),
			Endpoint:         aws.ToString(cluster.Endpoint),
			VPCConfig:        vpcConfig,
			EncryptionConfig: encryptionConfig,
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverElastiCacheResources discovers ElastiCache cluster resources
func (c *Client) discoverElastiCacheResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.elasticacheClient.DescribeCacheClusters(context.TODO(), &elasticache.DescribeCacheClustersInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list ElastiCache clusters: %w", err)
	}

	for _, cluster := range result.CacheClusters {
		// Skip clusters with invalid identifiers
		if cluster.CacheClusterId == nil || !isValidElastiCacheClusterId(*cluster.CacheClusterId) {
			continue
		}
		
		// Get tags
		tags, err := c.elasticacheClient.ListTagsForResource(context.TODO(), &elasticache.ListTagsForResourceInput{
			ResourceName: cluster.ARN,
		})
		if err != nil {
			continue
		}

		tagMap := make(map[string]string)
		for _, tag := range tags.TagList {
			if tag.Key != nil && tag.Value != nil {
				tagMap[*tag.Key] = *tag.Value
			}
		}

		// Convert security group IDs
		var securityGroupIDs []string
		if cluster.SecurityGroups != nil {
			for _, sg := range cluster.SecurityGroups {
				if sg.SecurityGroupId != nil {
					securityGroupIDs = append(securityGroupIDs, *sg.SecurityGroupId)
				}
			}
		}

		resource := &ElastiCacheResource{
			BaseResource: BaseResource{
				Type:   "elasticache",
				ID:     *cluster.CacheClusterId,
				Name:   *cluster.CacheClusterId,
				Region: c.region,
				Tags:   tagMap,
			},
			Engine:               aws.ToString(cluster.Engine),
			NodeType:             aws.ToString(cluster.CacheNodeType),
			NumCacheNodes:        aws.ToInt32(cluster.NumCacheNodes),
			Port:                 6379, // Default Redis port
			SubnetGroupName:      aws.ToString(cluster.CacheSubnetGroupName),
			SecurityGroupIDs:     securityGroupIDs,
			ParameterGroupName:   aws.ToString(cluster.CacheParameterGroup.CacheParameterGroupName),
			EngineVersion:        aws.ToString(cluster.EngineVersion),
			MultiAZEnabled:       false, // Default value
			AutomaticFailoverEnabled: false, // Default value
			AtRestEncryptionEnabled: false, // Default value
			TransitEncryptionEnabled: false, // Default value
			KMSKeyID:             "",
			SnapshotRetentionLimit: 0,
			SnapshotWindow:       "",
			MaintenanceWindow:    "",
			NotificationTopicARN: "",
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// isValidElastiCacheClusterId validates ElastiCache cluster ID format
func isValidElastiCacheClusterId(id string) bool {
	if id == "" {
		return false
	}
	
	// Check length (1-50 characters)
	if len(id) < 1 || len(id) > 50 {
		return false
	}
	
	// Check pattern: must begin with a letter, contain only ASCII letters, digits, and hyphens
	// and must not end with a hyphen or contain two consecutive hyphens
	matched, _ := regexp.MatchString(`^[a-zA-Z][a-zA-Z0-9-]*[a-zA-Z0-9]$`, id)
	return matched
}

// discoverDynamoDBResources discovers DynamoDB table resources
func (c *Client) discoverDynamoDBResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.dynamodbClient.ListTables(context.TODO(), &dynamodb.ListTablesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list DynamoDB tables: %w", err)
	}

	for _, tableName := range result.TableNames {
		// Get table details
		tableDetails, err := c.dynamodbClient.DescribeTable(context.TODO(), &dynamodb.DescribeTableInput{
			TableName: &tableName,
		})
		if err != nil {
			continue
		}

		table := tableDetails.Table

		// Get tags
		tags, err := c.dynamodbClient.ListTagsOfResource(context.TODO(), &dynamodb.ListTagsOfResourceInput{
			ResourceArn: table.TableArn,
		})
		if err != nil {
			continue
		}

		tagMap := make(map[string]string)
		for _, tag := range tags.Tags {
			if tag.Key != nil && tag.Value != nil {
				tagMap[*tag.Key] = *tag.Value
			}
		}

		// Extract billing mode
		billingMode := "PAY_PER_REQUEST"
		if table.BillingModeSummary != nil {
			billingMode = string(table.BillingModeSummary.BillingMode)
		}

		// Extract capacity settings
		var readCapacity, writeCapacity int64
		if table.ProvisionedThroughput != nil {
			if table.ProvisionedThroughput.ReadCapacityUnits != nil {
				readCapacity = *table.ProvisionedThroughput.ReadCapacityUnits
			}
			if table.ProvisionedThroughput.WriteCapacityUnits != nil {
				writeCapacity = *table.ProvisionedThroughput.WriteCapacityUnits
			}
		}

		// Extract hash key
		var hashKey string
		if len(table.KeySchema) > 0 && table.KeySchema[0].AttributeName != nil {
			hashKey = *table.KeySchema[0].AttributeName
		}

		// Extract range key if present
		var rangeKey string
		if len(table.KeySchema) > 1 && table.KeySchema[1].AttributeName != nil {
			rangeKey = *table.KeySchema[1].AttributeName
		}

		// Extract stream settings
		var streamEnabled bool
		var streamViewType string
		if table.StreamSpecification != nil {
			streamEnabled = *table.StreamSpecification.StreamEnabled
			streamViewType = string(table.StreamSpecification.StreamViewType)
		}

		// Extract server-side encryption
		var serverSideEncryption *ServerSideEncryption
		if table.SSEDescription != nil {
			serverSideEncryption = &ServerSideEncryption{
				Enabled:   true,
				KMSKeyARN: aws.ToString(table.SSEDescription.KMSMasterKeyArn),
			}
		}

		// Extract point-in-time recovery
		var pointInTimeRecovery *PointInTimeRecovery
		// PointInTimeRecoveryDescription is not available in the current AWS SDK version
		// Set default value for now
		pointInTimeRecovery = &PointInTimeRecovery{
			Enabled: false,
		}

		resource := &DynamoDBResource{
			BaseResource: BaseResource{
				Type:   "dynamodb",
				ID:     *table.TableName,
				Name:   *table.TableName,
				Region: c.region,
				Tags:   tagMap,
			},
			BillingMode:          billingMode,
			ReadCapacity:         readCapacity,
			WriteCapacity:        writeCapacity,
			HashKey:              hashKey,
			RangeKey:             rangeKey,
			StreamEnabled:        streamEnabled,
			StreamViewType:       streamViewType,
			ServerSideEncryption: serverSideEncryption,
			PointInTimeRecovery:  pointInTimeRecovery,
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverRedshiftResources discovers Redshift cluster resources
func (c *Client) discoverRedshiftResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.redshiftClient.DescribeClusters(context.TODO(), &redshift.DescribeClustersInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list Redshift clusters: %w", err)
	}

	for _, cluster := range result.Clusters {
		// Redshift doesn't have a direct ListTagsForResource method, skip tags for now
		tagMap := make(map[string]string)

		// Extract security group IDs
		var securityGroupIDs []string
		for _, sg := range cluster.VpcSecurityGroups {
			if sg.VpcSecurityGroupId != nil {
				securityGroupIDs = append(securityGroupIDs, *sg.VpcSecurityGroupId)
			}
		}

		// Extract endpoint port
		var port int32
		if cluster.Endpoint != nil && cluster.Endpoint.Port != nil {
			port = *cluster.Endpoint.Port
		}

		resource := &RedshiftResource{
			BaseResource: BaseResource{
				Type:   "redshift",
				ID:     *cluster.ClusterIdentifier,
				Name:   *cluster.ClusterIdentifier,
				Region: c.region,
				Tags:   tagMap,
			},
			ClusterIdentifier:     *cluster.ClusterIdentifier,
			NodeType:              *cluster.NodeType,
			NumberOfNodes:         *cluster.NumberOfNodes,
			MasterUsername:        *cluster.MasterUsername,
			Port:                  port,
			VpcSecurityGroupIds:   securityGroupIDs,
			ClusterSubnetGroupName: *cluster.ClusterSubnetGroupName,
			Encrypted:             *cluster.Encrypted,
			PubliclyAccessible:    *cluster.PubliclyAccessible,
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverRoute53Resources discovers Route53 hosted zones and records
func (c *Client) discoverRoute53Resources() ([]Resource, error) {
	var resources []Resource

	// Discover hosted zones
	zonesResult, err := c.route53Client.ListHostedZones(context.TODO(), &route53.ListHostedZonesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list Route53 hosted zones: %w", err)
	}

	for _, zone := range zonesResult.HostedZones {
		// Get tags
		tags, err := c.route53Client.ListTagsForResource(context.TODO(), &route53.ListTagsForResourceInput{
			ResourceType: "hostedzone",
			ResourceId:   zone.Id,
		})
		if err != nil {
			continue
		}

		tagMap := make(map[string]string)
		for _, tag := range tags.ResourceTagSet.Tags {
			if tag.Key != nil && tag.Value != nil {
				tagMap[*tag.Key] = *tag.Value
			}
		}

		// Extract comment if present
		var comment string
		if zone.Config != nil && zone.Config.Comment != nil {
			comment = *zone.Config.Comment
		}

		resource := &Route53Resource{
			BaseResource: BaseResource{
				Type:   "route53",
				ID:     *zone.Id,
				Name:   *zone.Name,
				Region: c.region,
				Tags:   tagMap,
			},
			ZoneID:  *zone.Id,
			Name:    *zone.Name,
			Comment: comment,
		}

		resources = append(resources, resource)

		// Discover records for this zone
		recordsResult, err := c.route53Client.ListResourceRecordSets(context.TODO(), &route53.ListResourceRecordSetsInput{
			HostedZoneId: zone.Id,
		})
		if err != nil {
			continue
		}

		for _, record := range recordsResult.ResourceRecordSets {
			// Extract records if present
			var records []string
			if len(record.ResourceRecords) > 0 {
				for _, rr := range record.ResourceRecords {
					if rr.Value != nil {
						records = append(records, *rr.Value)
					}
				}
			}

			// Extract TTL if present
			var ttl int64
			if record.TTL != nil {
				ttl = *record.TTL
			}

			recordResource := &Route53Record{
				Name:    *record.Name,
				Type:    string(record.Type),
				TTL:     ttl,
				Records: records,
			}

			// Add the record to the zone's records
			resource.Records = append(resource.Records, recordResource)
		}
	}

	return resources, nil
}

// discoverCloudFrontResources discovers CloudFront distribution resources
func (c *Client) discoverCloudFrontResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.cloudfrontClient.ListDistributions(context.TODO(), &cloudfront.ListDistributionsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list CloudFront distributions: %w", err)
	}

	for _, distribution := range result.DistributionList.Items {
		// Get tags
		tags, err := c.cloudfrontClient.ListTagsForResource(context.TODO(), &cloudfront.ListTagsForResourceInput{
			Resource: distribution.ARN,
		})
		if err != nil {
			continue
		}

		tagMap := make(map[string]string)
		for _, tag := range tags.Tags.Items {
			if tag.Key != nil && tag.Value != nil {
				tagMap[*tag.Key] = *tag.Value
			}
		}

		// Extract origins
		var origins []*Origin
		for _, origin := range distribution.Origins.Items {
			// Skip if required fields are nil
			if origin.DomainName == nil || origin.Id == nil {
				continue
			}
			
			originConfig := &Origin{
				DomainName: *origin.DomainName,
				OriginID:   *origin.Id,
			}
			
			// Add origin path if present
			if origin.OriginPath != nil {
				originConfig.OriginPath = *origin.OriginPath
			}
			
			// Add connection attempts if present
			if origin.ConnectionAttempts != nil {
				originConfig.ConnectionAttempts = *origin.ConnectionAttempts
			}
			
			// Add connection timeout if present
			if origin.ConnectionTimeout != nil {
				originConfig.ConnectionTimeout = *origin.ConnectionTimeout
			}
			
			// Add custom origin config if present
			if origin.CustomOriginConfig != nil {
				originConfig.CustomOriginConfig = &CustomOriginConfig{
					HTTPPort:             *origin.CustomOriginConfig.HTTPPort,
					HTTPSPort:            *origin.CustomOriginConfig.HTTPSPort,
					OriginProtocolPolicy: string(origin.CustomOriginConfig.OriginProtocolPolicy),
				}
				
				// Add SSL protocols if present
				if len(origin.CustomOriginConfig.OriginSslProtocols.Items) > 0 {
					for _, protocol := range origin.CustomOriginConfig.OriginSslProtocols.Items {
						originConfig.CustomOriginConfig.OriginSSLProtocols = append(originConfig.CustomOriginConfig.OriginSSLProtocols, string(protocol))
					}
				}
				
				// Add read timeout if present
				if origin.CustomOriginConfig.OriginReadTimeout != nil {
					originConfig.CustomOriginConfig.OriginReadTimeout = *origin.CustomOriginConfig.OriginReadTimeout
				}
				
				// Add keepalive timeout if present
				if origin.CustomOriginConfig.OriginKeepaliveTimeout != nil {
					originConfig.CustomOriginConfig.OriginKeepaliveTimeout = *origin.CustomOriginConfig.OriginKeepaliveTimeout
				}
			}
			
			// Add S3 origin config if present
			if origin.S3OriginConfig != nil && origin.S3OriginConfig.OriginAccessIdentity != nil {
				originConfig.S3OriginConfig = &S3OriginConfig{
					OriginAccessIdentity: *origin.S3OriginConfig.OriginAccessIdentity,
				}
			}
			
			// Add custom headers if present
			if origin.CustomHeaders != nil && len(origin.CustomHeaders.Items) > 0 {
				for _, header := range origin.CustomHeaders.Items {
					if header.HeaderName != nil && header.HeaderValue != nil {
						originConfig.CustomHeaders = append(originConfig.CustomHeaders, &CustomHeader{
							Name:  *header.HeaderName,
							Value: *header.HeaderValue,
						})
					}
				}
			}
			
			// Add origin shield if present
			if origin.OriginShield != nil && origin.OriginShield.Enabled != nil && origin.OriginShield.OriginShieldRegion != nil {
				originConfig.OriginShield = &OriginShield{
					Enabled:            *origin.OriginShield.Enabled,
					OriginShieldRegion: *origin.OriginShield.OriginShieldRegion,
				}
			}
			
			origins = append(origins, originConfig)
		}

		// Extract default cache behavior
		var defaultCacheBehavior *CacheBehavior
		if distribution.DefaultCacheBehavior != nil {
			// Skip if required fields are nil
			if distribution.DefaultCacheBehavior.TargetOriginId == nil || distribution.DefaultCacheBehavior.Compress == nil {
				continue
			}
			
			defaultCacheBehavior = &CacheBehavior{
				TargetOriginID:        *distribution.DefaultCacheBehavior.TargetOriginId,
				ViewerProtocolPolicy:  string(distribution.DefaultCacheBehavior.ViewerProtocolPolicy),
				Compress:              *distribution.DefaultCacheBehavior.Compress,
			}
			
			// Add allowed methods
			if distribution.DefaultCacheBehavior.AllowedMethods != nil && len(distribution.DefaultCacheBehavior.AllowedMethods.Items) > 0 {
				for _, method := range distribution.DefaultCacheBehavior.AllowedMethods.Items {
					defaultCacheBehavior.AllowedMethods = append(defaultCacheBehavior.AllowedMethods, string(method))
				}
			}
			
			// Note: CachedMethods not available in DistributionSummary
			
			// Add cache policy if present
			if distribution.DefaultCacheBehavior.CachePolicyId != nil {
				defaultCacheBehavior.CachePolicyID = *distribution.DefaultCacheBehavior.CachePolicyId
			}
			
			// Add origin request policy if present
			if distribution.DefaultCacheBehavior.OriginRequestPolicyId != nil {
				defaultCacheBehavior.OriginRequestPolicyID = *distribution.DefaultCacheBehavior.OriginRequestPolicyId
			}
			
			// Add response headers policy if present
			if distribution.DefaultCacheBehavior.ResponseHeadersPolicyId != nil {
				defaultCacheBehavior.ResponseHeadersPolicyID = *distribution.DefaultCacheBehavior.ResponseHeadersPolicyId
			}
			
			// Add TTL settings if present
			if distribution.DefaultCacheBehavior.DefaultTTL != nil {
				defaultCacheBehavior.DefaultTTL = *distribution.DefaultCacheBehavior.DefaultTTL
			}
			if distribution.DefaultCacheBehavior.MaxTTL != nil {
				defaultCacheBehavior.MaxTTL = *distribution.DefaultCacheBehavior.MaxTTL
			}
			if distribution.DefaultCacheBehavior.MinTTL != nil {
				defaultCacheBehavior.MinTTL = *distribution.DefaultCacheBehavior.MinTTL
			}
		}

		// Extract viewer certificate
		var viewerCertificate *ViewerCertificate
		if distribution.ViewerCertificate != nil {
			// Skip if required fields are nil
			if distribution.ViewerCertificate.CloudFrontDefaultCertificate == nil {
				continue
			}
			
			viewerCertificate = &ViewerCertificate{
				CloudFrontDefaultCertificate: *distribution.ViewerCertificate.CloudFrontDefaultCertificate,
				MinimumProtocolVersion:       string(distribution.ViewerCertificate.MinimumProtocolVersion),
				SSLSupportMethod:             string(distribution.ViewerCertificate.SSLSupportMethod),
			}
			
			// Add ACM certificate ARN if present
			if distribution.ViewerCertificate.ACMCertificateArn != nil {
				viewerCertificate.ACMCertificateARN = *distribution.ViewerCertificate.ACMCertificateArn
			}
		}

		// Extract aliases
		var aliases []string
		if distribution.Aliases != nil && len(distribution.Aliases.Items) > 0 {
			for _, alias := range distribution.Aliases.Items {
				aliases = append(aliases, alias)
			}
		}

		// Skip if required fields are nil
		if distribution.Id == nil || distribution.DomainName == nil || distribution.Enabled == nil || distribution.IsIPV6Enabled == nil {
			continue
		}
		
		resource := &CloudFrontResource{
			BaseResource: BaseResource{
				Type:   "cloudfront",
				ID:     *distribution.Id,
				Name:   *distribution.DomainName,
				Region: c.region,
				Tags:   tagMap,
			},
			Enabled:           *distribution.Enabled,
			PriceClass:        string(distribution.PriceClass),
			IsIPV6Enabled:     *distribution.IsIPV6Enabled,
			HttpVersion:       string(distribution.HttpVersion),
			Origins:           origins,
			DefaultCacheBehavior: defaultCacheBehavior,
			ViewerCertificate:    viewerCertificate,
			Aliases:              aliases,
		}

		// Note: DefaultRootObject not available in DistributionSummary

		// Add web ACL ID if present
		if distribution.WebACLId != nil {
			resource.WebACLID = *distribution.WebACLId
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverAPIGatewayResources discovers API Gateway resources
func (c *Client) discoverAPIGatewayResources() ([]Resource, error) {
	var resources []Resource

	// Discover HTTP APIs
	apisResult, err := c.apigatewayClient.GetApis(context.TODO(), &apigatewayv2.GetApisInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list API Gateway APIs: %w", err)
	}

	for _, api := range apisResult.Items {
		// Get tags
		tags, err := c.apigatewayClient.GetTags(context.TODO(), &apigatewayv2.GetTagsInput{
			ResourceArn: api.ApiId,
		})
		if err != nil {
			continue
		}

		tagMap := make(map[string]string)
		for key, value := range tags.Tags {
			tagMap[key] = value
		}

		resource := &APIGatewayResource{
			BaseResource: BaseResource{
				Type:   "apigateway",
				ID:     *api.ApiId,
				Name:   *api.Name,
				Region: c.region,
				Tags:   tagMap,
			},
			APIID:        *api.ApiId,
			Name:         *api.Name,
			ProtocolType: string(api.ProtocolType),
			Version:      *api.Version,
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverElasticsearchResources discovers OpenSearch/Elasticsearch domain resources
func (c *Client) discoverElasticsearchResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.elasticsearchClient.ListDomainNames(context.TODO(), &elasticsearchservice.ListDomainNamesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list Elasticsearch domains: %w", err)
	}

	for _, domainInfo := range result.DomainNames {
		// Get domain details
		domainDetails, err := c.elasticsearchClient.DescribeElasticsearchDomain(context.TODO(), &elasticsearchservice.DescribeElasticsearchDomainInput{
			DomainName: domainInfo.DomainName,
		})
		if err != nil {
			continue
		}

		domain := domainDetails.DomainStatus

		// Get tags
		tags, err := c.elasticsearchClient.ListTags(context.TODO(), &elasticsearchservice.ListTagsInput{
			ARN: domain.ARN,
		})
		if err != nil {
			continue
		}

		tagMap := make(map[string]string)
		for _, tag := range tags.TagList {
			if tag.Key != nil && tag.Value != nil {
				tagMap[*tag.Key] = *tag.Value
			}
		}

		// Build cluster config
		var clusterConfig *ElasticsearchClusterConfig
		if domain.ElasticsearchClusterConfig != nil {
			clusterConfig = &ElasticsearchClusterConfig{
				InstanceType:         string(domain.ElasticsearchClusterConfig.InstanceType),
				InstanceCount:        int32(aws.ToInt32(domain.ElasticsearchClusterConfig.InstanceCount)),
				DedicatedMasterEnabled: aws.ToBool(domain.ElasticsearchClusterConfig.DedicatedMasterEnabled),
				ZoneAwarenessEnabled: domain.ElasticsearchClusterConfig.ZoneAwarenessConfig != nil,
			}
		}

		// Build EBS options
		var ebsOptions *ElasticsearchEBSOptions
		if domain.EBSOptions != nil {
			ebsOptions = &ElasticsearchEBSOptions{
				EBSEnabled: aws.ToBool(domain.EBSOptions.EBSEnabled),
				VolumeType: string(domain.EBSOptions.VolumeType),
				VolumeSize: int32(aws.ToInt32(domain.EBSOptions.VolumeSize)),
			}
		}

		// Build encrypt at rest
		var encryptAtRest *ElasticsearchEncryptAtRest
		if domain.EncryptionAtRestOptions != nil {
			encryptAtRest = &ElasticsearchEncryptAtRest{
				Enabled: aws.ToBool(domain.EncryptionAtRestOptions.Enabled),
			}
		}

		// Build node to node encryption
		var nodeToNodeEncryption *ElasticsearchNodeToNodeEncryption
		if domain.NodeToNodeEncryptionOptions != nil {
			nodeToNodeEncryption = &ElasticsearchNodeToNodeEncryption{
				Enabled: aws.ToBool(domain.NodeToNodeEncryptionOptions.Enabled),
			}
		}

		resource := &ElasticsearchResource{
			BaseResource: BaseResource{
				Type:   "elasticsearch",
				ID:     *domain.DomainName,
				Name:   *domain.DomainName,
				Region: c.region,
				Tags:   tagMap,
			},
			DomainName:           *domain.DomainName,
			ElasticsearchVersion: aws.ToString(domain.ElasticsearchVersion),
			ClusterConfig:        clusterConfig,
			EBSOptions:           ebsOptions,
			EncryptAtRest:        encryptAtRest,
			NodeToNodeEncryption: nodeToNodeEncryption,
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverECRResources discovers ECR repository resources
func (c *Client) discoverECRResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.ecrClient.DescribeRepositories(context.TODO(), &ecr.DescribeRepositoriesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list ECR repositories: %w", err)
	}

	for _, repo := range result.Repositories {
		// Get tags
		tags, err := c.ecrClient.ListTagsForResource(context.TODO(), &ecr.ListTagsForResourceInput{
			ResourceArn: repo.RepositoryArn,
		})
		if err != nil {
			continue
		}

		tagMap := make(map[string]string)
		for _, tag := range tags.Tags {
			if tag.Key != nil && tag.Value != nil {
				tagMap[*tag.Key] = *tag.Value
			}
		}

		// Build image scanning configuration if present
		var imageScanConfig *ImageScanningConfiguration
		if repo.ImageScanningConfiguration != nil {
			imageScanConfig = &ImageScanningConfiguration{
				ScanOnPush: repo.ImageScanningConfiguration.ScanOnPush,
			}
		}

		// Build encryption configuration if present
		var encryptionConfig *EncryptionConfiguration
		if repo.EncryptionConfiguration != nil {
			encryptionConfig = &EncryptionConfiguration{
				EncryptionType: string(repo.EncryptionConfiguration.EncryptionType),
				KMSKey:         aws.ToString(repo.EncryptionConfiguration.KmsKey),
			}
		}

		// Note: Lifecycle policy would need additional API call to retrieve

		resource := &ECRResource{
			BaseResource: BaseResource{
				Type:   "ecr",
				ID:     *repo.RepositoryName,
				Name:   *repo.RepositoryName,
				Region: c.region,
				Tags:   tagMap,
			},
			RepositoryURI:              aws.ToString(repo.RepositoryUri),
			ImageTagMutability:         string(repo.ImageTagMutability),
			ScanOnPush:                 repo.ImageScanningConfiguration != nil && repo.ImageScanningConfiguration.ScanOnPush,
			ImageScanningConfiguration: imageScanConfig,
			EncryptionConfiguration:    encryptionConfig,
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverNeptuneResources discovers Neptune cluster resources
func (c *Client) discoverNeptuneResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.rdsClient.DescribeDBClusters(context.TODO(), &rds.DescribeDBClustersInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list Neptune clusters: %w", err)
	}

	for _, cluster := range result.DBClusters {
		// Get tags
		tags, err := c.rdsClient.ListTagsForResource(context.TODO(), &rds.ListTagsForResourceInput{
			ResourceName: cluster.DBClusterArn,
		})
		if err != nil {
			continue
		}

		tagMap := make(map[string]string)
		for _, tag := range tags.TagList {
			if tag.Key != nil && tag.Value != nil {
				tagMap[*tag.Key] = *tag.Value
			}
		}

		resource := &NeptuneResource{
			BaseResource: BaseResource{
				Type:   "neptune",
				ID:     *cluster.DBClusterIdentifier,
				Name:   *cluster.DBClusterIdentifier,
				Region: c.region,
				Tags:   tagMap,
			},
			ClusterIdentifier:     *cluster.DBClusterIdentifier,
			Engine:               aws.ToString(cluster.Engine),
			EngineVersion:        aws.ToString(cluster.EngineVersion),
			AvailabilityZones:    cluster.AvailabilityZones,
			BackupRetentionPeriod: int32(aws.ToInt32(cluster.BackupRetentionPeriod)),
			PreferredBackupWindow: aws.ToString(cluster.PreferredBackupWindow),
			PreferredMaintenanceWindow: aws.ToString(cluster.PreferredMaintenanceWindow),
			Port:                 int32(aws.ToInt32(cluster.Port)),
			DBSubnetGroupName:    aws.ToString(cluster.DBSubnetGroup),
			VpcSecurityGroupIds:  extractSecurityGroupIDs(cluster.VpcSecurityGroups),
			StorageEncrypted:     aws.ToBool(cluster.StorageEncrypted),
			KMSKeyARN:            aws.ToString(cluster.KmsKeyId),
			SkipFinalSnapshot:    aws.ToBool(cluster.DeletionProtection),
			DeletionProtection:   aws.ToBool(cluster.DeletionProtection),
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverDocDBResources discovers DocumentDB cluster resources
func (c *Client) discoverDocDBResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.rdsClient.DescribeDBClusters(context.TODO(), &rds.DescribeDBClustersInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list DocumentDB clusters: %w", err)
	}

	for _, cluster := range result.DBClusters {
		// Get tags
		tags, err := c.rdsClient.ListTagsForResource(context.TODO(), &rds.ListTagsForResourceInput{
			ResourceName: cluster.DBClusterArn,
		})
		if err != nil {
			continue
		}

		tagMap := make(map[string]string)
		for _, tag := range tags.TagList {
			if tag.Key != nil && tag.Value != nil {
				tagMap[*tag.Key] = *tag.Value
			}
		}

		resource := &DocDBResource{
			BaseResource: BaseResource{
				Type:   "docdb",
				ID:     *cluster.DBClusterIdentifier,
				Name:   *cluster.DBClusterIdentifier,
				Region: c.region,
				Tags:   tagMap,
			},
			ClusterIdentifier:     *cluster.DBClusterIdentifier,
			Engine:               aws.ToString(cluster.Engine),
			EngineVersion:        aws.ToString(cluster.EngineVersion),
			AvailabilityZones:    cluster.AvailabilityZones,
			BackupRetentionPeriod: int32(aws.ToInt32(cluster.BackupRetentionPeriod)),
			PreferredBackupWindow: aws.ToString(cluster.PreferredBackupWindow),
			PreferredMaintenanceWindow: aws.ToString(cluster.PreferredMaintenanceWindow),
			Port:                 int32(aws.ToInt32(cluster.Port)),
			DBSubnetGroupName:    aws.ToString(cluster.DBSubnetGroup),
			VpcSecurityGroupIds:  extractSecurityGroupIDs(cluster.VpcSecurityGroups),
			StorageEncrypted:     aws.ToBool(cluster.StorageEncrypted),
			KMSKeyARN:            aws.ToString(cluster.KmsKeyId),
			SkipFinalSnapshot:    aws.ToBool(cluster.DeletionProtection),
			DeletionProtection:   aws.ToBool(cluster.DeletionProtection),
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
// discoverElasticBeanstalkResources discovers Elastic Beanstalk environments
func (c *Client) discoverElasticBeanstalkResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.elasticbeanstalkClient.DescribeApplications(context.TODO(), &elasticbeanstalk.DescribeApplicationsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list Elastic Beanstalk applications: %w", err)
	}

	for _, app := range result.Applications {
		// Get tags
		tags, err := c.elasticbeanstalkClient.ListTagsForResource(context.TODO(), &elasticbeanstalk.ListTagsForResourceInput{
			ResourceArn: app.ApplicationArn,
		})
		if err != nil {
			continue
		}

		tagMap := make(map[string]string)
		for _, tag := range tags.ResourceTags {
			if tag.Key != nil && tag.Value != nil {
				tagMap[*tag.Key] = *tag.Value
			}
		}

		resource := &ElasticBeanstalkResource{
			BaseResource: BaseResource{
				Type:   "elasticbeanstalk",
				ID:     *app.ApplicationName,
				Name:   *app.ApplicationName,
				Region: c.region,
				Tags:   tagMap,
			},
			ApplicationName: *app.ApplicationName,
			EnvironmentName: *app.ApplicationName, // This will be updated when we get environment details
			Description:     aws.ToString(app.Description),
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
// discoverCodeBuildResources discovers CodeBuild project resources
func (c *Client) discoverCodeBuildResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.codebuildClient.ListProjects(context.TODO(), &codebuild.ListProjectsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list CodeBuild projects: %w", err)
	}

	for _, projectName := range result.Projects {
		// Get project details
		projectDetails, err := c.codebuildClient.BatchGetProjects(context.TODO(), &codebuild.BatchGetProjectsInput{
			Names: []string{projectName},
		})
		if err != nil {
			continue
		}

		if len(projectDetails.Projects) == 0 {
			continue
		}

		project := projectDetails.Projects[0]

		// CodeBuild doesn't have a direct ListTagsForResource method, skip tags for now
		tagMap := make(map[string]string)

		// Build source configuration
		source := &CodeBuildSource{
			Type:            string(project.Source.Type),
			Location:        aws.ToString(project.Source.Location),
			GitCloneDepth:   int32(aws.ToInt32(project.Source.GitCloneDepth)),
			Buildspec:       aws.ToString(project.Source.Buildspec),
			ReportBuildStatus: aws.ToBool(project.Source.ReportBuildStatus),
		}

		// Build artifacts configuration
		artifacts := &CodeBuildArtifacts{
			Type:                string(project.Artifacts.Type),
			Location:            aws.ToString(project.Artifacts.Location),
			Path:                aws.ToString(project.Artifacts.Path),
			NamespaceType:       string(project.Artifacts.NamespaceType),
			Name:                aws.ToString(project.Artifacts.Name),
			Packaging:           string(project.Artifacts.Packaging),
			OverrideArtifactName: aws.ToBool(project.Artifacts.OverrideArtifactName),
			EncryptionDisabled:  aws.ToBool(project.Artifacts.EncryptionDisabled),
		}

		// Build environment configuration
		environment := &CodeBuildEnvironment{
			Type:             string(project.Environment.Type),
			Image:            aws.ToString(project.Environment.Image),
			ComputeType:      string(project.Environment.ComputeType),
			PrivilegedMode:   aws.ToBool(project.Environment.PrivilegedMode),
		}

		resource := &CodeBuildResource{
			BaseResource: BaseResource{
				Type:   "codebuild",
				ID:     *project.Name,
				Name:   *project.Name,
				Region: c.region,
				Tags:   tagMap,
			},
			ProjectName:  *project.Name,
			Description:  aws.ToString(project.Description),
			BuildTimeout: int32(aws.ToInt32(project.TimeoutInMinutes)),
			ServiceRole:  aws.ToString(project.ServiceRole),
			Source:       source,
			Artifacts:    artifacts,
			Environment:  environment,
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverCodeDeployResources discovers CodeDeploy application resources
func (c *Client) discoverCodeDeployResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.codedeployClient.ListApplications(context.TODO(), &codedeploy.ListApplicationsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list CodeDeploy applications: %w", err)
	}

	for _, appName := range result.Applications {
		// Get application details
		appDetails, err := c.codedeployClient.GetApplication(context.TODO(), &codedeploy.GetApplicationInput{
			ApplicationName: &appName,
		})
		if err != nil {
			continue
		}

		app := appDetails.Application

		// CodeDeploy doesn't have ApplicationArn field, skip tags for now
		tagMap := make(map[string]string)

		resource := &CodeDeployResource{
			BaseResource: BaseResource{
				Type:   "codedeploy",
				ID:     *app.ApplicationName,
				Name:   *app.ApplicationName,
				Region: c.region,
				Tags:   tagMap,
			},
			ApplicationName: *app.ApplicationName,
			ComputePlatform: string(app.ComputePlatform),
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverSSMResources discovers SSM parameter resources
func (c *Client) discoverSSMResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.ssmClient.DescribeParameters(context.TODO(), &ssm.DescribeParametersInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list SSM parameters: %w", err)
	}

	for _, param := range result.Parameters {
		// Get tags
		tags, err := c.ssmClient.ListTagsForResource(context.TODO(), &ssm.ListTagsForResourceInput{
			ResourceType: "Parameter",
			ResourceId:   param.Name,
		})
		if err != nil {
			continue
		}

		tagMap := make(map[string]string)
		for _, tag := range tags.TagList {
			if tag.Key != nil && tag.Value != nil {
				tagMap[*tag.Key] = *tag.Value
			}
		}

		// Extract description if present
		var description string
		if param.Description != nil {
			description = *param.Description
		}

		// Extract tier if present
		var tier string
		if param.Tier != "" {
			tier = string(param.Tier)
		}

		resource := &SSMParameterResource{
			BaseResource: BaseResource{
				Type:   "ssm",
				ID:     *param.Name,
				Name:   *param.Name,
				Region: c.region,
				Tags:   tagMap,
			},
			Name:        *param.Name,
			Type:        string(param.Type),
			Description: description,
			Tier:        tier,
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverSecretsManagerResources discovers Secrets Manager secret resources
func (c *Client) discoverSecretsManagerResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.secretsmanagerClient.ListSecrets(context.TODO(), &secretsmanager.ListSecretsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list Secrets Manager secrets: %w", err)
	}

	for _, secret := range result.SecretList {
		// Secrets Manager doesn't have a direct ListTagsForResource method, skip tags for now
		tagMap := make(map[string]string)

		// Extract description if present
		var description string
		if secret.Description != nil {
			description = *secret.Description
		}

		// Extract KMS key ID if present
		var kmsKeyID string
		if secret.KmsKeyId != nil {
			kmsKeyID = *secret.KmsKeyId
		}

		resource := &SecretsManagerResource{
			BaseResource: BaseResource{
				Type:   "secretsmanager",
				ID:     *secret.Name,
				Name:   *secret.Name,
				Region: c.region,
				Tags:   tagMap,
			},
			Name:        *secret.Name,
			Description: description,
			KMSKeyID:    kmsKeyID,
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverKMSResources discovers KMS key resources
func (c *Client) discoverKMSResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.kmsClient.ListKeys(context.TODO(), &kms.ListKeysInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list KMS keys: %w", err)
	}

	for _, key := range result.Keys {
		// Get key details
		keyDetails, err := c.kmsClient.DescribeKey(context.TODO(), &kms.DescribeKeyInput{
			KeyId: key.KeyId,
		})
		if err != nil {
			continue
		}

		keyMetadata := keyDetails.KeyMetadata

		// Get tags
		tags, err := c.kmsClient.ListResourceTags(context.TODO(), &kms.ListResourceTagsInput{
			KeyId: key.KeyId,
		})
		if err != nil {
			continue
		}

		tagMap := make(map[string]string)
		for _, tag := range tags.Tags {
			if tag.TagKey != nil && tag.TagValue != nil {
				tagMap[*tag.TagKey] = *tag.TagValue
			}
		}

		// Extract description if present
		var description string
		if keyMetadata.Description != nil {
			description = *keyMetadata.Description
		}

		resource := &KMSResource{
			BaseResource: BaseResource{
				Type:   "kms",
				ID:     *keyMetadata.KeyId,
				Name:   description,
				Region: c.region,
				Tags:   tagMap,
			},
			KeyID:                *keyMetadata.KeyId,
			Description:          description,
			KeyUsage:             string(keyMetadata.KeyUsage),
			CustomerMasterKeySpec: string(keyMetadata.CustomerMasterKeySpec),
			DeletionWindowInDays: 7, // Default value
			EnableKeyRotation:    false, // Default value
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverCodeCommitResources discovers CodeCommit repository resources
func (c *Client) discoverCodeCommitResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.codecommitClient.ListRepositories(context.TODO(), &codecommit.ListRepositoriesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list CodeCommit repositories: %w", err)
	}

	for _, repo := range result.Repositories {
		// Get repository details
		repoDetails, err := c.codecommitClient.GetRepository(context.TODO(), &codecommit.GetRepositoryInput{
			RepositoryName: repo.RepositoryName,
		})
		if err != nil {
			continue
		}

		repository := repoDetails.RepositoryMetadata

		// Get tags
		tags, err := c.codecommitClient.ListTagsForResource(context.TODO(), &codecommit.ListTagsForResourceInput{
			ResourceArn: repository.Arn,
		})
		if err != nil {
			continue
		}

		tagMap := make(map[string]string)
		for key, value := range tags.Tags {
			tagMap[key] = value
		}

		// Extract description if present
		var description string
		if repository.RepositoryDescription != nil {
			description = *repository.RepositoryDescription
		}

		// Extract default branch if present
		var defaultBranch string
		if repository.DefaultBranch != nil {
			defaultBranch = *repository.DefaultBranch
		}

		resource := &CodeCommitResource{
			BaseResource: BaseResource{
				Type:   "codecommit",
				ID:     *repository.RepositoryName,
				Name:   *repository.RepositoryName,
				Region: c.region,
				Tags:   tagMap,
			},
			RepositoryName: *repository.RepositoryName,
			Description:    description,
			DefaultBranch:  defaultBranch,
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverCodePipelineResources discovers CodePipeline pipeline resources
func (c *Client) discoverCodePipelineResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.codepipelineClient.ListPipelines(context.TODO(), &codepipeline.ListPipelinesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list CodePipeline pipelines: %w", err)
	}

	for _, pipeline := range result.Pipelines {
		// Get pipeline details
		pipelineDetails, err := c.codepipelineClient.GetPipeline(context.TODO(), &codepipeline.GetPipelineInput{
			Name: pipeline.Name,
		})
		if err != nil {
			continue
		}

		pipelineInfo := pipelineDetails.Pipeline

		// Get tags - skip for now as ARN is not directly available
		tagMap := make(map[string]string)

		resource := &CodePipelineResource{
			BaseResource: BaseResource{
				Type:   "codepipeline",
				ID:     *pipeline.Name,
				Name:   *pipeline.Name,
				Region: c.region,
				Tags:   tagMap,
			},
			PipelineName: *pipeline.Name,
			RoleARN:      *pipelineInfo.RoleArn,
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverCloudFormationResources discovers CloudFormation stack resources
func (c *Client) discoverCloudFormationResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.cloudformationClient.ListStacks(context.TODO(), &cloudformation.ListStacksInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list CloudFormation stacks: %w", err)
	}

	for _, stack := range result.StackSummaries {
		// Get stack details
		stackDetails, err := c.cloudformationClient.DescribeStacks(context.TODO(), &cloudformation.DescribeStacksInput{
			StackName: stack.StackName,
		})
		if err != nil {
			continue
		}

		if len(stackDetails.Stacks) == 0 {
			continue
		}

		stackInfo := stackDetails.Stacks[0]

		// Get tags
		tagMap := make(map[string]string)
		for _, tag := range stackInfo.Tags {
			if tag.Key != nil && tag.Value != nil {
				tagMap[*tag.Key] = *tag.Value
			}
		}

		resource := &CloudFormationResource{
			BaseResource: BaseResource{
				Type:   "cloudformation",
				ID:     *stack.StackName,
				Name:   *stack.StackName,
				Region: c.region,
				Tags:   tagMap,
			},
			StackName: *stack.StackName,
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
// discoverConfigResources discovers AWS Config resources
func (c *Client) discoverConfigResources() ([]Resource, error) {
	var resources []Resource

	// Discover configuration recorders
	recordersResult, err := c.configClient.DescribeConfigurationRecorders(context.TODO(), &configservice.DescribeConfigurationRecordersInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list Config recorders: %w", err)
	}

	for _, recorder := range recordersResult.ConfigurationRecorders {
		// Build recording group if present
		var recordingGroup *ConfigRecordingGroup
		if recorder.RecordingGroup != nil {
			// Convert ResourceType slice to string slice
			var resourceTypes []string
			for _, rt := range recorder.RecordingGroup.ResourceTypes {
				resourceTypes = append(resourceTypes, string(rt))
			}
			
			recordingGroup = &ConfigRecordingGroup{
				AllSupported:           recorder.RecordingGroup.AllSupported,
				IncludeGlobalResources: recorder.RecordingGroup.IncludeGlobalResourceTypes,
				ResourceTypes:          resourceTypes,
			}
		}

		// Build recording mode if present
		var recordingMode *ConfigRecordingMode
		if recorder.RecordingMode != nil {
			recordingMode = &ConfigRecordingMode{
				RecordingFrequency: string(recorder.RecordingMode.RecordingFrequency),
				// MaximumExecutionFrequency is not available in the SDK response
			}
		}

		resource := &ConfigResource{
			BaseResource: BaseResource{
				Type:   "config",
				ID:     *recorder.Name,
				Name:   *recorder.Name,
				Region: c.region,
				Tags:   make(map[string]string),
			},
			RecorderName:  *recorder.Name,
			RoleARN:       aws.ToString(recorder.RoleARN),
			RecordingGroup: recordingGroup,
			RecordingMode:  recordingMode,
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverKinesisResources discovers Kinesis stream resources
func (c *Client) discoverKinesisResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.kinesisClient.ListStreams(context.TODO(), &kinesis.ListStreamsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list Kinesis streams: %w", err)
	}

	for _, streamName := range result.StreamNames {
		// Get stream details
		streamDetails, err := c.kinesisClient.DescribeStream(context.TODO(), &kinesis.DescribeStreamInput{
			StreamName: &streamName,
		})
		if err != nil {
			continue
		}

		stream := streamDetails.StreamDescription

		// Get tags
		tags, err := c.kinesisClient.ListTagsForStream(context.TODO(), &kinesis.ListTagsForStreamInput{
			StreamName: &streamName,
		})
		if err != nil {
			continue
		}

		tagMap := make(map[string]string)
		for _, tag := range tags.Tags {
			if tag.Key != nil && tag.Value != nil {
				tagMap[*tag.Key] = *tag.Value
			}
		}

		// Build stream mode details if present
		var streamModeDetails *KinesisStreamModeDetails
		if stream.StreamModeDetails != nil {
			streamModeDetails = &KinesisStreamModeDetails{
				StreamMode: string(stream.StreamModeDetails.StreamMode),
			}
		}

		resource := &KinesisResource{
			BaseResource: BaseResource{
				Type:   "kinesis",
				ID:     streamName,
				Name:   streamName,
				Region: c.region,
				Tags:   tagMap,
			},
			StreamName:           streamName,
			ShardCount:           int32(len(stream.Shards)),
			RetentionPeriodHours: int32(aws.ToInt32(stream.RetentionPeriodHours)),
			StreamARN:            aws.ToString(stream.StreamARN),
			EncryptionType:       string(stream.EncryptionType),
			KMSKeyID:            aws.ToString(stream.KeyId),
			StreamModeDetails:    streamModeDetails,
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverFSxResources discovers FSx file system resources
func (c *Client) discoverFSxResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.fsxClient.DescribeFileSystems(context.TODO(), &fsx.DescribeFileSystemsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list FSx file systems: %w", err)
	}

	for _, filesystem := range result.FileSystems {
		// Get tags
		tags, err := c.fsxClient.ListTagsForResource(context.TODO(), &fsx.ListTagsForResourceInput{
			ResourceARN: filesystem.ResourceARN,
		})
		if err != nil {
			continue
		}

		tagMap := make(map[string]string)
		for _, tag := range tags.Tags {
			if tag.Key != nil && tag.Value != nil {
				tagMap[*tag.Key] = *tag.Value
			}
		}

		resource := &FSxResource{
			BaseResource: BaseResource{
				Type:   "fsx",
				ID:     *filesystem.FileSystemId,
				Name:   *filesystem.FileSystemId,
				Region: c.region,
				Tags:   tagMap,
			},
			FileSystemID:         *filesystem.FileSystemId,
			FileSystemType:       string(filesystem.FileSystemType),
			StorageCapacity:      int32(aws.ToInt32(filesystem.StorageCapacity)),
			SubnetIDs:            filesystem.SubnetIds,
			// SecurityGroupIDs, KMSKeyID, StorageType, DeploymentType, PreferredSubnetID, RouteTableIDs
			// are not available in the basic FileSystem type - would need additional API calls
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverGuardDutyResources discovers GuardDuty detector resources
func (c *Client) discoverGuardDutyResources() ([]Resource, error) {
	// TODO: Implement GuardDuty resource discovery
	return []Resource{}, nil
}
// discoverBackupResources discovers AWS Backup vault resources
func (c *Client) discoverBackupResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.backupClient.ListBackupVaults(context.TODO(), &backup.ListBackupVaultsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list AWS Backup vaults: %w", err)
	}

	for _, vault := range result.BackupVaultList {
		// Get tags
		tags, err := c.backupClient.ListTags(context.TODO(), &backup.ListTagsInput{
			ResourceArn: vault.BackupVaultArn,
		})
		if err != nil {
			continue
		}

		tagMap := make(map[string]string)
		for key, value := range tags.Tags {
			tagMap[key] = value
		}

		resource := &BackupResource{
			BaseResource: BaseResource{
				Type:   "backup",
				ID:     *vault.BackupVaultName,
				Name:   *vault.BackupVaultName,
				Region: c.region,
				Tags:   tagMap,
			},
			BackupVaultName: *vault.BackupVaultName,
			BackupVaultARN:  aws.ToString(vault.BackupVaultArn),
			KMSKeyARN:       aws.ToString(vault.EncryptionKeyArn),
			ForceDestroy:    false, // Default value, would need additional API call to determine
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverGlacierResources discovers Glacier vault resources
func (c *Client) discoverGlacierResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.glacierClient.ListVaults(context.TODO(), &glacier.ListVaultsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list Glacier vaults: %w", err)
	}

	for _, vault := range result.VaultList {
		// Get tags
		tags, err := c.glacierClient.ListTagsForVault(context.TODO(), &glacier.ListTagsForVaultInput{
			VaultName: vault.VaultName,
		})
		if err != nil {
			continue
		}

		tagMap := make(map[string]string)
		for key, value := range tags.Tags {
			tagMap[key] = value
		}

		resource := &GlacierResource{
			BaseResource: BaseResource{
				Type:   "glacier",
				ID:     *vault.VaultName,
				Name:   *vault.VaultName,
				Region: c.region,
				Tags:   tagMap,
			},
			VaultName: *vault.VaultName,
			VaultARN:  aws.ToString(vault.VaultARN),
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverGlueResources discovers Glue catalog and ETL resources
func (c *Client) discoverGlueResources() ([]Resource, error) {
	var resources []Resource

	// Discover databases
	databases, err := c.glueClient.GetDatabases(context.TODO(), &glue.GetDatabasesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list Glue databases: %w", err)
	}

	for _, db := range databases.DatabaseList {
		// Get tags
		tags, err := c.glueClient.GetTags(context.TODO(), &glue.GetTagsInput{
			ResourceArn: db.CatalogId,
		})
		if err != nil {
			continue
		}

		tagMap := make(map[string]string)
		for key, value := range tags.Tags {
			tagMap[key] = value
		}

		resource := &GlueResource{
			BaseResource: BaseResource{
				Type:   "glue",
				ID:     *db.Name,
				Name:   *db.Name,
				Region: c.region,
				Tags:   tagMap,
			},
			DatabaseName: *db.Name,
			Description:  aws.ToString(db.Description),
			LocationURI:  aws.ToString(db.LocationUri),
			Parameters:   db.Parameters,
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverAthenaResources discovers Athena workgroup resources
func (c *Client) discoverAthenaResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.athenaClient.ListWorkGroups(context.TODO(), &athena.ListWorkGroupsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list Athena workgroups: %w", err)
	}

	for _, workgroup := range result.WorkGroups {
		// Get workgroup details
		workgroupDetails, err := c.athenaClient.GetWorkGroup(context.TODO(), &athena.GetWorkGroupInput{
			WorkGroup: workgroup.Name,
		})
		if err != nil {
			continue
		}

		workgroupInfo := workgroupDetails.WorkGroup

		// Get tags - skip for now as ARN is not directly available
		tagMap := make(map[string]string)

		// Build configuration if present
		var configuration *AthenaWorkgroupConfiguration
		if workgroupInfo.Configuration != nil {
			config := workgroupInfo.Configuration
			
			// Build engine version if present
			var engineVersion *AthenaEngineVersion
			if config.EngineVersion != nil {
				engineVersion = &AthenaEngineVersion{
					SelectedEngineVersion: aws.ToString(config.EngineVersion.SelectedEngineVersion),
					EffectiveEngineVersion: aws.ToString(config.EngineVersion.EffectiveEngineVersion),
				}
			}

			// Build result configuration if present
			var resultConfig *AthenaResultConfiguration
			if config.ResultConfiguration != nil {
				result := config.ResultConfiguration
				
				// Build encryption configuration if present
				var encryptionConfig *AthenaEncryptionConfiguration
				if result.EncryptionConfiguration != nil {
					encryptionConfig = &AthenaEncryptionConfiguration{
						EncryptionOption: string(result.EncryptionConfiguration.EncryptionOption),
						KMSKey:           aws.ToString(result.EncryptionConfiguration.KmsKey),
					}
				}

				resultConfig = &AthenaResultConfiguration{
					OutputLocation:          aws.ToString(result.OutputLocation),
					EncryptionConfiguration: encryptionConfig,
				}
			}

			configuration = &AthenaWorkgroupConfiguration{
				EnforceWorkgroupConfiguration:   aws.ToBool(config.EnforceWorkGroupConfiguration),
				PublishCloudwatchMetricsEnabled: aws.ToBool(config.PublishCloudWatchMetricsEnabled),
				BytesScannedCutoffPerQuery:     aws.ToInt64(config.BytesScannedCutoffPerQuery),
				RequesterPaysEnabled:           aws.ToBool(config.RequesterPaysEnabled),
				EngineVersion:                  engineVersion,
				ResultConfiguration:            resultConfig,
			}
		}

		resource := &AthenaResource{
			BaseResource: BaseResource{
				Type:   "athena",
				ID:     *workgroup.Name,
				Name:   *workgroup.Name,
				Region: c.region,
				Tags:   tagMap,
			},
			WorkgroupName: *workgroup.Name,
			Description:   aws.ToString(workgroupInfo.Description),
			State:         string(workgroupInfo.State),
			Configuration: configuration,
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
// discoverQuickSightResources discovers QuickSight user resources
func (c *Client) discoverQuickSightResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.quicksightClient.ListUsers(context.TODO(), &quicksight.ListUsersInput{
		AwsAccountId: aws.String("123456789012"), // This would need to be dynamic
		Namespace:    aws.String("default"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list QuickSight users: %w", err)
	}

	for _, user := range result.UserList {
		// Get tags - skip for now as method is not available
		tagMap := make(map[string]string)

		resource := &QuickSightResource{
			BaseResource: BaseResource{
				Type:   "quicksight",
				ID:     *user.UserName,
				Name:   *user.UserName,
				Region: c.region,
				Tags:   tagMap,
			},
			UserName:    *user.UserName,
			Email:       aws.ToString(user.Email),
			IdentityType: string(user.IdentityType),
			UserRole:    string(user.Role),
			Namespace:   "default", // Default namespace
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
// discoverWorkSpacesResources discovers WorkSpaces workspace resources
func (c *Client) discoverWorkSpacesResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.workspacesClient.DescribeWorkspaces(context.TODO(), &workspaces.DescribeWorkspacesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list WorkSpaces workspaces: %w", err)
	}

	for _, workspace := range result.Workspaces {
		// Get tags - skip for now as method is not available
		tagMap := make(map[string]string)

		resource := &WorkSpacesResource{
			BaseResource: BaseResource{
				Type:   "workspaces",
				ID:     *workspace.WorkspaceId,
				Name:   *workspace.WorkspaceId,
				Region: c.region,
				Tags:   tagMap,
			},
			WorkspaceID:    *workspace.WorkspaceId,
			DirectoryID:    aws.ToString(workspace.DirectoryId),
			BundleID:       aws.ToString(workspace.BundleId),
			UserName:       aws.ToString(workspace.UserName),
			RootVolumeSizeGib: int32(aws.ToInt32(workspace.WorkspaceProperties.RootVolumeSizeGib)),
			UserVolumeSizeGib: int32(aws.ToInt32(workspace.WorkspaceProperties.UserVolumeSizeGib)),
			ComputeTypeName: string(workspace.WorkspaceProperties.ComputeTypeName),
			UserVolumeEncryptionEnabled: false, // Not available in basic response
			RootVolumeEncryptionEnabled: false, // Not available in basic response
			RunningMode: string(workspace.WorkspaceProperties.RunningMode),
			AutoStopTimeoutInMinutes: 0, // Not available in basic response
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// Directory Service stub removed

// discoverStorageGatewayResources discovers Storage Gateway resources
func (c *Client) discoverStorageGatewayResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.storagegatewayClient.ListGateways(context.TODO(), &storagegateway.ListGatewaysInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list Storage Gateway gateways: %w", err)
	}

	for _, gateway := range result.Gateways {
		// Get gateway details
		gatewayDetails, err := c.storagegatewayClient.DescribeGatewayInformation(context.TODO(), &storagegateway.DescribeGatewayInformationInput{
			GatewayARN: gateway.GatewayARN,
		})
		if err != nil {
			continue
		}

		_ = gatewayDetails // Use gatewayDetails to avoid unused variable

		// Get tags
		tags, err := c.storagegatewayClient.ListTagsForResource(context.TODO(), &storagegateway.ListTagsForResourceInput{
			ResourceARN: gateway.GatewayARN,
		})
		if err != nil {
			continue
		}

		tagMap := make(map[string]string)
		for _, tag := range tags.Tags {
			if tag.Key != nil && tag.Value != nil {
				tagMap[*tag.Key] = *tag.Value
			}
		}

		resource := &StorageGatewayResource{
			BaseResource: BaseResource{
				Type:   "storagegateway",
				ID:     *gateway.GatewayARN,
				Name:   *gateway.GatewayName,
				Region: c.region,
				Tags:   tagMap,
			},
			GatewayName: *gateway.GatewayName,
			GatewayARN:  *gateway.GatewayARN,
			GatewayType: aws.ToString(gateway.GatewayType),
			GatewayTimezone: aws.ToString(gatewayDetails.GatewayTimezone),
			GatewayRegion: c.region, // Use client region
			GatewayVPCEndpoint: aws.ToString(gatewayDetails.VPCEndpoint),
			CloudWatchLogGroupARN: aws.ToString(gatewayDetails.CloudWatchLogGroupARN),
			AverageDownloadRateLimitInBitsPerSec: 0, // Not available in basic response
			AverageUploadRateLimitInBitsPerSec: 0, // Not available in basic response
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverTransferResources discovers AWS Transfer resources
func (c *Client) discoverTransferResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.transferClient.ListServers(context.TODO(), &transfer.ListServersInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list AWS Transfer servers: %w", err)
	}

	for _, server := range result.Servers {
		// Get tags
		tags, err := c.transferClient.ListTagsForResource(context.TODO(), &transfer.ListTagsForResourceInput{
			Arn: server.Arn,
		})
		if err != nil {
			continue
		}

		tagMap := make(map[string]string)
		for _, tag := range tags.Tags {
			if tag.Key != nil && tag.Value != nil {
				tagMap[*tag.Key] = *tag.Value
			}
		}

		// Build workflow details if present
		var workflowDetails *TransferWorkflowDetails
		// Note: Workflow details would need additional API calls

		resource := &TransferResource{
			BaseResource: BaseResource{
				Type:   "transfer",
				ID:     *server.ServerId,
				Name:   *server.ServerId,
				Region: c.region,
				Tags:   tagMap,
			},
			ServerID:            *server.ServerId,
			IdentityProviderType: string(server.IdentityProviderType),
			LoggingRole:         aws.ToString(server.LoggingRole),
			Protocols:           []string{}, // Not available in basic response
			EndpointType:        string(server.EndpointType),
			SecurityPolicyName:  "", // Not available in basic response
			WorkflowDetails:     workflowDetails,
			StructuredLogDestinations: []string{}, // Not available in basic response
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverMQResources discovers Amazon MQ broker resources
func (c *Client) discoverMQResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.mqClient.ListBrokers(context.TODO(), &mq.ListBrokersInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list Amazon MQ brokers: %w", err)
	}

	for _, broker := range result.BrokerSummaries {
		// Get broker details
		brokerDetails, err := c.mqClient.DescribeBroker(context.TODO(), &mq.DescribeBrokerInput{
			BrokerId: broker.BrokerId,
		})
		if err != nil {
			continue
		}

		brokerInfo := brokerDetails

		// Get tags
		tags, err := c.mqClient.ListTags(context.TODO(), &mq.ListTagsInput{
			ResourceArn: brokerInfo.BrokerArn,
		})
		if err != nil {
			continue
		}

		tagMap := make(map[string]string)
		for key, value := range tags.Tags {
			tagMap[key] = value
		}

		// Build maintenance window if present
		var maintenanceWindow *MQMaintenanceWindow
		if brokerInfo.MaintenanceWindowStartTime != nil {
			maintenanceWindow = &MQMaintenanceWindow{
				DayOfWeek: string(brokerInfo.MaintenanceWindowStartTime.DayOfWeek),
				TimeOfDay: aws.ToString(brokerInfo.MaintenanceWindowStartTime.TimeOfDay),
				TimeZone:  aws.ToString(brokerInfo.MaintenanceWindowStartTime.TimeZone),
			}
		}

		// Build logs configuration if present
		var logs *MQLogs
		if brokerInfo.Logs != nil {
			logs = &MQLogs{
				Audit:   aws.ToBool(brokerInfo.Logs.Audit),
				General: aws.ToBool(brokerInfo.Logs.General),
			}
		}

		// Build configuration if present
		var configuration *MQConfiguration
		if brokerInfo.Configurations != nil && brokerInfo.Configurations.Current != nil {
			config := brokerInfo.Configurations.Current
			configuration = &MQConfiguration{
				ID:       aws.ToString(config.Id),
				Revision: int32(aws.ToInt32(config.Revision)),
			}
		}

		resource := &MQResource{
			BaseResource: BaseResource{
				Type:   "mq",
				ID:     *broker.BrokerId,
				Name:   *broker.BrokerName,
				Region: c.region,
				Tags:   tagMap,
			},
			BrokerName:       *broker.BrokerName,
			BrokerID:         *broker.BrokerId,
			EngineType:       string(broker.EngineType),
			EngineVersion:    aws.ToString(brokerInfo.EngineVersion),
			HostInstanceType: aws.ToString(brokerInfo.HostInstanceType),
			DeploymentMode:   string(brokerInfo.DeploymentMode),
			SecurityGroups:   brokerInfo.SecurityGroups,
			SubnetIDs:        brokerInfo.SubnetIds,
			MaintenanceWindowStartTime: maintenanceWindow,
			Logs:                      logs,
			Configuration:              configuration,
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
// discoverFirehoseResources discovers Kinesis Firehose delivery stream resources
func (c *Client) discoverFirehoseResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.firehoseClient.ListDeliveryStreams(context.TODO(), &firehose.ListDeliveryStreamsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list Kinesis Firehose delivery streams: %w", err)
	}

	for _, streamName := range result.DeliveryStreamNames {
		// Skip empty or invalid stream names
		if streamName == "" || len(strings.TrimSpace(streamName)) == 0 {
			continue
		}

		// Validate stream name format
		if !isValidFirehoseStreamName(streamName) {
			continue
		}

		// Additional validation: ensure stream name is not just whitespace or special characters
		cleanName := strings.TrimSpace(streamName)
		if cleanName == "" || len(cleanName) == 0 {
			continue
		}

		// Get stream details
		streamDetails, err := c.firehoseClient.DescribeDeliveryStream(context.TODO(), &firehose.DescribeDeliveryStreamInput{
			DeliveryStreamName: &streamName,
		})
		if err != nil {
			continue
		}

		stream := streamDetails.DeliveryStreamDescription

		// Get tags
		tags, err := c.firehoseClient.ListTagsForDeliveryStream(context.TODO(), &firehose.ListTagsForDeliveryStreamInput{
			DeliveryStreamName: &streamName,
		})
		if err != nil {
			continue
		}

		tagMap := make(map[string]string)
		for _, tag := range tags.Tags {
			if tag.Key != nil && tag.Value != nil {
				tagMap[*tag.Key] = *tag.Value
			}
		}

		resource := &FirehoseResource{
			BaseResource: BaseResource{
				Type:   "firehose",
				ID:     streamName,
				Name:   streamName,
				Region: c.region,
				Tags:   tagMap,
			},
			Name:                 streamName,
			DeliveryStreamType:   string(stream.DeliveryStreamType),
			DeliveryStreamStatus: string(stream.DeliveryStreamStatus),
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// isValidFirehoseStreamName validates Firehose stream name format
func isValidFirehoseStreamName(name string) bool {
	if name == "" {
		return false
	}
	
	// Check length (1-64 characters)
	if len(name) < 1 || len(name) > 64 {
		return false
	}
	
	// Check pattern: [a-zA-Z0-9_.-]+
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_.-]+$`, name)
	return matched
}

// discoverMediaStoreResources discovers MediaStore container resources
func (c *Client) discoverMediaStoreResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.mediastoreClient.ListContainers(context.TODO(), &mediastore.ListContainersInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list MediaStore containers: %w", err)
	}

	for _, container := range result.Containers {
		// Get tags
		tags, err := c.mediastoreClient.ListTagsForResource(context.TODO(), &mediastore.ListTagsForResourceInput{
			Resource: container.ARN,
		})
		if err != nil {
			continue
		}

		tagMap := make(map[string]string)
		for _, tag := range tags.Tags {
			if tag.Key != nil && tag.Value != nil {
				tagMap[*tag.Key] = *tag.Value
			}
		}

		resource := &MediaStoreResource{
			BaseResource: BaseResource{
				Type:   "mediastore",
				ID:     *container.Name,
				Name:   *container.Name,
				Region: c.region,
				Tags:   tagMap,
			},
			ContainerName: *container.Name,
			ContainerARN:  aws.ToString(container.ARN),
			Status:        string(container.Status),
			AccessLoggingEnabled: false, // Would need additional API call to determine
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverMediaConvertResources discovers MediaConvert queue resources
func (c *Client) discoverMediaConvertResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.mediaconvertClient.ListQueues(context.TODO(), &mediaconvert.ListQueuesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list MediaConvert queues: %w", err)
	}

	for _, queue := range result.Queues {
		// Get tags - skip for now as ResourceTags type is complex
		tagMap := make(map[string]string)

		// Build reservation plan if present
		var reservationPlan *MediaConvertReservationPlan
		if queue.ReservationPlan != nil {
			reservationPlan = &MediaConvertReservationPlan{
				Commitment:    string(queue.ReservationPlan.Commitment),
				ReservedSlots: int32(aws.ToInt32(queue.ReservationPlan.ReservedSlots)),
				RenewalType:   string(queue.ReservationPlan.RenewalType),
			}
		}

		resource := &MediaConvertResource{
			BaseResource: BaseResource{
				Type:   "mediaconvert",
				ID:     *queue.Name,
				Name:   *queue.Name,
				Region: c.region,
				Tags:   tagMap,
			},
			QueueName:    *queue.Name,
			QueueARN:     aws.ToString(queue.Arn),
			Type:         string(queue.Type),
			Status:       string(queue.Status),
			PricingPlan:  string(queue.PricingPlan),
			ReservationPlan: reservationPlan,
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverMediaLiveResources discovers MediaLive channel resources
func (c *Client) discoverMediaLiveResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.medialiveClient.ListChannels(context.TODO(), &medialive.ListChannelsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list MediaLive channels: %w", err)
	}

	for _, channel := range result.Channels {
		// Get tags
		tags, err := c.medialiveClient.ListTagsForResource(context.TODO(), &medialive.ListTagsForResourceInput{
			ResourceArn: channel.Arn,
		})
		if err != nil {
			continue
		}

		tagMap := make(map[string]string)
		for key, value := range tags.Tags {
			tagMap[key] = value
		}

		// Build input specification if present
		var inputSpec *MediaLiveInputSpecification
		if channel.InputSpecification != nil {
			inputSpec = &MediaLiveInputSpecification{
				Codec:         string(channel.InputSpecification.Codec),
				Resolution:    string(channel.InputSpecification.Resolution),
				MaximumBitrate: string(channel.InputSpecification.MaximumBitrate),
			}
		}

		// Note: Encoder settings would need additional API calls to retrieve detailed configuration
		var encoderSettings *MediaLiveEncoderSettings

		resource := &MediaLiveResource{
			BaseResource: BaseResource{
				Type:   "medialive",
				ID:     *channel.Id,
				Name:   *channel.Name,
				Region: c.region,
				Tags:   tagMap,
			},
			ChannelID:    *channel.Id,
			ChannelName:  *channel.Name,
			ChannelARN:   aws.ToString(channel.Arn),
			State:        string(channel.State),
			ChannelClass: string(channel.ChannelClass),
			InputSpecification: inputSpec,
			EncoderSettings:    encoderSettings,
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverMediaTailorResources discovers MediaTailor configuration resources
func (c *Client) discoverMediaTailorResources() ([]Resource, error) {
	// TODO: Implement MediaTailor resource discovery
	return []Resource{}, nil
}
// discoverIoTResources discovers IoT Core resources
func (c *Client) discoverIoTResources() ([]Resource, error) {
	// TODO: Implement IoT resource discovery
	return []Resource{}, nil
}

// Greengrass stub services removed

// discoverIoTAnalyticsResources discovers IoT Analytics resources
func (c *Client) discoverIoTAnalyticsResources() ([]Resource, error) {
	// TODO: Implement IoT Analytics resource discovery
	return []Resource{}, nil
}
// discoverIoTEventsResources discovers IoT Events resources
func (c *Client) discoverIoTEventsResources() ([]Resource, error) {
	// TODO: Implement IoT Events resource discovery
	return []Resource{}, nil
}

// discoverIoTSiteWiseResources discovers IoT SiteWise resources
func (c *Client) discoverIoTSiteWiseResources() ([]Resource, error) {
	// TODO: Implement IoT SiteWise resource discovery
	return []Resource{}, nil
}

// discoverIoTThingsGraphResources discovers IoT Things Graph resources
func (c *Client) discoverIoTThingsGraphResources() ([]Resource, error) {
	// TODO: Implement IoT Things Graph resource discovery
	return []Resource{}, nil
}

// discoverIoTWirelessResources discovers IoT Wireless resources
func (c *Client) discoverIoTWirelessResources() ([]Resource, error) {
	// TODO: Implement IoT Wireless resource discovery
	return []Resource{}, nil
}
// IoT stub services removed 