# splitit

A no-login expense splitter. Create a group, share the link, add expenses and see who owes whom. The app works out the fewest payments needed to settle up.

## What it does

- Create a group and share its link. The link is the access, so there are no accounts.
- Add members and expenses, split equally or with custom amounts.
- See each person's balance and the settle-up payments.
- Money is stored as integer pence, so there are no rounding errors.

## Key components

| Component | What it is |
|-----------|------------|
| **API** | Go service. Handles groups, members, expenses and balance calculation. Runs as a non-root user in a distroless image |
| **Web** | React + Vite single-page app, built in a multi-stage image and served by unprivileged NGINX |
| **Postgres** | Stores groups, members, expenses and shares. A one-shot migration container creates the schema before the API starts |
| **Redis** | Caches calculated balances. The cache is cleared whenever an expense changes |
| **Gateway** | NGINX in compose, routing `/` to the web app and `/api` to the API, the same split Traefik makes in the cluster |
| **Kubernetes** | Manifests for every service: Postgres StatefulSet, Redis, migration Job, API and web Deployments, with health probes and resource requests and limits |
| **Traefik** | Gateway API ingress, routing `/` to web and `/api` to the API |
| **Terraform** | VPC and EKS cluster, with IRSA giving pods least-privilege AWS access |
| **Argo CD** | GitOps: Git is the source of truth for everything running in the cluster |
| **CI/CD** | GitHub Actions builds, scans with Trivy and pushes images, then updates the image tag in Git |
| **Monitoring** | Prometheus and Grafana dashboards for the cluster and the app |

## Build practices

- Multi-stage Dockerfiles for both images, with pinned base images.
- Both images pass Trivy with no high or critical findings.
- The database password is a mounted secret file, not an environment variable.
- Pre-commit hooks: gitleaks, hadolint, gofmt and basic file checks.

## Run it locally

```bash
docker compose up --build
```

Open http://localhost:8080.
