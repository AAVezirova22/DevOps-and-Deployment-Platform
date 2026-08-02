# Hetzner k3s Terraform Module

This example provisions one small Hetzner Cloud VM and installs k3s with public HTTP/HTTPS ports open. It is intentionally compact so the platform can be demonstrated cheaply, while still matching the operational shape of a self-hosted deployment platform.

## Usage

1. Create a Hetzner Cloud API token.
2. Copy `terraform.tfvars.example` to `terraform.tfvars`.
3. Set `hcloud_token`, `ssh_public_key`, and `server_location`.
4. Run:

```sh
terraform init
terraform apply
```

After the server is ready, SSH into it and copy `/etc/rancher/k3s/k3s.yaml` to your workstation. Replace `127.0.0.1` in that kubeconfig with the server IP. Then run `scripts/bootstrap-k3s-addons.sh` with `LETSENCRYPT_EMAIL` set.
