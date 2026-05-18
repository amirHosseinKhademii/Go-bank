# Getting Started - Bank API

Welcome! This guide will get you started with the Bank API service. Choose your path below.

## 🚀 Quick Start (5 minutes)

### Want to try locally first?

```bash
# 1. Start PostgreSQL
make postgres
make createdb
make migrateup

# 2. Run tests
make test

# 3. Start API server
make dev

# Server runs on http://localhost:8080
```

### Want to deploy to AWS?

```bash
# 1. Validate configuration
make validate-infra

# 2. Configure AWS
cp terraform/terraform.tfvars.example terraform/terraform.tfvars
# Edit with your AWS account ID and secure passwords

# 3. Deploy infrastructure
cd terraform
terraform init
terraform plan
terraform apply

# 4. Configure GitHub Repository
# Add 2 secrets in GitHub repo settings → Secrets:
# - AWS_ACCOUNT_ID: Your 12-digit AWS account ID
# - AWS_ROLE_TO_ASSUME: arn:aws:iam::YOUR_ACCOUNT:role/github-actions-role

# 5. Push code to trigger deployment
git push origin main
```

---

## 📚 Documentation

### For Different Audiences

**I want to understand the architecture:**
→ Read [README-INFRASTRUCTURE.md](README-INFRASTRUCTURE.md)

**I want to deploy this to AWS:**
→ Follow [DEPLOYMENT.md](DEPLOYMENT.md) step-by-step

**I need a command reference:**
→ Use [QUICK-REFERENCE.md](QUICK-REFERENCE.md)

**I want a complete overview:**
→ Check [IMPLEMENTATION-SUMMARY.md](IMPLEMENTATION-SUMMARY.md)

**I want to understand the project:**
→ Review [CLAUDE.md](CLAUDE.md) (development guide)

---

## 🏗️ What's Included

### Application (Go)
- RESTful API with Gin framework
- Account, Transfer, and User management
- PostgreSQL database with pgx/v5
- 50+ unit tests with comprehensive coverage
- Error handling for database constraints

### Infrastructure
- **Kubernetes**: EKS cluster with HPA, NetworkPolicy, RBAC, Ingress
- **Database**: PostgreSQL 15.4 Multi-AZ with automated backups
- **Container Registry**: ECR with image scanning
- **State Management**: S3 backend with DynamoDB locking

### Deployment
- **CI/CD**: GitHub Actions with full automation
- **Security**: OIDC federation, KMS encryption, non-root containers
- **Monitoring**: CloudWatch logs and alarms
- **Code Quality**: Tests, linting, and security scanning

---

## 🔍 Validation Checklist

Before deploying, ensure:

```bash
# ✅ Validate all configurations
make validate-infra

# ✅ Run tests locally
make test

# ✅ Check Docker build
docker build -t bank:test .

# ✅ Verify Kubernetes manifests
kubectl apply -f k8s/ --dry-run=client

# ✅ Validate Terraform
cd terraform && terraform validate && terraform fmt -check -recursive .
```

---

## 📋 Deployment Steps

### Step 1: Local Setup (1 minute)
```bash
git clone <repository>
cd bank
```

### Step 2: AWS Configuration (2 minutes)
```bash
# Configure AWS credentials
aws configure

# Copy and edit Terraform variables
cp terraform/terraform.tfvars.example terraform/terraform.tfvars
# Edit with your AWS account ID and secure passwords
```

### Step 3: Validate Infrastructure (1 minute)
```bash
make validate-infra
```

### Step 4: Deploy Infrastructure (15 minutes)
```bash
cd terraform
terraform init
terraform plan
terraform apply
```

### Step 5: GitHub Configuration (2 minutes)
Add 2 secrets to GitHub repository:
- `AWS_ACCOUNT_ID`: Your AWS account ID
- `AWS_ROLE_TO_ASSUME`: IAM role ARN (created by Terraform)

### Step 6: Deploy Application (5 minutes)
```bash
git push origin main
# GitHub Actions automatically deploys
```

**Total time: ~30 minutes** ⏱️

---

## 🧪 Testing

### Run All Tests
```bash
make test
```

### Run Specific Test
```bash
go test -v ./api -run TestCreateAccount
```

### Test Coverage
```bash
go test -cover ./...
```

---

## 📝 Common Tasks

### View Application Logs
```bash
kubectl logs -n bank-app -l app=bank-app --tail=50
```

### Scale Deployment
```bash
kubectl scale deployment bank-app -n bank-app --replicas=5
```

### Connect to Database
```bash
psql -h <rds-endpoint> -U admin -d bank
```

### Rollback Deployment
```bash
kubectl rollout undo deployment/bank-app -n bank-app
```

---

## 🆘 Troubleshooting

### Pod won't start
```bash
kubectl describe pod -n bank-app <pod-name>
kubectl logs -n bank-app <pod-name>
```

### Database connection fails
```bash
# Check security group allows RDS access
aws ec2 describe-security-groups --filters "Name=group-name,Values=bank-app-rds"
```

### GitHub Actions fails
- Check workflow logs: Go to repository → Actions
- Verify AWS secrets are configured
- Ensure Terraform can access AWS

---

## 💾 Key Files

### Application
- `cmd/main.go` — Application entry point
- `api/*.go` — HTTP handlers (accounts, transfers, users)
- `db/sqlc/` — Generated database code (auto-generated, do not edit)
- `db/query/` — SQL queries with sqlc annotations
- `db/migration/` — Database migrations

### Infrastructure
- `k8s/*.yaml` — Kubernetes manifests (3 files)
- `terraform/*.tf` — Infrastructure as code (6 files)
- `.github/workflows/cd.yml` — CI/CD pipeline

### Documentation
- `DEPLOYMENT.md` — Detailed deployment guide
- `README-INFRASTRUCTURE.md` — Architecture overview
- `QUICK-REFERENCE.md` — Command reference
- `IMPLEMENTATION-SUMMARY.md` — Complete overview

---

## 🔒 Security Highlights

✅ **11 security features implemented:**
- Non-root container user
- Read-only filesystem
- Pod Security Policy
- Network policies
- RBAC with least privileges
- Encrypted database (RDS)
- KMS encryption for storage
- TLS termination at ingress
- OIDC federation (no credentials in code)
- Health checks for pod monitoring
- Audit logging enabled

---

## 📊 Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     GitHub Repository                        │
│  (On push) → GitHub Actions CI/CD Pipeline                  │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                      AWS Infrastructure                      │
├─────────────────────────────────────────────────────────────┤
│  EKS Cluster (Kubernetes 1.28)                              │
│  ├─ 3 Bank API pods (auto-scales to 10)                     │
│  ├─ LoadBalancer service (external access)                  │
│  ├─ Ingress with TLS and rate limiting                      │
│  └─ NetworkPolicy restricting traffic                       │
│                                                              │
│  RDS PostgreSQL (Multi-AZ)                                  │
│  ├─ 20GB storage (encrypted)                                │
│  ├─ 30-day automated backups                                │
│  └─ CloudWatch monitoring                                   │
│                                                              │
│  ECR Repository                                             │
│  └─ Docker images with vulnerability scanning              │
│                                                              │
│  S3 + DynamoDB                                              │
│  └─ Terraform state management                              │
└─────────────────────────────────────────────────────────────┘
```

---

## 🎯 Next Steps

1. **Read the right guide** for your use case (see Documentation section above)
2. **Validate configurations** with `make validate-infra`
3. **Follow deployment steps** for AWS or run locally
4. **Monitor deployment** via GitHub Actions or `kubectl`
5. **Review logs** if anything goes wrong

---

## 💬 Need Help?

1. **Check QUICK-REFERENCE.md** for commands
2. **Review DEPLOYMENT.md** Troubleshooting section
3. **Look at logs**: `kubectl logs -n bank-app -l app=bank-app`
4. **Verify configuration**: `make validate-infra`

---

## 📈 Cost Estimate

**Monthly AWS bill (~$220):**
- EKS control plane: $73
- EC2 nodes (3): $40
- RDS database: $60
- NAT gateway: $32
- Other services: <$15

*Can be reduced 70% with spot instances*

---

## ✅ Deployment Checklist

Before going live:

- [ ] Read DEPLOYMENT.md
- [ ] Verified AWS credentials with `aws sts get-caller-identity`
- [ ] Created terraform/terraform.tfvars from example
- [ ] Ran `make validate-infra` successfully
- [ ] Deployed infrastructure with `terraform apply`
- [ ] Added AWS secrets to GitHub
- [ ] Pushed code to trigger first deployment
- [ ] Verified pods are running: `kubectl get pods -n bank-app`
- [ ] Checked logs for errors: `kubectl logs -n bank-app -f`
- [ ] Tested API health: `curl http://your-endpoint/health/live`

---

## 🎉 Success!

Your Bank API is now deployed and ready to use. 

**Continue learning:**
- Explore [QUICK-REFERENCE.md](QUICK-REFERENCE.md) for useful commands
- Review [README-INFRASTRUCTURE.md](README-INFRASTRUCTURE.md) for architecture details
- Check [DEPLOYMENT.md](DEPLOYMENT.md) for operational procedures

**Monitor your deployment:**
```bash
# Watch deployment status
kubectl rollout status deployment/bank-app -n bank-app -w

# Follow logs
kubectl logs -n bank-app -l app=bank-app -f

# Check HPA scaling
kubectl get hpa -n bank-app -w
```

**Get the service endpoint:**
```bash
kubectl get svc -n bank-app
```

Happy deploying! 🚀
