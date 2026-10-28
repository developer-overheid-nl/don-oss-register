package repositories

import (
	"context"
	"errors"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"gorm.io/gorm"
)

type ScansRepository interface {
	SaveScan(ctx context.Context, scan *models.RepositoryScan) error
	FindScanByIdempotencyKey(ctx context.Context, key string) (*models.RepositoryScan, error)
	// LatestScans returns the most recent scan per repository, keyed by repository ID.
	LatestScans(ctx context.Context, repositoryIDs []string) (map[string]models.RepositoryScan, error)
}

type scansRepository struct {
	db *gorm.DB
}

func NewScansRepository(db *gorm.DB) ScansRepository {
	return &scansRepository{db: db}
}

func (r *scansRepository) SaveScan(ctx context.Context, scan *models.RepositoryScan) error {
	return r.db.WithContext(ctx).Create(scan).Error
}

func (r *scansRepository) FindScanByIdempotencyKey(ctx context.Context, key string) (*models.RepositoryScan, error) {
	var scan models.RepositoryScan
	if err := r.db.WithContext(ctx).Where("idempotency_key = ?", key).First(&scan).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &scan, nil
}

func (r *scansRepository) LatestScans(ctx context.Context, repositoryIDs []string) (map[string]models.RepositoryScan, error) {
	result := make(map[string]models.RepositoryScan, len(repositoryIDs))
	if len(repositoryIDs) == 0 {
		return result, nil
	}
	var scans []models.RepositoryScan
	// The report is not needed to show a scan; the derived summary is enough.
	err := r.db.WithContext(ctx).
		Select("id", "repository_id", "status", "scanned_at", "score", "summary", "created_at").
		Where("repository_id IN ?", repositoryIDs).
		Where("scanned_at = (?)", r.db.Table("repository_scans AS latest").
			Select("MAX(latest.scanned_at)").
			Where("latest.repository_id = repository_scans.repository_id")).
		Order("created_at DESC").
		Find(&scans).Error
	if err != nil {
		return nil, err
	}
	for _, scan := range scans {
		if _, seen := result[scan.RepositoryID]; !seen {
			result[scan.RepositoryID] = scan
		}
	}
	return result, nil
}
