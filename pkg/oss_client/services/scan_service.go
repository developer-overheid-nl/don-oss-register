package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"slices"
	"strings"
	"time"

	problem "github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/helpers/problem"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/repositories"
	"github.com/google/uuid"
)

const scanSchemaVersion = 1

// vulnerabilityRule is reported per dependency. The register does not publish
// vulnerabilities, so this rule is left out of findings and the score.
const vulnerabilityRule = "VULNERABILITY_IN_DEPENDENCY"

var howToFixURLs = map[string]string{
	"MISSING_README_FILE":          "https://developer.overheid.nl/kennisbank/open-source/standaarden/readme-md",
	"MISSING_LICENSE_FILE":         "https://developer.overheid.nl/kennisbank/open-source/tutorials/open-source-software-licenties",
	"MISSING_PUBLICCODE_FILE":      "https://developer.overheid.nl/kennisbank/open-source/tutorials/publiccode-yml-toevoegen",
	"MISSING_CONTRIBUTING_FILE":    "https://developer.overheid.nl/kennisbank/open-source/standaarden/contributing-md",
	"MISSING_SECURITY_FILE":        "https://developer.overheid.nl/kennisbank/open-source/standaarden/security",
	"MISSING_CODE_OF_CONDUCT_FILE": "https://developer.overheid.nl/kennisbank/open-source/standaarden/code-of-conduct-md",
	"MISSING_CHANGELOG_FILE":       "https://developer.overheid.nl/kennisbank/open-source/tutorials/changelog-bijhouden",
}

// WithScans enables storing ORT scans and showing the latest scan per repository.
func (s *RepositoryService) WithScans(scans repositories.ScansRepository) *RepositoryService {
	s.scans = scans
	return s
}

// SubmitScan stores a scan delivered by the ort-runner. A delivery that is
// retried with the same idempotency key returns the scan stored the first time.
func (s *RepositoryService) SubmitScan(ctx context.Context, submission models.ScanSubmission) (*models.RepositoryScanDetail, error) {
	if s.scans == nil {
		return nil, problem.NewInternalServerError("Scans are not enabled")
	}
	if submission.SchemaVersion != scanSchemaVersion {
		return nil, problem.NewBadRequest("Invalid input",
			bodyError("schemaVersion", "enum", fmt.Sprintf("schemaVersion must be %d", scanSchemaVersion)),
		)
	}
	key := strings.TrimSpace(submission.IdempotencyKey)
	if key != "" {
		existing, err := s.scans.FindScanByIdempotencyKey(ctx, key)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return decodeScanDetail(*existing)
		}
	}

	var report models.RunReport
	if err := json.Unmarshal(submission.Scan, &report); err != nil {
		return nil, problem.NewBadRequest("Invalid input", bodyError("scan", "invalid", "scan must be an ort-runner run.json object"))
	}
	if !slices.Contains([]string{"completed", "incomplete", "failed"}, report.Status) {
		return nil, problem.NewBadRequest("Invalid input",
			bodyError("scan.status", "enum", "scan.status must be completed, incomplete or failed"),
		)
	}
	repositoryID := strings.TrimSpace(submission.RepositoryID)
	if err := validateRepositoryID(repositoryID); err != nil {
		return nil, err
	}
	repository, err := s.repo.GetRepositoryByID(ctx, repositoryID)
	if err != nil {
		return nil, err
	}
	if repository == nil {
		return nil, problem.NewNotFound("Resource does not exist")
	}

	detail := BuildScanDetail(report, time.Now().UTC())
	summary, err := json.Marshal(detail)
	if err != nil {
		return nil, err
	}
	scan := &models.RepositoryScan{
		ID:           uuid.NewString(),
		RepositoryID: repository.Id,
		Status:       detail.Status,
		Revision:     detail.Revision,
		ScannedAt:    detail.ScannedAt,
		Score:        detail.RepoScorePercentage,
		Report:       string(submission.Scan),
		Summary:      string(summary),
		CreatedAt:    time.Now().UTC(),
	}
	if key != "" {
		scan.IdempotencyKey = &key
	}
	if err := s.scans.SaveScan(ctx, scan); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "ORT scan stored",
		"component", "scan",
		"operation", "store",
		"repository_id", repository.Id,
		"scan_id", scan.ID,
		"status", scan.Status,
	)
	return &detail, nil
}

// BuildScanDetail derives what the register shows from an ort-runner report.
// receivedAt is used as scan time when the report has no finish time.
func BuildScanDetail(report models.RunReport, receivedAt time.Time) models.RepositoryScanDetail {
	detail := models.RepositoryScanDetail{
		RepositoryScanSummary: models.RepositoryScanSummary{
			Status:       report.Status,
			ScannedAt:    receivedAt,
			PackageCount: report.PackageCount,
		},
		StartedAt: report.StartedAt,
		Revision:  report.Revision,
		Error:     report.Error,
		Findings:  []models.ScanFinding{},
		Tooling: models.ScanTooling{
			ORTImage:      report.ORTImage,
			RunnerVersion: report.RunnerVersion,
			ConfigDigest:  report.ConfigDigest,
		},
	}
	if report.FinishedAt != nil {
		detail.ScannedAt = report.FinishedAt.UTC()
	}
	for _, stage := range report.Stages {
		if stage.Status == "failed" {
			detail.FailedStage = stage.Name
			break
		}
	}

	if dependencies := report.Dependencies; dependencies != nil {
		detail.DirectDependencyCount = &dependencies.DirectCount
		detail.TransitiveDependencyCount = &dependencies.TransitiveCount
		for _, project := range dependencies.Projects {
			detail.Projects = append(detail.Projects, models.ScanProject{
				PackageManager:            project.PackageManager,
				DefinitionFile:            project.DefinitionFile,
				DirectDependencyCount:     project.DirectCount,
				TransitiveDependencyCount: project.TransitiveCount,
			})
			if !slices.Contains(detail.PackageManagers, project.PackageManager) {
				detail.PackageManagers = append(detail.PackageManagers, project.PackageManager)
			}
		}
	}

	messages := make(map[string]models.RunFinding, len(report.Findings))
	for _, finding := range report.Findings {
		if _, seen := messages[finding.Rule]; !seen {
			messages[finding.Rule] = finding
		}
	}
	listed := make(map[string]bool, len(report.Rules))
	passed := 0
	for _, rule := range report.Rules {
		listed[rule.Rule] = true
		if rule.Scope != "repository" {
			continue
		}
		finding := models.ScanFinding{Rule: rule.Rule, Severity: rule.Severity, Passed: rule.Passed}
		if !rule.Passed {
			finding.Message = messages[rule.Rule].Message
		} else {
			passed++
		}
		detail.Findings = append(detail.Findings, finding)
	}
	if len(detail.Findings) > 0 {
		score := int(math.Round(float64(passed) * 100 / float64(len(detail.Findings))))
		detail.RepoScorePercentage = &score
	}
	// Without a rules manifest, or for rules missing from it, only violations are known.
	for _, finding := range report.Findings {
		if listed[finding.Rule] || finding.Rule == vulnerabilityRule {
			continue
		}
		listed[finding.Rule] = true
		detail.Findings = append(detail.Findings, models.ScanFinding{
			Rule: finding.Rule, Severity: finding.Severity, Message: finding.Message,
		})
	}

	for i := range detail.Findings {
		finding := &detail.Findings[i]
		finding.HowToFixUrl = howToFixURLs[finding.Rule]
		if finding.Passed {
			continue
		}
		switch strings.ToUpper(finding.Severity) {
		case "ERROR":
			detail.FindingCounts.Error++
		case "WARNING":
			detail.FindingCounts.Warning++
		case "HINT":
			detail.FindingCounts.Hint++
		}
	}
	return detail
}

// attachScanSummaries adds the latest scan to each repository in the list.
func (s *RepositoryService) attachScanSummaries(ctx context.Context, summaries []models.RepositorySummary) error {
	if s.scans == nil || len(summaries) == 0 {
		return nil
	}
	ids := make([]string, len(summaries))
	for i, summary := range summaries {
		ids[i] = summary.Id
	}
	latest, err := s.scans.LatestScans(ctx, ids)
	if err != nil {
		return err
	}
	for i := range summaries {
		scan, ok := latest[summaries[i].Id]
		if !ok {
			continue
		}
		detail, err := decodeScanDetail(scan)
		if err != nil {
			// One unreadable scan must not break the repository list.
			slog.WarnContext(ctx, "stored scan summary is invalid",
				"component", "scan",
				"operation", "decode_summary",
				"repository_id", scan.RepositoryID,
				"scan_id", scan.ID,
			)
			continue
		}
		summaries[i].ScanSummary = &detail.RepositoryScanSummary
	}
	return nil
}

// attachScanDetail adds the latest scan to a repository detail.
func (s *RepositoryService) attachScanDetail(ctx context.Context, detail *models.RepositoryDetail) error {
	if s.scans == nil || detail == nil {
		return nil
	}
	latest, err := s.scans.LatestScans(ctx, []string{detail.Id})
	if err != nil {
		return err
	}
	scan, ok := latest[detail.Id]
	if !ok {
		return nil
	}
	detail.Scan, err = decodeScanDetail(scan)
	return err
}

func decodeScanDetail(scan models.RepositoryScan) (*models.RepositoryScanDetail, error) {
	var detail models.RepositoryScanDetail
	if err := json.Unmarshal([]byte(scan.Summary), &detail); err != nil {
		return nil, problem.New(http.StatusInternalServerError, "Stored scan summary is invalid")
	}
	return &detail, nil
}
