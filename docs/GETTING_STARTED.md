# Panduan Memulai (Getting Started with GitOpsIdp)

Panduan ini menjelaskan cara menjalankan **GitOpsIdp** di komputer lokal, menguji fitur *self-service scaffolding*, dan melihat integrasi GitOps secara langsung.

---

## 1. Prasyarat Sistem
- **Go** 1.22+ (opsional untuk menjalankan API server langsung)
- **Docker** & **Docker Compose** (direkomendasikan untuk menjalankan seluruh platform)
- **Git**

---

## 2. Cara Menjalankan

### Opsi A: Menggunakan Docker Compose (Direkomendasikan)
Jalankan satu perintah dari root direktori proyek:

```bash
docker compose up --build -d
```

Layanan akan aktif di:
- **IDP Developer Portal (UI)**: [http://localhost:3000](http://localhost:3000)
- **IDP Control Plane API**: [http://localhost:8080](http://localhost:8080)
- **API Health Check**: [http://localhost:8080/healthz](http://localhost:8080/healthz)

### Opsi B: Menjalankan Langsung via Go
Jika ingin menjalankan tanpa Docker:

```bash
# Jalankan API Server (otomatis melayani UI portal jika diakses di browser)
cd idp-api
go run ./cmd/server
```
Buka browser di [http://localhost:8080](http://localhost:8080).

---

## 3. Langkah Pengujian Self-Service Scaffolding

1. Buka browser di [http://localhost:3000](http://localhost:3000) (atau `http://localhost:8080`).
2. Masuk ke tab **"🚀 Golden Path Wizard"**.
3. Isi data formulir:
   - **Service Name**: `payment-service`
   - **Golden Path Template**: `Go Microservice (Go 1.24 HTTP Server)`
   - **Team / Owner**: `Fintech Squad`
   - **HTTP Port**: `8080`
   - Centang opsi `Auto-commit changes & GitOps overlays to Git repository`.
4. Klik tombol **"🚀 Scaffold & Push to GitOps"**.
5. Amati apa yang terjadi secara otomatis:
   - Kode sumber aplikasi baru dibuat di: `services/payment-service/`
   - Manifest Kubernetes base dibuat di: `gitops/apps/payment-service/base/`
   - Overlay untuk `dev`, `staging`, dan `prod` dibuat di: `gitops/apps/payment-service/overlays/`
   - Pendaftaran overlay otomatis ke `gitops/environments/dev/kustomization.yaml`
   - Perubahan langsung di-*commit* ke Git lokal secara otomatis.
   - Layanan langsung muncul di tab **Service Catalog** dan **Environments Topology**.

---

## 4. Pengujian API via cURL

Anda juga dapat memanggil API langsung:

### Melihat Katalog Layanan
```bash
curl http://localhost:8080/api/v1/services | jq .
```

### Melihat Status GitOps & Drift
```bash
curl http://localhost:8080/api/v1/gitops/status | jq .
```

### Melakukan Scaffolding Layanan via API
```bash
curl -X POST http://localhost:8080/api/v1/scaffold \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "notification-service",
    "language": "python",
    "owner": "Messaging Team",
    "port": "8000",
    "description": "Email and push notifications dispatch service",
    "auto_commit": true
  }' | jq .
```

### Memicu Sinkronisasi GitOps Manual
```bash
curl -X POST http://localhost:8080/api/v1/gitops/sync \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "sample-service",
    "environment": "dev"
  }' | jq .
```

---

## 5. Menghubungkan ke Cluster Kubernetes & ArgoCD Nyata

Jika Anda memiliki cluster Kubernetes aktif (Minikube, Kind, K3s, EKS, GKE):
1. Install ArgoCD di namespace `argocd`:
   ```bash
   kubectl create namespace argocd
   kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
   ```
2. Terapkan Root App-of-Apps:
   ```bash
   kubectl apply -f gitops/platform/argocd-apps/root-app-of-apps.yaml
   ```
3. Terapkan ApplicationSet otomatis:
   ```bash
   kubectl apply -f gitops/platform/argocd-apps/applicationset-services.yaml
   ```
Setiap kali ada layanan baru yang di-*scaffold* oleh GitOpsIdp, ArgoCD ApplicationSet akan otomatis mendeteksi foldernya dan menyebarkannya ke namespace yang bersangkutan!
