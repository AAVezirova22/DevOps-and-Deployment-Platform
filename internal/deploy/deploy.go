package deploy

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/template"
	"time"

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

	if err := d.recordRelease(ctx, cfg, image, time.Now().UTC()); err != nil {
		return err
	}
	fmt.Fprintf(d.stdout, "deployment succeeded: https://%s\n", cfg.Domain)
	return nil
}

func (d Deployer) Status(ctx context.Context, cfg config.Config) error {
	commands := [][]string{
		{"kubectl", "get", "deployment/" + cfg.Name, "-n", cfg.Namespace, "-o", "wide"},
		{"kubectl", "get", "service/" + cfg.Name, "-n", cfg.Namespace},
		{"kubectl", "get", "ingress/" + cfg.Name, "-n", cfg.Namespace},
		{"kubectl", "get", "configmap", "-n", cfg.Namespace, "-l", "app.kubernetes.io/name=" + cfg.Name + ",deploykit.io/release=true", "--sort-by=.metadata.creationTimestamp"},
	}

	for _, command := range commands {
		result, err := d.runner.Run(ctx, nil, command[0], command[1:]...)
		if err != nil {
			return err
		}
		fmt.Fprint(d.stdout, result.Stdout)
		if result.Stderr != "" {
			fmt.Fprint(d.stderr, result.Stderr)
		}
	}

	certificate, err := d.runner.Run(ctx, nil, "kubectl", "get", "certificate/"+cfg.Name+"-tls", "-n", cfg.Namespace)
	if err != nil {
		fmt.Fprintf(d.stderr, "warning: certificate status unavailable; cert-manager may not be installed or the certificate may not exist yet\n")
		return nil
	}
	fmt.Fprint(d.stdout, certificate.Stdout)
	return nil
}

func (d Deployer) Rollback(ctx context.Context, cfg config.Config) error {
	fmt.Fprintf(d.stderr, "rollback triggered: manual rollback requested\n")
	if _, err := d.runner.Run(ctx, nil, "kubectl", "rollout", "undo", "deployment/"+cfg.Name, "-n", cfg.Namespace); err != nil {
		return err
	}
	if _, err := d.runner.Run(ctx, nil, "kubectl", "rollout", "status", "deployment/"+cfg.Name, "-n", cfg.Namespace, "--timeout=180s"); err != nil {
		return err
	}
	fmt.Fprintf(d.stdout, "rollback completed for %s in namespace %s\n", cfg.Name, cfg.Namespace)
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

func (d Deployer) recordRelease(ctx context.Context, cfg config.Config, image string, deployedAt time.Time) error {
	record := RenderReleaseRecord(cfg, image, deployedAt)
	if _, err := d.runner.Run(ctx, strings.NewReader(record), "kubectl", "apply", "-f", "-"); err != nil {
		return fmt.Errorf("record release history: %w", err)
	}
	return nil
}

type manifestData struct {
	config.Config
	ImageRef string
	Env      []envVar
	Secrets  []string
}

type envVar struct {
	Name  string
	Value string
}

func RenderManifests(cfg config.Config) string {
	data := manifestData{Config: cfg, ImageRef: cfg.ImageRef(), Env: sortedEnv(cfg.Env), Secrets: sortedStrings(cfg.SecretRefs)}
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

func sortedStrings(values []string) []string {
	out := append([]string(nil), values...)
	sort.Strings(out)
	return out
}

func RenderReleaseRecord(cfg config.Config, image string, deployedAt time.Time) string {
	data := releaseRecordData{
		Name:       releaseRecordName(cfg.Name, cfg.Tag),
		AppName:    cfg.Name,
		Namespace:  cfg.Namespace,
		Image:      image,
		Tag:        cfg.Tag,
		Domain:     cfg.Domain,
		DeployedAt: deployedAt.UTC().Format(time.RFC3339),
	}
	var buf bytes.Buffer
	if err := releaseRecordTemplate.Execute(&buf, data); err != nil {
		panic(err)
	}
	return strings.TrimSpace(buf.String()) + "\n"
}

type releaseRecordData struct {
	Name       string
	AppName    string
	Namespace  string
	Image      string
	Tag        string
	Domain     string
	DeployedAt string
}

func releaseRecordName(appName, tag string) string {
	base := sanitizeName(appName + "-release-" + tag)
	if len(base) <= 63 {
		return base
	}
	sum := sha1.Sum([]byte(base))
	suffix := "-" + hex.EncodeToString(sum[:4])
	return strings.TrimSuffix(base[:63-len(suffix)], "-") + suffix
}

func sanitizeName(value string) string {
	value = strings.ToLower(value)
	var b strings.Builder
	lastDash := false
	for _, r := range value {
		ok := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if ok {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "release"
	}
	return out
}

var releaseRecordTemplate = template.Must(template.New("releaseRecord").Parse(`apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Name }}
  namespace: {{ .Namespace }}
  labels:
    app.kubernetes.io/name: {{ .AppName }}
    app.kubernetes.io/managed-by: deployctl
    deploykit.io/release: "true"
data:
  image: {{ printf "%q" .Image }}
  tag: {{ printf "%q" .Tag }}
  domain: {{ printf "%q" .Domain }}
  deployedAt: {{ printf "%q" .DeployedAt }}
`))

var manifestTemplate = template.Must(template.New("manifests").Parse(`apiVersion: v1
kind: Namespace
metadata:
  name: {{ .Namespace }}
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ .ServiceAccount }}
  namespace: {{ .Namespace }}
  labels:
    app.kubernetes.io/name: {{ .Name }}
    app.kubernetes.io/managed-by: deployctl
automountServiceAccountToken: false
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: {{ .Name }}-runtime
  namespace: {{ .Namespace }}
  labels:
    app.kubernetes.io/name: {{ .Name }}
    app.kubernetes.io/managed-by: deployctl
rules: []
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: {{ .Name }}-runtime
  namespace: {{ .Namespace }}
  labels:
    app.kubernetes.io/name: {{ .Name }}
    app.kubernetes.io/managed-by: deployctl
subjects:
  - kind: ServiceAccount
    name: {{ .ServiceAccount }}
    namespace: {{ .Namespace }}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: {{ .Name }}-runtime
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
      serviceAccountName: {{ .ServiceAccount }}
      automountServiceAccountToken: false
      securityContext:
        runAsNonRoot: true
        runAsUser: 10001
        runAsGroup: 10001
        fsGroup: 10001
        seccompProfile:
          type: RuntimeDefault
      containers:
        - name: {{ .Name }}
          image: {{ .ImageRef }}
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: {{ .Port }}
          securityContext:
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: true
            capabilities:
              drop:
                - ALL
          resources:
            requests:
              cpu: {{ printf "%q" .Resources.CPURequest }}
              memory: {{ printf "%q" .Resources.MemoryRequest }}
            limits:
              cpu: {{ printf "%q" .Resources.CPULimit }}
              memory: {{ printf "%q" .Resources.MemoryLimit }}
{{- if .Env }}
          env:
{{- range .Env }}
            - name: {{ .Name }}
              value: {{ printf "%q" .Value }}
{{- end }}
{{- end }}
{{- if .Secrets }}
          envFrom:
{{- range .Secrets }}
            - secretRef:
                name: {{ . }}
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
