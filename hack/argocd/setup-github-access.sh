#!/bin/bash

# Setup script for configuring ArgoCD to access private GitHub repository
# Usage: ./setup-github-access.sh <GITHUB_PAT>

set -e

if [ -z "$1" ]; then
  echo "Usage: $0 <GITHUB_PAT>"
  echo ""
  echo "Create a GitHub PAT at: https://github.com/settings/tokens"
  echo "Required scopes: repo (all)"
  exit 1
fi

GITHUB_PAT="$1"
REPO_URL="https://github.com/brmorris/bss-operator.git"
GITHUB_USER="brmorris"

echo "Creating repository secret in ArgoCD namespace..."

kubectl create secret generic github-repo-creds \
  --from-literal=type=git \
  --from-literal=url="${REPO_URL}" \
  --from-literal=password="${GITHUB_PAT}" \
  --from-literal=username="${GITHUB_USER}" \
  -n argocd \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl label secret github-repo-creds \
  argocd.argoproj.io/secret-type=repository \
  -n argocd \
  --overwrite

echo ""
echo "✓ Repository credentials configured successfully!"
echo ""

