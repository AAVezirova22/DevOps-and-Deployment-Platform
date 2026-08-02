# Architecture

DeployKit is a compact deployment platform built around one operator action: run `deployctl deploy` and receive a live HTTPS endpoint. The implementation keeps the control plane in a CLI and delegates durable runtime behavior to Kubernetes.

## System Context

```text
developer or CI runner
        |
        | deployctl deploy
        v
local Docker daemon or CI builder
        |
        | image: registry/service:tag
        v
container registry
        |
        | kubectl apply
        v
k3s/Kubernetes cluster
        |
        | ingress-nginx + cert-manager
        v
public HTTPS URL
```

## Core Flow

1. `deployctl` loads `deploykit.yaml`.
2. It applies defaults and validates required fields such as service name, domain, port, and replica count.
3. Unless `--no-build` is used, it runs `docker build -t <image> -f <dockerfile> <context>`.
4. It renders Kubernetes YAML for Namespace, Deployment, Service, and Ingress.
5. It applies the manifests with `kubectl apply -f -`.
6. It waits for `kubectl rollout status`.
7. It reads recent deployment logs and scans for configured failure signatures.
8. If verification fails and rollback is enabled, it runs `kubectl rollout undo`.
9. On success, it records the release as a Kubernetes ConfigMap.
10. It prints the HTTPS URL derived from the configured domain.

## Isolation Model

Each service can deploy into its own Kubernetes namespace. That gives a lightweight environment boundary for:

- service resources,
- service logs,
- rollout history,
- Kubernetes RBAC in future versions,
- cleanup and teardown,
- per-environment configuration.

This is intentionally simpler than creating a whole cluster per service. For the target customer, namespace isolation provides practical operational separation while keeping infrastructure cost and complexity low.

## Runtime Resources

DeployKit renders these Kubernetes resources:

- `Namespace`: isolates service resources.
- `ServiceAccount`: gives each service an explicit runtime identity.
- `Role` and `RoleBinding`: provide a namespace-scoped RBAC extension point with no default runtime privileges.
- `Deployment`: runs application replicas with rolling updates.
- `Service`: exposes pods inside the cluster.
- `Ingress`: routes public HTTP/HTTPS traffic to the service.

The Deployment uses readiness and liveness probes against the configured health path. The rollout strategy sets `maxUnavailable: 0` and `maxSurge: 1` so updates keep old pods available while new pods become ready. The pod template also includes resource requests/limits and a hardened non-root security context.

## TLS and Networking

TLS is handled through cert-manager. DeployKit annotates the Ingress with:

```yaml
cert-manager.io/cluster-issuer: letsencrypt-prod
```

The cluster must have:

- an ingress controller such as ingress-nginx,
- cert-manager,
- a matching `ClusterIssuer`,
- DNS pointed at the ingress load balancer or node IP.

This division keeps certificate lifecycle management in the cluster, where renewal controllers can operate continuously.

## Rollback Strategy

Rollback has two triggers:

- Kubernetes rollout failure, detected by `kubectl rollout status`.
- suspicious log output, detected by scanning recent deployment logs for configured terms such as `panic:`, `fatal`, or `imagepullbackoff`.

When triggered, DeployKit runs:

```sh
kubectl rollout undo deployment/<service> -n <namespace>
```

This uses Kubernetes revision history instead of a custom release database. That keeps the first version reliable and understandable.

## Release Records

After a successful rollout and log scan, DeployKit writes a ConfigMap with release metadata. This creates a durable cluster-local audit trail without requiring a database.

## Infrastructure Layer

The Terraform example provisions a Hetzner Cloud VM and installs k3s through cloud-init. This gives a small but realistic self-hosted cluster target.

The bootstrap script installs:

- ingress-nginx,
- cert-manager,
- a production Let's Encrypt `ClusterIssuer`.

The same CLI can target managed Kubernetes services such as EKS, GKE, AKS, or DigitalOcean Kubernetes as long as `kubectl` is configured.

## Security Boundaries

The current implementation relies on the user's local Docker and Kubernetes credentials. Production deployments should add:

- scoped kubeconfigs per environment,
- registry token handling through CI secrets,
- namespace-level RBAC,
- sealed secrets or external secret management,
- admission policies for images and resource limits.

The code is structured so these concerns can be added behind the existing deploy workflow.
