package gitops

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/opeteer/gitops-idp/idp-api/internal/catalog"
	"github.com/opeteer/gitops-idp/idp-api/internal/model"
)

type Engine struct {
	mu          sync.RWMutex
	projectRoot string
	catalog     *catalog.Catalog
}

func NewEngine(projectRoot string, cat *catalog.Catalog) *Engine {
	return &Engine{
		projectRoot: projectRoot,
		catalog:     cat,
	}
}

func (e *Engine) GetCurrentGitCommit() (string, string) {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	cmd.Dir = e.projectRoot
	out, err := cmd.Output()
	head := "unknown"
	if err == nil {
		head = strings.TrimSpace(string(out))
	}

	cmdBranch := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmdBranch.Dir = e.projectRoot
	outBranch, err := cmdBranch.Output()
	branch := "main"
	if err == nil {
		branch = strings.TrimSpace(string(outBranch))
	}

	return head, branch
}

func (e *Engine) CommitChanges(message string) (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// git add
	addCmd := exec.Command("git", "add", "gitops", "services")
	addCmd.Dir = e.projectRoot
	if out, err := addCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git add failed: %s: %w", string(out), err)
	}

	// git commit
	commitCmd := exec.Command("git", "commit", "-m", message)
	commitCmd.Dir = e.projectRoot
	var stderr bytes.Buffer
	commitCmd.Stderr = &stderr
	if err := commitCmd.Run(); err != nil {
		// If nothing to commit, return current HEAD
		if strings.Contains(stderr.String(), "nothing to commit") {
			head, _ := e.GetCurrentGitCommit()
			return head, nil
		}
		return "", fmt.Errorf("git commit failed: %s: %w", stderr.String(), err)
	}

	head, _ := e.GetCurrentGitCommit()
	return head, nil
}

func (e *Engine) GetStatus() model.GitOpsStatus {
	e.mu.RLock()
	defer e.mu.RUnlock()

	headCommit, branch := e.GetCurrentGitCommit()
	services := e.catalog.ListServices()

	appStatuses := make([]model.AppSyncStatus, 0)
	syncedCount := 0
	outOfSyncCount := 0

	for _, svc := range services {
		for envName, envDeploy := range svc.Environments {
			syncStatus := envDeploy.SyncStatus
			if syncStatus == "" {
				syncStatus = "Synced"
			}
			healthStatus := envDeploy.HealthStatus
			if healthStatus == "" {
				healthStatus = "Healthy"
			}

			if syncStatus == "Synced" {
				syncedCount++
			} else {
				outOfSyncCount++
			}

			appStatuses = append(appStatuses, model.AppSyncStatus{
				Name:            svc.Name,
				Environment:     envName,
				SyncStatus:      syncStatus,
				HealthStatus:    healthStatus,
				TargetRevision:  headCommit,
				CurrentRevision: headCommit,
				DriftDetected:   syncStatus != "Synced",
				LastSyncedAt:    envDeploy.LastSyncedAt,
			})
		}
	}

	return model.GitOpsStatus{
		RepoURL:       "https://github.com/opeteer/GitOpsIdp_Project.git",
		Branch:        branch,
		HeadCommit:    headCommit,
		TotalApps:     len(appStatuses),
		SyncedApps:    syncedCount,
		OutOfSyncApps: outOfSyncCount,
		LastCheckedAt: time.Now(),
		Applications:  appStatuses,
	}
}

func (e *Engine) TriggerSync(appName, envName string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Validate service existence
	_, exists := e.catalog.GetService(appName)
	if !exists {
		return fmt.Errorf("service %s not found in catalog", appName)
	}

	// Update status to Synced & Healthy
	e.catalog.UpdateEnvStatus(appName, envName, "Synced", "Healthy")
	return nil
}
