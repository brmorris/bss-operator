# make commands

make generate
make manifests
make docker-build docker-push IMG="localhost:5000/bss-operator:v0.0.1"
make bundle IMG="localhost:5000/bss-operator:v0.0.1"
make bundle-build bundle-push BUNDLE_IMG="localhost:5000/bss-operator-bundle:v0.0.1"

run bundle:

operator-sdk run bundle localhost:5000/memcached-operator-bundle:v0.0.1


# kind

kind delete cluster && kind create cluster # optionally, --config kind-config.yml


## kind registry

The registry container is connected to the Kind network
Containerd inside the Kind cluster is configured to use HTTP for localhost:5000
TLS verification is disabled for that registry:

```
docker network connect kind $(docker ps | grep 'registry:' | awk '{print $1}')

### docker exec kind-control-plane mkdir -p /etc/containerd/certs.d/localhost:5000
docker exec kind-control-plane sh -c "cat >> /etc/containerd/config.toml << 'EOF'

        [plugins.\"io.containerd.grpc.v1.cri\".registry.mirrors.\"docker-registry-registry-1:5000\"]
          endpoint = [\"http://docker-registry-registry-1:5000\"]
        [plugins.\"io.containerd.grpc.v1.cri\".registry.configs.\"docker-registry-registry-1:5000\"]
          [plugins.\"io.containerd.grpc.v1.cri\".registry.configs.\"docker-registry-registry-1:5000\".tls]
            insecure_skip_verify = true
EOF"
docker exec kind-control-plane systemctl restart containerd
```

# deploy

make deploy

Here's what make deploy did:

1. Generated manifests (manifests dependency):

Ran controller-gen to generate RBAC roles, CRDs, and webhook configurations from your Go code
2. Set the controller image (kustomize edit):

Updated kustomization.yaml to use controller=controller:latest as the image
3. Built and applied Kubernetes manifests:

Used kustomize to build all manifests from default
Applied them to your Kind cluster with kubectl apply
What was created in your cluster:

Namespace: bss-operator-system - isolated namespace for your operator
CRD: bssclusters.bss.localhost - the custom resource definition for BssCluster
ServiceAccount: for the operator to use
RBAC:
Leader election role (for operator HA)
Admin/editor/viewer roles for BssCluster resources
Manager role for the operator itself
Metrics roles for Prometheus integration
Service: Metrics endpoint for monitoring
Deployment: The actual controller-manager pod running your operator
This is a direct deployment approach (not using OLM), which is simpler and doesn't require dealing with bundle images or private registries.

# headlamp

```
kubectl apply -f https://raw.githubusercontent.com/kinvolk/headlamp/main/kubernetes-headlamp.yaml

*NOTE*: port fwding must be per session:
kubectl port-forward -n kube-system service/headlamp 8080:80, better still: nohup kubectl port-forward -n kube-system svc/headlamp 8080:80 > /tmp/headlamp-port-forward.log 2>&1 &


kubectl -n kube-system create serviceaccount headlamp-admin
kubectl create clusterrolebinding headlamp-admin --serviceaccount=kube-system:headlamp-admin --clusterrole=cluster-admin
kubectl create token headlamp-admin -n kube-system
```


# create and watch bsscluster crd

kubectl apply -f config/samples/bss_v1alpha1_bsscluster.yaml
kubectl logs -f -n bss-operator-system deployment/bss-operator-controller-manager
kubectl get bsscluster

## argocd

### setup

```bash
# Install ArgoCD
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
kubectl wait --for=condition=available --timeout=300s deployment/argocd-server -n argocd

# Configure GitHub access (requires PAT with repo scope)
./argocd/setup-github-access.sh <GITHUB_PAT>

# Push changes and deploy app-of-apps
git add -A && git commit -m "Update" && git push origin main
kubectl apply -f argocd/app-of-apps.yaml

# Get admin password
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d && echo

# port forwarding
nohup kubectl port-forward svc/argocd-server -n argocd 8081:80 --address=0.0.0.0 > /tmp/argocd-port-forward.log 2>&1 &

# Access UI
kubectl port-forward svc/argocd-server -n argocd 8081:443
# Open https://localhost:8080 (admin / password-from-above)
```

### gitops workflow

```bash
# Make changes to code or manifests
vim config/samples/bss_v1alpha1_bsscluster.yaml

# Commit and push - ArgoCD auto-syncs
git add . && git commit -m "Update cluster config" && git push

# Check status
kubectl get applications -n argocd
argocd app sync bss-clusters  # manual sync if needed
```

# debugging notes

```
10187  kubectl get bsscluster -A
10188  kubectl get bsscluster bsscluster-sample -o yaml
10189  kubectl get pods -n bss-operator-system
10190  kubectl logs -n bss-operator-system bss-operator-controller-manager-b855f588d-qqg5x --tail=50
10191  kubectl logs -n bss-operator-system bss-operator-controller-manager-b855f588d-qqg5x --tail=100 | grep -i "reconcil\|bsscluster\|deploy\|error" | tail -20
10192  kubectl get deployment,service -n default -l app.kubernetes.io/instance=bsscluster-sample
```

# custom resources

```
# List all resources with creation timestamps
kubectl get bsscluster,bssquery -o wide

# See events related to these resources
kubectl get events --field-selector involvedObject.kind=BSSQuery
kubectl get events --field-selector involvedObject.kind=BssCluster


# portforwarding:

kubectl port-forward -n default deployment/bsscluster-sample 8880:8880 &

# Check for BssCluster resources
kubectl get bsscluster
# or short form
kubectl get bssc

# Check for BSSQuery resources
kubectl get bssquery
# or short form
kubectl get bssq

# See all at once
kubectl get bsscluster,bssquery

# Get detailed info about a specific resource
kubectl describe bsscluster bsscluster-sample
kubectl describe bssquery bssquery-sample-clusters

# Get YAML output to see full status
kubectl get bssquery bssquery-sample-clusters -o yaml

# Check the status field specifically (shows query results)
kubectl get bssquery bssquery-sample-clusters -o jsonpath='{.status}' | jq .

# Check if the samples from the files exist
kubectl get bsscluster bsscluster-sample 2>/dev/null && echo "✅ BssCluster sample exists" || echo "❌ BssCluster sample not found"
kubectl get bssquery bssquery-sample-cluster 2>/dev/null && echo "✅ BSSQuery (single) sample exists" || echo "❌ BSSQuery (single) sample not found"
kubectl get bssquery bssquery-sample-clusters 2>/dev/null && echo "✅ BSSQuery (list) sample exists" || echo "❌ BSSQuery (list) sample not found"

# List all resources with creation timestamps
kubectl get bsscluster,bssquery -o wide

# See events related to these resources
kubectl get events --field-selector involvedObject.kind=BSSQuery
kubectl get events --field-selector involvedObject.kind=BssCluster
```
