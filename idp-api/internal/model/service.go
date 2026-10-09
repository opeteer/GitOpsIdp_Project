package model

import "time"

type EnvDeployment struct {
	Environment  string    `json:"environment"` // dev, staging, prod
	Replicas     int       `json:"replicas"`
	ImageTag     string    `json:"image_tag"`
	SyncStatus   string    `json:"sync_status"`   // Synced, OutOfSync, Progressing
	HealthStatus string    `json:"health_status"` // Healthy, Degraded, Unknown
	LastSyncedAt time.Time `json:"last_synced_at"`
	EndpointURL  string    `json:"endpoint_url"`
}

type Service struct {
	Name         string                   `json:"name"`
	Owner        string                   `json:"owner"`
	Language     string                   `json:"language"`
	Description  string                   `json:"description"`
	RepoPath     string                   `json:"repo_path"`
	CreatedAt    time.Time                `json:"created_at"`
	Environments map[string]EnvDeployment `json:"environments"`
}

type ScaffoldRequest struct {
	ServiceName   string `json:"service_name"`
	Owner         string `json:"owner"`
	Language      string `json:"language"` // go, node, python
	Description   string `json:"description"`
	Port          string `json:"port"`
	CpuRequest    string `json:"cpu_request"`
	CpuLimit      string `json:"cpu_limit"`
	MemoryRequest string `json:"memory_request"`
	MemoryLimit   string `json:"memory_limit"`
	AutoCommit    bool   `json:"auto_commit"`
}

type ScaffoldResponse struct {
	Success        bool     `json:"success"`
	Message        string   `json:"message"`
	ServiceName    string   `json:"service_name"`
	CommitSHA      string   `json:"commit_sha,omitempty"`
	GeneratedFiles []string `json:"generated_files"`
}

type AppSyncStatus struct {
	Name            string    `json:"name"`
	Environment     string    `json:"environment"`
	SyncStatus      string    `json:"sync_status"` // Synced, OutOfSync, Progressing
	HealthStatus    string    `json:"health_status"` // Healthy, Degraded
	TargetRevision  string    `json:"target_revision"`
	CurrentRevision string    `json:"current_revision"`
	DriftDetected   bool      `json:"drift_detected"`
	LastSyncedAt    time.Time `json:"last_synced_at"`
}

type GitOpsStatus struct {
	RepoURL       string          `json:"repo_url"`
	Branch        string          `json:"branch"`
	HeadCommit    string          `json:"head_commit"`
	TotalApps     int             `json:"total_apps"`
	SyncedApps    int             `json:"synced_apps"`
	OutOfSyncApps int             `json:"out_of_sync_apps"`
	LastCheckedAt time.Time       `json:"last_checked_at"`
	Applications  []AppSyncStatus `json:"applications"`
}

type TemplateInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Language    string `json:"language"`
	Description string `json:"description"`
	DefaultPort string `json:"default_port"`
}
