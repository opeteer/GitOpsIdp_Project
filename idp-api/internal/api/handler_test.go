package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/opeteer/gitops-idp/idp-api/internal/api"
	"github.com/opeteer/gitops-idp/idp-api/internal/catalog"
	"github.com/opeteer/gitops-idp/idp-api/internal/gitops"
	"github.com/opeteer/gitops-idp/idp-api/internal/model"
	"github.com/opeteer/gitops-idp/idp-api/internal/scaffolder"
)

func TestAPIEndpoints(t *testing.T) {
	projectRoot, _ := filepath.Abs("../../..")
	portalDir := filepath.Join(projectRoot, "idp-portal")
	cat := catalog.NewCatalog(filepath.Join(projectRoot, "gitops", "apps"))
	scaf := scaffolder.NewScaffolder(projectRoot)
	git := gitops.NewEngine(projectRoot, cat)

	server := api.NewServer(cat, scaf, git, portalDir)
	handler := server.Routes()

	// 1. Test /healthz
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected /healthz to return 200, got %d", rr.Code)
	}

	// 2. Test /api/v1/templates
	reqTmpl := httptest.NewRequest(http.MethodGet, "/api/v1/templates", nil)
	rrTmpl := httptest.NewRecorder()
	handler.ServeHTTP(rrTmpl, reqTmpl)
	if rrTmpl.Code != http.StatusOK {
		t.Fatalf("expected /api/v1/templates to return 200, got %d", rrTmpl.Code)
	}

	// 3. Test /api/v1/services
	reqServices := httptest.NewRequest(http.MethodGet, "/api/v1/services", nil)
	rrServices := httptest.NewRecorder()
	handler.ServeHTTP(rrServices, reqServices)
	if rrServices.Code != http.StatusOK {
		t.Fatalf("expected /api/v1/services to return 200, got %d", rrServices.Code)
	}

	// 4. Test /api/v1/gitops/status
	reqStatus := httptest.NewRequest(http.MethodGet, "/api/v1/gitops/status", nil)
	rrStatus := httptest.NewRecorder()
	handler.ServeHTTP(rrStatus, reqStatus)
	if rrStatus.Code != http.StatusOK {
		t.Fatalf("expected /api/v1/gitops/status to return 200, got %d", rrStatus.Code)
	}

	// 5. Test Static Portal Serving on /
	reqPortal := httptest.NewRequest(http.MethodGet, "/", nil)
	rrPortal := httptest.NewRecorder()
	handler.ServeHTTP(rrPortal, reqPortal)
	if rrPortal.Code != http.StatusOK {
		t.Fatalf("expected / to return 200 (portal index.html), got %d", rrPortal.Code)
	}

	// 6. Test Scaffolding Endpoint
	testSvc := "api-test-service"
	defer func() {
		_ = os.RemoveAll(filepath.Join(projectRoot, "services", testSvc))
		_ = os.RemoveAll(filepath.Join(projectRoot, "gitops", "apps", testSvc))
	}()

	scaffoldPayload := model.ScaffoldRequest{
		ServiceName: testSvc,
		Language:    "node",
		Owner:       "Core API Team",
		Port:        "3000",
		AutoCommit:  false,
	}
	bodyBytes, _ := json.Marshal(scaffoldPayload)
	reqScaffold := httptest.NewRequest(http.MethodPost, "/api/v1/scaffold", bytes.NewReader(bodyBytes))
	rrScaffold := httptest.NewRecorder()
	handler.ServeHTTP(rrScaffold, reqScaffold)
	if rrScaffold.Code != http.StatusCreated {
		t.Fatalf("expected /api/v1/scaffold to return 201, got %d (body: %s)", rrScaffold.Code, rrScaffold.Body.String())
	}

	// 7. Test Fetch Specific Service
	reqSingle := httptest.NewRequest(http.MethodGet, "/api/v1/services/"+testSvc, nil)
	rrSingle := httptest.NewRecorder()
	handler.ServeHTTP(rrSingle, reqSingle)
	if rrSingle.Code != http.StatusOK {
		t.Fatalf("expected /api/v1/services/%s to return 200, got %d", testSvc, rrSingle.Code)
	}

	// 8. Test GitOps Manual Sync Endpoint
	syncPayload := map[string]string{
		"service_name": testSvc,
		"environment":  "dev",
	}
	syncBytes, _ := json.Marshal(syncPayload)
	reqSync := httptest.NewRequest(http.MethodPost, "/api/v1/gitops/sync", bytes.NewReader(syncBytes))
	rrSync := httptest.NewRecorder()
	handler.ServeHTTP(rrSync, reqSync)
	if rrSync.Code != http.StatusOK {
		t.Fatalf("expected /api/v1/gitops/sync to return 200, got %d", rrSync.Code)
	}
}
