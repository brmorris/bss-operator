# BSS API GraphQL Support

This document describes the GraphQL support added to the BSS API and the BSSQuery custom resource for consuming it.

## Overview

The BSS API now supports both REST and GraphQL endpoints:
- **REST API**: `http://localhost:8880/api/v1/clusters`
- **GraphQL API**: `http://localhost:8880/graphql`
- **GraphiQL Playground**: `http://localhost:8880/graphql` (when accessed via browser)

## GraphQL Schema

### Queries

#### Get a single cluster
```graphql
query GetCluster($id: String!) {
  cluster(id: $id) {
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

#### List all clusters
```graphql
query ListClusters {
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

### Mutations

#### Create a cluster
```graphql
mutation CreateCluster($name: String!, $replicas: Int!, $version: String!) {
  createCluster(name: $name, replicas: $replicas, version: $version) {
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

#### Delete a cluster
```graphql
mutation DeleteCluster($id: String!) {
  deleteCluster(id: $id)
}
```

## BSSQuery Custom Resource

The `BSSQuery` CR allows you to consume the GraphQL API from within Kubernetes, with automatic reconciliation and status updates.

### Example: Query a specific cluster

```yaml
apiVersion: bss.localhost/v1alpha1
kind: BSSQuery
metadata:
  name: my-cluster-query
spec:
  apiEndpoint: "http://localhost:8880/graphql"
  query: cluster
  clusterID: "cluster-id-123"
  refreshInterval: 30
```

### Example: List all clusters

```yaml
apiVersion: bss.localhost/v1alpha1
kind: BSSQuery
metadata:
  name: all-clusters-query
spec:
  apiEndpoint: "http://localhost:8880/graphql"
  query: clusters
  refreshInterval: 60
```

## BSSQuery Controller

The BSSQuery controller:
1. Validates the query configuration
2. Executes GraphQL queries against the BSS API endpoint
3. Updates the status with query results
4. Re-queries at the specified refresh interval
5. Reports errors via Kubernetes conditions

### Status Fields

- `lastQueryTime`: Timestamp of the last successful query
- `result`: JSON-encoded result from the GraphQL query
- `clusterCount`: Number of clusters returned
- `conditions`: Standard Kubernetes conditions (Available, Degraded)
- `observedGeneration`: Generation of the last reconciled spec

## Testing

### Start the BSS API

```bash
cd hack/bss-api
go run main.go
```

The API will be available at:
- REST: http://localhost:8880/api/v1/clusters
- GraphQL: http://localhost:8880/graphql

### Test GraphQL Queries (using curl)

Create a cluster:
```bash
curl -X POST http://localhost:8880/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "mutation { createCluster(name: \"test\", replicas: 3, version: \"1.0.0\") { id name state } }"
  }'
```

List clusters:
```bash
curl -X POST http://localhost:8880/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "query { clusters { id name state replicas } }"
  }'
```

Get a specific cluster:
```bash
curl -X POST http://localhost:8880/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "query { cluster(id: \"YOUR_CLUSTER_ID\") { id name state replicas } }"
  }'
```

### Deploy BSSQuery to Kubernetes

1. Install the CRDs:
```bash
make install
```

2. Run the operator:
```bash
make run
```

3. Create a BSSQuery:
```bash
# First create a cluster via the API (see above)
# Then update the sample with the cluster ID
kubectl apply -f config/samples/bss_v1alpha1_bssquery_cluster.yaml
```

4. Check the status:
```bash
kubectl get bssquery -o yaml
kubectl describe bssquery bssquery-sample-cluster
```

## Architecture

```
┌──────────────┐
│   BSSQuery   │
│      CR      │
└──────┬───────┘
       │
       │ reconcile
       ▼
┌──────────────────┐      GraphQL         ┌──────────────┐
│    BSSQuery      │─────────────────────▶│   BSS API    │
│   Controller     │      HTTP/POST       │   GraphQL    │
└──────────────────┘                      │   Endpoint   │
                                          └──────────────┘
```

## Internal Client

The operator includes a GraphQL client package at `internal/client/graphql.go` that provides:

- `NewGraphQLClient(endpoint string)`: Create a new client
- `GetCluster(id string)`: Retrieve a single cluster
- `ListClusters()`: Retrieve all clusters
- `CreateCluster(name, replicas, version)`: Create a new cluster
- `DeleteCluster(id string)`: Delete a cluster

This client is used by the BSSQuery controller and can also be used in other controllers or tools.
