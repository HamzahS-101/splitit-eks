# splitit on AWS EKS

A no-login expense splitter, deployed to Amazon EKS with Terraform, GitHub Actions and Argo CD.

Create a group, share the link, add expenses and see who owes whom. A Go API and a React frontend run on Kubernetes behind Traefik, with RDS PostgreSQL and a Redis cache. Git is the source of truth for the cluster: a merged app change was live in under 4 minutes, with no manual deploy step.

![Balances and settle-up payments](docs/images/app-balances.png)

## Architecture

| Service | Built with | What it does |
|---------|------------|--------------|
| **web** | React, NGINX | Serves the static frontend to the browser |
| **api** | Go | Handles groups, members, expenses and balances |
| **redis** | Redis 7 | Caches balances for 5 minutes. Cleared when a group changes |
| **database** | RDS PostgreSQL 16 | Stores all the data |

**Request path:** Route 53 → Network Load Balancer (TCP) → Traefik pods (TLS ends here, HTTP redirects to HTTPS) → `/api` goes to the API and everything else to web. The browser calls the API on the same hostname, so web and API never talk to each other. Only the API uses Redis and Postgres.

**Network:** one VPC across two Availability Zones in eu-west-2, with three subnet tiers.

- **Public:** the load balancer and one NAT gateway.
- **App:** the EKS nodes and pods, with no public IPs.
- **Data:** RDS, which only accepts connections from the nodes.

**Who owns what:**

- **Terraform** creates the AWS infrastructure, installs Argo CD and creates the root application.
- **Argo CD** runs everything else in the cluster from Git.
- **GitHub Actions** builds images and updates their tags in Git. It never deploys the app directly.

**Healthy Traefik targets in both Availability Zones:**

![Load balancer targets](docs/images/nlb-targets.png)

**HTTPS certificate issued by Let's Encrypt through cert-manager:**

![HTTPS certificate](docs/images/https-certificate.png)

## Contents

- [Repo layout](#repo-layout)
- [Key components](#key-components)
- [Deployment workflow](#deployment-workflow)
- [Decisions and trade-offs](#decisions-and-trade-offs)
- [Build practices](#build-practices)
- [Proof it works](#proof-it-works)
- [Troubleshooting](#troubleshooting)
- [Security practices](#security-practices)
- [Run it locally](#run-it-locally)
- [Deploy to AWS](#deploy-to-aws)
- [Teardown](#teardown)
- [Planned improvements](#planned-improvements)

## Repo layout

```
.
├── .github/workflows/
├── api/
│   ├── cmd/splitit/
│   ├── internal/
│   ├── migrations/
│   └── Dockerfile
├── web/
│   ├── src/
│   ├── nginx.conf
│   └── Dockerfile
├── splitit-helm/
│   ├── templates/
│   └── values.yaml
├── argocd/
│   ├── apps/
│   ├── cluster-config/
│   ├── root-app.yaml
│   └── values.yaml
├── terraform/
│   └── modules/
│       ├── network/
│       ├── eks/
│       ├── irsa/
│       ├── database/
│       └── argocd/
├── compose/
├── docs/images/
├── docker-compose.yml
└── .pre-commit-config.yaml
```

## Key components

| Component | What it does | Where |
|-----------|--------------|-------|
| **Cluster** | EKS 1.37 with two ARM64 `t4g.medium` nodes across two AZs | [`terraform/modules/eks`](terraform/modules/eks) |
| **Network** | Three subnet tiers and one NAT gateway | [`terraform/modules/network`](terraform/modules/network) |
| **Database** | Private, encrypted, single-AZ RDS PostgreSQL | [`terraform/modules/database`](terraform/modules/database) |
| **Pod permissions** | Five IRSA roles, one for each component that calls AWS | [`terraform/modules/irsa`](terraform/modules/irsa) |
| **GitOps** | A root application managing 11 child applications, with pruning and self-heal | [`argocd`](argocd) |
| **App chart** | Deployments, probes, disruption budgets, quota, API autoscaling (2 to 4 pods) and a migration Job | [`splitit-helm`](splitit-helm) |
| **Ingress, DNS and TLS** | Traefik with the Gateway API, AWS Load Balancer Controller, ExternalDNS and cert-manager | [`argocd/apps`](argocd/apps) |
| **Secrets** | External Secrets copies values from Secrets Manager and Parameter Store | [`externalsecret.yaml`](splitit-helm/templates/externalsecret.yaml) |
| **Monitoring** | Prometheus scrapes the cluster, API and Traefik. Grafana has cluster and Traefik dashboards | [`kube-prometheus-stack.yaml`](argocd/apps/kube-prometheus-stack.yaml) |
| **Scheduled job** | Kubernetes CronJob for inactive groups | [`cleanup.yaml`](splitit-helm/templates/cleanup.yaml) |
| **CI/CD** | Image, GitOps and Terraform workflows, using OIDC to reach AWS | [`.github/workflows`](.github/workflows) |

## Deployment workflow

**1. Pull request**

Checks run for whatever changed:

- **App code:** both images are built and scanned with Trivy. Fixable criticals fail the check. Nothing is pushed.
- **Helm and Argo CD:** Gitleaks scans for secrets, and kube-linter and Checkov scan the chart. Checkov only reports.
- **Terraform:** format, TFLint, Checkov, validate and plan.

**2. Merge to main**

The image workflow builds and scans again, pushes to ECR with the commit SHA as the tag, then commits that tag to the Helm values.

**3. Argo CD sync**

Argo CD checks Git every 2 to 3 minutes. It sees the tag commit and rolls out the new pods. A new pod must be ready before an old one stops, and migrations run after the sync.

**Argo CD synced to the pipeline's tag commit:**

![Argo CD synced](docs/images/argocd-synced.png)

**Before and after a merge.** The only change was the headline. It merged at 15:45:38 and was live at 15:49:28.

| Before | After |
|--------|-------|
| ![Before the deploy](docs/images/deploy-before.png) | ![After the deploy](docs/images/deploy-after.png) |

**Infrastructure changes**

Terraform plans on `main`, then apply waits for approval on the protected `prod` environment.

**Bootstrap order**

On a new cluster, sync waves create the applications in dependency order:

| Wave | Applications |
|------|--------------|
| 0 | Gateway API and Prometheus Operator CRDs |
| 1 | cert-manager, External Secrets, ExternalDNS, metrics-server |
| 2 | AWS Load Balancer Controller |
| 3 | Traefik, cluster config |
| 4 | splitit, Prometheus and Grafana |

![App of apps](docs/images/argocd-app-of-apps.png)

## Decisions and trade-offs

| Decision | Why | Trade-off |
|----------|-----|-----------|
| **App of apps, bootstrapped by Terraform** | One apply gives a working cluster, then Git manages what runs in it | Terraform needs access to the Kubernetes API |
| **Gateway API with Traefik, not Ingress** | The standard successor to Ingress. Each app owns its routes and listeners | Extra CRDs, and Traefik is another component to run |
| **RDS, not Postgres in the cluster** | The data isn't tied to a node or volume, and AWS handles backups and patching | Costs more, and single-AZ is a single point of failure |
| **IRSA** | Each component gets its own temporary AWS credentials | Each cluster needs an OIDC provider. EKS Pod Identity is a simpler alternative |
| **External Secrets and a write-only password** | The password is in neither Git nor Terraform state | Another controller, and rotation isn't automatic |
| **cert-manager, not ACM** | TLS ends at Traefik, and each Gateway gets a certificate that renews itself | Certificates and keys live in the cluster |
| **One NAT gateway** | About half the cost of one per AZ | If its AZ fails, both AZs lose outbound internet |
| **Public EKS endpoint** | GitHub-hosted runners can reach it. IAM and access entries control who gets in | Reachable from the internet. A private endpoint with runners in the VPC would close it |

## Build practices

- **Pre-commit hooks:** gitleaks, `terraform fmt`, `validate` and TFLint, hadolint, `helm lint`, gofmt and basic file checks.
- **Git:** feature branches for each section, merged through pull requests.
- **Images:** multi-stage, non-root, built natively for ARM64 and tagged with the commit SHA.
- **Versions:** Helm charts, GitHub Actions and base images use version tags, never `latest`.
- **Terraform:** one module per area, encrypted remote state with locking, and default tags.
- **CI:** no stored AWS keys, and approval before every apply or destroy.

## Proof it works

**Zero downtime during a node drain:** all 368 requests returned 200. The app and Traefik pods moved to the remaining node.

![Node drain](docs/images/node-drain.png)

**Disruption budget:** draining the second node was refused, because it would have taken down the last ready web pod.

![Disruption budget refusing an eviction](docs/images/pdb-refusal.png)

**Self-heal:** a manual scale from 2 to 5 replicas was reverted by Argo CD within about a second.

![Self-heal](docs/images/self-heal.png)

**Autoscaling:** under load, the API went from 2 to 4 pods in about 30 seconds.

![API autoscaling](docs/images/hpa-scaling.png)

**Monitoring:** Grafana showing the splitit pods. CPU limits show no data because none are set, to avoid throttling.

![Grafana pod dashboard](docs/images/grafana-pods.png)

## Troubleshooting

### The load balancer controller's webhook rejected requests

The chart generated a new webhook certificate on every render. Argo CD renders often, so the certificate Kubernetes trusted stopped matching the one the pods served, and Traefik never got a load balancer.

**Fix:** `enableCertManager: true`, so cert-manager issues and manages the certificate.

### splitit deadlocked on first boot

splitit synced before the External Secrets webhook was ready, so its secret was never created. The API waited for the secret, and the sync waited for the API, so it never finished or retried.

**Fix:** sync waves, plus an [Application health check](argocd/values.yaml) so the root app waits on each wave's health.

### The load balancer blocked destroy

The load balancer is created by a controller, so Terraform doesn't know about it. Left behind, it blocks deletion of the VPC.

**Fix:** the destroy workflow deletes the Argo CD applications and Traefik's Service first, and stops if the load balancer still exists.

## Security practices

**Access**

- GitHub Actions assumes an AWS role through OIDC, and the role only trusts this repository.
- IRSA roles are tied to one service account each. ExternalDNS and cert-manager can only change records in this hosted zone.
- Cluster access is granted through explicit EKS access entries.

**Secrets**

- Terraform generates the database password as an ephemeral value and writes it with write-only arguments.
- Pods read the password from a mounted file, not an environment variable.

**Network and encryption**

- The load balancer is the only public way into the app. Nodes and RDS are private.
- RDS storage and Terraform state are encrypted. Traffic uses HTTPS to Traefik and TLS to RDS.

**Containers**

- All splitit containers run as non-root with read-only root filesystems.
- The namespace has a resource quota and default limits.
- Images with fixable critical vulnerabilities aren't pushed to ECR.

**What the scans changed**

- kube-linter's 9 findings went to 0 with no exclusions: read-only filesystems, safer eviction policies for unhealthy pods, anti-affinity for web and automatic removal of finished migration Jobs.
- Those fixes took Checkov from 70 findings to 60. The rest are false positives, deliberate choices such as no CPU limits, or listed in [Planned improvements](#planned-improvements).
- Trivy failed the web image, which led to a newer NGINX base image.

## Run it locally

```bash
git clone https://github.com/HamzahS-101/splitit-eks.git
cd splitit-eks
docker compose up --build
```

Open http://localhost:8080. `docker compose down` stops it. Add `-v` to also delete the database.

## Deploy to AWS

**One-time setup:**

1. An encrypted S3 bucket for Terraform state.
2. ECR repositories `splitit-api` and `splitit-web`, with immutable tags.
3. A Route 53 hosted zone for your domain.
4. A GitHub OIDC provider and an IAM role that trusts your repository.
5. A Secrets Manager secret `splitit/grafana/admin` with the keys `admin-user` and `admin-password`.
6. A GitHub environment called `prod` with a required reviewer.

Replace the account ID, role ARNs, bucket, domain, zone ID and repo URL in `.github/workflows/`, `terraform/`, `argocd/` and `splitit-helm/values.yaml`.

**Deploy:**

1. Run **Docker build, scan and push to ECR**.
2. Run **Terraform plan**, review it, then approve **Terraform apply**. It creates 76 resources in about 18 minutes, and Argo CD installs the rest.

**Cost:** about $0.29 an hour while running, and about $1.50 a month for what stays after a teardown.

![Pipelines passing](docs/images/pipelines-green.png)

## Teardown

Run **Terraform destroy** and approve it. It takes about 13 minutes.

The one-time setup stays. DNS records created by ExternalDNS also stay, so delete those by hand if you aren't redeploying.

## Planned improvements

**Reliability**

- Multi-AZ RDS, a NAT gateway per AZ and a Redis replica.
- Node autoscaling.
- Finish and end-to-end test the cleanup job.

**Deployments**

- Branch protection with required checks, and schema validation with kubeconform.
- An Argo CD webhook instead of polling.
- Canary releases with Argo Rollouts.
- Cluster config one wave after Traefik.

**Security**

- A private EKS endpoint and a narrower GitHub Actions role.
- NetworkPolicies, rate limiting and data subnets with no internet route.
- Dropped capabilities, seccomp, high UIDs and no mounted service account tokens.
- An admission policy, secret rotation and SSO for Argo CD and Grafana.

**Observability**

- Alertmanager, log shipping and persistent Prometheus storage.
