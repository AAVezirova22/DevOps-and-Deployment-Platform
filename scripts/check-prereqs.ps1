$ErrorActionPreference = "Stop"

$missing = $false

function Test-Tool {
    param(
        [string] $Name,
        [string] $Command,
        [string] $VersionCommand
    )

    if (Get-Command $Command -ErrorAction SilentlyContinue) {
        Write-Host "ok: $Name - " -NoNewline
        try {
            Invoke-Expression $VersionCommand | Select-Object -First 1
        } catch {
            Write-Host "installed"
        }
    } else {
        Write-Host "missing: $Name ($Command)"
        $script:missing = $true
    }
}

Test-Tool "Go 1.22+" "go" "go version"
Test-Tool "Docker Desktop / Docker CLI" "docker" "docker --version"
Test-Tool "kubectl" "kubectl" "kubectl version --client=true"
Test-Tool "kind" "kind" "kind version"
Test-Tool "Terraform" "terraform" "terraform version"
Test-Tool "Git" "git" "git --version"

if (Get-Command docker -ErrorAction SilentlyContinue) {
    try {
        docker info *> $null
        Write-Host "ok: Docker daemon is running"
    } catch {
        Write-Host "warning: Docker CLI is installed, but the Docker daemon is not reachable"
        Write-Host "         Start Docker Desktop before running deploys or kind smoke tests."
    }
}

if ($missing) {
    Write-Host ""
    Write-Host "Install missing tools using docs/prerequisites.md, then rerun this script."
    exit 1
}

Write-Host ""
Write-Host "All required DeployKit prerequisites are installed."
