# DeployKit

DeployKit is a lightweight self-hosted deployment platform for small engineering teams that want the experience of `deploy` producing a live HTTPS URL without handing their runtime to a proprietary platform. It combines a Go CLI, Docker image builds, Kubernetes rollout management, cert-manager TLS, and Terraform-provisioned k3s infrastructure.

The project is intentionally compact, but it demonstrates the same concerns senior engineers deal with in production platforms: repeatable infrastructure, isolated deploy environments, observable rollouts, automated rollback, CI checks, and clear operational documentation.

## What It Does

- `deployctl init` creates a project deployment config.
- `deployctl deploy` builds and pushes a Docker image for the current repository.
- The CLI creates an isolated Kubernetes namespace per service/environment.
- It applies Deployment, Service, and Ingress resources.
- TLS is delegated to cert-manager and Let's Encrypt through an annotated Ingress.
- It waits for `kubectl rollout status`.
- It scans recent deployment logs for configured failure text.
- It automatically runs `kubectl rollout undo` when rollout verification fails.
- It prints the final `https://...` URL when deployment succeeds.

## Repository Layout

```text
cmd/deployctl/                 CLI entrypoint
internal/cli/                  command parsing and UX
internal/config/               deploykit.yaml parser and validation
internal/deploy/               Docker/Kubernetes deployment workflow
internal/logscan/              rollback failure pattern detection
internal/runner/               command execution abstraction
infra/terraform/hetzner-k3s/   single-node k3s infrastructure example
scripts/bootstrap-k3s-addons.sh ingress-nginx and cert-manager bootstrap
.github/workflows/             CI and manual deployment workflows
docs/                          architecture, customers, operations, decisions
```

## Quick Start

Build the CLI:

```sh
make build
```

Create a deployment config in another application repository:

```sh
deployctl init
```

Edit `deploykit.yaml` with your registry, domain, port, and environment variables:

```yaml
name: demo-api
namespace: demo-api-prod
registry: ghcr.io/example
tag: latest
domain: demo-api.example.com
port: 8080
replicas: 2
context: .
dockerfile: Dockerfile
ingressClass: nginx
clusterIssuer: letsencrypt-prod
healthPath: /healthz
env:
  APP_ENV: production
rollback:
  enabled: true
  failureText: [panic:, fatal, exception, crashloopbackoff, imagepullbackoff]
```

Preview Kubernetes resources before applying them:

```sh
deployctl render --config deploykit.yaml
```

Deploy:

```sh
deployctl deploy --config deploykit.yaml --tag "$(git rev-parse --short HEAD)"
```

For local clusters that can directly access the local Docker image, use `--no-push`.

## Platform Prerequisites

The target cluster needs:

- Kubernetes, with k3s recommended for a low-cost self-hosted setup.
- `kubectl` access from the machine running `deployctl`.
- A container registry such as GHCR, ECR, or Docker Hub.
- ingress-nginx installed in the cluster.
- cert-manager installed with a `letsencrypt-prod` `ClusterIssuer`.
- DNS A/AAAA records pointing the configured domain to the cluster ingress address.

The `infra/terraform/hetzner-k3s` directory provisions a small Hetzner Cloud VM and installs k3s. The add-on bootstrap script installs ingress-nginx, cert-manager, and a Let's Encrypt issuer.

## Architecture

DeployKit has four layers:

1. CLI control plane: validates intent from `deploykit.yaml`, resolves image tags, and coordinates the deployment workflow.
2. Build layer: runs Docker locally or in CI to produce an immutable image.
3. Runtime layer: applies Kubernetes resources into a service-specific namespace.
4. Safety layer: verifies rollouts, scans logs, and triggers Kubernetes rollback if the new release is unhealthy.

See [docs/architecture.md](docs/architecture.md) for the full architecture.

## CI/CD Pipeline

The repository includes:

- `ci.yml` for tests, CLI build, manifest rendering, and container image publishing.
- `deploy.yml` for manual GitHub Actions deployments using a base64 kubeconfig secret.

See [docs/ci-cd.md](docs/ci-cd.md) for pipeline details and recommended branch rules.

## Why These Technologies

- Go keeps the CLI portable, fast to start, and easy to ship as one binary.
- Docker is the standard packaging boundary for heterogeneous application stacks.
- Kubernetes gives a stable deployment API, rollout history, service discovery, and health checks.
- k3s provides the Kubernetes API with lower operational cost for small teams and portfolio demos.
- Terraform captures cloud infrastructure as reviewable, repeatable code.
- GitHub Actions gives an accessible CI/CD path for tests, image publishing, and manifest validation.
- cert-manager and Let's Encrypt automate HTTPS without manual certificate handling.

See [docs/technology-decisions.md](docs/technology-decisions.md) for tradeoffs.

## Target Customers

DeployKit is aimed at:

- solo developers shipping multiple side projects,
- startups that need low-cost deployments before platform engineering headcount exists,
- agencies hosting many small client services,
- internal tools teams that want repeatable environments without a large platform stack,
- engineering candidates demonstrating DevOps maturity with a concrete system.

See [docs/target-customers.md](docs/target-customers.md) for a deeper customer breakdown.

## Development

Run tests:

```sh
make test
```

Render example manifests:

```sh
make render-example
```

Build the container image:

```sh
docker build -t deploykit/deployctl:local .
```

## Current Scope

DeployKit currently focuses on a single-cluster Kubernetes deployment workflow. The extension points are intentionally visible: registry pushes, cloud-provider DNS automation, multi-environment state, Pulumi support, and a dashboard API can be added without replacing the CLI contract.
