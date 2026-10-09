package catalog

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/opeteer/gitops-idp/idp-api/internal/model"
)

type Catalog struct {
	mu            sync.RWMutex
	services      map[string]model.Service
	gitopsAppsDir string
}

func NewCatalog(gitopsAppsDir string) *Catalog {
	c := &Catalog{
		services:      make(map[string]model.Service),
		gitopsAppsDir: gitopsAppsDir,
	}
	c.RefreshFromGitOps()
	return c
}

// RefreshFromGitOps scans the gitops/apps directory to discover all registered services.
func (c *Catalog) RefreshFromGitOps() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, err := os.Stat(c.gitopsAppsDir); os.IsNotExist(err) {
		return nil
	}

	entries, err := os.ReadDir(c.gitopsAppsDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		serviceName := entry.Name()
		serviceDir := filepath.Join(c.gitopsAppsDir, serviceName)

		// Check if base directory exists
		baseDir := filepath.Join(serviceDir, "base")
		if _, err := os.Stat(baseDir); os.IsNotExist(err) {
			continue
		}

		// Detect environments
		envs := make(map[string]model.EnvDeployment)
		for _, envName := range []string{"dev", "staging", "prod"} {
			overlayDir := filepath.Join(serviceDir, "overlays", envName)
			if _, err := os.Stat(overlayDir); err == nil {
				replicas := 1
				imageTag := "latest"
				if envName == "staging" {
					replicas = 2
					imageTag = "rc-latest"
				} else if envName == "prod" {
					replicas = 3
					imageTag = "v1.0.0"
				}

				envs[envName] = model.EnvDeployment{
					Environment:  envName,
					Replicas:     replicas,
					ImageTag:     imageTag,
					SyncStatus:   "Synced",
					HealthStatus: "Healthy",
					LastSyncedAt: time.Now().Add(-5 * time.Minute),
					EndpointURL:  "http://" + serviceName + "." + envName + ".cluster.local",
				}
			}
		}

		existing, exists := c.services[serviceName]
		owner := "Platform Team"
		language := "Go"
		description := "Declarative microservice managed via GitOpsIdp"
		createdAt := time.Now()

		if exists {
			owner = existing.Owner
			language = existing.Language
			description = existing.Description
			createdAt = existing.CreatedAt
		} else {
			// Infer language or defaults
			if strings.Contains(serviceName, "node") {
				language = "Node.js"
			} else if strings.Contains(serviceName, "python") || strings.Contains(serviceName, "fastapi") {
				language = "Python"
			}
		}

		c.services[serviceName] = model.Service{
			Name:         serviceName,
			Owner:        owner,
			Language:     language,
			Description:  description,
			RepoPath:     filepath.Join("gitops/apps", serviceName),
			CreatedAt:    createdAt,
			Environments: envs,
		}
	}

	return nil
}

func (c *Catalog) ListServices() []model.Service {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]model.Service, 0, len(c.services))
	for _, s := range c.services {
		result = append(result, s)
	}
	return result
}

func (c *Catalog) GetService(name string) (model.Service, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	svc, exists := c.services[name]
	return svc, exists
}

func (c *Catalog) RegisterService(svc model.Service) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.services[svc.Name] = svc
}

func (c *Catalog) UpdateEnvStatus(serviceName, envName, syncStatus, healthStatus string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if svc, exists := c.services[serviceName]; exists {
		if envDeploy, ok := svc.Environments[envName]; ok {
			envDeploy.SyncStatus = syncStatus
			envDeploy.HealthStatus = healthStatus
			envDeploy.LastSyncedAt = time.Now()
			svc.Environments[envName] = envDeploy
			c.services[serviceName] = svc
		}
	}
}
