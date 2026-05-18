# Local Testing with Minikube

Complete guide for testing the Bank API service locally using Minikube.

## Prerequisites

### Install Required Tools

```bash
# Install Minikube (macOS)
brew install minikube

# Install kubectl
brew install kubectl

# Install Docker (required for minikube docker driver)
brew install docker

# Or use Docker Desktop for macOS
# https://www.docker.com/products/docker-desktop
```

### Verify Installation

```bash
minikube version
kubectl version --client
docker --version
```

## Quick Start (5 minutes)

### 1. Start Minikube

```bash
minikube start --cpus=4 --memory=4096 --addons=ingress,metrics-server
```

### 2. Start Local PostgreSQL

```bash
# In a separate terminal
docker run --name postgres \
  -e POSTGRES_USER=root \
  -e POSTGRES_PASSWORD=admin \
  -p 5432:5432 \
  -d postgres:15

# Wait a moment for postgres to start
sleep 10

# Create bank database
docker exec -it postgres createdb -U root -d bank

# Run migrations
make migrateup
```

### 3. Build Docker Image

```bash
# Build image with local tag
docker build -t bank-app:local .

# Load into minikube
minikube image load bank-app:local
```

### 4. Deploy to Minikube

```bash
# Apply the minikube overlay
kubectl apply -k k8s/overlays/minikube

# Wait for rollout
kubectl rollout status deployment/local-bank-app -n bank-app --timeout=2m
```

### 5. Test the API

```bash
# Port-forward service
kubectl port-forward -n bank-app svc/local-bank-app-service 8080:8080 &

# In another terminal, test the API
curl http://localhost:8080/health/live

# Should return: {"status":"alive"}
```

## Step-by-Step Guide

### Step 1: Configure Minikube

```bash
# Start with recommended settings for bank-app
minikube start \
  --cpus=4 \
  --memory=4096 \
  --disk-size=20gb \
  --vm-driver=docker \
  --addons=ingress,metrics-server \
  --kubernetes-version=v1.28.0

# Verify cluster is running
minikube status

# Get cluster info
kubectl cluster-info
```

### Step 2: Set up Local PostgreSQL

```bash
# Start PostgreSQL container
docker run --name postgres \
  -e POSTGRES_USER=root \
  -e POSTGRES_PASSWORD=admin \
  -e POSTGRES_DB=bank \
  -p 5432:5432 \
  -d postgres:15

# Verify container is running
docker ps | grep postgres

# Test connection
psql -h localhost -U root -d bank -c "SELECT version();"

# Run migrations
make migrateup

# Verify schema was created
docker exec -it postgres psql -U root -d bank -c "\dt"
```

### Step 3: Build Application Image

```bash
# Build image optimized for local testing
docker build -t bank-app:local .

# Verify image was built
docker images | grep bank-app

# Load image into minikube
minikube image load bank-app:local

# Verify image is available in minikube
minikube image ls | grep bank
```

### Step 4: Deploy to Minikube

```bash
# Create namespace and deploy
kubectl apply -k k8s/overlays/minikube

# Wait a bit for resources to be created
sleep 5

# Check namespace
kubectl get namespace bank-app

# Check deployment
kubectl get deployment -n bank-app

# Check pods (should be 1 replica)
kubectl get pods -n bank-app

# Check services
kubectl get svc -n bank-app
```

### Step 5: Monitor Deployment

```bash
# Watch deployment status
kubectl rollout status deployment/local-bank-app -n bank-app -w

# Check logs
kubectl logs -n bank-app -l app=bank-app -f

# Describe pod (if issues)
kubectl describe pod -n bank-app local-bank-app-xxxxx
```

### Step 6: Access the Service

```bash
# Option 1: Port-forward to local machine
kubectl port-forward -n bank-app svc/local-bank-app-service 8080:8080 &

# Test API
curl -s http://localhost:8080/health/live | jq .

# Option 2: Get minikube IP and NodePort
MINIKUBE_IP=$(minikube ip)
NODE_PORT=$(kubectl get svc -n bank-app local-bank-app-service -o jsonpath='{.spec.ports[0].nodePort}')
echo "Service available at: http://$MINIKUBE_IP:$NODE_PORT"
curl -s http://$MINIKUBE_IP:$NODE_PORT/health/live | jq .

# Option 3: Use minikube service command
minikube service local-bank-app-service -n bank-app
```

## Testing Workflows

### Test 1: Health Checks

```bash
# Port-forward
kubectl port-forward -n bank-app svc/local-bank-app-service 8080:8080 &

# Test health endpoints
curl http://localhost:8080/health/live
curl http://localhost:8080/health/ready

# Expected: HTTP 200 OK
```

### Test 2: Create User

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice",
    "password": "SecurePassword123"
  }' | jq .
```

### Test 3: Create Account

```bash
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{
    "owner": "alice",
    "currency": "USD",
    "initial_balance": 1000
  }' | jq .
```

### Test 4: List Accounts

```bash
curl http://localhost:8080/accounts?page_id=1&page_size=10 | jq .
```

### Test 5: Create Transfer

```bash
# First create two users and accounts
# Then test transfer between accounts

curl -X POST http://localhost:8080/transfers \
  -H "Content-Type: application/json" \
  -d '{
    "from_account_id": 1,
    "to_account_id": 2,
    "amount": 100
  }' | jq .
```

## View Logs and Metrics

### Application Logs

```bash
# Stream logs
kubectl logs -n bank-app -l app=bank-app -f

# View logs from all pods
kubectl logs -n bank-app -l app=bank-app --all-containers=true

# View logs with timestamps
kubectl logs -n bank-app -l app=bank-app -f --timestamps=true

# View last 100 lines
kubectl logs -n bank-app -l app=bank-app --tail=100
```

### Resource Metrics

```bash
# Pod resource usage
kubectl top pod -n bank-app

# Node resource usage
kubectl top nodes

# HPA status
kubectl get hpa -n bank-app
kubectl describe hpa -n bank-app bank-app-hpa
```

### Debugging Commands

```bash
# Describe pod
kubectl describe pod -n bank-app local-bank-app-xxxxx

# Get pod events
kubectl get events -n bank-app --sort-by='.lastTimestamp'

# Execute command in pod
kubectl exec -it -n bank-app local-bank-app-xxxxx -- /bin/sh

# Port-forward for debugging
kubectl port-forward -n bank-app local-bank-app-xxxxx 2000:2000
```

## Cleaning Up

### Stop Services

```bash
# Kill port-forward processes
pkill -f "kubectl port-forward"

# Stop minikube
minikube stop

# Stop PostgreSQL
docker stop postgres
docker rm postgres
```

### Full Cleanup

```bash
# Delete all minikube resources
kubectl delete -k k8s/overlays/minikube

# Delete namespace
kubectl delete namespace bank-app

# Stop minikube
minikube stop

# Delete minikube cluster
minikube delete

# Remove Docker containers
docker stop postgres
docker rm postgres
docker rmi bank-app:local
```

## Troubleshooting

### Minikube won't start

```bash
# Check Docker is running
docker ps

# Check minikube status
minikube status

# If stuck, delete and recreate
minikube delete
minikube start --cpus=4 --memory=4096

# Check logs
minikube logs
```

### Pod won't start / ImagePullBackOff

```bash
# Image must be loaded into minikube before deploying
minikube image load bank-app:local

# Verify image is loaded
minikube image ls | grep bank

# Rebuild and reload image if necessary
docker build -t bank-app:local .
minikube image load bank-app:local

# Restart pod to pull new image
kubectl rollout restart deployment/local-bank-app -n bank-app
```

### Database connection fails

```bash
# Verify PostgreSQL is running
docker ps | grep postgres

# Test connectivity from minikube pod
kubectl run -it --rm debug --image=postgres -- \
  psql -h host.minikube.internal -U root -d bank -c "SELECT 1"

# Check environment variables in pod
kubectl exec -it -n bank-app local-bank-app-xxxxx -- env | grep DB_

# Inspect pod logs for DB errors
kubectl logs -n bank-app local-bank-app-xxxxx | grep -i "database\|connection"
```

### Out of resources

```bash
# Check node capacity
kubectl describe node minikube

# Check pod resource requests
kubectl describe pod -n bank-app local-bank-app-xxxxx | grep -A 5 "Requests"

# Increase minikube resources
minikube stop
minikube start --cpus=6 --memory=8192

# Or reduce pod resource requests in minikube overlay
# Edit: k8s/overlays/minikube/kustomization.yaml
# Modify resource requests/limits
```

### Port-forward not working

```bash
# Use minikube service instead
minikube service local-bank-app-service -n bank-app

# Or use NodePort directly
MINIKUBE_IP=$(minikube ip)
NODE_PORT=$(kubectl get svc -n bank-app local-bank-app-service -o jsonpath='{.spec.ports[0].nodePort}')
curl http://$MINIKUBE_IP:$NODE_PORT/health/live
```

### Logs show "CrashLoopBackOff"

```bash
# Check pod status
kubectl describe pod -n bank-app local-bank-app-xxxxx

# View logs
kubectl logs -n bank-app local-bank-app-xxxxx

# Common issues:
# 1. Database connection failed - check DB is running
# 2. Image not found - reload image into minikube
# 3. Port already in use - change port in deployment
# 4. Missing environment variables - check ConfigMap
```

## Configuration Changes

### Update Environment Variables (Minikube)

```bash
# Edit minikube environment
vim k8s/overlays/minikube/env.properties

# Apply changes
kubectl apply -k k8s/overlays/minikube

# Restart deployment to apply
kubectl rollout restart deployment/local-bank-app -n bank-app
```

### Update Database Credentials (Minikube)

```bash
# Edit database secret
vim k8s/overlays/minikube/db-secret.properties

# Apply changes
kubectl apply -k k8s/overlays/minikube

# Restart deployment
kubectl rollout restart deployment/local-bank-app -n bank-app
```

### Change Replica Count

```bash
# Manually scale
kubectl scale deployment/local-bank-app -n bank-app --replicas=2

# Or edit overlay
vim k8s/overlays/minikube/kustomization.yaml
# Change: replicas: count: 2

kubectl apply -k k8s/overlays/minikube
```

## Performance Tips

### Reduce Resource Usage

```bash
# Edit minikube overlay kustomization.yaml
# Reduce resource requests/limits:
# - CPU requests: 50m → 25m
# - Memory requests: 64Mi → 32Mi
# - CPU limits: 250m → 100m
# - Memory limits: 256Mi → 128Mi

kubectl apply -k k8s/overlays/minikube
```

### Speed Up Image Loading

```bash
# Use eval for Docker daemon inside minikube
eval $(minikube docker-env)

# Then build image directly in minikube (no need to load)
docker build -t bank-app:local .
```

## Next Steps

1. **Modify code** and rebuild image
2. **Reload image** into minikube
3. **Restart deployment** to test changes
4. **Review logs** for any errors
5. **When ready**, push to GitHub to trigger AWS deployment

## Useful Resources

- [Minikube Documentation](https://minikube.sigs.k8s.io/)
- [kubectl Cheat Sheet](https://kubernetes.io/docs/reference/kubectl/cheatsheet/)
- [Kustomize Tutorial](https://kubectl.docs.kubernetes.io/guides/kustomizing/)
- [PostgreSQL Docker Hub](https://hub.docker.com/_/postgres)
