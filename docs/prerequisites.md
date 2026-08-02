# Prerequisites

DeployKit needs a small set of local tools. All of them are available without a paid subscription.

## Required Tools

- Go 1.22 or newer: builds and tests the `deployctl` CLI.
- Docker Desktop or a reachable Docker daemon: builds application images and runs kind clusters.
- kubectl: applies and inspects Kubernetes resources.
- kind: runs the no-cost local Kubernetes smoke test.
- Terraform: provisions the optional Hetzner k3s infrastructure example.
- Git: provides commit SHAs for immutable image tags.
- Make: runs the convenience targets on macOS/Linux. Windows users can run the PowerShell commands directly.

## Check Your Machine

Run these commands from the DeployKit project folder after cloning this repository.

macOS or Linux:

```sh
scripts/check-prereqs.sh
```

Windows PowerShell:

```powershell
.\scripts\check-prereqs.ps1
```

## macOS Install

Using Homebrew:

```sh
brew install go kubectl kind
brew tap hashicorp/tap
brew install hashicorp/tap/terraform
```

Install Docker Desktop from:

```text
https://www.docker.com/products/docker-desktop/
```

After installing Docker Desktop, start it once before running smoke tests or deployments.

## Windows Install

Using Chocolatey:

```powershell
choco install golang docker-desktop kubernetes-cli kind git -y
choco install terraform -y
```

Using winget:

```powershell
winget install GoLang.Go
winget install Docker.DockerDesktop
winget install Kubernetes.kubectl
winget install Kubernetes.kind
winget install Git.Git
winget install Hashicorp.Terraform
```

Restart PowerShell after installation so the new commands are available on `PATH`.

## Linux Install

Install Go, Docker, kubectl, kind, Terraform, Git, and Make through your distribution package manager where available.

For Ubuntu-style systems:

```sh
sudo apt-get update
sudo apt-get install -y git make ca-certificates curl gnupg
```

Install Docker Engine from Docker's Linux documentation, then install Go, kubectl, kind, and Terraform using their official package instructions or your distribution packages.

## Optional Cloud Requirements

The local smoke test does not need cloud credentials. Real public deployments need:

- a Kubernetes cluster or the optional Hetzner k3s Terraform example,
- a container registry such as GHCR,
- DNS records pointing to the cluster ingress address,
- a reachable ingress controller,
- cert-manager configured with a Let's Encrypt issuer.
