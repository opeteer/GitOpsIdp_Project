# ⚡ GitOpsIdp: GitOps-Driven Internal Developer Platform

[![Platform Engineering](https://img.shields.io/badge/Platform%20Engineering-IDP-blue.svg)](https://internaldeveloperplatform.org/)
[![GitOps](https://img.shields.io/badge/GitOps-ArgoCD%20%7C%20Kustomize-orange.svg)](https://argo-cd.readthedocs.io/)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8.svg?logo=go)](https://go.dev/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED.svg?logo=docker)](https://www.docker.com/)

**GitOpsIdp** adalah implementasi *turnkey* **Internal Developer Platform (IDP)** berbasis arsitektur **GitOps**. Platform ini dirancang untuk menyederhanakan alur kerja pengembang (*Developer Experience / DevEx*) dengan menyediakan portal *self-service*, otomatisasi pembuatan repositori dan mikroservis (*Golden Path Scaffolding*), serta integrasi deklaratif ke manifest Kubernetes dan ArgoCD.

---

## 🌟 Fitur Utama

- 📋 **Service Catalog**: Katalog terpadu yang melacak seluruh mikroservis di organisasi, kepemilikan tim (*team ownership*), endpoint, dan status deployment multi-environment (`dev`, `staging`, `prod`).
- 🚀 **Golden Path Scaffolding Wizard**: Generator layanan instan berbasis template teruji (*Go*, *Node.js Express*, *Python FastAPI*) yang otomatis membuat kode sumber, Dockerfile multi-stage, dan manifest K8s/Kustomize.
- 🔀 **Automated GitOps Integration**: Setiap layanan baru yang dibuat langsung menghasilkan struktur deklaratif Kubernetes (`base` dan `overlays/{dev,staging,prod}`) serta melakukan *commit* otomatis ke repositori Git.
- 🔄 **GitOps & ArgoCD Monitor**: Memantau commit SHA terbaru, status sinkronisasi deklaratif (*Synced*, *OutOfSync*), deteksi *drift*, dan menyediakan pemicu rekonsiliasi instan (*⚡ Sync Now*).
- 🌐 **Environments Topology**: Visualisasi promosi aplikasi antar namespace/cluster (`dev` -> `staging` -> `prod`).
- 🐳 **Zero-Friction Local Experience**: Dapat dijalankan sepenuhnya secara lokal menggunakan `docker compose` atau binary Go mandiri.

---

## 🏗️ Arsitektur Proyek

```
GitOpsIdp_Project/
├── idp-portal/              # Frontend Web Developer Portal (SPA + Nginx)
│   ├── index.html           # UI Dashboard: Catalog, Scaffolder, GitOps Monitor
│   ├── app.js               # Dynamic client-side logic & API integration
│   ├── styles.css           # Dark theme modern design system
│   ├── Dockerfile
│   └── nginx.conf
│
├── idp-api/                 # Backend Control Plane API (Go)
│   ├── cmd/server/          # HTTP Server entrypoint
│   ├── internal/
│   │   ├── catalog/         # Service discovery & GitOps scanner
│   │   ├── scaffolder/      # Template generation engine
│   │   ├── gitops/          # Git commit automation & ArgoCD bridge
│   │   ├── api/             # REST endpoints (/services, /scaffold, /gitops)
│   │   └── model/           # Data definitions
│   ├── Dockerfile
│   └── go.mod
│
├── gitops/                  # Repositori Deklaratif GitOps
│   ├── apps/                # Manifest per-service (base & overlays)
│   │   └── sample-service/
│   ├── environments/        # Kustomize agregator (dev, staging, prod)
│   └── platform/            # Manifest ArgoCD App-of-Apps & ApplicationSet
│       └── argocd-apps/
│
├── templates/               # Golden Path blueprints
│   ├── go-service/          # Go HTTP API blueprint
│   ├── node-service/        # Node.js Express blueprint
│   ├── python-service/      # Python FastAPI blueprint
│   └── gitops-manifests/    # K8s Deployment, Service, Kustomize blueprints
│
├── services/                # Layanan mikroservis yang di-scaffold
│   └── sample-service/      # Mikroservis contoh yang sudah terdaftar
│
├── docs/                    # Dokumentasi lengkap
│   ├── ARCHITECTURE.md      # Detail arsitektur & diagram alur
│   └── GETTING_STARTED.md   # Panduan instalasi dan pengujian
│
└── docker-compose.yml       # Orkestrasi lokal satu perintah
```

---

## ⚡ Quickstart

### 1. Menjalankan dengan Docker Compose
```bash
docker compose up --build -d
```
Akses platform di:
- **Developer Portal**: [http://localhost:3000](http://localhost:3000)
- **Control Plane API**: [http://localhost:8080](http://localhost:8080)

### 2. Menjalankan secara Native (Go)
```bash
cd idp-api
go run ./cmd/server
```
Buka [http://localhost:8080](http://localhost:8080) di browser Anda.

---

## 📖 Dokumentasi Lengkap
- [Detail Arsitektur & Pola GitOps](docs/ARCHITECTURE.md)
- [Panduan Pengujian & API Reference](docs/GETTING_STARTED.md)

---

## 📄 Lisensi
Didistribusikan di bawah lisensi MIT.
