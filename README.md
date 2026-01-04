# epochcloud-test

Test repository for EpochCloud Kubernetes cluster CI/CD pipeline testing.

## Quick Links

| 🌐 Live Sites | 📦 Repos |
| :------------- | :-------- |
| [🧪 Test (Prod)](https://test.<your-domain>) | [☁️ EpochCloud Infra](https://github.com/EpochBoy/epochcloud) |
| [🔬 Staging](https://test-staging.<your-domain>) | |
| [🧑‍💻 Dev](https://test-dev.<your-domain>) | |

## Purpose

This is a **proof-of-concept app** demonstrating the complete EpochCloud deployment flow.

App repos should be **minimal** - just source code and a Dockerfile. Everything else (deployment manifests, CI pipelines, monitoring) lives in the **infra repo**.

## What's in this repo (app concerns)

```text
epochcloud-test/
├── Dockerfile              # How to build the app
├── main.go, go.mod         # Source code
├── VERSION                 # App version
└── README.md               # This file
```

## What's in the infra repo (platform concerns)

```text
epochcloud/
├── kubernetes/apps/epochcloud-test/    # Deployment manifests
└── kubernetes/infrastructure/       # CI pipelines (Argo Workflows)
```

## Complete Deployment Flow

```text
1. DEVELOPER PUSHES CODE
   └── Push to EpochBoy/epochcloud-test main branch

2. ARGO WORKFLOWS CI (webhook triggered)
   └── GitHub App EventSource triggers app-baseline pipeline:
       ├── Pre-build: Semgrep SAST, TruffleHog secrets, OSV-Scanner SCA
       ├── Build: Buildah container build + push to Harbor
       └── Post-build: Trivy scan, Grype CVE, Syft SBOM, Cosign signing

3. IMAGE PUSHED TO HARBOR
   └── registry.<your-domain>/epochcloud/epochcloud-test:<sha>

4. KARGO DETECTS NEW IMAGE (Warehouse polls Harbor)
   └── Auto-promotes to dev environment

5. KARGO PROMOTES TO STAGING
   └── Auto-promotion policy triggers staging deployment
   └── OWASP ZAP DAST scan runs as verification gate

6. KARGO PROMOTES TO PRODUCTION
   └── Manual promotion required (click in Kargo UI)
   └── ArgoCD syncs production deployment
```

## Deployment Tools

| Tool | What it does | When it runs |
| ---- | ------------ | ------------ |
| **Renovate** | Updates Dockerfile base images (alpine, golang) | Creates PRs for external deps |
| **Kargo** | Promotes images through dev→staging→prod | After Argo Workflows pushes to Harbor |
| **ArgoCD** | Syncs deployments to cluster | When Kargo updates image tags |

## Local Development

```bash
# Run locally
go run main.go

# Build container
docker build -t epochcloud-test .

# Test locally
curl http://localhost:8080/health
curl http://localhost:8080/version
```

## Endpoints

| Endpoint | Description |
| -------- | ----------- |
| `GET /health` | Health check (for probes) |
| `GET /version` | Version info (git commit, build time) |
| `GET /metrics` | Prometheus metrics (scraped automatically) |
| `GET /` | Welcome page |

## Observability

This app demonstrates full observability integration with the EpochCloud platform:

### Prometheus Metrics

The `/metrics` endpoint exposes:

| Metric | Type | Description |
| ------ | ---- | ----------- |
| `epochcloud_http_requests_total` | Counter | Total HTTP requests by method, path, and status |
| `epochcloud_http_request_duration_seconds` | Histogram | Request latency distribution (p50, p95, p99) |
| `epochcloud_app_info` | Gauge | App metadata (version, commit, start time) |

### Platform Integration

- **PodMonitor** auto-discovers all pods with `app.kubernetes.io/part-of: epochcloud` label
- **Grafana Dashboard** shows request rates, latency percentiles, and running instances
- **Promtail** collects logs → **Loki** for aggregation
- **Tempo** for distributed tracing (OTLP instrumentation ready)
