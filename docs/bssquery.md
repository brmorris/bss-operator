# BSSQuery Custom Resource

The BSSQuery custom resource allows you to query the BSS API GraphQL endpoint from within Kubernetes and store the results in the resource status.

## Overview

BSSQuery provides a declarative way to:
- Query BSS cluster information via GraphQL
- Automatically refresh query results at configurable intervals
- Monitor query status through Kubernetes conditions
- Access query results through the resource status

## Quick Start

### 1. Start the BSS API

```bash
cd hack/bss-api
go run main.go
```

The API will be available at:
- GraphQL: http://localhost:8880/graphql
- REST: http://localhost:8880/api/v1/clusters

### 2. Install the CRD

```bash
make install
```

### 3. Run the operator

```bash
make run
```

### 4. Create a BSSQuery

List all clusters:
```bash
kubectl apply -f config/samples/bss_v1alpha1_bssquery_clusters.yaml
```

Query a specific cluster:
```bash
# First, get a cluster ID from the API
CLUSTER_ID=$(curl -s -X POST http://localhost:8880/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"mutation { createCluster(name: \"test\", replicas: 3, version: \"1.0.0\") { id } }"}' \
  | grep -oP '"id"\s*:\s*"\K[^"]+')

# Create the BSSQuery
cat <<EOF | kubectl apply -f -
apiVersion: bss.localhost/v1alpha1
kind: BSSQuery
metadata:
  name: my-cluster
spec:
  apiEndpoint: "http://localhost:8880/graphql"
  query: cluster
  clusterID: "${CLUSTER_ID}"
  refreshInterval: 30
EOF
```

### 5. Check the results

```bash
# List all BSSQuery resources
kubectl get bssquery

# View detailed status
kubectl describe bssquery my-cluster

# Get the JSON result
kubectl get bssquery my-cluster -o jsonpath='{.status.result}' | jq .
```

## API Reference

### BSSQuerySpec

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `apiEndpoint` | string | Yes | URL of the BSS API GraphQL endpoint |
| `query` | BSSQueryType | Yes | Type of query: `cluster` or `clusters` |
| `clusterID` | string | Conditional | Required when `query` is `cluster` |
| `refreshInterval` | int32 | No | How often to refresh results (seconds), default: 30 |

### BSSQueryType

- `cluster`: Query a single cluster by ID
- `clusters`: Query all clusters

### BSSQueryStatus

| Field | Type | Description |
|-------|------|-------------|
| `lastQueryTime` | *metav1.Time | Timestamp of the last successful query |
| `result` | string | JSON-encoded result from the GraphQL query |
| `clusterCount` | int | Number of clusters in the result |
| `conditions` | []metav1.Condition | Standard Kubernetes conditions |
| `observedGeneration` | int64 | Generation of the last reconciled spec |

### Conditions

- **Available**: Query is executing successfully
- **Degraded**: Query is failing or configuration is invalid

## Examples

### List All Clusters

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

### Monitor a Specific Cluster

```yaml
apiVersion: bss.localhost/v1alpha1
kind: BSSQuery
metadata:
  name: prod-cluster
spec:
  apiEndpoint: "http://bss-api.production.svc.cluster.local/graphql"
  query: cluster
  clusterID: "abc-123-def-456"
  refreshInterval: 15
```

### Status Example

```yaml
status:
  clusterCount: 1
  conditions:
  - lastTransitionTime: "2026-01-04T20:00:00Z"
    message: Query executed successfully
    reason: QuerySuccess
    status: "True"
    type: Available
  - lastTransitionTime: "2026-01-04T20:00:00Z"
    message: Query executed successfully
    reason: QuerySuccess
    status: "False"
    type: Degraded
  lastQueryTime: "2026-01-04T20:00:00Z"
  observedGeneration: 1
  result: '{"id":"abc-123","name":"prod-cluster","replicas":3,"version":"1.0.0","state":"ready","readyReplicas":3}'
```

## Use Cases

### 1. Monitoring Cluster State

Create BSSQuery resources to continuously monitor cluster states and use them in other controllers or external monitoring systems.

```yaml
apiVersion: bss.localhost/v1alpha1
kind: BSSQuery
metadata:
  name: monitor-cluster-abc
spec:
  apiEndpoint: "http://localhost:8880/graphql"
  query: cluster
  clusterID: "abc-123"
  refreshInterval: 10  # Check every 10 seconds
```

### 2. Dashboard Data Source

Use BSSQuery resources as a data source for Kubernetes dashboards or custom controllers that need cluster information.

```bash
# Get all clusters and their states
kubectl get bssquery all-clusters -o jsonpath='{.status.result}' | jq '.[] | {name, state, replicas}'
```

### 3. GitOps Integration

Commit BSSQuery resources to Git as part of your GitOps workflow to track which clusters are being monitored.

## Troubleshooting

### Query Keeps Failing

Check the conditions:
```bash
kubectl get bssquery my-cluster -o jsonpath='{.status.conditions}' | jq .
```

Common issues:
- **APIEndpoint unreachable**: Ensure the BSS API is running and accessible
- **Invalid ClusterID**: Verify the cluster exists
- **Network policies**: Check if the operator can reach the API endpoint

### Status Not Updating

1. Check if the controller is running:
```bash
kubectl logs -n bss-operator-system deployment/bss-operator-controller-manager
```

2. Verify the refresh interval:
```bash
kubectl get bssquery my-cluster -o jsonpath='{.spec.refreshInterval}'
```

### Get Controller Logs

```bash
# If running with make run
# Check the terminal output

# If deployed to cluster
kubectl logs -n bss-operator-system -l control-plane=controller-manager -f
```

## Demo

Run the interactive demo:
```bash
./hack/demo-bssquery.sh
```

This will:
1. Create test clusters in the BSS API
2. Create BSSQuery resources
3. Show the status and results
4. Watch for updates
5. Clean up resources

## Architecture

The BSSQuery controller uses the GraphQL client in `internal/client/graphql.go` to:
1. Execute GraphQL queries against the BSS API
2. Parse and validate responses
3. Update the BSSQuery status
4. Schedule the next reconciliation based on `refreshInterval`

## Development

### Add New Query Types

1. Add the query type to the enum in `api/v1alpha1/bssquery_types.go`
2. Implement the query logic in the controller's `executeQuery` method
3. Add a corresponding method to `internal/client/graphql.go`
4. Update the GraphQL schema in `hack/bss-api/graphql/schema.go`

### Testing

Run unit tests:
```bash
make test
```

Run the GraphQL API tests:
```bash
cd hack/bss-api
./test-graphql.sh
```

## Related Documentation

- [GraphQL API Documentation](./graphql.md)
- [BSS API README](../hack/bss-api/README.md)
- [Controller Architecture](./controller_architecture.md)
