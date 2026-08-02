# Target Customers

DeployKit targets engineers and teams that need reliable deployments without a large platform engineering investment.

## Solo Developers

Solo developers often have several projects but do not want to configure hosting differently for each one. DeployKit gives them a repeatable path:

- write a Dockerfile,
- set a domain,
- run `deployctl deploy`,
- receive a live HTTPS endpoint.

The benefit is less time spent remembering deployment steps and more consistency across projects.

## Early-Stage Startups

Small startups may not need a full internal developer platform, but they still need production basics: rollback, health checks, TLS, logs, and infrastructure as code.

DeployKit helps by providing an understandable deployment layer on top of Kubernetes. It can start on a low-cost k3s node and later move to managed Kubernetes.

## Agencies and Freelancers

Agencies often host many small client applications. Namespace isolation and a config-per-service model make it easier to separate applications while using a shared cluster.

The project is useful when each client service needs:

- its own domain,
- independent rollout history,
- environment variables,
- simple recovery from bad deploys.

## Internal Tools Teams

Internal tools are often important but not large enough to justify heavyweight infrastructure. DeployKit gives these teams a pragmatic platform:

- simple CLI workflow,
- Kubernetes-native runtime,
- automated HTTPS,
- repeatable infrastructure bootstrap.

## Engineering Portfolio Reviewers

For hiring managers and senior engineers, this project demonstrates more than CRUD application development. It shows practical knowledge of:

- container builds,
- Kubernetes deployment primitives,
- rollout safety,
- cloud provisioning,
- CI/CD automation,
- operational documentation,
- customer and product positioning.

That combination is why the project is useful as a senior-level portfolio piece.
