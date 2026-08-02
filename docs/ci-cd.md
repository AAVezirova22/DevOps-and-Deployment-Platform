# CI/CD Pipeline

DeployKit includes two GitHub Actions workflows.

## Continuous Integration

`.github/workflows/ci.yml` runs on pushes and pull requests. It:

- checks out the repository,
- installs Go,
- runs `go test ./...`,
- builds the `deployctl` binary,
- renders example Kubernetes manifests,
- publishes the CLI container image to GHCR on pushes to `main`.

This catches broken CLI code, invalid manifest rendering, and container packaging regressions before deployment.

## Manual Deployment

`.github/workflows/deploy.yml` is a manual deployment workflow. It uses `workflow_dispatch` so an operator can choose the config file and optional image tag.

Required secret:

- `KUBECONFIG_B64`: base64-encoded kubeconfig for the target cluster.

The workflow:

1. Checks out the application repository.
2. Logs in to GitHub Container Registry.
3. Writes the kubeconfig from `KUBECONFIG_B64`.
4. Runs `go run ./cmd/deployctl deploy`.
5. Lets DeployKit build the image, push it, apply Kubernetes manifests, verify rollout, and rollback on failure.

For application repositories that consume a released `deployctl` binary instead of this source tree, replace the final step with a binary download or container invocation.

## Failure Handling

Deployment failure handling happens inside `deployctl`, not in GitHub Actions YAML. This keeps rollback behavior identical whether a developer deploys locally or CI deploys from a protected branch.

The workflow fails if:

- Docker build fails,
- Docker push fails,
- `kubectl apply` fails,
- rollout verification times out,
- logs match configured failure text,
- rollback itself fails after a failed release.

## Recommended Branch Rules

For production repositories:

- require the `ci` workflow before merging,
- restrict `deploy` workflow execution to trusted maintainers,
- protect environment secrets with GitHub Environments,
- use immutable tags such as commit SHAs,
- restrict Kubernetes credentials to the namespaces managed by the project.
