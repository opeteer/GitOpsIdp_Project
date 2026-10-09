package scaffolder_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/opeteer/gitops-idp/idp-api/internal/model"
	"github.com/opeteer/gitops-idp/idp-api/internal/scaffolder"
)

func TestScaffolder(t *testing.T) {
	// Root of the GitOpsIdp_Project
	projectRoot, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatalf("failed to resolve project root: %v", err)
	}

	scaf := scaffolder.NewScaffolder(projectRoot)

	testServiceName := "test-order-service"
	// Clean up after test
	defer func() {
		os.RemoveAll(filepath.Join(projectRoot, "services", testServiceName))
		os.RemoveAll(filepath.Join(projectRoot, "gitops", "apps", testServiceName))
	}()

	req := model.ScaffoldRequest{
		ServiceName: testServiceName,
		Owner:       "E-Commerce Team",
		Language:    "go",
		Port:        "9090",
		AutoCommit:  false,
	}

	resp, err := scaf.Scaffold(req)
	if err != nil {
		t.Fatalf("scaffolder failed: %v", err)
	}

	if !resp.Success {
		t.Fatalf("expected success, got false")
	}

	// Verify generated service files
	mainGoPath := filepath.Join(projectRoot, "services", testServiceName, "main.go")
	if _, err := os.Stat(mainGoPath); os.IsNotExist(err) {
		t.Errorf("expected %s to exist", mainGoPath)
	}

	// Verify generated gitops manifests
	kustPath := filepath.Join(projectRoot, "gitops", "apps", testServiceName, "base", "kustomization.yaml")
	if _, err := os.Stat(kustPath); os.IsNotExist(err) {
		t.Errorf("expected %s to exist", kustPath)
	}

	devOverlayPath := filepath.Join(projectRoot, "gitops", "apps", testServiceName, "overlays", "dev", "kustomization.yaml")
	if _, err := os.Stat(devOverlayPath); os.IsNotExist(err) {
		t.Errorf("expected %s to exist", devOverlayPath)
	}
}
