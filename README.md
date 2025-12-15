# homelab-test

Test repository for homelab Kubernetes cluster CI/CD pipeline testing.

## Quick Links

| 🌐 Live Sites | 📦 Repos |
|:--------------|:---------|
| [🧪 Test (Prod)](https://test.epoch.engineering) | [🏠 Homelab Infra](https://github.com/EpochBoy/homelab) |
| [🔬 Staging](https://staging.test.epoch.engineering) | |
| [🧑‍💻 Dev](https://dev.test.epoch.engineering) | |

## Purpose

This is a **proof-of-concept app** demonstrating the complete homelab deployment flow.

App repos should be **minimal** - just source code and a Dockerfile. Everything else (deployment manifests, CI pipelines, monitoring) lives in the **infra repo**.

## What's in this repo (app concerns)

```text
homelab-test/
├── Dockerfile              # How to build the app
├── main.go, go.mod         # Source code
├── VERSION                  # App version
└── README.md               # This file
```

## What's in the infra repo (platform concerns)

```text
homelab/
├── kubernetes/apps/homelab-test/   # Deployment manifests
├── ansible/tasks/homelab-test-deploy.yml  # ArgoCD Application
└── Woodpecker Config Service       # Auto-generates CI pipeline
```

## Complete Deployment Flow

```text
1. DEVELOPER PUSHES CODE
   └── Push to EpochBoy/homelab-test main branch

2. WOODPECKER CI (auto-triggered)
   └── Config Service auto-generates pipeline:
       ├── Pre-build: SAST, secret scan, SCA, IaC scan
       ├── Build: Detected Dockerfile → buildah build/push
       └── Post-build: SBOM, image scan, image signing

3. IMAGE PUSHED TO HARBOR
   └── registry.epoch.engineering/homelab/homelab-test:<sha>

4. RENOVATE DETECTS NEW IMAGE
   └── Creates PR to homelab repo updating image tag

5. ARGOCD SYNCS
   └── Deploys to cluster with new image
```

## Local Development

```bash
# Run locally
go run main.go

# Build container
docker build -t homelab-test .

# Test locally
curl http://localhost:8080/health
curl http://localhost:8080/version
```

## Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /health` | Health check (for probes) |
| `GET /version` | Version info (git commit, build time) |
| `GET /` | Welcome page |
