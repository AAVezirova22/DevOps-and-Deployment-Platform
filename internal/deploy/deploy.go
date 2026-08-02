package deploy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/template"

	"github.com/ani/devops-deployment-platform/internal/config"
	"github.com/ani/devops-deployment-platform/internal/logscan"
	"github.com/ani/devops-deployment-platform/internal/runner"
)

type Options struct {
	DryRun  bool
	NoBuild bool
	NoPush  bool
}

type Deployer struct {
	runner runner.CommandRunner
	stdout io.Writer
	stderr io.Writer
}

func New(commandRunner runner.CommandRunner, stdout, stderr io.Writer) Deployer {
	return Deployer{runner: commandRunner, stdout: stdout, stderr: stderr}
}

func (d Deployer) Deploy(ctx context.Context, cfg config.Config, opts Options) error {
	image := cfg.ImageRef()
	if !opts.NoBuild {
		if err := d.buildImage(ctx, cfg, image, opts.DryRun); err != nil {
			return err
		}
	}
	if !opts.NoPush && shouldPushImage(image) {
		if err := d.pushImage(ctx, image, opts.DryRun); err != nil {
			return err
		}
	}

	manifests := RenderManifests(cfg)
	if opts.DryRun {
		fmt.Fprintln(d.stdout, "---")
		fmt.Fprint(d.stdout, manifests)
		fmt.Fprintf(d.stdout, "\nplanned URL: https://%s\n", cfg.Domain)
		return nil
	}

	if _, err := d.runner.Run(ctx, strings.NewReader(manifests), "kubectl", "apply", "-f", "-"); err != nil {
		return fmt.Errorf("apply kubernetes manifests: %w", err)
	}

	if _, err := d.runner.Run(ctx, nil, "kubectl", "rollout", "status", "deployment/"+cfg.Name, "-n", cfg.Namespace, "--timeout=180s"); err != nil {
		d.rollback(ctx, cfg, "rollout status failed")
		return err
	}

	logs, err := d.runner.Run(ctx, nil, "kubectl", "logs", "deployment/"+cfg.Name, "-n", cfg.Namespace, "--tail=200")
	if err == nil {
		if pattern, ok := logscan.Match(logs.Stdout+"\n"+logs.Stderr, cfg.Rollback.FailureText); ok {
			d.rollback(ctx, cfg, "failure pattern detected in logs: "+pattern)
			return fmt.Errorf("deployment failed log scan: %s", pattern)
		}
	} else {
		fmt.Fprintf(d.stderr, "warning: unable to read rollout logs: %v\n", err)
	}

	fmt.Fprintf(d.stdout, "deployment succeeded: https://%s\n", cfg.Domain)
	return nil
}

func (d Deployer) buildImage(ctx context.Context, cfg config.Config, image string, dryRun bool) error {
	args := []string{"build", "-t", image, "-f", cfg.Dockerfile, cfg.Context}
	if dryRun {
		fmt.Fprintf(d.stdout, "docker %s\n", strings.Join(args, " "))
		return nil
	}
	if _, err := d.runner.Run(ctx, nil, "docker", args...); err != nil {
		return fmt.Errorf("build image %s: %w", image, err)
	}
	return nil
}

func (d Deployer) pushImage(ctx context.Context, image string, dryRun bool) error {
	args := []string{"push", image}
	if dryRun {
		fmt.Fprintf(d.stdout, "docker %s\n", strings.Join(args, " "))
		return nil
	}
	if _, err := d.runner.Run(ctx, nil, "docker", args...); err != nil {
		return fmt.Errorf("push image %s: %w", image, err)
	}
	return nil
}

func shouldPushImage(image string) bool {
	repository := image
	if idx := strings.LastIndex(repository, ":"); idx > strings.LastIndex(repository, "/") {
		repository = repository[:idx]
	}
	return strings.Contains(repository, "/")
}

func (d Deployer) rollback(ctx context.Context, cfg config.Config, reason string) {
	if !cfg.Rollback.Enabled {
		fmt.Fprintf(d.stderr, "rollback skipped: %s\n", reason)
		return
	}
	fmt.Fprintf(d.stderr, "rollback triggered: %s\n", reason)
	if _, err := d.runner.Run(ctx, nil, "kubectl", "rollout", "undo", "deployment/"+cfg.Name, "-n", cfg.Namespace); err != nil {
		fmt.Fprintf(d.stderr, "rollback failed: %v\n", err)
	}
}

type manifestData struct {
	config.Config
	ImageRef string
	Env      []envVar
}

type envVar struct {
	Name  string
	Value string
}

func RenderManifests(cfg config.Config) string {
	data := manifestData{Config: cfg, ImageRef: cfg.ImageRef(), Env: sortedEnv(cfg.Env)}
	var buf bytes.Buffer
	if err := manifestTemplate.Execute(&buf, data); err != nil {
		panic(err)
	}
	return strings.TrimSpace(buf.String()) + "\n"
}

func sortedEnv(values map[string]string) []envVar {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	out := make([]envVar, 0, len(keys))
	for _, key := range keys {
		out = append(out, envVar{Name: key, Value: values[key]})
	}
	return out
}

var manifestTemplate = template.Must(template.New("manifests").Parse(`apiVersion: v1
kind: Namespace
metadata:
  name: {{ .Namespace }}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Name }}
  namespace: {{ .Namespace }}
  labels:
    app.kubernetes.io/name: {{ .Name }}
    app.kubernetes.io/managed-by: deployctl
spec:
  replicas: {{ .Replicas }}
  revisionHistoryLimit: 5
  selector:
    matchLabels:
      app.kubernetes.io/name: {{ .Name }}
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 0
      maxSurge: 1
  template:
    metadata:
      labels:
        app.kubernetes.io/name: {{ .Name }}
    spec:
      containers:
        - name: {{ .Name }}
          image: {{ .ImageRef }}
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: {{ .Port }}
{{- if .Env }}
          env:
{{- range .Env }}
            - name: {{ .Name }}
              value: {{ printf "%q" .Value }}
{{- end }}
{{- end }}
          readinessProbe:
            httpGet:
              path: {{ .HealthPath }}
              port: {{ .Port }}
            initialDelaySeconds: 5
            periodSeconds: 10
          livenessProbe:
            httpGet:
              path: {{ .HealthPath }}
              port: {{ .Port }}
            initialDelaySeconds: 15
            periodSeconds: 20
---
apiVersion: v1
kind: Service
metadata:
  name: {{ .Name }}
  namespace: {{ .Namespace }}
  labels:
    app.kubernetes.io/name: {{ .Name }}
spec:
  type: ClusterIP
  selector:
    app.kubernetes.io/name: {{ .Name }}
  ports:
    - name: http
      port: 80
      targetPort: {{ .Port }}
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: {{ .Name }}
  namespace: {{ .Namespace }}
  annotations:
    cert-manager.io/cluster-issuer: {{ .ClusterIssuer }}
spec:
  ingressClassName: {{ .IngressClass }}
  tls:
    - hosts:
        - {{ .Domain }}
      secretName: {{ .Name }}-tls
  rules:
    - host: {{ .Domain }}
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: {{ .Name }}
                port:
                  number: 80
`))
