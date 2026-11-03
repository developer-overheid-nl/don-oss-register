package models

import (
	"encoding/json"
	"time"
)

// RepositoryScan stores every ORT scan the ort-runner delivers, so the history
// of a repository stays available. Report holds the complete run.json; Summary
// holds the RepositoryScanDetail derived from it when the scan was received.
type RepositoryScan struct {
	ID             string    `gorm:"column:id;primaryKey"`
	RepositoryID   string    `gorm:"column:repository_id;index;not null"`
	IdempotencyKey *string   `gorm:"column:idempotency_key;uniqueIndex"`
	Status         string    `gorm:"column:status;not null"`
	Revision       string    `gorm:"column:revision"`
	ScannedAt      time.Time `gorm:"column:scanned_at;index"`
	Score          *int      `gorm:"column:score"`
	Report         string    `gorm:"column:report;type:text"`
	Summary        string    `gorm:"column:summary;type:text"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

// ScanSubmission is the message the ort-runner posts for each scanned repository.
type ScanSubmission struct {
	IdempotencyKey string          `header:"Idempotency-Key"`
	SchemaVersion  int             `json:"schemaVersion" binding:"required"`
	RepositoryID   string          `json:"repositoryId" binding:"required"`
	Scan           json.RawMessage `json:"scan" binding:"required"`
}

// RunReport is the part of the ort-runner run.json the register reads.
type RunReport struct {
	Status        string           `json:"status"`
	Error         string           `json:"error"`
	Revision      string           `json:"revision"`
	ConfigDigest  string           `json:"config_digest"`
	ORTImage      string           `json:"ort_image"`
	RunnerVersion string           `json:"runner_version"`
	StartedAt     time.Time        `json:"started_at"`
	FinishedAt    *time.Time       `json:"finished_at"`
	PackageCount  *int             `json:"package_count"`
	Dependencies  *RunDependencies `json:"dependencies"`
	Stages        []RunStage       `json:"stages"`
	Findings      []RunFinding     `json:"findings"`
	Rules         []RunRule        `json:"rules"`
}

type RunDependencies struct {
	DirectCount     int          `json:"direct_count"`
	TransitiveCount int          `json:"transitive_count"`
	Projects        []RunProject `json:"projects"`
}

type RunProject struct {
	PackageManager  string `json:"package_manager"`
	DefinitionFile  string `json:"definition_file"`
	DirectCount     int    `json:"direct_count"`
	TransitiveCount int    `json:"transitive_count"`
}

type RunStage struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type RunFinding struct {
	Rule     string `json:"rule"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

type RunRule struct {
	Rule     string `json:"rule"`
	Severity string `json:"severity"`
	Scope    string `json:"scope"`
	Passed   bool   `json:"passed"`
}

// RepositoryScanSummary is the latest scan as shown in the repository list.
// Vulnerabilities are left out on purpose; they are only kept in the stored report.
type RepositoryScanSummary struct {
	Status                    string            `json:"status"`
	ScannedAt                 time.Time         `json:"scannedAt"`
	RepoScorePercentage       *int              `json:"repoScorePercentage,omitempty"`
	PackageCount              *int              `json:"packageCount,omitempty"`
	DirectDependencyCount     *int              `json:"directDependencyCount,omitempty"`
	TransitiveDependencyCount *int              `json:"transitiveDependencyCount,omitempty"`
	PackageManagers           []string          `json:"packageManagers,omitempty"`
	FindingCounts             ScanFindingCounts `json:"findingCounts"`
}

type ScanFindingCounts struct {
	Error   int `json:"error"`
	Warning int `json:"warning"`
	Hint    int `json:"hint"`
}

// RepositoryScanDetail is the latest scan as shown on the repository detail page.
type RepositoryScanDetail struct {
	RepositoryScanSummary
	StartedAt   time.Time     `json:"startedAt"`
	Revision    string        `json:"revision,omitempty"`
	Error       string        `json:"error,omitempty"`
	FailedStage string        `json:"failedStage,omitempty"`
	Projects    []ScanProject `json:"projects,omitempty"`
	Findings    []ScanFinding `json:"findings"`
	Tooling     ScanTooling   `json:"tooling"`
}

type ScanProject struct {
	PackageManager            string `json:"packageManager"`
	DefinitionFile            string `json:"definitionFile"`
	DirectDependencyCount     int    `json:"directDependencyCount"`
	TransitiveDependencyCount int    `json:"transitiveDependencyCount"`
}

// ScanFinding is one ORT repository rule: passed, or failed with ORT's message.
type ScanFinding struct {
	Rule        string `json:"rule"`
	Severity    string `json:"severity"`
	Passed      bool   `json:"passed"`
	Message     string `json:"message,omitempty"`
	HowToFixUrl string `json:"howToFixUrl,omitempty"`
}

type ScanTooling struct {
	ORTImage      string `json:"ortImage,omitempty"`
	RunnerVersion string `json:"runnerVersion,omitempty"`
	ConfigDigest  string `json:"configDigest,omitempty"`
}
