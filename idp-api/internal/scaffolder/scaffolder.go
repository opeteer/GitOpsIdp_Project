package scaffolder

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/opeteer/gitops-idp/idp-api/internal/model"
)

type Scaffolder struct {
	projectRoot string
}

func NewScaffolder(projectRoot string) *Scaffolder {
	return &Scaffolder{
		projectRoot: projectRoot,
	}
}

type TemplateData struct {
	ServiceName   string
	Owner         string
	Port          string
	CpuRequest    string
	CpuLimit      string
	MemoryRequest string
	MemoryLimit   string
	Image         string
	Env           string
	LogLevel      string
	Replicas      int
}

func (s *Scaffolder) Scaffold(req model.ScaffoldRequest) (*model.ScaffoldResponse, error) {
	// Defaults
	if req.ServiceName == "" {
		return nil, fmt.Errorf("service_name is required")
	}
	req.ServiceName = strings.ToLower(strings.TrimSpace(req.ServiceName))

	if req.Port == "" {
		req.Port = "8080"
	}
	if req.Owner == "" {
		req.Owner = "Platform Engineering"
	}
	if req.CpuRequest == "" {
		req.CpuRequest = "50m"
	}
	if req.CpuLimit == "" {
		req.CpuLimit = "200m"
	}
	if req.MemoryRequest == "" {
		req.MemoryRequest = "64Mi"
	}
	if req.MemoryLimit == "" {
		req.MemoryLimit = "256Mi"
	}

	langDir := "go-service"
	switch strings.ToLower(req.Language) {
	case "node", "nodejs", "javascript", "typescript":
		langDir = "node-service"
		if req.Port == "8080" {
			req.Port = "3000"
		}
	case "python", "fastapi":
		langDir = "python-service"
	default:
		langDir = "go-service"
	}

	generatedFiles := make([]string, 0)

	// 1. Generate Application Source Code in services/<service-name>
	appTargetDir := filepath.Join(s.projectRoot, "services", req.ServiceName)
	if err := os.MkdirAll(appTargetDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create service directory: %w", err)
	}

	tmplDir := filepath.Join(s.projectRoot, "templates", langDir)
	entries, err := os.ReadDir(tmplDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read template directory %s: %w", tmplDir, err)
	}

	data := TemplateData{
		ServiceName:   req.ServiceName,
		Owner:         req.Owner,
		Port:          req.Port,
		CpuRequest:    req.CpuRequest,
		CpuLimit:      req.CpuLimit,
		MemoryRequest: req.MemoryRequest,
		MemoryLimit:   req.MemoryLimit,
		Image:         fmt.Sprintf("ghcr.io/opeteer/%s:v1.0.0", req.ServiceName),
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tmpl") {
			continue
		}
		targetFileName := strings.TrimSuffix(entry.Name(), ".tmpl")
		srcPath := filepath.Join(tmplDir, entry.Name())
		destPath := filepath.Join(appTargetDir, targetFileName)

		if err := renderTemplateFile(srcPath, destPath, data); err != nil {
			return nil, fmt.Errorf("failed rendering %s: %w", entry.Name(), err)
		}
		rel, _ := filepath.Rel(s.projectRoot, destPath)
		generatedFiles = append(generatedFiles, rel)
	}

	// 2. Generate GitOps manifests in gitops/apps/<service-name>
	gitopsAppDir := filepath.Join(s.projectRoot, "gitops", "apps", req.ServiceName)
	baseDir := filepath.Join(gitopsAppDir, "base")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create gitops base dir: %w", err)
	}

	gitopsTmplDir := filepath.Join(s.projectRoot, "templates", "gitops-manifests")

	// Base deployment, service, configmap, kustomization
	baseData := data
	baseData.Env = "base"
	baseData.LogLevel = "info"
	baseData.Replicas = 1

	baseFiles := map[string]string{
		"deployment.yaml.tmpl":         "deployment.yaml",
		"service.yaml.tmpl":            "service.yaml",
		"configmap.yaml.tmpl":          "configmap.yaml",
		"base-kustomization.yaml.tmpl": "kustomization.yaml",
	}

	for tmplName, outName := range baseFiles {
		srcPath := filepath.Join(gitopsTmplDir, tmplName)
		destPath := filepath.Join(baseDir, outName)
		if err := renderTemplateFile(srcPath, destPath, baseData); err != nil {
			return nil, fmt.Errorf("failed rendering gitops base %s: %w", tmplName, err)
		}
		rel, _ := filepath.Rel(s.projectRoot, destPath)
		generatedFiles = append(generatedFiles, rel)
	}

	// Overlays for dev, staging, prod
	envConfigs := []struct {
		env      string
		replicas int
		imageTag string
		logLevel string
	}{
		{env: "dev", replicas: 1, imageTag: "dev-latest", logLevel: "debug"},
		{env: "staging", replicas: 2, imageTag: "rc-latest", logLevel: "info"},
		{env: "prod", replicas: 3, imageTag: "v1.0.0", logLevel: "warn"},
	}

	for _, cfg := range envConfigs {
		overlayDir := filepath.Join(gitopsAppDir, "overlays", cfg.env)
		if err := os.MkdirAll(overlayDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create overlay dir %s: %w", cfg.env, err)
		}

		overlayData := data
		overlayData.Env = cfg.env
		overlayData.Replicas = cfg.replicas
		overlayData.Image = fmt.Sprintf("ghcr.io/opeteer/%s:%s", req.ServiceName, cfg.imageTag)
		overlayData.LogLevel = cfg.logLevel

		srcPath := filepath.Join(gitopsTmplDir, "overlay-kustomization.yaml.tmpl")
		destPath := filepath.Join(overlayDir, "kustomization.yaml")
		if err := renderTemplateFile(srcPath, destPath, overlayData); err != nil {
			return nil, fmt.Errorf("failed rendering overlay %s: %w", cfg.env, err)
		}
		rel, _ := filepath.Rel(s.projectRoot, destPath)
		generatedFiles = append(generatedFiles, rel)

		// Also register the overlay in gitops/environments/<env>/kustomization.yaml
		envKustPath := filepath.Join(s.projectRoot, "gitops", "environments", cfg.env, "kustomization.yaml")
		s.appendResourceToKustomization(envKustPath, fmt.Sprintf("../../apps/%s/overlays/%s", req.ServiceName, cfg.env))
	}

	return &model.ScaffoldResponse{
		Success:        true,
		Message:        fmt.Sprintf("Service '%s' successfully scaffolded and integrated with GitOps manifests!", req.ServiceName),
		ServiceName:    req.ServiceName,
		GeneratedFiles: generatedFiles,
	}, nil
}

func renderTemplateFile(srcPath, destPath string, data interface{}) error {
	content, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}

	t, err := template.New(filepath.Base(srcPath)).Parse(string(content))
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return err
	}

	return os.WriteFile(destPath, buf.Bytes(), 0644)
}

func (s *Scaffolder) appendResourceToKustomization(kustPath, resourceEntry string) error {
	content, err := os.ReadFile(kustPath)
	if err != nil {
		return err
	}

	lines := string(content)
	if strings.Contains(lines, resourceEntry) {
		return nil
	}

	// Append resource entry under resources
	trimmed := strings.TrimRight(lines, "\n")
	newContent := fmt.Sprintf("%s\n  - %s\n", trimmed, resourceEntry)
	return os.WriteFile(kustPath, []byte(newContent), 0644)
}
