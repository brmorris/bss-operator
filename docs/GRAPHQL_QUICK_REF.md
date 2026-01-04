# GraphQL Quick Reference

## BSS API GraphQL Endpoint

**URL:** `http://localhost:8880/graphql`
**GraphiQL Playground:** Open the URL in a browser

## Quick Commands

### Start the API
```bash
cd hack/bss-api
go run main.go
```

### Test the API
```bash
./hack/bss-api/test-graphql.sh
```

### Use with Kubernetes
```bash
# Install CRD
make install

# Run controller
make run

# Create BSSQuery
kubectl apply -f config/samples/bss_v1alpha1_bssquery_clusters.yaml

# Check status
kubectl get bssquery
kubectl describe bssquery <name>
```

## GraphQL Queries

### Create a Cluster
```graphql
mutation {
  createCluster(name: "my-cluster", replicas: 3, version: "1.0.0") {
    id
    name
    state
    replicas
  }
}
```

### Get a Cluster
```graphql
query {
  cluster(id: "your-cluster-id") {
    id
    name
    state
    replicas
    readyReplicas
  }
}
```

### List All Clusters
```graphql
query {
  clusters {
    id
    name
    state
    replicas
    version
  }
}
```

### Delete a Cluster
```graphql
mutation {
  deleteCluster(id: "your-cluster-id")
}
```

## BSSQuery Examples

### List all clusters
```yaml
apiVersion: bss.localhost/v1alpha1
kind: BSSQuery
metadata:
  name: all-clusters
spec:
  apiEndpoint: "http://localhost:8880/graphql"
  query: clusters
  refreshInterval: 60
```

### Monitor a specific cluster
```yaml
apiVersion: bss.localhost/v1alpha1
kind: BSSQuery
metadata:
  name: my-cluster
spec:
  apiEndpoint: "http://localhost:8880/graphql"
  query: cluster
  clusterID: "abc-123"
  refreshInterval: 30
```

## Useful kubectl Commands

```bash
# List BSSQuery resources
kubectl get bssquery
kubectl get bssq  # short name

# Get detailed info
kubectl describe bssquery <name>

# Get JSON result
kubectl get bssquery <name> -o jsonpath='{.status.result}' | jq .

# Watch for changes
kubectl get bssquery -w

# Check conditions
kubectl get bssquery <name> -o jsonpath='{.status.conditions}' | jq .

# Delete
kubectl delete bssquery <name>
```

## Curl Examples

### Create Cluster
```bash
curl -X POST http://localhost:8880/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "mutation CreateCluster($name: String!, $replicas: Int!, $version: String!) { createCluster(name: $name, replicas: $replicas, version: $version) { id name state } }",
    "variables": {"name": "test", "replicas": 3, "version": "1.0.0"}
  }'
```

### List Clusters
```bash
curl -X POST http://localhost:8880/graphql \
  -H "Content-Type: application/json" \
  -d '{"query": "query { clusters { id name state replicas } }"}'
```

### Get Cluster
```bash
curl -X POST http://localhost:8880/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "query GetCluster($id: String!) { cluster(id: $id) { id name state } }",
    "variables": {"id": "your-cluster-id"}
  }'
```

## Documentation

- **GraphQL API**: [docs/graphql.md](./graphql.md)
- **BSSQuery CR**: [docs/bssquery.md](./bssquery.md)
- **Implementation**: [docs/GRAPHQL_IMPLEMENTATION.md](./GRAPHQL_IMPLEMENTATION.md)

## Troubleshooting

### API not responding
```bash
# Check if running
lsof -i :8880

# Check logs (if using nohup)
tail -f /tmp/bss-api.log

# Restart
cd hack/bss-api && go run main.go
```

### Controller not reconciling
```bash
# Check controller logs
kubectl logs -n bss-operator-system -l control-plane=controller-manager -f

# Or if running locally with make run, check terminal
```

### Query failing
```bash
# Check BSSQuery status
kubectl get bssquery <name> -o yaml | grep -A 20 status:

# Check conditions
kubectl describe bssquery <name>
```
