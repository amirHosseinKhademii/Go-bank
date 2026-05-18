# Kubernetes Manifests with Kustomize Overlays

This directory contains Kubernetes manifests organized with Kustomize for easy management of multiple environments.

## Directory Structure

```
k8s/
├── base/                          # Base manifests (shared across all environments)
│   ├── kustomization.yaml
│   ├── namespace.yaml
│   ├── deployment.yaml
│   └── ingress.yaml
│
└── overlays/
    ├── minikube/                  # Local development (minikube)
    │   ├── kustomization.yaml     # Patches for minikube
    │   ├── env.properties         # Environment variables (debug mode)
    │   └── db-secret.properties   # Database secrets (local postgres)
    │
    └── production/                # AWS production environment
        ├── kustomization.yaml     # Patches for AWS
        ├── env.properties         # Environment variables (release mode)
        └── db-secret.properties   # Database secrets (from CI/CD)
```

## Usage

### Deploy to Minikube (Local Development)

```bash
# Build image locally
docker build -t bank-app:local .

# Load image into minikube
minikube image load bank-app:local

# Deploy using minikube overlay
kubectl apply -k k8s/overlays/minikube

# Or use kustomize with kubectl
kustomize build k8s/overlays/minikube | kubectl apply -f -
```

### Deploy to Production (AWS EKS)

```bash
# Update secrets with AWS values
kubectl apply -k k8s/overlays/production

# Or via CI/CD (GitHub Actions handles this)
```

## Minikube Overlay Features

### Environment Optimizations
- **Single replica** (vs 3 in production)
- **Lower resource requests**: 50m CPU / 64Mi memory
- **Lower resource limits**: 250m CPU / 256Mi memory
- **Shorter startup delays**: 10s liveness, 5s readiness (vs 30s/10s)
- **Debug mode**: GIN_MODE=debug, LOG_LEVEL=debug

### Configuration
- **Database host**: `host.minikube.internal` (access host machine)
- **Database port**: 5432
- **Database user**: root
- **Database password**: admin
- **Service type**: NodePort (instead of LoadBalancer)
- **TLS**: Disabled for local testing
- **Image pull policy**: Never (use locally built images)

### Deployment
- Reduced HPA: 1-2 replicas (vs 3-10 in production)
- No ingress TLS
- NGINX ingress class
- Local service naming: `local-bank-app-service`

## Production Overlay Features

### Environment Optimizations
- **3 replicas** for high availability
- **Standard resource requests/limits** as per base manifest
- **Production startup delays**: 30s liveness, 10s readiness
- **Release mode**: GIN_MODE=release, LOG_LEVEL=info

### Configuration
- **Database host**: AWS RDS endpoint
- **Database secrets**: From CI/CD GitHub Actions
- **Service type**: LoadBalancer
- **TLS**: Enabled with proper certificates
- **Image pull policy**: IfNotPresent (from ECR)

### Deployment
- HPA: 3-10 replicas based on metrics
- Network policies enforcement
- Full security hardening

## Minikube Setup Guide

### 1. Start Minikube with Proper Resources

```bash
minikube start \
  --cpus=4 \
  --memory=4096 \
  --vm-driver=docker \
  --addons=ingress,metrics-server
```

### 2. Start Local PostgreSQL

```bash
# In a separate terminal
docker run --name postgres \
  -e POSTGRES_USER=root \
  -e POSTGRES_PASSWORD=admin \
  -p 5432:5432 \
  -d postgres:15

# Create database
docker exec -it postgres createdb -U root -d bank
```

### 3. Build and Load Image

```bash
# Build image
docker build -t bank-app:local .

# Load into minikube
minikube image load bank-app:local
```

### 4. Deploy to Minikube

```bash
# Apply overlay
kubectl apply -k k8s/overlays/minikube

# Wait for deployment
kubectl rollout status deployment/local-bank-app -n bank-app --timeout=2m
```

### 5. Access the Service

```bash
# Get service info
kubectl get svc -n bank-app

# Port forward to local machine
kubectl port-forward -n bank-app svc/local-bank-app-service 8080:8080

# Test API
curl http://localhost:8080/health/live
```

### 6. View Logs

```bash
# Follow deployment logs
kubectl logs -n bank-app -l app=bank-app -f

# Get pod details
kubectl describe pod -n bank-app <pod-name>

# Get all events
kubectl get events -n bank-app --sort-by='.lastTimestamp'
```

## Scaling in Minikube

Minikube has limited resources. If pods are pending:

```bash
# Check resource usage
kubectl top nodes
kubectl top pods -n bank-app

# Increase minikube resources
minikube stop
minikube start --cpus=6 --memory=8192

# Reduce replica count manually
kubectl patch deployment local-bank-app -n bank-app -p '{"spec":{"replicas":1}}'
```

## Updating Configuration

### Change Environment Variables

**For Minikube:**
```bash
# Edit env.properties
vim k8s/overlays/minikube/env.properties

# Reapply overlay
kubectl apply -k k8s/overlays/minikube
```

**For Production:**
```bash
# Edit env.properties
vim k8s/overlays/production/env.properties

# Commit and push (triggers CI/CD)
git add k8s/overlays/production/env.properties
git commit -m "Update production environment"
git push origin main
```

### Change Database Secrets

**For Minikube:**
```bash
# Edit db-secret.properties
vim k8s/overlays/minikube/db-secret.properties

# Reapply overlay
kubectl apply -k k8s/overlays/minikube
```

**For Production:**
Secrets are managed by GitHub Actions. Update GitHub repository secrets instead.

## Troubleshooting

### Pod stays in Pending

```bash
# Check resource requests vs available
kubectl describe node

# Check resource availability
kubectl top nodes

# Reduce resource requests in minikube overlay
vim k8s/overlays/minikube/kustomization.yaml
```

### Database connection fails

```bash
# Verify database is running
docker ps | grep postgres

# Test connectivity from pod
kubectl run -it --rm debug --image=postgres -- \
  psql -h host.minikube.internal -U root -d bank -c "SELECT 1"

# Check environment variables in pod
kubectl exec -it <pod-name> -n bank-app -- env | grep DB_
```

### Image pull fails

```bash
# For minikube, images must be loaded locally
minikube image load bank-app:local

# Verify image is loaded
minikube image ls | grep bank

# Update deployment image
kubectl patch deployment local-bank-app \
  -n bank-app \
  -p '{"spec":{"template":{"spec":{"containers":[{"name":"bank-app","image":"bank-app:local"}]}}}}'
```

### Ingress not working

```bash
# Check ingress status
kubectl get ingress -n bank-app
kubectl describe ingress -n bank-app

# Verify NGINX ingress controller is running
kubectl get pods -n ingress-nginx

# Test with port-forward instead
kubectl port-forward -n bank-app svc/local-bank-app-service 8080:8080
```

## Comparison: Minikube vs Production

| Feature | Minikube | Production |
|---------|----------|-----------|
| Replicas | 1 | 3 |
| HPA Min/Max | 1/2 | 3/10 |
| CPU Requests | 50m | 100m |
| Memory Requests | 64Mi | 128Mi |
| CPU Limits | 250m | 500m |
| Memory Limits | 256Mi | 512Mi |
| Database | Local postgres | AWS RDS |
| Service Type | NodePort | LoadBalancer |
| TLS | Disabled | Enabled |
| Startup Delays | 10s/5s | 30s/10s |
| Image Pull | Never | IfNotPresent |
| Mode | Debug | Release |

## Common Commands

```bash
# View kustomized manifests (without applying)
kustomize build k8s/overlays/minikube

# Dry-run deployment
kubectl apply -k k8s/overlays/minikube --dry-run=client -o yaml

# Check differences between overlays
diff <(kustomize build k8s/overlays/minikube) \
     <(kustomize build k8s/overlays/production)

# Watch deployment
kubectl rollout status deployment/local-bank-app -n bank-app -w

# Delete overlay deployment
kubectl delete -k k8s/overlays/minikube
```

## References

- [Kustomize Documentation](https://kubectl.docs.kubernetes.io/docs/tasks/manage-kubernetes-objects/declarative-config/)
- [Minikube Documentation](https://minikube.sigs.k8s.io/)
- [kubectl Kustomize Guide](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/)
