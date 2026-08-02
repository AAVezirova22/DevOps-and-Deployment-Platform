# Technology Decisions

## Go for the CLI

Go is a strong fit for deployment tooling because it produces small static binaries, starts quickly, handles subprocess orchestration well, and is familiar to infrastructure teams. The current CLI uses only the standard library, which keeps setup friction low and avoids supply-chain weight for the first version.

Alternatives considered:

- Node.js: better for frontend-heavy tooling, but less attractive for single-binary distribution.
- Python: productive for automation, but packaging a reliable cross-platform CLI is more involved.
- Bash: simple for prototypes, but error handling, testing, and config parsing degrade quickly.

## Docker as the Build Boundary

Docker provides a consistent artifact format regardless of whether the application is written in Go, Node, Python, Java, or another stack. DeployKit treats the Docker image as the unit of release.

This is useful for employers and teams because it shows understanding of build reproducibility, dependency isolation, and runtime portability.

## Kubernetes for Runtime Orchestration

Kubernetes is used because it already solves the hard runtime pieces:

- rolling deployments,
- health probes,
- service discovery,
- ingress integration,
- rollout history,
- namespace isolation,
- declarative reconciliation.

DeployKit intentionally does not invent a scheduler or process supervisor. It wraps the Kubernetes API in a simpler product experience.

## k3s for Self-Hosted Clusters

k3s is Kubernetes packaged for smaller footprints. It is appropriate for:

- demos,
- small internal tools,
- edge deployments,
- cost-sensitive startup environments,
- local or single-node platform experiments.

The same manifests work on larger managed clusters when the user outgrows one VM.

## Terraform for Infrastructure

Terraform is used because cloud infrastructure should be declared, reviewed, and reproducible. The Hetzner example shows the minimum viable platform host: a VM, SSH key, firewall, and cloud-init k3s installation.

Pulumi would also be reasonable, especially for teams that prefer general-purpose languages. Terraform is the default here because it remains widely recognized across DevOps and platform teams.

## GitHub Actions for CI/CD

GitHub Actions gives a common integration path for open-source and portfolio repositories. The workflow runs tests, builds the CLI, renders manifests, and publishes a container image on pushes to `main`.

This demonstrates that the project is not only a local script. It has repeatable validation and a path to automated release.

## cert-manager and Let's Encrypt for TLS

Automated TLS is a core platform feature. cert-manager handles certificate issuance and renewal inside the cluster, while Let's Encrypt provides public certificates without manual purchasing or rotation.

This keeps `deployctl` focused on deployment orchestration and lets a dedicated controller own certificate lifecycle.
