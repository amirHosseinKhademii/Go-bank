# Quick Reference Guide

Common commands for development, testing, and deployment.

## Local Development

```bash
# Start PostgreSQL container
make postgres

# Create database
make createdb

# Run migrations
make migrateup

# Drop database
make dropdb

# Run all tests
make test

# Run specific test
go test -v ./db/sqlc -run TestCreateAccount

# Start development server
make dev

# Build binary
make build

# Run linter
make lint
```

## Database Operations

```bash
# Run migrations
make migrateup
make migratedown
make migrateup1    # Only up 1
make migratedown1  # Only down 1

# Connect to local database
psql -U root -d bank -h localhost

# Regenerate Go code from SQL
make sqlc
```

## Docker

```bash
# Build image
docker build -t bank:latest .

# Run container
docker run -p 8080:8080 \
  -e DB_SOURCE=postgresql://... \
  bank:latest

# View image layers
docker history bank:latest

# Scan for vulnerabilities
docker scout cves bank:latest
```

## Kubernetes - Local Testing

```bash
# Apply Kubernetes manifests (requires K8s cluster)
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/ingress.yaml

# Check deployment status
kubectl get deployments -n bank-app
kubectl rollout status deployment/bank-app -n bank-app

# View pod logs
kubectl logs -n bank-app -l app=bank-app

# Port forward to service
kubectl port-forward -n bank-app svc/bank-app-service 8080:8080

# Delete all resources
kubectl delete -f k8s/
```

## Terraform

```bash
# Initialize Terraform
cd terraform
terraform init

# Validate configuration
terraform validate

# Format files
terraform fmt -recursive .

# Plan changes
terraform plan -out=tfplan

# Apply changes
terraform apply tfplan

# Destroy infrastructure
terraform destroy

# Check outputs
terraform output

# Get specific output
terraform output -raw eks_cluster_name

# View state
terraform state list
terraform state show aws_eks_cluster.main

# Refresh state
terraform refresh
```

## AWS

```bash
# Configure credentials
aws configure

# Verify access
aws sts get-caller-identity

# Get RDS endpoint
aws rds describe-db-instances \
  --query 'DBInstances[0].Endpoint.Address' --output text

# Get EKS cluster info
aws eks describe-cluster --name bank-app

# Update kubeconfig
aws eks update-kubeconfig --name bank-app --region us-east-1

# View ECR repositories
aws ecr describe-repositories

# List ECR images
aws ecr describe-images --repository-name bank-app

# View CloudWatch logs
aws logs tail /aws/eks/bank-app/cluster --follow
aws logs tail /aws/rds/instance/bank-app-db/postgresql --follow

# Check alarms
aws cloudwatch describe-alarms --alarm-names bank-app-rds-cpu
```

## GitHub Actions

```bash
# View workflow status
gh run list

# Watch workflow in progress
gh run watch

# View workflow logs
gh run view <run-id>

# Re-run workflow
gh run rerun <run-id>

# View action logs
gh run view <run-id> -j <job-id>
```

## Kubernetes Operations

```bash
# Get all resources
kubectl get all -n bank-app

# Get pod details
kubectl describe pod -n bank-app <pod-name>

# Get logs
kubectl logs -n bank-app <pod-name>
kubectl logs -n bank-app -l app=bank-app --tail=100
kubectl logs -n bank-app -f <pod-name>  # Follow logs

# Execute command on pod
kubectl exec -it -n bank-app <pod-name> -- /bin/sh

# Scale deployment
kubectl scale deployment bank-app -n bank-app --replicas=5

# Update image
kubectl set image deployment/bank-app \
  bank-app=<new-image> -n bank-app

# Rollout status
kubectl rollout status deployment/bank-app -n bank-app

# Rollout history
kubectl rollout history deployment/bank-app -n bank-app

# Rollback to previous version
kubectl rollout undo deployment/bank-app -n bank-app

# Rollback to specific revision
kubectl rollout undo deployment/bank-app --to-revision=2 -n bank-app

# Get HPA status
kubectl get hpa -n bank-app
kubectl describe hpa bank-app-hpa -n bank-app

# Get events
kubectl get events -n bank-app --sort-by='.lastTimestamp'

# Check node status
kubectl get nodes
kubectl describe node <node-name>

# Check service endpoints
kubectl get svc -n bank-app
kubectl get endpoints -n bank-app

# Port forward
kubectl port-forward -n bank-app pod/<pod-name> 8080:8080
kubectl port-forward -n bank-app svc/bank-app-service 8080:8080
```

## Validation

```bash
# Validate all infrastructure files
make validate-infra

# Validate Kubernetes manifests
kubectl apply -f k8s/ --dry-run=client

# Validate Terraform
terraform validate

# Run tests and security checks
make test
golangci-lint run ./...
gosec ./...
```

## Debugging

```bash
# Check pod logs for errors
kubectl logs -n bank-app -l app=bank-app | grep -i error

# Debug pod connectivity
kubectl run -it --rm debug --image=busybox -- /bin/sh
# Inside shell: ping <service-name>.bank-app.svc.cluster.local

# Check DNS
kubectl run -it --rm debug --image=busybox -- nslookup bank-app-service.bank-app

# Connect to database from pod
kubectl run -it --rm debug --image=postgres -- \
  psql -h <rds-endpoint> -U admin -d bank

# View pod events
kubectl describe pod -n bank-app <pod-name> | grep -A 20 "Events:"

# Check ingress configuration
kubectl get ingress -n bank-app
kubectl describe ingress -n bank-app

# Check network policies
kubectl get networkpolicies -n bank-app
kubectl describe networkpolicy -n bank-app <policy-name>
```

## CI/CD Pipeline

### On Local Machine
```bash
# Run tests (same as CI)
make test

# Run linting (same as CI)
golangci-lint run ./...

# Run security scanner (same as CI)
gosec -fmt json -out gosec-results.json ./...

# Build Docker image (same as CI)
docker build -t bank:local .

# Dry-run deployment manifests
kubectl apply -f k8s/ --dry-run=client -o yaml
```

### On GitHub
```bash
# Push code to trigger workflow
git push origin main

# View workflow in GitHub
# Visit: github.com/your-repo/actions

# Monitor deployment
gh run watch

# Check specific job logs
gh run view <run-id> -j <job-id>
```

## Common Tasks

### Deploy New Version
```bash
# 1. Make code changes
git add .
git commit -m "feat: update API"

# 2. Push to trigger CI/CD
git push origin main

# 3. Monitor deployment
gh run watch

# 4. Verify health
kubectl rollout status deployment/bank-app -n bank-app
```

### Scale Deployment
```bash
# Manual scale
kubectl scale deployment bank-app -n bank-app --replicas=5

# Check HPA status
kubectl get hpa -n bank-app
```

### Update Infrastructure
```bash
# 1. Make Terraform changes
vim terraform/variables.tf
vim terraform/main.tf

# 2. Plan and review
terraform plan

# 3. Apply changes
terraform apply

# 4. Or let CI/CD handle it
git add terraform/
git commit -m "infra: update configuration"
git push origin main
```

### Connect to Database
```bash
# From local machine (if network allows)
psql -h <rds-endpoint> -U admin -d bank

# From Kubernetes pod
kubectl run -it --rm debug --image=postgres -- \
  psql -h <rds-endpoint> -U admin -d bank
```

### View Application Logs
```bash
# Current logs
kubectl logs -n bank-app -l app=bank-app

# Last 100 lines
kubectl logs -n bank-app -l app=bank-app --tail=100

# Follow logs (like tail -f)
kubectl logs -n bank-app -l app=bank-app -f

# Logs from all containers in pod
kubectl logs -n bank-app <pod-name> --all-containers=true

# Logs from specific timestamp
kubectl logs -n bank-app -l app=bank-app --since=5m
```

### Rollback Deployment
```bash
# Immediate rollback
kubectl rollout undo deployment/bank-app -n bank-app

# View available versions
kubectl rollout history deployment/bank-app -n bank-app

# Rollback to specific version
kubectl rollout undo deployment/bank-app --to-revision=2 -n bank-app
```

### Debug Network Issues
```bash
# From inside cluster
kubectl run -it --rm debug --image=nicolaka/netshoot -- /bin/bash

# Common network tests
nslookup bank-app-service.bank-app
ping bank-app-service.bank-app
curl http://bank-app-service.bank-app:8080/health/live
```

## Environment Setup

```bash
# First-time setup
git clone <repository>
cd bank

# Copy environment sample
cp .env.example .env
# Edit .env with your values

# Set up development environment
make postgres
make createdb
make migrateup

# Prepare Terraform
cp terraform/terraform.tfvars.example terraform/terraform.tfvars
# Edit terraform/terraform.tfvars with AWS details

# Validate setup
make validate-infra
make test
```

## Monitoring

```bash
# Pod metrics (requires metrics-server)
kubectl top pod -n bank-app
kubectl top nodes

# Pod resource requests
kubectl describe pod -n bank-app <pod-name> | grep -A 5 "Limits\|Requests"

# HPA metrics
kubectl get hpa -n bank-app
watch kubectl get hpa -n bank-app  # Auto-refresh

# Check pod restarts
kubectl get pods -n bank-app -o jsonpath='{.items[*].status.containerStatuses[*].restartCount}'

# View CloudWatch metrics
aws cloudwatch get-metric-statistics \
  --metric-name CPUUtilization \
  --namespace AWS/RDS \
  --dimensions Name=DBInstanceIdentifier,Value=bank-app-db \
  --start-time $(date -u -d '1 hour ago' +%Y-%m-%dT%H:%M:%S) \
  --end-time $(date -u +%Y-%m-%dT%H:%M:%S) \
  --period 300 \
  --statistics Average
```

## Security

```bash
# Scan images for vulnerabilities
docker scout cves bank:latest

# Run security linter
gosec ./...

# Check secrets (should not be in code)
git log -p | grep -i "password\|secret\|key" OR
git show <commit> | grep -i "password"

# Verify RBAC permissions
kubectl auth can-i get pods -n bank-app

# Check pod security policies
kubectl get psp
kubectl describe psp restricted

# View network policies
kubectl get networkpolicies -n bank-app
kubectl describe networkpolicy -n bank-app
```

## Cleanup

```bash
# Stop local containers
docker stop postgres
docker rm postgres

# Delete local database
make dropdb

# Remove Kubernetes resources
kubectl delete namespace bank-app  # or
kubectl delete -f k8s/

# Destroy AWS infrastructure
terraform destroy

# Clean Docker images
docker rmi bank:latest
docker system prune -a
```
