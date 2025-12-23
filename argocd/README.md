# ArgoCD Workflow for BSS Operator

## Overview

This directory contains ArgoCD Application manifests for deploying and managing the BSS Operator platform.

## Applications

1. **bss-platform** (App of Apps) - Parent application that manages all child applications
2. **bss-operator** - Deploys the operator itself (CRDs, RBAC, controller)
3. **bss-clusters** - Manages BssCluster CR instances

## Setup

### Prerequisites

- ArgoCD installed in the cluster
- This repository pushed to GitHub

### Installation

1. **Apply the App of Apps**:
   ```bash
   kubectl apply -f argocd/app-of-apps.yaml
   ```

   This will automatically create and sync all child applications.

2. **Access ArgoCD UI**:
   ```bash
   # Port-forward to ArgoCD server
   kubectl port-forward svc/argocd-server -n argocd 8080:443

   # Get admin password
   kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
   ```

   Then open: https://localhost:8080
   - Username: `admin`
   - Password: (from command above)

## Development Workflow

### GitOps-Driven Development

1. **Make changes to the code**:
   ```bash
   # Edit controller code in internal/controller/
   vim internal/controller/bsscluster_controller.go
   ```

2. **Build and push image**:
   ```bash
   make docker-build IMG=bss-operator:v0.0.2
   kind load docker-image bss-operator:v0.0.2
   ```

3. **Update kustomization**:
   ```bash
   # Update config/manager/kustomization.yaml with new tag
   sed -i 's/newTag: .*/newTag: v0.0.2/' config/manager/kustomization.yaml
   ```

4. **Commit and push**:
   ```bash
   git add .
   git commit -m "Update controller image to v0.0.2"
   git push origin main
   ```

5. **ArgoCD auto-syncs** (if automated sync is enabled) or manually sync via UI/CLI

### Managing BssCluster Instances

Add new cluster instances by creating YAML files in `config/samples/`:

```bash
# Create a new cluster manifest
cat > config/samples/bss-production.yaml <<EOF
apiVersion: bss.localhost/v1alpha1
kind: BssCluster
metadata:
  name: bss-production
spec:
  replicas: 5
  image: my-bss-app:v1.0.0
EOF

git add config/samples/bss-production.yaml
git commit -m "Add production BSS cluster"
git push
```

ArgoCD will automatically deploy the new cluster.

## Local Development

For rapid local development without pushing to Git, you can:

1. **Disable auto-sync**:
   ```bash
   kubectl patch app bss-operator -n argocd --type=json \
     -p='[{"op": "remove", "path": "/spec/syncPolicy/automated"}]'
   ```

2. **Apply changes directly**:
   ```bash
   kubectl apply -k config/default
   ```

3. **Manually sync when ready**:
   ```bash
   argocd app sync bss-operator
   ```

## Commands

```bash
# List all applications
kubectl get applications -n argocd

# Check sync status
argocd app list

# Manually sync an application
argocd app sync bss-operator

# View application details
argocd app get bss-operator

# View application diff
argocd app diff bss-operator

# Rollback to previous version
argocd app rollback bss-operator

# Delete an application (but keep resources)
argocd app delete bss-operator --cascade=false
```

## Troubleshooting

### ArgoCD can't access GitHub

If using a private repository, create a secret:

```bash
kubectl create secret generic github-creds \
  --from-literal=type=git \
  --from-literal=url=https://github.com/brmorris/logscale \
  --from-literal=password=<github-token> \
  --from-literal=username=brmorris \
  -n argocd

kubectl label secret github-creds argocd.argoproj.io/secret-type=repository -n argocd
```

### Sync issues

```bash
# Force sync (ignore differences)
argocd app sync bss-operator --force

# Prune resources not in Git
argocd app sync bss-operator --prune
```

### Image not found in kind

Make sure to load images into kind after building:

```bash
kind load docker-image bss-operator:latest
```
