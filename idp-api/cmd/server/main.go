package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/opeteer/gitops-idp/idp-api/internal/api"
	"github.com/opeteer/gitops-idp/idp-api/internal/catalog"
	"github.com/opeteer/gitops-idp/idp-api/internal/gitops"
	"github.com/opeteer/gitops-idp/idp-api/internal/scaffolder"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot == "" {
		// Default to directory above if running from idp-api, or current dir
		wd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get working directory: %v", err)
		}
		if filepath.Base(wd) == "idp-api" || filepath.Base(wd) == "server" {
			projectRoot = filepath.Clean(filepath.Join(wd, ".."))
			if filepath.Base(wd) == "server" {
				projectRoot = filepath.Clean(filepath.Join(wd, "../.."))
			}
		} else {
			projectRoot = wd
		}
	}

	log.Printf("==================================================")
	log.Printf("   🚀 GitOpsIdp Control Plane API Server")
	log.Printf("==================================================")
	log.Printf("Project Root: %s", projectRoot)
	log.Printf("Listening on: http://0.0.0.0:%s", port)

	gitopsAppsDir := filepath.Join(projectRoot, "gitops", "apps")
	cat := catalog.NewCatalog(gitopsAppsDir)
	scaf := scaffolder.NewScaffolder(projectRoot)
	gitEngine := gitops.NewEngine(projectRoot, cat)

	portalDir := filepath.Join(projectRoot, "idp-portal")
	if _, err := os.Stat(portalDir); os.IsNotExist(err) {
		portalDir = ""
	}

	srv := api.NewServer(cat, scaf, gitEngine, portalDir)
	routes := srv.Routes()

	log.Printf("Endpoints initialized:")
	log.Printf(" - GET  /healthz")
	log.Printf(" - GET  /api/v1/services")
	log.Printf(" - GET  /api/v1/services/{name}")
	log.Printf(" - POST /api/v1/scaffold")
	log.Printf(" - GET  /api/v1/gitops/status")
	log.Printf(" - POST /api/v1/gitops/sync")
	log.Printf(" - GET  /api/v1/templates")

	if err := http.ListenAndServe(":"+port, routes); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
