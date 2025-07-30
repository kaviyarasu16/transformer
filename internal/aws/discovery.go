package aws

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
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
	"github.com/aws/aws-sdk-go-v2/service/guardduty"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	"github.com/aws/aws-sdk-go-v2/service/iotanalytics"
	"github.com/aws/aws-sdk-go-v2/service/iotevents"
	"github.com/aws/aws-sdk-go-v2/service/iotsitewise"
	"github.com/aws/aws-sdk-go-v2/service/iotthingsgraph"
	"github.com/aws/aws-sdk-go-v2/service/iotwireless"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/mediaconvert"
	"github.com/aws/aws-sdk-go-v2/service/medialive"
	"github.com/aws/aws-sdk-go-v2/service/mediastore"
	"github.com/aws/aws-sdk-go-v2/service/mediatailor"
	"github.com/aws/aws-sdk-go-v2/service/mq"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
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
	return decoded
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "asg",
				ID:     *asg.AutoScalingGroupName,
				Name:   *asg.AutoScalingGroupName,
				Region: c.region,
				Tags:   tags,
			},
			ResourceType: "aws_autoscaling_group",
			Attributes: map[string]interface{}{
				"name":                *asg.AutoScalingGroupName,
				"max_size":            asg.MaxSize,
				"min_size":            asg.MinSize,
				"desired_capacity":    asg.DesiredCapacity,
				"health_check_type":   asg.HealthCheckType,
				"health_check_grace_period": asg.HealthCheckGracePeriod,
			},
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
	var resources []Resource

	result, err := c.sqsClient.ListQueues(context.TODO(), &sqs.ListQueuesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list SQS queues: %w", err)
	}

	for _, queueURL := range result.QueueUrls {
		// Get queue attributes
		attributes, err := c.sqsClient.GetQueueAttributes(context.TODO(), &sqs.GetQueueAttributesInput{
			QueueUrl: &queueURL,
			AttributeNames: []types.QueueAttributeName{types.QueueAttributeNameAll},
		})
		if err != nil {
			continue
		}

		// Extract queue name from URL
		parts := strings.Split(queueURL, "/")
		queueName := parts[len(parts)-1]

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "sqs",
				ID:     queueName,
				Name:   queueName,
				Region: c.region,
				Tags:   make(map[string]string), // SQS tags would need separate call
			},
			ResourceType: "aws_sqs_queue",
			Attributes: map[string]interface{}{
				"name": queueName,
				"url":  queueURL,
			},
		}

		// Add attributes if available
		if attributes.Attributes != nil {
			for k, v := range attributes.Attributes {
				resource.Attributes[k] = v
			}
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverSNSResources discovers SNS topic resources
func (c *Client) discoverSNSResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.snsClient.ListTopics(context.TODO(), &sns.ListTopicsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list SNS topics: %w", err)
	}

	for _, topic := range result.Topics {
		// Extract topic name from ARN
		parts := strings.Split(*topic.TopicArn, ":")
		topicName := parts[len(parts)-1]

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "sns",
				ID:     topicName,
				Name:   topicName,
				Region: c.region,
				Tags:   make(map[string]string), // SNS tags would need separate call
			},
			ResourceType: "aws_sns_topic",
			Attributes: map[string]interface{}{
				"name": topicName,
				"arn":  *topic.TopicArn,
			},
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverCloudWatchResources discovers CloudWatch resources
func (c *Client) discoverCloudWatchResources() ([]Resource, error) {
	var resources []Resource

	// Discover CloudWatch alarms
	alarms, err := c.cloudwatchClient.DescribeAlarms(context.TODO(), &cloudwatch.DescribeAlarmsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe CloudWatch alarms: %w", err)
	}

	for _, alarm := range alarms.MetricAlarms {
		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "cloudwatch",
				ID:     *alarm.AlarmName,
				Name:   *alarm.AlarmName,
				Region: c.region,
				Tags:   make(map[string]string), // CloudWatch tags would need separate call
			},
			ResourceType: "aws_cloudwatch_metric_alarm",
			Attributes: map[string]interface{}{
				"alarm_name":          *alarm.AlarmName,
				"comparison_operator": alarm.ComparisonOperator,
				"evaluation_periods":  alarm.EvaluationPeriods,
				"metric_name":         alarm.MetricName,
				"namespace":           alarm.Namespace,
				"period":              alarm.Period,
				"statistic":           alarm.Statistic,
				"threshold":           alarm.Threshold,
			},
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverCloudTrailResources discovers CloudTrail resources
func (c *Client) discoverCloudTrailResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.cloudtrailClient.ListTrails(context.TODO(), &cloudtrail.ListTrailsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list CloudTrail trails: %w", err)
	}

	for _, trail := range result.Trails {
		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "cloudtrail",
				ID:     *trail.Name,
				Name:   *trail.Name,
				Region: c.region,
				Tags:   make(map[string]string), // CloudTrail tags would need separate call
			},
			ResourceType: "aws_cloudtrail",
			Attributes: map[string]interface{}{
				"name": *trail.Name,
				"s3_bucket_name": "unknown", // Would need separate API call to get S3 bucket name
			},
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "ecs",
				ID:     *service.ServiceName,
				Name:   *service.ServiceName,
				Region: c.region,
				Tags:   make(map[string]string), // ECS tags would need separate call
			},
			ResourceType: "aws_ecs_service",
			Attributes: map[string]interface{}{
				"name":            *service.ServiceName,
				"cluster":         service.ClusterArn,
				"task_definition": service.TaskDefinition,
				"desired_count":   service.DesiredCount,
			},
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "eks",
				ID:     *cluster.Name,
				Name:   *cluster.Name,
				Region: c.region,
				Tags:   cluster.Tags,
			},
			ResourceType: "aws_eks_cluster",
			Attributes: map[string]interface{}{
				"name":     *cluster.Name,
				"version":  cluster.Version,
				"platform_version": cluster.PlatformVersion,
				"endpoint": cluster.Endpoint,
			},
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "elasticache",
				ID:     *cluster.CacheClusterId,
				Name:   *cluster.CacheClusterId,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_elasticache_cluster",
			Attributes: map[string]interface{}{
				"cluster_id":           *cluster.CacheClusterId,
				"engine":               cluster.Engine,
				"node_type":            cluster.CacheNodeType,
				"num_cache_nodes":      cluster.NumCacheNodes,
				"subnet_group_name":    cluster.CacheSubnetGroupName,
			},
		}

		resources = append(resources, resource)
	}

	return resources, nil
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "dynamodb",
				ID:     *table.TableName,
				Name:   *table.TableName,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_dynamodb_table",
			Attributes: map[string]interface{}{
				"name": *table.TableName,
			},
		}

				// Add optional attributes with nil checks
		if table.BillingModeSummary != nil {
			resource.Attributes["billing_mode"] = table.BillingModeSummary.BillingMode
		}
		if table.ProvisionedThroughput != nil {
			if table.ProvisionedThroughput.ReadCapacityUnits != nil {
				resource.Attributes["read_capacity"] = *table.ProvisionedThroughput.ReadCapacityUnits
			}
			if table.ProvisionedThroughput.WriteCapacityUnits != nil {
				resource.Attributes["write_capacity"] = *table.ProvisionedThroughput.WriteCapacityUnits
			}
		}
		if len(table.KeySchema) > 0 && table.KeySchema[0].AttributeName != nil {
			resource.Attributes["hash_key"] = *table.KeySchema[0].AttributeName
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "redshift",
				ID:     *cluster.ClusterIdentifier,
				Name:   *cluster.ClusterIdentifier,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_redshift_cluster",
			Attributes: map[string]interface{}{
				"cluster_identifier":     *cluster.ClusterIdentifier,
				"node_type":              *cluster.NodeType,
				"number_of_nodes":        cluster.NumberOfNodes,
				"master_username":        *cluster.MasterUsername,
				"port":                   cluster.Endpoint.Port,
				"vpc_security_group_ids": cluster.VpcSecurityGroups,
				"cluster_subnet_group_name": cluster.ClusterSubnetGroupName,
			},
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "route53",
				ID:     *zone.Id,
				Name:   *zone.Name,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_route53_zone",
			Attributes: map[string]interface{}{
				"name": *zone.Name,
				"comment": zone.Config.Comment,
			},
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
			recordResource := &GenericResource{
				BaseResource: BaseResource{
					Type:   "route53",
					ID:     fmt.Sprintf("%s-%s", *zone.Id, *record.Name),
					Name:   *record.Name,
					Region: c.region,
					Tags:   make(map[string]string),
				},
				ResourceType: "aws_route53_record",
				Attributes: map[string]interface{}{
					"zone_id": *zone.Id,
					"name":    *record.Name,
					"type":    record.Type,
					"ttl":     record.TTL,
				},
			}

			resources = append(resources, recordResource)
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "cloudfront",
				ID:     *distribution.Id,
				Name:   *distribution.DomainName,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_cloudfront_distribution",
			Attributes: map[string]interface{}{
				"id":              *distribution.Id,
				"domain_name":     *distribution.DomainName,
				"enabled":         distribution.Enabled,
				"price_class":     distribution.PriceClass,
				"origin_domain_name": distribution.Origins.Items[0].DomainName,
			},
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "apigateway",
				ID:     *api.ApiId,
				Name:   *api.Name,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_apigatewayv2_api",
			Attributes: map[string]interface{}{
				"api_id":      *api.ApiId,
				"name":        *api.Name,
				"protocol_type": api.ProtocolType,
				"version":     api.Version,
			},
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "elasticsearch",
				ID:     *domain.DomainName,
				Name:   *domain.DomainName,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_elasticsearch_domain",
			Attributes: map[string]interface{}{
				"domain_name":           *domain.DomainName,
				"elasticsearch_version": domain.ElasticsearchVersion,
				"instance_type":         domain.ElasticsearchClusterConfig.InstanceType,
				"instance_count":        domain.ElasticsearchClusterConfig.InstanceCount,
				"vpc_options":           domain.VPCOptions != nil,
			},
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "ecr",
				ID:     *repo.RepositoryName,
				Name:   *repo.RepositoryName,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_ecr_repository",
			Attributes: map[string]interface{}{
				"name":                 *repo.RepositoryName,
				"image_tag_mutability": repo.ImageTagMutability,
				"scan_on_push":         repo.ImageScanningConfiguration.ScanOnPush,
			},
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "neptune",
				ID:     *cluster.DBClusterIdentifier,
				Name:   *cluster.DBClusterIdentifier,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_neptune_cluster",
			Attributes: map[string]interface{}{
				"cluster_identifier":     *cluster.DBClusterIdentifier,
				"engine":                 cluster.Engine,
				"engine_version":         cluster.EngineVersion,
				"availability_zones":     cluster.AvailabilityZones,
				"backup_retention_period": cluster.BackupRetentionPeriod,
				"storage_encrypted":      cluster.StorageEncrypted,
			},
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "docdb",
				ID:     *cluster.DBClusterIdentifier,
				Name:   *cluster.DBClusterIdentifier,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_docdb_cluster",
			Attributes: map[string]interface{}{
				"cluster_identifier":     *cluster.DBClusterIdentifier,
				"engine":                 cluster.Engine,
				"engine_version":         cluster.EngineVersion,
				"availability_zones":     cluster.AvailabilityZones,
				"backup_retention_period": cluster.BackupRetentionPeriod,
				"storage_encrypted":      cluster.StorageEncrypted,
			},
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "elasticbeanstalk",
				ID:     *app.ApplicationName,
				Name:   *app.ApplicationName,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_elastic_beanstalk_application",
			Attributes: map[string]interface{}{
				"name":        *app.ApplicationName,
				"description": app.Description,
			},
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "codebuild",
				ID:     *project.Name,
				Name:   *project.Name,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_codebuild_project",
			Attributes: map[string]interface{}{
				"name":           *project.Name,
				"service_role":   project.ServiceRole,
				"build_timeout":  project.TimeoutInMinutes,
				"environment":    project.Environment.Type,
				"source_type":    project.Source.Type,
			},
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "codedeploy",
				ID:     *app.ApplicationName,
				Name:   *app.ApplicationName,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_codedeploy_app",
			Attributes: map[string]interface{}{
				"name":             *app.ApplicationName,
				"compute_platform": app.ComputePlatform,
			},
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "ssm",
				ID:     *param.Name,
				Name:   *param.Name,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_ssm_parameter",
			Attributes: map[string]interface{}{
				"name":        *param.Name,
				"type":        param.Type,
				"description": param.Description,
				"tier":        param.Tier,
			},
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "secretsmanager",
				ID:     *secret.Name,
				Name:   *secret.Name,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_secretsmanager_secret",
			Attributes: map[string]interface{}{
				"name":        *secret.Name,
				"description": secret.Description,
				"kms_key_id":  secret.KmsKeyId,
			},
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "kms",
				ID:     *keyMetadata.KeyId,
				Name:   *keyMetadata.Description,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_kms_key",
			Attributes: map[string]interface{}{
				"key_id":         *keyMetadata.KeyId,
				"description":    keyMetadata.Description,
				"key_usage":      keyMetadata.KeyUsage,
				"origin":         keyMetadata.Origin,
				"customer_master_key_spec": keyMetadata.CustomerMasterKeySpec,
			},
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "codecommit",
				ID:     *repository.RepositoryName,
				Name:   *repository.RepositoryName,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_codecommit_repository",
			Attributes: map[string]interface{}{
				"repository_name": *repository.RepositoryName,
				"description":     repository.RepositoryDescription,
				"default_branch":  repository.DefaultBranch,
			},
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "codepipeline",
				ID:     *pipeline.Name,
				Name:   *pipeline.Name,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_codepipeline",
			Attributes: map[string]interface{}{
				"name": *pipeline.Name,
				"role_arn": pipelineInfo.RoleArn,
			},
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "cloudformation",
				ID:     *stack.StackName,
				Name:   *stack.StackName,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_cloudformation_stack",
			Attributes: map[string]interface{}{
				"name": *stack.StackName,
				"capabilities": stackInfo.Capabilities,
			},
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
		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "config",
				ID:     *recorder.Name,
				Name:   *recorder.Name,
				Region: c.region,
				Tags:   make(map[string]string),
			},
			ResourceType: "aws_config_configuration_recorder",
			Attributes: map[string]interface{}{
				"name": *recorder.Name,
				"role_arn": recorder.RoleARN,
			},
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "kinesis",
				ID:     streamName,
				Name:   streamName,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_kinesis_stream",
			Attributes: map[string]interface{}{
				"name":             streamName,
				"shard_count":      stream.Shards,
				"retention_period": stream.RetentionPeriodHours,
				"stream_arn":       stream.StreamARN,
			},
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "fsx",
				ID:     *filesystem.FileSystemId,
				Name:   *filesystem.FileSystemId,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_fsx_windows_file_system",
			Attributes: map[string]interface{}{
				"file_system_id": *filesystem.FileSystemId,
				"file_system_type": filesystem.FileSystemType,
				"storage_capacity": filesystem.StorageCapacity,
				"subnet_ids": filesystem.SubnetIds,
			},
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverGuardDutyResources discovers GuardDuty detector resources
func (c *Client) discoverGuardDutyResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.guarddutyClient.ListDetectors(context.TODO(), &guardduty.ListDetectorsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list GuardDuty detectors: %w", err)
	}

	for _, detectorId := range result.DetectorIds {
		// Get detector details
		detectorDetails, err := c.guarddutyClient.GetDetector(context.TODO(), &guardduty.GetDetectorInput{
			DetectorId: &detectorId,
		})
		if err != nil {
			continue
		}

		detector := detectorDetails

		// Get tags - skip for now as ARN is not directly available
		tagMap := make(map[string]string)

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "guardduty",
				ID:     detectorId,
				Name:   detectorId,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_guardduty_detector",
			Attributes: map[string]interface{}{
				"detector_id": detectorId,
				"status":      detector.Status,
			},
		}

		resources = append(resources, resource)
	}

	return resources, nil
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "backup",
				ID:     *vault.BackupVaultName,
				Name:   *vault.BackupVaultName,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_backup_vault",
			Attributes: map[string]interface{}{
				"name": *vault.BackupVaultName,
			},
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "glacier",
				ID:     *vault.VaultName,
				Name:   *vault.VaultName,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_glacier_vault",
			Attributes: map[string]interface{}{
				"name": *vault.VaultName,
			},
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "glue",
				ID:     *db.Name,
				Name:   *db.Name,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_glue_catalog_database",
			Attributes: map[string]interface{}{
				"name":        *db.Name,
				"description": db.Description,
			},
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "athena",
				ID:     *workgroup.Name,
				Name:   *workgroup.Name,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_athena_workgroup",
			Attributes: map[string]interface{}{
				"name":        *workgroup.Name,
				"description": workgroupInfo.Description,
				"state":       workgroupInfo.State,
			},
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
// discoverQuickSightResources discovers QuickSight dashboard resources
func (c *Client) discoverQuickSightResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.quicksightClient.ListDashboards(context.TODO(), &quicksight.ListDashboardsInput{
		AwsAccountId: aws.String("123456789012"), // This would need to be dynamic
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list QuickSight dashboards: %w", err)
	}

	for _, dashboard := range result.DashboardSummaryList {
		// Get tags
		tags, err := c.quicksightClient.ListTagsForResource(context.TODO(), &quicksight.ListTagsForResourceInput{
			ResourceArn: dashboard.Arn,
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "quicksight",
				ID:     *dashboard.DashboardId,
				Name:   *dashboard.Name,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_quicksight_dashboard",
			Attributes: map[string]interface{}{
				"dashboard_id": *dashboard.DashboardId,
				"name":         *dashboard.Name,
			},
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
// discoverWorkSpacesResources discovers WorkSpaces directory resources
func (c *Client) discoverWorkSpacesResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.workspacesClient.DescribeWorkspaceDirectories(context.TODO(), &workspaces.DescribeWorkspaceDirectoriesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list WorkSpaces directories: %w", err)
	}

	for _, directory := range result.Directories {
		// Get tags - skip for now as method is not available
		tagMap := make(map[string]string)

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "workspaces",
				ID:     *directory.DirectoryId,
				Name:   *directory.DirectoryName,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_workspaces_directory",
			Attributes: map[string]interface{}{
				"directory_id":   *directory.DirectoryId,
				"directory_name": *directory.DirectoryName,
				"directory_type": directory.DirectoryType,
			},
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "storagegateway",
				ID:     *gateway.GatewayARN,
				Name:   *gateway.GatewayName,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_storagegateway_gateway",
			Attributes: map[string]interface{}{
				"gateway_arn":  *gateway.GatewayARN,
				"gateway_name": *gateway.GatewayName,
				"gateway_type": gateway.GatewayType,
			},
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "transfer",
				ID:     *server.ServerId,
				Name:   *server.ServerId,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_transfer_server",
			Attributes: map[string]interface{}{
				"server_id": *server.ServerId,
				"identity_provider_type": server.IdentityProviderType,
				"endpoint_type": server.EndpointType,
			},
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "mq",
				ID:     *broker.BrokerId,
				Name:   *broker.BrokerName,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_mq_broker",
			Attributes: map[string]interface{}{
				"broker_id":   *broker.BrokerId,
				"broker_name": *broker.BrokerName,
				"engine_type": broker.EngineType,
			},
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "firehose",
				ID:     streamName,
				Name:   streamName,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_kinesis_firehose_delivery_stream",
			Attributes: map[string]interface{}{
				"name": streamName,
				"delivery_stream_type": stream.DeliveryStreamType,
				"delivery_stream_status": stream.DeliveryStreamStatus,
			},
		}

		resources = append(resources, resource)
	}

	return resources, nil
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "mediastore",
				ID:     *container.Name,
				Name:   *container.Name,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_mediastore_container",
			Attributes: map[string]interface{}{
				"name": *container.Name,
				"status": container.Status,
			},
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverMediaConvertResources discovers MediaConvert job template resources
func (c *Client) discoverMediaConvertResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.mediaconvertClient.ListJobTemplates(context.TODO(), &mediaconvert.ListJobTemplatesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list MediaConvert job templates: %w", err)
	}

	for _, template := range result.JobTemplates {
		// Get tags - skip for now as ResourceTags type is complex

		tagMap := make(map[string]string)
		// Skip tags for now as ResourceTags type is complex

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "mediaconvert",
				ID:     *template.Name,
				Name:   *template.Name,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_mediaconvert_job_template",
			Attributes: map[string]interface{}{
				"name": *template.Name,
				"category": template.Category,
			},
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "medialive",
				ID:     *channel.Id,
				Name:   *channel.Name,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_medialive_channel",
			Attributes: map[string]interface{}{
				"channel_id": *channel.Id,
				"name":       *channel.Name,
				"state":      channel.State,
			},
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverMediaTailorResources discovers MediaTailor configuration resources
func (c *Client) discoverMediaTailorResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.mediatailorClient.ListPlaybackConfigurations(context.TODO(), &mediatailor.ListPlaybackConfigurationsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list MediaTailor playback configurations: %w", err)
	}

	for _, config := range result.Items {
		// Get tags
		tags, err := c.mediatailorClient.ListTagsForResource(context.TODO(), &mediatailor.ListTagsForResourceInput{
			ResourceArn: config.PlaybackConfigurationArn,
		})
		if err != nil {
			continue
		}

		tagMap := make(map[string]string)
		for key, value := range tags.Tags {
			tagMap[key] = value
		}

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "mediatailor",
				ID:     *config.Name,
				Name:   *config.Name,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_mediatailor_playback_configuration",
			Attributes: map[string]interface{}{
				"name": *config.Name,
				"playback_endpoint_prefix": config.PlaybackEndpointPrefix,
			},
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
// discoverIoTResources discovers IoT Core resources
func (c *Client) discoverIoTResources() ([]Resource, error) {
	var resources []Resource

	// Discover IoT things
	things, err := c.iotClient.ListThings(context.TODO(), &iot.ListThingsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list IoT things: %w", err)
	}

	for _, thing := range things.Things {
		// Get thing details
		thingDetails, err := c.iotClient.DescribeThing(context.TODO(), &iot.DescribeThingInput{
			ThingName: thing.ThingName,
		})
		if err != nil {
			continue
		}

		thingInfo := thingDetails

		// Get tags
		tags, err := c.iotClient.ListTagsForResource(context.TODO(), &iot.ListTagsForResourceInput{
			ResourceArn: thingInfo.ThingArn,
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "iot",
				ID:     *thing.ThingName,
				Name:   *thing.ThingName,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_iot_thing",
			Attributes: map[string]interface{}{
				"thing_name": *thing.ThingName,
				"thing_type_name": thingInfo.ThingTypeName,
			},
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// Greengrass stub services removed

// discoverIoTAnalyticsResources discovers IoT Analytics resources
func (c *Client) discoverIoTAnalyticsResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.iotanalyticsClient.ListDatasets(context.TODO(), &iotanalytics.ListDatasetsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list IoT Analytics datasets: %w", err)
	}

	for _, dataset := range result.DatasetSummaries {
		// Get tags - skip for now as ARN is not directly available
		tagMap := make(map[string]string)

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "iotanalytics",
				ID:     *dataset.DatasetName,
				Name:   *dataset.DatasetName,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_iotanalytics_dataset",
			Attributes: map[string]interface{}{
				"dataset_name": *dataset.DatasetName,
				"status":       dataset.Status,
			},
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
// discoverIoTEventsResources discovers IoT Events resources
func (c *Client) discoverIoTEventsResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.ioteventsClient.ListDetectorModels(context.TODO(), &iotevents.ListDetectorModelsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list IoT Events detector models: %w", err)
	}

	for _, detector := range result.DetectorModelSummaries {
		// Get tags - skip for now as ARN is not directly available
		tagMap := make(map[string]string)

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "iotevents",
				ID:     *detector.DetectorModelName,
				Name:   *detector.DetectorModelName,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_iotevents_detector_model",
			Attributes: map[string]interface{}{
				"detector_model_name": *detector.DetectorModelName,
			},
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverIoTSiteWiseResources discovers IoT SiteWise resources
func (c *Client) discoverIoTSiteWiseResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.iotsitewiseClient.ListPortals(context.TODO(), &iotsitewise.ListPortalsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list IoT SiteWise portals: %w", err)
	}

	for _, portal := range result.PortalSummaries {
		// Get tags - skip for now as ARN is not directly available
		tagMap := make(map[string]string)

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "iotsitewise",
				ID:     *portal.Id,
				Name:   *portal.Name,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_iotsitewise_portal",
			Attributes: map[string]interface{}{
				"portal_id": *portal.Id,
				"name":      *portal.Name,
			},
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverIoTThingsGraphResources discovers IoT Things Graph resources
func (c *Client) discoverIoTThingsGraphResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.iotthingsgraphClient.SearchThings(context.TODO(), &iotthingsgraph.SearchThingsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list IoT Things Graph things: %w", err)
	}

	for _, thing := range result.Things {
		// Get tags
		tags, err := c.iotthingsgraphClient.ListTagsForResource(context.TODO(), &iotthingsgraph.ListTagsForResourceInput{
			MaxResults: aws.Int32(50),
			ResourceArn: thing.ThingArn,
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "iotthingsgraph",
				ID:     *thing.ThingName,
				Name:   *thing.ThingName,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_iotthingsgraph_thing",
			Attributes: map[string]interface{}{
				"thing_name": *thing.ThingName,
				"thing_arn":  *thing.ThingArn,
			},
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverIoTWirelessResources discovers IoT Wireless resources
func (c *Client) discoverIoTWirelessResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.iotwirelessClient.ListWirelessGateways(context.TODO(), &iotwireless.ListWirelessGatewaysInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list IoT Wireless gateways: %w", err)
	}

	for _, gateway := range result.WirelessGatewayList {
		// Get tags
		tags, err := c.iotwirelessClient.ListTagsForResource(context.TODO(), &iotwireless.ListTagsForResourceInput{
			ResourceArn: gateway.Arn,
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

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "iotwireless",
				ID:     *gateway.Id,
				Name:   *gateway.Name,
				Region: c.region,
				Tags:   tagMap,
			},
			ResourceType: "aws_iotwireless_wireless_gateway",
			Attributes: map[string]interface{}{
				"wireless_gateway_id": *gateway.Id,
				"name":                *gateway.Name,
				"description":         gateway.Description,
			},
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
// IoT stub services removed 