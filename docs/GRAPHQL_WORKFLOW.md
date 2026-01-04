# Complete GraphQL Integration Workflow

This document provides a step-by-step guide to using the new GraphQL functionality.

## Prerequisites

- Go 1.22+
- kubectl configured
- Kubernetes cluster (or kind)

## Step 1: Start the BSS API

Open a terminal and start the BSS API server:

```bash
cd hack/bss-api
go run main.go
```

You should see:
```
2026/01/04 20:13:35 BSS API listening on :8880
2026/01/04 20:13:35 REST API: http://localhost:8880/api/v1/clusters
2026/01/04 20:13:35 GraphQL: http://localhost:8880/graphql
```

**Tip:** Open http://localhost:8880/graphql in your browser to use the GraphiQL playground!

## Step 2: Test GraphQL API Directly

In a new terminal, test the GraphQL API:

```bash
cd hack/bss-api
./test-graphql.sh
```

You should see all tests pass ✅

Or test manually with curl:

```bash
# Create a cluster
curl -X POST http://localhost:8880/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "mutation { createCluster(name: \"demo\", replicas: 3, version: \"1.0.0\") { id name state replicas } }"
  }'

# List all clusters
curl -X POST http://localhost:8880/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "query { clusters { id name state replicas readyReplicas } }"
  }'
```

## Step 3: Install the BSSQuery CRD

In a new terminal:

```bash
cd /path/to/bss-operator
make install
```

Verify the CRD was created:

```bash
kubectl get crd bssqueries.bss.localhost
```

## Step 4: Start the Operator

In the same terminal as Step 3:

```bash
make run
```

You should see the operator start and the BSSQuery controller initialize.

## Step 5: Create a BSSQuery Resource

In a new terminal, create a BSSQuery to list all clusters:

```bash
kubectl apply -f config/samples/bss_v1alpha1_bssquery_clusters.yaml
```

Or create one manually:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: bss.localhost/v1alpha1
kind: BSSQuery
metadata:
  name: monitor-clusters
  namespace: default
spec:
  apiEndpoint: "http://localhost:8880/graphql"
  query: clusters
  refreshInterval: 30
EOF
```

## Step 6: Monitor the BSSQuery

Watch the BSSQuery resource:

```bash
# List all BSSQueries
kubectl get bssquery

# Get detailed information
kubectl describe bssquery monitor-clusters

# Watch for updates
kubectl get bssquery -w
```

View the query results:

```bash
# Get the JSON result
kubectl get bssquery monitor-clusters -o jsonpath='{.status.result}' | jq .

# Get cluster count
kubectl get bssquery monitor-clusters -o jsonpath='{.status.clusterCount}'

# Check last query time
kubectl get bssquery monitor-clusters -o jsonpath='{.status.lastQueryTime}'
```

Check conditions:

```bash
kubectl get bssquery monitor-clusters -o jsonpath='{.status.conditions}' | jq .
```

## Step 7: Create a Query for a Specific Cluster

First, get a cluster ID from the previous results:

```bash
CLUSTER_ID=$(kubectl get bssquery monitor-clusters -o jsonpath='{.status.result}' | jq -r '.[0].id')
echo "Using cluster ID: ${CLUSTER_ID}"
```

Create a BSSQuery for that specific cluster:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: bss.localhost/v1alpha1
kind: BSSQuery
metadata:
  name: monitor-specific-cluster
  namespace: default
spec:
  apiEndpoint: "http://localhost:8880/graphql"
  query: cluster
  clusterID: "${CLUSTER_ID}"
  refreshInterval: 15
EOF
```

Monitor it:

```bash
kubectl get bssquery monitor-specific-cluster -o yaml
```

## Step 8: Interactive Demo

Run the complete interactive demo:

```bash
./hack/demo-bssquery.sh
```

This will:
1. Create test clusters via GraphQL
2. Create BSSQuery resources
3. Show status and results
4. Watch for updates (15 seconds)
5. Clean up everything

## Step 9: Experiment with GraphiQL

Open http://localhost:8880/graphql in your browser to use the GraphiQL playground.

Try these queries in the GraphiQL interface:

### Query 1: List all clusters
```graphql
query {
  clusters {
    id
    name
    replicas
    version
    state
    readyReplicas
    createdAt
    lastUpdateTime
  }
}
```

### Query 2: Create a new cluster
```graphql
mutation {
  createCluster(name: "graphiql-test", replicas: 5, version: "2.0.0") {
    id
    name
    state
    replicas
  }
}
```

### Query 3: Get the cluster you just created
(Replace `YOUR_CLUSTER_ID` with the ID from the mutation response)
```graphql
query {
  cluster(id: "YOUR_CLUSTER_ID") {
    id
    name
    state
    replicas
    readyReplicas
  }
}
```

### Query 4: Delete the cluster
```graphql
mutation {
  deleteCluster(id: "YOUR_CLUSTER_ID")
}
```

## Step 10: Clean Up

Stop the operator (Ctrl+C in the terminal running `make run`)

Delete the BSSQuery resources:

```bash
kubectl delete bssquery --all
```

Uninstall the CRD (optional):

```bash
make uninstall
```

Stop the BSS API (Ctrl+C in the terminal running the API)

## Using in Your Own Controllers

You can use the GraphQL client in your own controllers:

```go
import (
    bssclient "github.com/brmorris/bss-operator/internal/client"
)

func myReconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // Create a GraphQL client
    client := bssclient.NewGraphQLClient("http://localhost:8880/graphql")

    // List all clusters
    clusters, err := client.ListClusters()
    if err != nil {
        return ctrl.Result{}, err
    }

    // Use the cluster data
    for _, cluster := range clusters {
        log.Info("Found cluster", "id", cluster.ID, "name", cluster.Name, "state", cluster.State)
    }

    // Get a specific cluster
    cluster, err := client.GetCluster("some-id")
    if err != nil {
        return ctrl.Result{}, err
    }

    // Create a new cluster
    newCluster, err := client.CreateCluster("new-cluster", 3, "1.0.0")
    if err != nil {
        return ctrl.Result{}, err
    }

    return ctrl.Result{}, nil
}
```

## Advanced Usage

### Multiple API Endpoints

You can create multiple BSSQuery resources pointing to different API endpoints:

```yaml
---
apiVersion: bss.localhost/v1alpha1
kind: BSSQuery
metadata:
  name: prod-clusters
spec:
  apiEndpoint: "http://bss-api.production.svc.cluster.local/graphql"
  query: clusters
  refreshInterval: 60
---
apiVersion: bss.localhost/v1alpha1
kind: BSSQuery
metadata:
  name: staging-clusters
spec:
  apiEndpoint: "http://bss-api.staging.svc.cluster.local/graphql"
  query: clusters
  refreshInterval: 120
```

### Using with External Monitoring

Export BSSQuery status to monitoring systems:

```bash
# Export cluster count as a metric
kubectl get bssquery -o json | jq -r '.items[] | "\(.metadata.name) \(.status.clusterCount)"'

# Export last query time
kubectl get bssquery -o json | jq -r '.items[] | "\(.metadata.name) \(.status.lastQueryTime)"'

# Check if queries are succeeding
kubectl get bssquery -o json | jq -r '.items[] | select(.status.conditions[]? | select(.type=="Degraded" and .status=="True")) | .metadata.name'
```

### Automated Workflows

Use BSSQuery in GitOps workflows:

1. Commit BSSQuery resources to Git
2. ArgoCD/Flux syncs them to clusters
3. Operators automatically query BSS API
4. Other controllers consume the BSSQuery status

## Troubleshooting

### "connection refused" errors

Make sure the BSS API is running:
```bash
lsof -i :8880
```

### BSSQuery status not updating

Check controller logs:
```bash
# If using make run, check the terminal output

# If deployed to cluster:
kubectl logs -n bss-operator-system -l control-plane=controller-manager
```

### Invalid ClusterID error

Verify the cluster exists:
```bash
curl -X POST http://localhost:8880/graphql \
  -H "Content-Type: application/json" \
  -d '{"query": "query { clusters { id name } }"}'
```

## Next Steps

- Read the [GraphQL API documentation](./graphql.md)
- Read the [BSSQuery user guide](./bssquery.md)
- Check the [implementation details](./GRAPHQL_IMPLEMENTATION.md)
- Review the [quick reference](./GRAPHQL_QUICK_REF.md)
- Explore creating your own controllers that consume BSSQuery resources
