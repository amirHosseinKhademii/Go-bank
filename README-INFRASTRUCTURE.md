# Bank API - Infrastructure & Deployment

Production-ready infrastructure for the Bank API service with Kubernetes, Terraform, and GitHub Actions CI/CD.

## Quick Start

### Local Development
```bash
# Start dependencies
make postgres
make createdb
make migrateup

# Run tests
make test

# Start API server
make dev
```

### Validate Infrastructure Files
```bash
make validate-infra
```

### Deploy to AWS

See [DEPLOYMENT.md](DEPLOYMENT.md) for comprehensive deployment steps.

```bash
# 1. Prepare Terraform variables
cp terraform/terraform.tfvars.example terraform/terraform.tfvars
vim terraform/terraform.tfvars

# 2. Deploy infrastructure
cd terraform
terraform init
terraform plan
terraform apply

# 3. Push code to trigger GitHub Actions CD
git push origin main
```

## Architecture

### Components

#### Application (Go)
- HTTP API server with Gin framework
- PostgreSQL database access layer using pgx/v5
- Account, Transfer, and User management
- Comprehensive test coverage

#### Docker
- Multi-stage build with test execution
- Alpine-based runtime (3.19)
- Non-root user for security
- Health checks for Kubernetes probes

#### Kubernetes (EKS)
- **Deployment**: 3 replicas with RollingUpdate strategy
- **HPA**: Auto-scales 3-10 replicas based on CPU/memory
- **Service**: LoadBalancer for external access
- **Ingress**: TLS termination, rate limiting, host routing
- **NetworkPolicy**: Restrict pod-to-pod communication
- **RBAC**: Service accounts and roles with least privileges
- **ConfigMaps & Secrets**: Environment and sensitive data

#### Database (RDS)
- PostgreSQL 15.4 on db.t3.micro
- Multi-AZ deployment for high availability
- Automated backups (30-day retention)
- Encryption at rest with KMS
- Enhanced monitoring enabled

#### Container Registry (ECR)
- Private Docker image repository
- Automated vulnerability scanning on push
- Lifecycle policy (keep 10 tagged images)
- KMS encryption

#### Infrastructure State (Terraform)
- S3 backend with versioning
- DynamoDB table for state locking
- KMS key rotation enabled
- Comprehensive outputs for integration

#### CI/CD (GitHub Actions)
- Automated tests and security scanning
- Docker build and ECR push
- Terraform plan and apply
- EKS deployment with health checks
- Automatic rollback on failure

## Directory Structure

```
bank/
├── api/                          # HTTP handlers
│   ├── account.go               # Account endpoints
│   ├── account_test.go          # Account tests
│   ├── transfer.go              # Transfer endpoints
│   ├── transfer_test.go         # Transfer tests
│   ├── user.go                  # User endpoints
│   └── *_test.go                # Handler tests
├── cmd/                         # Application entry points
│   └── main.go
├── db/
│   ├── migration/               # Database migrations
│   │   └── 000001_init_schema.*
│   ├── query/                   # sqlc queries
│   │   ├── account.sql
│   │   ├── entry.sql
│   │   ├── transfer.sql
│   │   └── user.sql
│   └── sqlc/                    # Generated Go code (do not edit)
│       ├── db.go
│       ├── models.go
│       ├── account.sql.go
│       └── *_test.go
├── utils/                       # Utility functions
│   ├── password.go             # Password hashing
│   └── password_test.go        # Password tests
├── .github/
│   └── workflows/
│       └── cd.yml              # GitHub Actions CI/CD pipeline
├── k8s/                        # Kubernetes manifests
│   ├── namespace.yaml          # Namespace, ConfigMap, Secret
│   ├── deployment.yaml         # Deployment, HPA, Service
│   └── ingress.yaml            # Ingress, NetworkPolicy, RBAC
├── terraform/                  # Infrastructure as Code
│   ├── main.tf                # Provider, backend, state setup
│   ├── vpc.tf                 # VPC, subnets, gateways
│   ├── eks.tf                 # EKS cluster, nodes, ECR
│   ├── rds.tf                 # PostgreSQL instance
│   ├── variables.tf           # Input variables
│   ├── outputs.tf             # Outputs for integration
│   └── terraform.tfvars.example
├── scripts/
│   └── validate-infrastructure.sh  # Configuration validation
├── Dockerfile                 # Multi-stage Docker build
├── docker-compose.yml         # Local dev dependencies
├── Makefile                   # Common tasks
├── go.mod / go.sum           # Go dependencies
├── sqlc.yaml                 # sqlc configuration
├── DEPLOYMENT.md             # Detailed deployment guide
├── README.md                 # Application documentation
└── README-INFRASTRUCTURE.md  # This file
```

## Configuration

### Environment Variables

Set in `.env` for local development:
```env
DB_DRIVER=postgresql
DB_SOURCE=postgresql://root:admin@localhost:5432/bank?sslmode=disable
ANTHROPIC_API_KEY=your-api-key
GIN_MODE=debug
```

Kubernetes deployment uses ConfigMap (namespace.yaml):
- `DB_DRIVER`: postgresql
- `GIN_MODE`: release

Database credentials stored in Secret:
- `DB_USERNAME`: admin
- `DB_PASSWORD`: *(set during Terraform apply)*
- `API_KEY`: *(optional, for AI services)*

### Terraform Variables

Key variables in `terraform/terraform.tfvars`:
```hcl
aws_region           = "us-east-1"
environment          = "production"
app_name             = "bank-app"
eks_desired_size     = 3
eks_min_size         = 1
eks_max_size         = 10
instance_types       = ["t3.medium"]
db_username          = "admin"
db_password          = "your-secure-password"
enable_logging       = true
```

## Deployment Workflow

### 1. Local Development & Testing
- Make changes to application code
- Run `make test` to verify changes
- Test locally with `make dev`

### 2. Build & Push
- On push to main branch, GitHub Actions:
  - Runs go test + security scans
  - Builds Docker image using Dockerfile
  - Pushes image to ECR
  - Scans image for vulnerabilities

### 3. Infrastructure Update
- Terraform plan step validates changes
- Terraform apply step updates infrastructure
- Database migrations run as part of deployment

### 4. Deployment to EKS
- Updates deployment with new image URI
- Monitors rollout status
- Runs health checks on new pods
- Automatic rollback if deployment fails

### 5. Post-Deployment
- Smoke tests verify API health
- CloudWatch alarms monitor performance
- Logs and metrics aggregated for debugging

## Security Features

✅ **Implemented**
- Non-root container user (UID 1000)
- Read-only root filesystem in containers
- Pod Security Policy enforcement
- Network policies restricting pod communication
- RBAC with least-privilege principles
- Encrypted secrets in Kubernetes
- TLS termination at ingress layer
- KMS encryption for RDS and EBS volumes
- S3 state backend with versioning
- OIDC federation for GitHub Actions
- IMDSv2 enforcement on EC2 instances
- Health checks for pod liveness/readiness

⚠️ **Recommended**
- Set up AWS VPN or bastion host for database access
- Enable GuardDuty for threat detection
- Configure CloudTrail for audit logging
- Use AWS Secrets Manager instead of Kubernetes secrets
- Implement API authentication/authorization
- Enable rate limiting at ALB level
- Set up EKS audit logging to CloudWatch
- Configure CloudWatch anomaly detection

## Monitoring & Observability

### Logs
- **EKS cluster**: CloudWatch Logs (with KMS encryption)
- **Pod logs**: `kubectl logs -n bank-app <pod-name>`
- **RDS logs**: PostgreSQL logs exported to CloudWatch

### Metrics
- **CPU/Memory**: CloudWatch | Kubernetes metrics
- **RDS**: CPU utilization, database connections
- **ECR**: Image scan results, lifecycle events

### Alarms
- RDS CPU utilization > 80%
- Database connections > 80%
- Deployment replica failures
- Pod restart loops

### Health Checks
- Liveness probe: `/health/live` (30s initial delay)
- Readiness probe: `/health/ready` (10s initial delay)
- Kubernetes monitors and auto-restarts unhealthy pods

## Scaling

### Horizontal Pod Autoscaling
The HPA automatically scales replicas based on:
- **CPU threshold**: 70% utilization
- **Memory threshold**: 80% utilization
- **Min replicas**: 3
- **Max replicas**: 10

Monitor with:
```bash
kubectl get hpa -n bank-app -w
kubectl describe hpa bank-app-hpa -n bank-app
```

### Node Group Scaling
EKS automatically scales node group based on resource requests:
```bash
# Check node group status
aws eks describe-nodegroup \
  --cluster-name bank-app \
  --nodegroup-name bank-app-node-group
```

## Backup & Disaster Recovery

### Database Backups
- Automated backups: Daily (30-day retention)
- Manual backup before major changes:
```bash
aws rds create-db-snapshot \
  --db-instance-identifier bank-app-db \
  --db-snapshot-identifier bank-app-backup-$(date +%Y%m%d)
```

### Terraform State
- S3 backend with object versioning
- DynamoDB locking prevents concurrent modifications
- KMS encryption protects state file

### Container Images
- ECR lifecycle policy keeps 10 tagged versions
- Earlier versions can be restored via kubectl rollout history

## Cost Optimization

Current monthly cost estimate:
- EKS control plane: ~$73
- EC2 nodes (3x t3.medium): ~$40
- RDS (db.t3.micro multi-AZ): ~$60
- NAT Gateway: ~$32
- ECR: <$5
- Data transfer: ~$10
- **Total: ~$220/month**

To reduce costs:
1. Use spot instances (`mixed_instances_policy` in Terraform)
2. Reduce `eks_desired_size` to 1-2 nodes
3. Downgrade RDS instance for non-production
4. Enable S3 lifecycle for old backups
5. Consolidate CloudWatch log retention

## Troubleshooting

### Pod not starting
```bash
kubectl describe pod -n bank-app <pod-name>
kubectl logs -n bank-app <pod-name>
```

### Database connection fails
```bash
# Check RDS status
aws rds describe-db-instances --query 'DBInstances[0].DBInstanceStatus'

# Check security group rules
aws ec2 describe-security-groups \
  --filters "Name=group-name,Values=bank-app-rds" \
  --query 'SecurityGroups[0].IpPermissions'
```

### Terraform apply fails
```bash
# Check state lock
aws dynamodb scan --table-name terraform-lock

# Remove stuck lock if necessary
aws dynamodb delete-item \
  --table-name terraform-lock \
  --key '{"LockID": {"S": "bank/terraform.tfstate"}}'
```

### GitHub Actions workflow fails
- Check workflow logs in GitHub repository
- Verify secrets are configured correctly
- Ensure IAM role has required permissions
- Check Terraform plan output for configuration errors

## Important Notes

⚠️ **Before Production**
- [ ] Change all default passwords
- [ ] Set up proper AWS credential management (not IAM user keys)
- [ ] Enable CloudTrail for audit logging
- [ ] Set up VPN/bastion for database access
- [ ] Configure proper backup retention policy
- [ ] Set up monitoring and alerting
- [ ] Enable AWS Config for compliance
- [ ] Review and customize security groups
- [ ] Set up disaster recovery procedures
- [ ] Test failure scenarios and recovery

⚠️ **Database**
- The schema has intentional misspellings (`entires` table, `ammount` column) that are source control. Models use correct names.
- Migrations run automatically in CI/CD. Manual migrations rarely needed.
- Always backup before applying migrations to production.

⚠️ **Kubernetes**
- NetworkPolicy restricts ingress to port 8080 and egress to RDS (5432), external APIs (443/80), and DNS (53)
- RBAC uses service account with minimal permissions
- Pods run as non-root user (UID 1000)

## Next Steps

1. **Configure AWS credentials**: `aws configure`
2. **Set up Terraform variables**: `cp terraform/terraform.tfvars.example terraform/terraform.tfvars` and edit
3. **Validate configuration**: `make validate-infra`
4. **Deploy infrastructure**: `cd terraform && terraform init && terraform apply`
5. **Configure GitHub**: Add AWS_ACCOUNT_ID and AWS_ROLE_TO_ASSUME secrets
6. **Push code**: `git push origin main` to trigger CI/CD
7. **Monitor deployment**: Check GitHub Actions workflow and EKS cluster

For detailed steps, see [DEPLOYMENT.md](DEPLOYMENT.md).
