#!/bin/bash

# Demo script for BSSQuery Custom Resource
# This script demonstrates creating and monitoring BSSQuery resources

set -e

echo "======================================"
echo "BSSQuery Controller Demo"
echo "======================================"
echo ""

# Check prerequisites
echo "1. Checking prerequisites..."
if ! kubectl cluster-info > /dev/null 2>&1; then
    echo "❌ Error: kubectl is not configured or cluster is not accessible"
    exit 1
fi
echo "✅ Kubernetes cluster is accessible"
echo ""

# Check if BSS API is running
echo "2. Checking BSS API availability..."
if ! curl -s http://localhost:8880/graphql > /dev/null 2>&1; then
    echo "⚠️  Warning: BSS API is not accessible at http://localhost:8880"
    echo "   Please ensure the BSS API is running:"
    echo "   cd hack/bss-api && go run main.go"
    echo ""
fi

# Create a test cluster via GraphQL first
echo "3. Creating test clusters via GraphQL API..."
CLUSTER1_ID=$(curl -s -X POST http://localhost:8880/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "mutation { createCluster(name: \"demo-cluster-1\", replicas: 3, version: \"1.0.0\") { id } }"
  }' | grep -oP '"id"\s*:\s*"\K[^"]+' || echo "")

CLUSTER2_ID=$(curl -s -X POST http://localhost:8880/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "mutation { createCluster(name: \"demo-cluster-2\", replicas: 5, version: \"1.1.0\") { id } }"
  }' | grep -oP '"id"\s*:\s*"\K[^"]+' || echo "")

if [ -z "${CLUSTER1_ID}" ]; then
    echo "⚠️  Could not create test clusters (API might not be running)"
else
    echo "✅ Created test clusters"
    echo "   Cluster 1 ID: ${CLUSTER1_ID}"
    echo "   Cluster 2 ID: ${CLUSTER2_ID}"
fi
echo ""

# Create BSSQuery for listing all clusters
echo "4. Creating BSSQuery to list all clusters..."
cat <<EOF | kubectl apply -f -
apiVersion: bss.localhost/v1alpha1
kind: BSSQuery
metadata:
  name: all-clusters
  namespace: default
spec:
  apiEndpoint: "http://localhost:8880/graphql"
  query: clusters
  refreshInterval: 10
EOF
echo "✅ BSSQuery 'all-clusters' created"
echo ""

# Create BSSQuery for a specific cluster (if we have a cluster ID)
if [ -n "${CLUSTER1_ID}" ]; then
    echo "5. Creating BSSQuery for specific cluster..."
    cat <<EOF | kubectl apply -f -
apiVersion: bss.localhost/v1alpha1
kind: BSSQuery
metadata:
  name: specific-cluster
  namespace: default
spec:
  apiEndpoint: "http://localhost:8880/graphql"
  query: cluster
  clusterID: "${CLUSTER1_ID}"
  refreshInterval: 5
EOF
    echo "✅ BSSQuery 'specific-cluster' created"
    echo ""
fi

# Wait for reconciliation
echo "6. Waiting for controller to reconcile..."
sleep 5

# Show the BSSQuery resources
echo "7. BSSQuery Resources:"
echo "---"
kubectl get bssquery
echo ""

# Show detailed status
echo "8. Detailed status of 'all-clusters' BSSQuery:"
echo "---"
kubectl get bssquery all-clusters -o yaml | grep -A 30 "status:"
echo ""

if [ -n "${CLUSTER1_ID}" ]; then
    echo "9. Detailed status of 'specific-cluster' BSSQuery:"
    echo "---"
    kubectl get bssquery specific-cluster -o yaml | grep -A 30 "status:"
    echo ""
fi

# Show the JSON results
echo "10. Query Results:"
echo "---"
echo "All Clusters Result:"
kubectl get bssquery all-clusters -o jsonpath='{.status.result}' | jq . 2>/dev/null || kubectl get bssquery all-clusters -o jsonpath='{.status.result}'
echo ""

if [ -n "${CLUSTER1_ID}" ]; then
    echo "Specific Cluster Result:"
    kubectl get bssquery specific-cluster -o jsonpath='{.status.result}' | jq . 2>/dev/null || kubectl get bssquery specific-cluster -o jsonpath='{.status.result}'
    echo ""
fi

# Watch for updates
echo "11. Watching BSSQuery resources (Ctrl+C to stop)..."
echo "    The controller will refresh every 5-10 seconds based on refreshInterval"
echo "---"
sleep 2
kubectl get bssquery -w &
WATCH_PID=$!

# Wait a bit to show some updates
sleep 15

# Clean up
kill $WATCH_PID 2>/dev/null || true
echo ""
echo ""
echo "12. Cleaning up..."
kubectl delete bssquery all-clusters --ignore-not-found=true
kubectl delete bssquery specific-cluster --ignore-not-found=true 2>/dev/null || true

if [ -n "${CLUSTER1_ID}" ]; then
    curl -s -X POST http://localhost:8880/graphql \
      -H "Content-Type: application/json" \
      -d "{\"query\": \"mutation { deleteCluster(id: \\\"${CLUSTER1_ID}\\\") }\"}" > /dev/null 2>&1 || true
fi

if [ -n "${CLUSTER2_ID}" ]; then
    curl -s -X POST http://localhost:8880/graphql \
      -H "Content-Type: application/json" \
      -d "{\"query\": \"mutation { deleteCluster(id: \\\"${CLUSTER2_ID}\\\") }\"}" > /dev/null 2>&1 || true
fi

echo "✅ Demo completed and resources cleaned up"
echo ""
echo "======================================"
echo "Summary"
echo "======================================"
echo "The BSSQuery controller:"
echo "  ✅ Periodically queries the GraphQL API"
echo "  ✅ Updates the status with query results"
echo "  ✅ Reports conditions for monitoring"
echo "  ✅ Supports both single cluster and list queries"
echo ""
echo "To see the CRD definition:"
echo "  kubectl get crd bssqueries.bss.localhost -o yaml"
echo ""
echo "To create your own BSSQuery:"
echo "  kubectl apply -f config/samples/bss_v1alpha1_bssquery_clusters.yaml"
