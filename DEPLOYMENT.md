# Bank API Deployment Guide

Complete guide for deploying the Bank API service to AWS with Kubernetes and Terraform.

## Prerequisites

### Local Development
- Go 1.25+
- Docker & Docker Buildx
- kubectl
- Terraform 1.6.0+
- PostgreSQL client (psql)
- AWS CLI v2

### AWS Account
- AWS account with appropriate permissions (EC2, EKS, RDS, ECR, IAM, VPC)
- AWS credentials configured (`aws configure` or environment variables)
- GitHub repository with repository secrets configured

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     GitHub Actions CI/CD                     │
│  (Tests → Build → Push to ECR → Terraform → Deploy to EKS)  │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                      AWS Infrastructure                      │
├───────────────────────────────────────────────────────────────┤
│  EKS Cluster (1.28)                                           │
│  ├─ Node Group: 3 t3.medium nodes (1-10 auto-scaling)       │
│  ├─ Deployment: 3 replicas with HPA                           │
│  ├─ Service: LoadBalancer type                                │
│  ├─ Ingress: Rate limiting, TLS termination                   │
│  └─ NetworkPolicy: Restrict ingress/egress                    │
│                                                               │
│  RDS PostgreSQL (Multi-AZ)                                    │
│  ├─ Version: 15.4                                             │
│  ├─ Instance: db.t3.micro                                     │
│  ├─ Storage: 20GB gp3 encrypted                               │
│  ├─ Backup: 30-day retention                                  │
│  └─ Enhanced Monitoring: Enabled                              │
│                                                               │
│  ECR Repository                                              │
│  ├─ Image Scanning: On push                                   │
│  ├─ Lifecycle: Keep 10 tagged, remove untagged after 7d      │
│  └─ Encryption: KMS                                           │
│                                                               │
│  VPC (10.0.0.0/16)                                           │
│  ├─ Public Subnets: 10.0.1.0/24, 10.0.2.0/24 (NAT gateways)│
│  ├─ Private Subnets: 10.0.10.0/24, 10.0.11.0/24            │
│  └─ Security Groups: EKS, RDS, ALB                           │
│                                                               │
│  Terraform State Management                                  │
│  ├─ Backend: S3 bucket with versioning                       │
│  ├─ Locking: DynamoDB table                                  │
│  └─ Encryption: KMS key with rotation                        │
└─────────────────────────────────────────────────────────────┘
```

## Local Development Setup

### 1. Start PostgreSQL Container
```bash
make postgres
make createdb
make migrateup
```

### 2. Run Tests
```bash
make test
```

### 3. Build and Run Locally
```bash
go build -o bank ./cmd
./bank
```

The API will be available at `http://localhost:8080`.

## AWS Deployment

### Phase 1: Infrastructure Preparation

#### 1.1 Clone and Setup
```bash
git clone <repository>
cd bank
```

#### 1.2 Validate Infrastructure Files
```bash
./scripts/validate-infrastructure.sh
```

This checks:
- Terraform syntax and formatting
- Kubernetes YAML validity
- Docker configuration
- Required tools availability

#### 1.3 Configure AWS
```bash
# Set up AWS credentials
aws configure

# Verify access
aws sts get-caller-identity
```

#### 1.4 Configure Terraform Variables
```bash
# Copy example configuration
cp terraform/terraform.tfvars.example terraform/terraform.tfvars

# Edit with your values
vim terraform/terraform.tfvars
```

Required values:
- `aws_region`: Target AWS region (default: us-east-1)
- `environment`: Environment name (default: production)
- `app_name`: Application name (default: bank-app)
- `container_image`: Will be set by CI/CD pipeline
- `db_username`: Database admin user (generate a secure password)
- `db_password`: Database password (minimum 15 characters, complex)
- `eks_desired_size`: Initial node count (default: 3)

#### 1.5 Initialize Terraform Backend
```bash
cd terraform

# Initialize with backend configuration
terraform init

# Plan infrastructure
terraform plan -out=tfplan

# Review the plan carefully
# This will create:
# - VPC with subnets and NAT gateways
# - RDS PostgreSQL instance
# - EKS cluster with node group
# - ECR repository
# - IAM roles and policies
# - Security groups
# - KMS keys and CloudWatch logs
```

#### 1.6 Apply Terraform Configuration
```bash
# Apply the planned changes
terraform apply tfplan

# Wait for EKS cluster to be ready (10-15 minutes)
# Get outputs for next steps
terraform output

# Extract kubeconfig
aws eks update-kubeconfig --name bank-app --region us-east-1
```

### Phase 2: GitHub Actions Setup

#### 2.1 Configure Repository Secrets
Add the following secrets to GitHub repository settings:

```
AWS_ACCOUNT_ID: <your-12-digit-account-id>
AWS_ROLE_TO_ASSUME: arn:aws:iam::<account-id>:role/github-actions-role
```

#### 2.2 Set Up OIDC Provider

The workflow uses OIDC federation. Create the IAM role:

```bash
# Create OIDC provider for GitHub (one-time setup)
aws iam create-open-id-connect-provider \
  --url https://token.actions.githubusercontent.com \
  --client-id-list sts.amazonaws.com \
  --thumbprint-list 6938fd4d98bab03faadb97b34396831e3780aea1 \
  2>/dev/null || echo "OIDC provider already exists"

# Create IAM role for GitHub Actions
cat > /tmp/trust-policy.json << 'EOF'
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::ACCOUNT_ID:oidc-provider/token.actions.githubusercontent.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
        },
        "StringLike": {
          "token.actions.githubusercontent.com:sub": "repo:YOUR_GITHUB_ORG/bank:ref:refs/heads/main"
        }
      }
    }
  ]
}
EOF

# Replace placeholders
sed -i 's/ACCOUNT_ID/<your-account-id>/g' /tmp/trust-policy.json
sed -i 's|YOUR_GITHUB_ORG|<your-org>|g' /tmp/trust-policy.json

# Create the role
aws iam create-role \
  --role-name github-actions-role \
  --assume-role-policy-document file:///tmp/trust-policy.json

# Attach permissions
aws iam attach-role-policy \
  --role-name github-actions-role \
  --policy-arn arn:aws:iam::aws:policy/AdministratorAccess
```

#### 2.3 Configure Repository Secrets
1. Go to GitHub repository settings → Secrets and variables → Actions
2. Add `AWS_ROLE_TO_ASSUME`: `arn:aws:iam::<account-id>:role/github-actions-role`
3. Add `AWS_ACCOUNT_ID`: Your 12-digit AWS account ID

### Phase 3: First Deployment

#### 3.1 Push Code to Trigger Pipeline
```bash
git add .
git commit -m "Deploy infrastructure"
git push origin main
```

The GitHub Actions workflow will automatically:
1. Run tests and security scans
2. Build Docker image
3. Push to ECR
4. Apply Terraform changes
5. Deploy to EKS
6. Run smoke tests

#### 3.2 Monitor Deployment
```bash
# View workflow status
gh run list

# Stream logs
gh run watch

# Or manually check deployment
kubectl get pods -n bank-app
kubectl logs -n bank-app -l app=bank-app --tail=50
```

#### 3.3 Verify Health
```bash
# Check deployment status
kubectl rollout status deployment/bank-app -n bank-app

# Get service endpoint
kubectl get svc -n bank-app

# Test API
ENDPOINT=$(kubectl get svc bank-app-service -n bank-app -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
curl http://$ENDPOINT/health/live
```

## Ongoing Operations

### Database Management

#### Connect to RDS
```bash
# Get database endpoint
DB_ENDPOINT=$(terraform output -raw rds_endpoint)

# Connect using psql
psql -h $DB_ENDPOINT -U admin -d bank

# Or through a port-forward (if database is private)
kubectl run -it --rm debug --image=postgres:15 -- \
  psql -h <rds-endpoint> -U admin -d bank
```

#### Run Migrations
Migrations run automatically in the Dockerfile during build. For manual migrations:

```bash
# Create migration
migrate create -ext sql -dir db/migration -seq <name>

# Apply migrations
make migrateup

# Rollback
make migratedown
```

### Scaling

#### Manual Scaling
```bash
# Scale replicas
kubectl scale deployment bank-app -n bank-app --replicas=5

# Check HPA status
kubectl get hpa -n bank-app
```

#### Auto-scaling Configuration
The HPA is configured to scale based on:
- CPU utilization: 70% threshold
- Memory utilization: 80% threshold
- Min replicas: 3
- Max replicas: 10

Monitor scaling events:
```bash
kubectl get hpa -n bank-app -w
kubectl describe hpa bank-app-hpa -n bank-app
```

### Monitoring

#### CloudWatch Logs
```bash
# View EKS cluster logs
aws logs tail /aws/eks/bank-app/cluster --follow

# View RDS logs
aws logs tail /aws/rds/instance/bank-app-db/postgresql --follow
```

#### RDS Alarms
CloudWatch alarms monitor:
- CPU utilization > 80%
- Database connections > 80%

View alarms:
```bash
aws cloudwatch describe-alarms --alarm-names bank-app-rds-cpu bank-app-rds-connections
```

### Updates

#### Update Code
```bash
# Make code changes
git add .
git commit -m "Update: description"
git push origin main

# GitHub Actions automatically deploys
```

#### Update Infrastructure
```bash
cd terraform

# Make changes to .tf files
# Plan changes
terraform plan

# Apply when ready
terraform apply
```

### Rollback

#### Automatic Rollback
The CD pipeline includes automatic rollback on deployment failure. If deployment fails:

```bash
# Manually rollback to previous version
kubectl rollout undo deployment/bank-app -n bank-app

# Check rollout history
kubectl rollout history deployment/bank-app -n bank-app
```

#### Manual Rollback Infrastructure
```bash
cd terraform

# View previous state
terraform state list

# Rollback to previous version
terraform destroy # Review carefully first!
```

## Security Best Practices

✅ Implemented:
- [x] Non-root container user (UID 1000)
- [x] Read-only filesystem
- [x] Pod Security Policy
- [x] Network Policy (ingress/egress restrictions)
- [x] RBAC with least-privilege roles
- [x] Encrypted secrets in ConfigMap/Secret
- [x] TLS termination at ingress
- [x] KMS encryption for RDS and EBS
- [x] S3 backend with versioning for Terraform state
- [x] OIDC federation for GitHub Actions
- [x] IMDSv2 enforcement on EC2 instances
- [x] Health checks for liveness and readiness

⚠️ Recommended:
- Set up VPN/bastion host for private database access
- Enable AWS GuardDuty for threat detection
- Implement CloudTrail for audit logging
- Set up AWS Config for compliance monitoring
- Use AWS Secrets Manager for sensitive data
- Implement API rate limiting at the ALB level
- Enable EKS audit logging to CloudWatch
- Set up CloudWatch alarms for anomalies

## Troubleshooting

### Deployment Fails
```bash
# Check logs
kubectl logs -n bank-app -l app=bank-app --tail=100

# Check pod status
kubectl describe pod -n bank-app <pod-name>

# Check events
kubectl get events -n bank-app --sort-by='.lastTimestamp'
```

### Database Connection Issues
```bash
# Verify RDS is accessible
aws rds describe-db-instances --query 'DBInstances[0].[DBInstanceIdentifier,DBInstanceStatus]'

# Check security group rules
aws ec2 describe-security-groups --filters "Name=group-name,Values=bank-app-rds"
```

### Node Issues
```bash
# Check node status
kubectl get nodes

# Describe problematic node
kubectl describe node <node-name>

# Check node logs
aws ec2 get-console-output --instance-id <instance-id>
```

## Cost Optimization

Current infrastructure estimate (monthly):
- EKS cluster: ~$73 (control plane)
- EC2 nodes (3x t3.medium): ~$40
- RDS (db.t3.micro multi-AZ): ~$60
- NAT Gateway: ~$32
- ECR storage: <$5 (100 images)
- Data transfer: ~$10

**Total: ~$220/month** (varies by region and usage)

To reduce costs:
- Use spot instances for node groups (savings up to 70%)
- Reduce desired_size to 1-2 nodes (RDS is the cost driver)
- Use db.t3.micro for non-production environments
- Enable S3 lifecycle policies for log retention

## Cleanup

To destroy all infrastructure:

```bash
# Backup database
aws rds create-db-snapshot --db-instance-identifier bank-app-db --db-snapshot-identifier bank-app-final-backup

# Destroy infrastructure
cd terraform
terraform destroy

# Confirm destruction prompts carefully
```

## Support

For issues or questions:
1. Check `/scripts/validate-infrastructure.sh` output for configuration issues
2. Review CloudWatch logs for runtime errors
3. Check GitHub Actions workflow logs for deployment issues
4. Consult AWS documentation for specific service issues
