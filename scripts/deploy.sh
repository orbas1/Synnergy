#!/usr/bin/env bash
# deploy.sh - Deploy Synnergy to Kubernetes with canary and rollback support.
set -euo pipefail

ENVIRONMENT="${1:-staging}"
OVERLAY="infrastructure/k8s/overlays/${ENVIRONMENT}"

if [[ ! -d "$OVERLAY" ]]; then
  echo "Unknown environment: $ENVIRONMENT" >&2
  echo "Usage: $0 [staging|production]" >&2
  exit 1
fi

IMAGE="${IMAGE:-ghcr.io/example/synnergy:latest}"

# Apply base manifests for the target environment
kubectl kustomize "$OVERLAY" | kubectl apply -f -

# Determine desired replicas after apply
DESIRED_REPLICAS=$(kubectl get deployment synnergy -o jsonpath='{.spec.replicas}')

# Update deployment image and rollout canary
kubectl set image deployment/synnergy synnergy="$IMAGE" --record
kubectl scale deployment/synnergy --replicas=1

if ! kubectl rollout status deployment/synnergy --timeout=120s; then
  echo "Canary failed, rolling back" >&2
  kubectl rollout undo deployment/synnergy
  exit 1
fi

# Promote canary to full rollout
kubectl scale deployment/synnergy --replicas="$DESIRED_REPLICAS"

if ! kubectl rollout status deployment/synnergy --timeout=120s; then
  echo "Rollout failed, rolling back" >&2
  kubectl rollout undo deployment/synnergy
  exit 1
fi

echo "Deployment of $ENVIRONMENT environment succeeded"
