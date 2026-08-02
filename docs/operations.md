# Operations Guide

## Provision a Cluster

Use the Hetzner Terraform example for a low-cost k3s host:

```sh
cd infra/terraform/hetzner-k3s
cp terraform.tfvars.example terraform.tfvars
terraform init
terraform apply
```

Copy the kubeconfig from the server:

```sh
scp root@<server-ip>:/etc/rancher/k3s/k3s.yaml ./kubeconfig
```

Replace `127.0.0.1` in `./kubeconfig` with the server IP and export:

```sh
export KUBECONFIG="$PWD/kubeconfig"
```

## Install Cluster Add-ons

Install ingress-nginx, cert-manager, and the Let's Encrypt issuer:

```sh
LETSENCRYPT_EMAIL=platform@example.com scripts/bootstrap-k3s-addons.sh
```

## Deploy an App

Inside an application repository:

```sh
deployctl init
deployctl deploy --tag "$(git rev-parse --short HEAD)"
```

For CI deployments, set the tag from the commit SHA and make sure the runner can access Docker, the registry, and the target kubeconfig.

## DNS

Create an A record from the configured `domain` to the ingress address. On the single-node Hetzner example, this is the server IPv4 address.

## Rollback

DeployKit rolls back automatically when:

- Kubernetes cannot complete the rollout in time,
- recent deployment logs contain one of the configured `rollback.failureText` strings.

Manual rollback is also available:

```sh
kubectl rollout undo deployment/<name> -n <namespace>
```

## Troubleshooting

Check rollout status:

```sh
kubectl rollout status deployment/<name> -n <namespace>
```

Inspect pods:

```sh
kubectl get pods -n <namespace>
kubectl describe pod <pod> -n <namespace>
```

Read logs:

```sh
kubectl logs deployment/<name> -n <namespace> --tail=200
```

Check certificate status:

```sh
kubectl get certificate -n <namespace>
kubectl describe certificate <name>-tls -n <namespace>
```

Common failure causes:

- DNS does not point to the cluster ingress address.
- cert-manager issuer email was not configured.
- Docker image tag does not exist in the registry.
- Cluster nodes cannot pull from a private registry.
- Application health endpoint does not match `healthPath`.
