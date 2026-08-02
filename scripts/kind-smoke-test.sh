#!/usr/bin/env sh
set -eu

CLUSTER_NAME="${CLUSTER_NAME:-deploykit-smoke}"
CONFIG_FILE="$(mktemp)"

cleanup() {
  rm -f "$CONFIG_FILE"
  if [ "${KEEP_KIND_CLUSTER:-}" = "" ]; then
    kind delete cluster --name "$CLUSTER_NAME" >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT

command -v kind >/dev/null 2>&1 || {
  echo "kind is required for this smoke test" >&2
  exit 1
}
command -v kubectl >/dev/null 2>&1 || {
  echo "kubectl is required for this smoke test" >&2
  exit 1
}
command -v go >/dev/null 2>&1 || {
  echo "go is required for this smoke test" >&2
  exit 1
}

cat > "$CONFIG_FILE" <<EOF
name: smoke-api
namespace: smoke-api-prod
image: registry.k8s.io/pause:3.9
domain: smoke.localhost
port: 8080
replicas: 2
healthPath: /
serviceAccount: smoke-api
secretRefs: []
env:
  APP_ENV: smoke
resources:
  cpuRequest: 25m
  memoryRequest: 32Mi
  cpuLimit: 100m
  memoryLimit: 128Mi
rollback:
  enabled: true
  failureText: [panic:, fatal, exception, crashloopbackoff, imagepullbackoff]
EOF

if ! kind get clusters | grep -qx "$CLUSTER_NAME"; then
  kind create cluster --name "$CLUSTER_NAME"
fi

kubectl create namespace smoke-api-prod --dry-run=client -o yaml | kubectl apply -f -
go run ./cmd/deployctl render --config "$CONFIG_FILE" > /tmp/deploykit-smoke.yaml
kubectl apply --dry-run=server -f /tmp/deploykit-smoke.yaml
