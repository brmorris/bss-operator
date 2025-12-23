# bss-operator

Kubernetes operator for managing BSS (Brad's Super Simple) clusters using StatefulSets.

## Description

Implements a custom `BssCluster` CRD that deploys and manages stateful application clusters with headless services. Built with Kubebuilder and managed via ArgoCD GitOps.

## Getting Started

### Prerequisites
- Go 1.24+, Docker, kubectl
- kind: `kind create cluster`

### Local Development

```sh
# Build and load into kind
make docker-build IMG=bss-operator:latest
kind load docker-image bss-operator:latest

# Deploy
make deploy
kubectl apply -f config/samples/bss_v1alpha1_bsscluster.yaml

# Watch
kubectl logs -f -n bss-operator-system deployment/bss-operator-controller-manager
```

### GitOps with ArgoCD

```sh
# Setup (one-time)
./argocd/setup-github-access.sh <GITHUB_PAT>
kubectl apply -f argocd/app-of-apps.yaml

# Workflow: edit → commit → push (ArgoCD auto-syncs)
vim config/samples/bss_v1alpha1_bsscluster.yaml
git add . && git commit -m "Update" && git push
```

### BssCluster Example

```yaml
apiVersion: bss.localhost/v1alpha1
kind: BssCluster
metadata:
  name: my-cluster
spec:
  replicas: 3
  image: nginx:latest  # required
```

Creates StatefulSet + headless Service (port 8080).

### Development Commands

```sh
make generate manifests  # after API changes
make test
kubectl get bsscluster
```

See [docs/command_reference.md](docs/command_reference.md) and [argocd/README.md](argocd/README.md) for details.
