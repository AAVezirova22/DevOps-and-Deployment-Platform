package config

import "testing"

func TestParseYAMLAppliesStructuredFields(t *testing.T) {
	cfg, err := ParseYAML(`name: api
domain: api.example.com
port: 3000
replicas: 3
serviceAccount: api-runtime
secretRefs: [api-secrets, registry-secret]
env:
  APP_ENV: production
  LOG_LEVEL: info
resources:
  cpuRequest: 200m
  memoryRequest: 256Mi
  cpuLimit: 1
  memoryLimit: 1Gi
rollback:
  enabled: true
  failureText: [panic:, fatal]
`)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ApplyDefaults()
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}

	if cfg.Name != "api" || cfg.Namespace != "api" || cfg.Port != 3000 || cfg.Replicas != 3 {
		t.Fatalf("unexpected config: %+v", cfg)
	}
	if cfg.Env["APP_ENV"] != "production" || cfg.Rollback.FailureText[1] != "fatal" {
		t.Fatalf("nested values were not parsed: %+v", cfg)
	}
	if cfg.ServiceAccount != "api-runtime" || cfg.SecretRefs[1] != "registry-secret" {
		t.Fatalf("production identity values were not parsed: %+v", cfg)
	}
	if cfg.Resources.CPURequest != "200m" || cfg.Resources.MemoryLimit != "1Gi" {
		t.Fatalf("resource values were not parsed: %+v", cfg.Resources)
	}
}

func TestValidateRequiresDomain(t *testing.T) {
	cfg := Config{Name: "api"}
	cfg.ApplyDefaults()
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected missing domain validation error")
	}
}

func TestImageRefLowercasesRepositoryPath(t *testing.T) {
	cfg := Config{Name: "Demo-API", Registry: "ghcr.io/AAVezirova22/DevOps-and-Deployment-Platform", Tag: "ABC123"}
	got := cfg.ImageRef()
	want := "ghcr.io/aavezirova22/devops-and-deployment-platform/demo-api:ABC123"
	if got != want {
		t.Fatalf("ImageRef() = %q, want %q", got, want)
	}
}

func TestImageRefLowercasesExplicitImageRepositoryButPreservesTag(t *testing.T) {
	cfg := Config{Image: "GHCR.io/AAVezirova22/API:Release-Candidate"}
	got := cfg.ImageRef()
	want := "ghcr.io/aavezirova22/api:Release-Candidate"
	if got != want {
		t.Fatalf("ImageRef() = %q, want %q", got, want)
	}
}

func TestValidateRejectsInvalidKubernetesNames(t *testing.T) {
	cfg := Config{Name: "Demo_API", Domain: "demo.example.com"}
	cfg.ApplyDefaults()
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected invalid Kubernetes name validation error")
	}
}
