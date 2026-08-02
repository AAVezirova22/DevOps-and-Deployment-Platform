package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ani/devops-deployment-platform/internal/config"
	"github.com/ani/devops-deployment-platform/internal/deploy"
	"github.com/ani/devops-deployment-platform/internal/runner"
)

const usage = `deployctl is a lightweight self-hosted deployment platform CLI.

Usage:
  deployctl init [--force]
  deployctl deploy [--config deploykit.yaml] [--dry-run] [--no-build] [--no-push] [--tag v1]
  deployctl status [--config deploykit.yaml]
  deployctl rollback [--config deploykit.yaml]
  deployctl render [--config deploykit.yaml]

Commands:
  init      create a starter deploykit.yaml for the current repository
  deploy    build a container, apply Kubernetes resources, verify rollout, rollback on failure
  status    show deployment, ingress, certificate, and release-record status
  rollback  manually roll back the Kubernetes deployment to the previous revision
  render    print the Kubernetes manifests without applying them
`

// Run executes the deployctl command line interface.
func Run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Fprint(stdout, usage)
		return nil
	}

	switch args[0] {
	case "init":
		return runInit(args[1:], stdout)
	case "deploy":
		return runDeploy(ctx, args[1:], stdout, stderr)
	case "status":
		return runStatus(ctx, args[1:], stdout, stderr)
	case "rollback":
		return runRollback(ctx, args[1:], stdout, stderr)
	case "render":
		return runRender(args[1:], stdout)
	default:
		return fmt.Errorf("unknown command %q\n\n%s", args[0], usage)
	}
}

func runInit(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	force := fs.Bool("force", false, "overwrite existing deploykit.yaml")
	if err := fs.Parse(args); err != nil {
		return err
	}

	path := "deploykit.yaml"
	if _, err := os.Stat(path); err == nil && !*force {
		return fmt.Errorf("%s already exists; rerun with --force to overwrite", path)
	}

	projectName := filepath.Base(mustGetwd())
	projectName = strings.ToLower(strings.ReplaceAll(projectName, " ", "-"))
	starter := config.StarterYAML(projectName)
	if err := os.WriteFile(path, []byte(starter), 0o644); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "created %s\n", path)
	return nil
}

func loadConfigFromFlags(command string, args []string) (config.Config, error) {
	fs := flag.NewFlagSet(command, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	configPath := fs.String("config", "deploykit.yaml", "path to deployment config")
	if err := fs.Parse(args); err != nil {
		return config.Config{}, err
	}

	cfg, err := config.LoadFile(*configPath)
	if err != nil {
		return config.Config{}, err
	}
	cfg.ApplyDefaults()
	if err := cfg.Validate(); err != nil {
		return config.Config{}, err
	}
	return cfg, nil
}

func runDeploy(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("deploy", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	configPath := fs.String("config", "deploykit.yaml", "path to deployment config")
	dryRun := fs.Bool("dry-run", false, "print commands and manifests without changing the cluster")
	noBuild := fs.Bool("no-build", false, "skip docker build")
	noPush := fs.Bool("no-push", false, "skip docker push")
	tag := fs.String("tag", "", "image tag override")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.LoadFile(*configPath)
	if err != nil {
		return err
	}
	if *tag != "" {
		cfg.Tag = *tag
	}
	cfg.ApplyDefaults()
	if err := cfg.Validate(); err != nil {
		return err
	}

	return deploy.New(runner.OSCommandRunner{}, stdout, stderr).Deploy(ctx, cfg, deploy.Options{
		DryRun:  *dryRun,
		NoBuild: *noBuild,
		NoPush:  *noPush,
	})
}

func runStatus(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	cfg, err := loadConfigFromFlags("status", args)
	if err != nil {
		return err
	}
	return deploy.New(runner.OSCommandRunner{}, stdout, stderr).Status(ctx, cfg)
}

func runRollback(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	cfg, err := loadConfigFromFlags("rollback", args)
	if err != nil {
		return err
	}
	return deploy.New(runner.OSCommandRunner{}, stdout, stderr).Rollback(ctx, cfg)
}

func runRender(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("render", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	configPath := fs.String("config", "deploykit.yaml", "path to deployment config")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.LoadFile(*configPath)
	if err != nil {
		return err
	}
	cfg.ApplyDefaults()
	if err := cfg.Validate(); err != nil {
		return err
	}
	fmt.Fprint(stdout, deploy.RenderManifests(cfg))
	return nil
}

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return "app"
	}
	return wd
}
