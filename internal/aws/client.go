package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// Client represents an AWS client for resource discovery
type Client struct {
	region           string
	cfg              aws.Config
	ec2Client        *ec2.Client
	iamClient        *iam.Client
	rdsClient        *rds.Client
	s3Client         *s3.Client
	elbv2Client      *elasticloadbalancingv2.Client
	asgClient        *autoscaling.Client
	lambdaClient     *lambda.Client
	sqsClient        *sqs.Client
	snsClient        *sns.Client
	cloudwatchClient *cloudwatch.Client
	cloudtrailClient *cloudtrail.Client
	ecsClient        *ecs.Client
	eksClient        *eks.Client
}

// NewClient creates a new AWS client
func NewClient(region string) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &Client{
		region:           region,
		cfg:              cfg,
		ec2Client:        ec2.NewFromConfig(cfg),
		iamClient:        iam.NewFromConfig(cfg),
		rdsClient:        rds.NewFromConfig(cfg),
		s3Client:         s3.NewFromConfig(cfg),
		elbv2Client:      elasticloadbalancingv2.NewFromConfig(cfg),
		asgClient:        autoscaling.NewFromConfig(cfg),
		lambdaClient:     lambda.NewFromConfig(cfg),
		sqsClient:        sqs.NewFromConfig(cfg),
		snsClient:        sns.NewFromConfig(cfg),
		cloudwatchClient: cloudwatch.NewFromConfig(cfg),
		cloudtrailClient: cloudtrail.NewFromConfig(cfg),
		ecsClient:        ecs.NewFromConfig(cfg),
		eksClient:        eks.NewFromConfig(cfg),
	}, nil
}

// DiscoverResources discovers AWS resources of the specified types
func (c *Client) DiscoverResources(resourceTypes []string) ([]Resource, error) {
	var allResources []Resource

	for _, resourceType := range resourceTypes {
		resources, err := c.discoverResourceType(resourceType)
		if err != nil {
			return nil, fmt.Errorf("failed to discover %s resources: %w", resourceType, err)
		}
		allResources = append(allResources, resources...)
	}

	return allResources, nil
}

// discoverResourceType discovers resources of a specific type
func (c *Client) discoverResourceType(resourceType string) ([]Resource, error) {
	switch resourceType {
	case "vpc":
		return c.discoverVPCResources()
	case "ec2":
		return c.discoverEC2Resources()
	case "iam":
		return c.discoverIAMResources()
	case "rds":
		return c.discoverRDSResources()
	case "s3":
		return c.discoverS3Resources()
	case "alb", "elb":
		return c.discoverLoadBalancerResources()
	case "asg":
		return c.discoverAutoScalingResources()
	case "lambda":
		return c.discoverLambdaResources()
	case "sqs":
		return c.discoverSQSResources()
	case "sns":
		return c.discoverSNSResources()
	case "cloudwatch":
		return c.discoverCloudWatchResources()
	case "cloudtrail":
		return c.discoverCloudTrailResources()
	case "ecs":
		return c.discoverECSResources()
	case "eks":
		return c.discoverEKSResources()
	case "elasticache":
		return c.discoverElastiCacheResources()
	case "dynamodb":
		return c.discoverDynamoDBResources()
	case "redshift":
		return c.discoverRedshiftResources()
	case "route53":
		return c.discoverRoute53Resources()
	case "cloudfront":
		return c.discoverCloudFrontResources()
	case "apigateway":
		return c.discoverAPIGatewayResources()
	case "elasticsearch":
		return c.discoverElasticsearchResources()
	case "neptune":
		return c.discoverNeptuneResources()
	case "docdb":
		return c.discoverDocDBResources()
	case "elasticbeanstalk":
		return c.discoverElasticBeanstalkResources()
	case "ecr":
		return c.discoverECRResources()
	case "codecommit":
		return c.discoverCodeCommitResources()
	case "codebuild":
		return c.discoverCodeBuildResources()
	case "codedeploy":
		return c.discoverCodeDeployResources()
	case "codepipeline":
		return c.discoverCodePipelineResources()
	case "cloudformation":
		return c.discoverCloudFormationResources()
	case "ssm":
		return c.discoverSSMResources()
	case "secretsmanager":
		return c.discoverSecretsManagerResources()
	case "kms":
		return c.discoverKMSResources()
	case "guardduty":
		return c.discoverGuardDutyResources()
	case "config":
		return c.discoverConfigResources()
	case "backup":
		return c.discoverBackupResources()
	case "glacier":
		return c.discoverGlacierResources()
	case "glue":
		return c.discoverGlueResources()
	case "athena":
		return c.discoverAthenaResources()
	case "quicksight":
		return c.discoverQuickSightResources()
	case "workspaces":
		return c.discoverWorkSpacesResources()
	case "directoryservice":
		return c.discoverDirectoryServiceResources()
	case "fsx":
		return c.discoverFSxResources()
	case "storagegateway":
		return c.discoverStorageGatewayResources()
	case "transfer":
		return c.discoverTransferResources()
	case "mq":
		return c.discoverMQResources()
	case "kinesis":
		return c.discoverKinesisResources()
	case "firehose":
		return c.discoverFirehoseResources()
	case "mediastore":
		return c.discoverMediaStoreResources()
	case "mediaconvert":
		return c.discoverMediaConvertResources()
	case "medialive":
		return c.discoverMediaLiveResources()
	case "mediatailor":
		return c.discoverMediaTailorResources()
	case "iot":
		return c.discoverIoTResources()
	case "greengrass":
		return c.discoverGreengrassResources()
	case "greengrassv2":
		return c.discoverGreengrassV2Resources()
	case "iotanalytics":
		return c.discoverIoTAnalyticsResources()
	case "iotevents":
		return c.discoverIoTEventsResources()
	case "iotsitewise":
		return c.discoverIoTSiteWiseResources()
	case "iotthingsgraph":
		return c.discoverIoTThingsGraphResources()
	case "iotwireless":
		return c.discoverIoTWirelessResources()
	case "iotdeviceadvisor":
		return c.discoverIoTDeviceAdvisorResources()
	case "iotfleethub":
		return c.discoverIoTFleetHubResources()
	case "iotsecuretunneling":
		return c.discoverIoTSecureTunnelingResources()
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

// GetAllSupportedResources returns all supported resource types
func GetAllSupportedResources() []string {
	return []string{
		"vpc", "ec2", "iam", "rds", "s3", "alb", "elb", "asg", "cloudwatch", "cloudtrail", "ecs", "eks", "lambda", "sqs", "sns",
		"elasticache", "redshift", "route53", "cloudfront", "apigateway", "dynamodb", "elasticsearch", "neptune", "docdb",
		"elasticbeanstalk", "ecr", "codecommit", "codebuild", "codedeploy", "codepipeline", "cloudformation", "ssm",
		"secretsmanager", "kms", "guardduty", "config", "backup", "glacier", "glue", "athena", "quicksight", "workspaces",
		"directoryservice", "fsx", "storagegateway", "transfer", "mq", "kinesis", "firehose", "mediastore", "mediaconvert",
		"medialive", "mediatailor", "iot", "greengrass", "greengrassv2", "iotanalytics", "iotevents", "iotsitewise",
		"iotthingsgraph", "iotwireless", "iotdeviceadvisor", "iotfleethub", "iotsecuretunneling",
	}
} 