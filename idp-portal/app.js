const API_BASE = window.location.port === "3000" ? "http://localhost:8080" : "";

// State
let services = [];
let gitopsStatus = null;
let currentTab = "catalog";

// Elements
const tabCatalog = document.getElementById("tab-catalog");
const tabScaffold = document.getElementById("tab-scaffold");
const tabGitops = document.getElementById("tab-gitops");
const tabEnvs = document.getElementById("tab-envs");

const viewCatalog = document.getElementById("view-catalog");
const viewScaffold = document.getElementById("view-scaffold");
const viewGitops = document.getElementById("view-gitops");
const viewEnvs = document.getElementById("view-envs");

const catalogList = document.getElementById("catalog-list");
const scaffoldForm = document.getElementById("scaffold-form");
const gitopsTableBody = document.getElementById("gitops-table-body");
const envTopology = document.getElementById("env-topology");
const toast = document.getElementById("toast");

// Setup Navigation
function setupNavigation() {
  const tabs = [
    { btn: tabCatalog, view: viewCatalog, name: "catalog" },
    { btn: tabScaffold, view: viewScaffold, name: "scaffold" },
    { btn: tabGitops, view: viewGitops, name: "gitops" },
    { btn: tabEnvs, view: viewEnvs, name: "envs" },
  ];

  tabs.forEach(t => {
    t.btn.addEventListener("click", () => {
      tabs.forEach(item => {
        item.btn.classList.remove("active");
        item.view.classList.remove("active");
      });
      t.btn.classList.add("active");
      t.view.classList.add("active");
      currentTab = t.name;
    });
  });
}

function showToast(message, isError = false) {
  toast.innerText = message;
  toast.style.borderColor = isError ? "var(--danger)" : "var(--primary)";
  toast.style.display = "block";
  setTimeout(() => {
    toast.style.display = "none";
  }, 4000);
}

// Fetch Services
async function fetchServices() {
  try {
    const res = await fetch(`${API_BASE}/api/v1/services`);
    if (!res.ok) throw new Error("Failed to load services");
    services = await res.json();
    renderCatalog();
    renderEnvironments();
  } catch (err) {
    console.error("Error fetching services:", err);
  }
}

// Fetch GitOps Status
async function fetchGitOpsStatus() {
  try {
    const res = await fetch(`${API_BASE}/api/v1/gitops/status`);
    if (!res.ok) throw new Error("Failed to load GitOps status");
    gitopsStatus = await res.json();
    renderHeaderStatus();
    renderGitOpsInspector();
  } catch (err) {
    console.error("Error fetching gitops status:", err);
  }
}

function renderHeaderStatus() {
  const badge = document.getElementById("cluster-status-badge");
  const commitSpan = document.getElementById("header-commit");
  if (!gitopsStatus) return;

  commitSpan.innerText = `Git HEAD: ${gitopsStatus.head_commit} (${gitopsStatus.branch})`;

  if (gitopsStatus.out_of_sync_apps > 0) {
    badge.className = "status-badge status-outofsync";
    badge.innerText = `● ${gitopsStatus.out_of_sync_apps} Drift Detected`;
  } else {
    badge.className = "status-badge status-synced";
    badge.innerText = `● ArgoCD Synced (${gitopsStatus.synced_apps}/${gitopsStatus.total_apps})`;
  }
}

function renderCatalog() {
  if (!services || services.length === 0) {
    catalogList.innerHTML = `<div style="grid-column: 1/-1; text-align: center; color: var(--text-muted); padding: 3rem;">No services found in GitOps repository. Use the Golden Path Scaffolder to create one!</div>`;
    return;
  }

  catalogList.innerHTML = services.map(svc => {
    const envEntries = svc.environments ? Object.entries(svc.environments) : [];
    const envPills = envEntries.map(([env, dep]) => {
      const isSynced = dep.sync_status === "Synced";
      const statusColor = isSynced ? "var(--success)" : "var(--warning)";
      return `
        <div class="env-pill">
          <div class="env-pill-name">${env}</div>
          <div style="color: ${statusColor}; font-weight: 600;">${dep.sync_status}</div>
          <div style="color: var(--text-muted); font-size: 0.65rem;">${dep.replicas} replica(s)</div>
        </div>
      `;
    }).join("");

    return `
      <div class="card">
        <div class="card-title-group">
          <h3 class="card-title">${svc.name}</h3>
          <span class="badge-lang">${svc.language}</span>
        </div>
        <p class="card-desc">${svc.description || "Microservice managed via GitOpsIdp"}</p>
        <div class="card-meta">
          <div class="meta-row">
            <span>Owner</span>
            <span>${svc.owner}</span>
          </div>
          <div class="meta-row">
            <span>Repo Path</span>
            <span style="font-family: monospace; font-size: 0.75rem;">${svc.repo_path}</span>
          </div>
        </div>
        <div class="env-pills">
          ${envPills}
        </div>
      </div>
    `;
  }).join("");
}

function renderGitOpsInspector() {
  if (!gitopsStatus || !gitopsStatus.applications) return;

  gitopsTableBody.innerHTML = gitopsStatus.applications.map(app => {
    const isSynced = app.sync_status === "Synced";
    const statusClass = isSynced ? "status-synced" : "status-outofsync";
    const healthClass = app.health_status === "Healthy" ? "status-synced" : "status-degraded";

    return `
      <tr>
        <td style="font-weight: 600;">${app.name}</td>
        <td><span style="text-transform: uppercase; font-size: 0.75rem; background: #334155; padding: 2px 6px; border-radius: 4px;">${app.environment}</span></td>
        <td><span class="status-badge ${statusClass}">● ${app.sync_status}</span></td>
        <td><span class="status-badge ${healthClass}">● ${app.health_status}</span></td>
        <td style="font-family: monospace; font-size: 0.8rem;">${app.target_revision}</td>
        <td>
          <button class="btn btn-secondary btn-sm" onclick="triggerSync('${app.name}', '${app.environment}')">
            ⚡ Sync Now
          </button>
        </td>
      </tr>
    `;
  }).join("");
}

function renderEnvironments() {
  const envs = ["dev", "staging", "prod"];
  envTopology.innerHTML = envs.map(env => {
    const envServices = services.filter(s => s.environments && s.environments[env]);
    return `
      <div class="card" style="margin-bottom: 1.5rem;">
        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 1rem; border-bottom: 1px solid var(--border-color); padding-bottom: 0.75rem;">
          <h3 style="font-size: 1.25rem; text-transform: uppercase; font-weight: 700; color: #fff;">Environment: ${env}</h3>
          <span style="font-size: 0.8rem; color: var(--text-muted);">${envServices.length} Deployed Service(s)</span>
        </div>
        <div style="display: flex; flex-direction: column; gap: 0.75rem;">
          ${envServices.map(s => {
            const dep = s.environments[env];
            return `
              <div style="display: flex; justify-content: space-between; align-items: center; background: #0f172a; padding: 0.75rem 1rem; border-radius: 6px; border: 1px solid var(--border-color);">
                <div>
                  <div style="font-weight: 600; font-size: 0.9rem;">${s.name}</div>
                  <div style="font-size: 0.75rem; color: var(--text-muted);">${dep.endpoint_url} | Image: ${dep.image_tag}</div>
                </div>
                <div style="display: flex; align-items: center; gap: 0.5rem;">
                  <span class="status-badge status-synced" style="font-size: 0.75rem;">${dep.replicas} Pods</span>
                  <span class="status-badge status-synced" style="font-size: 0.75rem;">${dep.sync_status}</span>
                </div>
              </div>
            `;
          }).join("")}
        </div>
      </div>
    `;
  }).join("");
}

// Scaffolding Handler
async function handleScaffoldSubmit(e) {
  e.preventDefault();

  const submitBtn = document.getElementById("scaffold-submit-btn");
  submitBtn.disabled = true;
  submitBtn.innerText = "Generating & Committing to Git...";

  const payload = {
    service_name: document.getElementById("svc-name").value.trim(),
    language: document.getElementById("svc-lang").value,
    owner: document.getElementById("svc-owner").value.trim(),
    port: document.getElementById("svc-port").value.trim(),
    description: document.getElementById("svc-desc").value.trim(),
    cpu_request: document.getElementById("svc-cpu-req").value.trim(),
    cpu_limit: document.getElementById("svc-cpu-limit").value.trim(),
    memory_request: document.getElementById("svc-mem-req").value.trim(),
    memory_limit: document.getElementById("svc-mem-limit").value.trim(),
    auto_commit: document.getElementById("svc-autocommit").checked,
  };

  try {
    const res = await fetch(`${API_BASE}/api/v1/scaffold`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });

    const data = await res.json();
    if (!res.ok) throw new Error(data.error || "Scaffolding failed");

    showToast(`✓ Service '${data.service_name}' successfully generated & committed!`);
    scaffoldForm.reset();

    // Refresh and go to catalog
    await fetchServices();
    await fetchGitOpsStatus();
    tabCatalog.click();
  } catch (err) {
    showToast(`Error: ${err.message}`, true);
  } finally {
    submitBtn.disabled = false;
    submitBtn.innerText = "🚀 Scaffold & Push to GitOps";
  }
}

// Sync trigger
window.triggerSync = async function(appName, env) {
  try {
    const res = await fetch(`${API_BASE}/api/v1/gitops/sync`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ service_name: appName, environment: env }),
    });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || "Sync failed");

    showToast(`⚡ GitOps reconciled: ${appName} [${env}] is now Synced!`);
    await fetchServices();
    await fetchGitOpsStatus();
  } catch (err) {
    showToast(`Sync Error: ${err.message}`, true);
  }
};

// Language dropdown port helper
document.getElementById("svc-lang")?.addEventListener("change", (e) => {
  const portInput = document.getElementById("svc-port");
  if (e.target.value === "node") {
    portInput.value = "3000";
  } else if (e.target.value === "python") {
    portInput.value = "8000";
  } else {
    portInput.value = "8080";
  }
});

// Initialization
document.addEventListener("DOMContentLoaded", () => {
  setupNavigation();
  scaffoldForm.addEventListener("submit", handleScaffoldSubmit);

  fetchServices();
  fetchGitOpsStatus();

  // Poll every 5s
  setInterval(() => {
    fetchServices();
    fetchGitOpsStatus();
  }, 5000);
});
