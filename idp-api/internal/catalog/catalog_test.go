package catalog_test

import (
	"path/filepath"
	"testing"

	"github.com/opeteer/gitops-idp/idp-api/internal/catalog"
	"github.com/opeteer/gitops-idp/idp-api/internal/model"
)

func TestCatalogDiscovery(t *testing.T) {
	// gitops/apps relative to idp-api
	gitopsAppsDir := filepath.Join("..", "..", "..", "gitops", "apps")
	cat := catalog.NewCatalog(gitopsAppsDir)

	services := cat.ListServices()
	if len(services) == 0 {
		t.Logf("Notice: No services found yet under %s or relative path", gitopsAppsDir)
	}

	// Register a mock service
	mockSvc := model.Service{
		Name:        "payment-service",
		Owner:       "Fintech Team",
		Language:    "Go",
		Description: "Core payment gateway",
		Environments: map[string]model.EnvDeployment{
			"dev": {
				Environment:  "dev",
				Replicas:     1,
				ImageTag:     "dev-latest",
				SyncStatus:   "Synced",
				HealthStatus: "Healthy",
			},
		},
	}
	cat.RegisterService(mockSvc)

	got, found := cat.GetService("payment-service")
	if !found {
		t.Fatalf("expected payment-service to be found in catalog")
	}
	if got.Owner != "Fintech Team" {
		t.Fatalf("expected owner 'Fintech Team', got '%s'", got.Owner)
	}

	// Update sync status
	cat.UpdateEnvStatus("payment-service", "dev", "OutOfSync", "Degraded")
	updated, _ := cat.GetService("payment-service")
	if updated.Environments["dev"].SyncStatus != "OutOfSync" {
		t.Fatalf("expected status 'OutOfSync', got '%s'", updated.Environments["dev"].SyncStatus)
	}
}
