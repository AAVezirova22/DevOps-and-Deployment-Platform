package deploy

import (
	"context"
	"io"
	"strings"
	"testing"

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
		"kind: Deployment",
		"kind: Ingress",
		"cert-manager.io/cluster-issuer: letsencrypt-prod",
		"secretName: api-tls",
		"readinessProbe:",
		"livenessProbe:",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("rendered manifest missing %q:\n%s", want, out)
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
