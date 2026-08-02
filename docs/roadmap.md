# Roadmap

## Near Term

- Push images to the configured registry after build.
- Support config overlays for staging and production.
- Add configurable pod security profiles for applications that cannot run with the hardened defaults.
- Add first-class secret creation helpers for teams that do not use Sealed Secrets, SOPS, or External Secrets Operator.

## Platform Features

- Store deployment records in a small SQLite or Postgres control-plane database.
- Add a web dashboard for services, releases, logs, and rollback actions.
- Integrate ExternalDNS for automated DNS records.
- Support sealed-secrets or external secret stores.
- Add GitHub App integration for commit status updates.

## Cloud Providers

- Add AWS EKS Terraform examples.
- Add GCP GKE Terraform examples.
- Add Pulumi examples for teams that prefer general-purpose languages.
- Add Hetzner multi-node k3s support.
