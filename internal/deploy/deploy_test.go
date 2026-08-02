package deploy

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/ani/devops-deployment-platform/internal/config"
	"github.com/ani/devops-deployment-platform/internal/runner"
)

type fakeRunner struct {
	calls []string
	logs  string
	fail  map[string]error
}

func (f *fakeRunner) Run(_ context.Context, _ io.Reader, name string, args ...string) (runner.Result, error) {
	call := name + " " + strings.Join(args, " ")
	f.calls = append(f.calls, call)
	if err := f.fail[call]; err != nil {
		return runner.Result{}, err
	}
	if strings.HasPrefix(call, "kubectl logs") {
		return runner.Result{Stdout: f.logs}, nil
	}
	return runner.Result{}, nil
}

func TestRenderManifestsIncludesTLSIngressAndProbes(t *testing.T) {
	cfg := validConfig()
	out := RenderManifests(cfg)
	for _, want := range []string{
		"kind: ServiceAccount",
		"automountServiceAccountToken: false",
		"kind: Role",
		"serviceAccountName: api",
		"kind: Deployment",
		"kind: Ingress",
		"cert-manager.io/cluster-issuer: letsencrypt-prod",
		"secretName: api-tls",
		"runAsNonRoot: true",
		"readOnlyRootFilesystem: true",
		"resources:",
		"readinessProbe:",
		"livenessProbe:",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("rendered manifest missing %q:\n%s", want, out)
		}
	}
}

func TestRenderManifestsIncludesSecretRefs(t *testing.T) {
	cfg := validConfig()
	cfg.SecretRefs = []string{"api-secrets"}
	out := RenderManifests(cfg)
	if !strings.Contains(out, "envFrom:") || !strings.Contains(out, "name: api-secrets") {
		t.Fatalf("rendered manifest missing secret refs:\n%s", out)
	}
}

func TestRenderReleaseRecordIncludesImageAndMetadata(t *testing.T) {
	cfg := validConfig()
	out := RenderReleaseRecord(cfg, "ghcr.io/example/api:latest", time.Date(2026, 8, 3, 10, 0, 0, 0, time.UTC))
	for _, want := range []string{
		"kind: ConfigMap",
		"deploykit.io/release: \"true\"",
		"image: \"ghcr.io/example/api:latest\"",
		"deployedAt: \"2026-08-03T10:00:00Z\"",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("release record missing %q:\n%s", want, out)
		}
	}
}

func TestDeployRollsBackWhenLogsContainFailureText(t *testing.T) {
	cfg := validConfig()
	cfg.Rollback.Enabled = true
	fr := &fakeRunner{logs: "panic: boot failed", fail: map[string]error{}}
	err := New(fr, io.Discard, io.Discard).Deploy(context.Background(), cfg, Options{NoBuild: true})
	if err == nil {
		t.Fatal("expected deployment error")
	}

	foundUndo := false
	for _, call := range fr.calls {
		if call == "kubectl rollout undo deployment/api -n api-prod" {
			foundUndo = true
		}
	}
	if !foundUndo {
		t.Fatalf("expected rollback call, got %#v", fr.calls)
	}
}

func TestDeployPushesRegistryBackedImages(t *testing.T) {
	cfg := validConfig()
	fr := &fakeRunner{fail: map[string]error{}}
	err := New(fr, io.Discard, io.Discard).Deploy(context.Background(), cfg, Options{})
	if err != nil {
		t.Fatal(err)
	}

	foundPush := false
	for _, call := range fr.calls {
		if call == "docker push ghcr.io/example/api:latest" {
			foundPush = true
		}
	}
	if !foundPush {
		t.Fatalf("expected image push call, got %#v", fr.calls)
	}
}

func TestDeployRecordsSuccessfulRelease(t *testing.T) {
	cfg := validConfig()
	fr := &fakeRunner{fail: map[string]error{}}
	err := New(fr, io.Discard, io.Discard).Deploy(context.Background(), cfg, Options{NoBuild: true, NoPush: true})
	if err != nil {
		t.Fatal(err)
	}

	applyCount := 0
	for _, call := range fr.calls {
		if call == "kubectl apply -f -" {
			applyCount++
		}
	}
	if applyCount != 2 {
		t.Fatalf("expected manifest apply and release record apply, got %d calls: %#v", applyCount, fr.calls)
	}
}

func TestManualRollbackIgnoresAutomaticRollbackToggle(t *testing.T) {
	cfg := validConfig()
	cfg.Rollback.Enabled = false
	fr := &fakeRunner{fail: map[string]error{}}
	if err := New(fr, io.Discard, io.Discard).Rollback(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}

	if len(fr.calls) < 1 || fr.calls[0] != "kubectl rollout undo deployment/api -n api-prod" {
		t.Fatalf("expected direct rollback, got %#v", fr.calls)
	}
}

func validConfig() config.Config {
	cfg := config.Config{
		Name:       "api",
		Namespace:  "api-prod",
		Registry:   "ghcr.io/example",
		Domain:     "api.example.com",
		Port:       8080,
		Replicas:   2,
		Context:    ".",
		Dockerfile: "Dockerfile",
		Env:        map[string]string{"APP_ENV": "production"},
	}
	cfg.ApplyDefaults()
	return cfg
}
