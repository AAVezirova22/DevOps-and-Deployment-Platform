#!/usr/bin/env sh
set -eu

missing=0

check() {
  name="$1"
  command="$2"
  version_command="$3"

  if command -v "$command" >/dev/null 2>&1; then
    printf "ok: %s - " "$name"
    sh -c "$version_command" 2>/dev/null | head -n 1 || true
  else
    printf "missing: %s (%s)\n" "$name" "$command"
    missing=1
  fi
}

check "Go 1.22+" "go" "go version"
check "Docker Desktop / Docker CLI" "docker" "docker --version"
check "kubectl" "kubectl" "kubectl version --client=true"
check "kind" "kind" "kind version"
check "Terraform" "terraform" "terraform version"
check "Git" "git" "git --version"
check "Make" "make" "make --version"

if command -v docker >/dev/null 2>&1; then
  if docker info >/dev/null 2>&1; then
    echo "ok: Docker daemon is running"
  else
    echo "warning: Docker CLI is installed, but the Docker daemon is not reachable"
    echo "         Start Docker Desktop before running deploys or kind smoke tests."
  fi
fi

if [ "$missing" -ne 0 ]; then
  echo
  echo "Install missing tools using docs/prerequisites.md, then rerun this script."
  exit 1
fi

echo
echo "All required DeployKit prerequisites are installed."
