package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/opeteer/gitops-idp/idp-api/internal/catalog"
	"github.com/opeteer/gitops-idp/idp-api/internal/gitops"
	"github.com/opeteer/gitops-idp/idp-api/internal/model"
	"github.com/opeteer/gitops-idp/idp-api/internal/scaffolder"
)

type Server struct {
	catalog    *catalog.Catalog
	scaffolder *scaffolder.Scaffolder
	gitops     *gitops.Engine
	portalDir  string
}

func NewServer(cat *catalog.Catalog, scaf *scaffolder.Scaffolder, git *gitops.Engine, portalDir string) *Server {
	return &Server{
		catalog:    cat,
		scaffolder: scaf,
		gitops:     git,
		portalDir:  portalDir,
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", s.handleHealthz)
	mux.HandleFunc("/api/v1/services", s.handleServices)
	mux.HandleFunc("/api/v1/services/", s.handleServiceByName)
	mux.HandleFunc("/api/v1/scaffold", s.handleScaffold)
	mux.HandleFunc("/api/v1/gitops/status", s.handleGitOpsStatus)
	mux.HandleFunc("/api/v1/gitops/sync", s.handleGitOpsSync)
	mux.HandleFunc("/api/v1/templates", s.handleTemplates)

	if s.portalDir != "" {
		fs := http.FileServer(http.Dir(s.portalDir))
		mux.Handle("/", fs)
	}

	return enableCORS(mux)
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "idp-api",
		"version": "1.0.0",
	})
}

func (s *Server) handleServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	_ = s.catalog.RefreshFromGitOps()
	services := s.catalog.ListServices()
	writeJSON(w, http.StatusOK, services)
}

func (s *Server) handleServiceByName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	name := strings.TrimPrefix(r.URL.Path, "/api/v1/services/")
	if name == "" {
		writeError(w, http.StatusBadRequest, "Service name missing")
		return
	}

	svc, exists := s.catalog.GetService(name)
	if !exists {
		writeError(w, http.StatusNotFound, fmt.Sprintf("Service '%s' not found", name))
		return
	}

	writeJSON(w, http.StatusOK, svc)
}

func (s *Server) handleScaffold(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req model.ScaffoldRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON request body: "+err.Error())
		return
	}

	resp, err := s.scaffolder.Scaffold(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Scaffolding failed: "+err.Error())
		return
	}

	// Auto commit to Git if requested
	if req.AutoCommit {
		commitMsg := fmt.Sprintf("feat(idp): scaffold service '%s' and gitops manifests", req.ServiceName)
		sha, commitErr := s.gitops.CommitChanges(commitMsg)
		if commitErr == nil {
			resp.CommitSHA = sha
		}
	}

	// Refresh catalog
	_ = s.catalog.RefreshFromGitOps()

	writeJSON(w, http.StatusCreated, resp)
}

func (s *Server) handleGitOpsStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	status := s.gitops.GetStatus()
	writeJSON(w, http.StatusOK, status)
}

func (s *Server) handleGitOpsSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var payload struct {
		ServiceName string `json:"service_name"`
		Environment string `json:"environment"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if payload.ServiceName == "" || payload.Environment == "" {
		writeError(w, http.StatusBadRequest, "service_name and environment are required")
		return
	}

	if err := s.gitops.TriggerSync(payload.ServiceName, payload.Environment); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Triggered GitOps synchronization for %s [%s]", payload.ServiceName, payload.Environment),
	})
}

func (s *Server) handleTemplates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	templates := []model.TemplateInfo{
		{
			ID:          "go-service",
			Name:        "Go Microservice API",
			Language:    "Go",
			Description: "High-performance, lightweight Cloud-Native HTTP microservice with health probes and structured logs.",
			DefaultPort: "8080",
		},
		{
			ID:          "node-service",
			Name:        "Node.js Express Service",
			Language:    "JavaScript / Node.js",
			Description: "Node.js REST service built on Express with JSON middleware and health probes.",
			DefaultPort: "3000",
		},
		{
			ID:          "python-service",
			Name:        "Python FastAPI Service",
			Language:    "Python",
			Description: "Modern, asynchronous FastAPI service with automatic OpenAPI documentation and Pydantic validation.",
			DefaultPort: "8000",
		},
	}

	writeJSON(w, http.StatusOK, templates)
}
