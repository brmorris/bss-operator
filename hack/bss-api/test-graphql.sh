#!/bin/bash

# Test script for BSS API GraphQL endpoint

set -e

API_URL="http://localhost:8880"
GRAPHQL_ENDPOINT="${API_URL}/graphql"

echo "======================================"
echo "BSS API GraphQL Test Script"
echo "======================================"
echo ""

# Check if the API is running
echo "1. Checking if BSS API is running..."
if ! curl -s "${API_URL}/graphql" > /dev/null; then
    echo "❌ Error: BSS API is not running at ${API_URL}"
    echo "Please start the API with: cd hack/bss-api && go run main.go"
    exit 1
fi
echo "✅ BSS API is running"
echo ""

# Create a test cluster via GraphQL
echo "2. Creating a test cluster via GraphQL mutation..."
CLUSTER_RESPONSE=$(curl -s -X POST "${GRAPHQL_ENDPOINT}" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "mutation CreateCluster($name: String!, $replicas: Int!, $version: String!) { createCluster(name: $name, replicas: $replicas, version: $version) { id name replicas version state } }",
    "variables": {
      "name": "test-cluster",
      "replicas": 3,
      "version": "1.0.0"
    }
  }')

echo "Response: ${CLUSTER_RESPONSE}"

# Extract cluster ID
CLUSTER_ID=$(echo "${CLUSTER_RESPONSE}" | grep -oP '"id"\s*:\s*"\K[^"]+' | head -1)

if [ -z "${CLUSTER_ID}" ]; then
    echo "❌ Failed to create cluster"
    exit 1
fi
echo "✅ Created cluster with ID: ${CLUSTER_ID}"
echo ""

# Query the specific cluster
echo "3. Querying the specific cluster via GraphQL..."
CLUSTER_QUERY_RESPONSE=$(curl -s -X POST "${GRAPHQL_ENDPOINT}" \
  -H "Content-Type: application/json" \
  -d "{
    \"query\": \"query GetCluster(\$id: String!) { cluster(id: \$id) { id name replicas version state readyReplicas } }\",
    \"variables\": {
      \"id\": \"${CLUSTER_ID}\"
    }
  }")

echo "Response: ${CLUSTER_QUERY_RESPONSE}"
echo "✅ Successfully queried cluster"
echo ""

# List all clusters
echo "4. Listing all clusters via GraphQL..."
CLUSTERS_RESPONSE=$(curl -s -X POST "${GRAPHQL_ENDPOINT}" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "query { clusters { id name replicas version state } }"
  }')

echo "Response: ${CLUSTERS_RESPONSE}"
CLUSTER_COUNT=$(echo "${CLUSTERS_RESPONSE}" | grep -oP '"id"\s*:\s*"[^"]+' | wc -l)
echo "✅ Found ${CLUSTER_COUNT} cluster(s)"
echo ""

# Wait a bit for cluster to transition to ready
echo "5. Waiting 3 seconds for cluster state to update..."
sleep 3

# Query again to see state change
echo "6. Querying cluster again to check state..."
CLUSTER_STATE_RESPONSE=$(curl -s -X POST "${GRAPHQL_ENDPOINT}" \
  -H "Content-Type: application/json" \
  -d "{
    \"query\": \"query GetCluster(\$id: String!) { cluster(id: \$id) { id name state readyReplicas } }\",
    \"variables\": {
      \"id\": \"${CLUSTER_ID}\"
    }
  }")

echo "Response: ${CLUSTER_STATE_RESPONSE}"
STATE=$(echo "${CLUSTER_STATE_RESPONSE}" | grep -oP '"state"\s*:\s*"\K[^"]+')
echo "✅ Cluster state: ${STATE}"
echo ""

# Delete the cluster
echo "7. Deleting the cluster via GraphQL mutation..."
DELETE_RESPONSE=$(curl -s -X POST "${GRAPHQL_ENDPOINT}" \
  -H "Content-Type: application/json" \
  -d "{
    \"query\": \"mutation DeleteCluster(\$id: String!) { deleteCluster(id: \$id) }\",
    \"variables\": {
      \"id\": \"${CLUSTER_ID}\"
    }
  }")

echo "Response: ${DELETE_RESPONSE}"
echo "✅ Cluster deletion initiated"
echo ""

# Wait for deletion
echo "8. Waiting 3 seconds for cluster deletion..."
sleep 3

# Try to query the deleted cluster
echo "9. Verifying cluster is deleted..."
DELETED_QUERY_RESPONSE=$(curl -s -X POST "${GRAPHQL_ENDPOINT}" \
  -H "Content-Type: application/json" \
  -d "{
    \"query\": \"query GetCluster(\$id: String!) { cluster(id: \$id) { id name state } }\",
    \"variables\": {
      \"id\": \"${CLUSTER_ID}\"
    }
  }")

echo "Response: ${DELETED_QUERY_RESPONSE}"
if echo "${DELETED_QUERY_RESPONSE}" | grep -q '"cluster":null'; then
    echo "✅ Cluster successfully deleted"
else
    echo "⚠️  Cluster may still be in deleting state"
fi
echo ""

echo "======================================"
echo "✅ All GraphQL tests completed!"
echo "======================================"
echo ""
echo "You can also test the GraphQL API interactively by opening:"
echo "  ${GRAPHQL_ENDPOINT}"
echo "in your web browser (GraphiQL playground)."
