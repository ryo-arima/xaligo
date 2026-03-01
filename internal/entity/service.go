package entity

import "strings"

// ServiceEntry represents a single AWS service in a list or catalog.
type ServiceEntry struct {
	CatalogID    int
	OfficialName string
	Abbreviation string
	Summary      string
	Usage        string
	Note         string
}

// itemAbbreviations maps the service name (after stripping "Amazon "/"AWS " prefix)
// to its well-known abbreviation. Entries here are the single source of truth shared
// by both the render path (scene.go) and the add-service path (add.go).
var itemAbbreviations = map[string]string{
	// Networking & Content Delivery
	"CloudFront":                     "CF",
	"Route 53":                       "R53",
	"Virtual Private Cloud":          "VPC",
	"Elastic Load Balancing":         "ELB",
	"VPC Internet Gateway":           "IGW",
	"VPC NAT Gateway":                "NATGW",
	"App Mesh":                       "AppMesh",
	"Private 5G":                     "P5G",
	"Direct Connect":                 "DX",
	"API Gateway":                    "APIGW",
	"Transit Gateway":                "TGW",
	"Global Accelerator":             "GA",
	"PrivateLink":                    "PL",
	// Compute
	"EC2":                            "EC2",
	"EC2 Auto Scaling":               "ASG",
	"Lambda":                         "Lambda",
	"Elastic Container Service":      "ECS",
	"Elastic Kubernetes Service":     "EKS",
	"Fargate":                        "Fargate",
	"Elastic Beanstalk":              "EB",
	"Batch":                          "Batch",
	// Storage
	"Simple Storage Service":         "S3",
	"Elastic File System":            "EFS",
	"S3 Glacier":                     "Glacier",
	"Storage Gateway":                "SGW",
	"Backup":                         "Backup",
	// Database
	"RDS":                            "RDS",
	"DynamoDB":                       "DDB",
	"ElastiCache":                    "EC",
	"Aurora":                         "Aurora",
	"Redshift":                       "RS",
	"Neptune":                        "Neptune",
	"DocumentDB":                     "DocDB",
	"QLDB":                           "QLDB",
	// Analytics
	"Kinesis":                        "Kinesis",
	"Athena":                         "Athena",
	"Glue":                           "Glue",
	"EMR":                            "EMR",
	"OpenSearch Service":             "OSS",
	"QuickSight":                     "QS",
	"Lake Formation":                 "LF",
	"MSK":                            "MSK",
	// Application Integration
	"Simple Queue Service":           "SQS",
	"Simple Notification Service":    "SNS",
	"EventBridge":                    "EB",
	"Step Functions":                 "SF",
	"MQ":                             "MQ",
	"AppSync":                        "AppSync",
	// Management & Governance
	"CloudWatch":                     "CW",
	"CloudFormation":                 "CFn",
	"CloudTrail":                     "CT",
	"Systems Manager":                "SSM",
	"Organizations":                  "Orgs",
	"Control Tower":                  "CT",
	"Service Catalog":                "SC",
	"Trusted Advisor":                "TA",
	// Security, Identity & Compliance
	"Identity and Access Management": "IAM",
	"Cognito":                        "Cognito",
	"Secrets Manager":                "SM",
	"Key Management Service":         "KMS",
	"Certificate Manager":            "ACM",
	"WAF":                            "WAF",
	"Shield":                         "Shield",
	"GuardDuty":                      "GD",
	"Security Hub":                   "SH",
	"Macie":                          "Macie",
	// Developer Tools
	"CodeDeploy":                     "CD",
	"CodePipeline":                   "CP",
	"CodeBuild":                      "CB",
	"CodeCommit":                     "CC",
	"CodeArtifact":                   "CA",
	"CodeStar":                       "CS",
	// Machine Learning
	"SageMaker":                      "SM",
	"Rekognition":                    "Rekog",
	"Bedrock":                        "Bedrock",
	// Containers
	"Elastic Container Registry":     "ECR",
	// Migration
	"Database Migration Service":     "DMS",
	"DataSync":                       "DS",
	"Transfer Family":                "TF",
	// End User Computing
	"WorkSpaces Family":              "WorkSpaces",
	"AppStream 2":                    "AppStream",
}

// ItemShortName returns a compact abbreviation for an AWS service name.
// It strips the "Amazon " or "AWS " prefix, then looks up the result in the
// abbreviations table. If no entry is found, the prefix-stripped name is returned.
//
// This is the single canonical abbreviation function used by both the
// render path (scene.go) and the add-service path (add.go).
func ItemShortName(name string) string {
	short := name
	for _, pfx := range []string{"Amazon ", "AWS "} {
		if strings.HasPrefix(name, pfx) {
			short = name[len(pfx):]
			break
		}
	}
	if abbr, ok := itemAbbreviations[short]; ok {
		return abbr
	}
	return short
}

// ShortLabel returns the abbreviation if set; otherwise delegates to ItemShortName
// so that the same abbreviation table is used consistently across all rendering paths.
func (s ServiceEntry) ShortLabel() string {
	if s.Abbreviation != "" {
		return s.Abbreviation
	}
	return ItemShortName(s.OfficialName)
}
