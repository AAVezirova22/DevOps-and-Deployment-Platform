# Production Readiness

DeployKit is now production-oriented for a small self-hosted platform, using only free and open-source components. It is not equivalent to a fully managed commercial platform, but the remaining gaps require real infrastructure, provider credentials, a domain, or paid/owned compute to prove end to end.

## Hardened Without Paid Services

The project includes these production controls:

- Kubernetes namespace isolation for each service/environment.
- Explicit service accounts with token automount disabled.
- Minimal namespace Role and RoleBinding objects for runtime identity.
- Non-root pod security context.
- Container hardening with privilege escalation disabled and all Linux capabilities dropped.
- Read-only root filesystem.
- CPU and memory requests/limits.
- Readiness and liveness probes.
- TLS automation through cert-manager and Let's Encrypt.
- Secret references through `secretRefs`, so secrets are injected from Kubernetes Secret objects instead of stored in `deploykit.yaml`.
- Persistent release records stored as Kubernetes ConfigMaps.
- Automatic rollback when rollout verification fails or logs match configured failure text.
- Manual `deployctl rollback`.
- `deployctl status` for Deployment, Service, Ingress, certificate, and release-record visibility.
- CI tests, build checks, manifest rendering, image publishing, and kind-based Kubernetes API validation.

## No-Cost Validation

The `scripts/kind-smoke-test.sh` script creates a local kind cluster and validates rendered manifests with Kubernetes server-side dry-run:

```sh
make smoke-test
```

This verifies that the Kubernetes API accepts the generated objects without requiring AWS, GCP, Hetzner, DNS records, or a paid cluster.

## Secrets Model

DeployKit intentionally does not place secret values in `deploykit.yaml`. Production users should create Kubernetes Secrets with `kubectl`, Sealed Secrets, External Secrets Operator, SOPS, or their CI/CD secret store.

The application config references existing secrets:

```yaml
secretRefs: [demo-api-secrets]
```

The rendered Deployment injects them with `envFrom.secretRef`.

## Release History

Every successful deployment writes a ConfigMap labeled:

```yaml
deploykit.io/release: "true"
```

The record stores:

- image,
- tag,
- domain,
- deployment timestamp.

This keeps an auditable trail inside the cluster without running a database.

## RBAC Boundary

Application pods run with a dedicated service account and no mounted Kubernetes API token by default. The Role intentionally starts with no permissions. Teams can extend it only when an application has a real need to call the Kubernetes API.

The operator running `deployctl` still needs credentials that can create and update resources in the target namespace. In production, those credentials should be scoped with Kubernetes RBAC and stored in GitHub Environments or another protected secret store.

## What Still Requires Real Infrastructure

These items cannot be fully proven for free in a repository alone:

- Public DNS automation, because it requires an owned domain and DNS provider credentials.
- Real Let's Encrypt issuance, because it requires public DNS and reachable HTTP ingress.
- Multi-node high availability, because it requires multiple VMs or machines.
- Cloud-provider integration tests, because they require cloud credentials and may create billable resources.
- Registry pull testing for private production images, because it depends on real registry credentials.

DeployKit documents these boundaries instead of pretending local tests prove them.
