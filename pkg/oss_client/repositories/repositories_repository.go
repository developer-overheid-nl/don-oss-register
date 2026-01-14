package repositories

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"gorm.io/gorm"
)

type RepositoriesRepository interface {
	GetRepositorys(ctx context.Context, page, perPage int, organisation *string) ([]models.Repository, models.Pagination, error)
	GetRepositoryByID(ctx context.Context, oasUrl string) (*models.Repository, error)
	SaveRepository(ctx context.Context, repository *models.Repository) error
	SearchRepositorys(ctx context.Context, page, perPage int, organisation *string, query string) ([]models.Repository, models.Pagination, error)
	SaveOrganisatie(organisation *models.Organisation) error
	AllRepositorys(ctx context.Context) ([]models.Repository, error)
	GetOrganisations(ctx context.Context) ([]models.Organisation, error)
	FindOrganisationByURI(ctx context.Context, uri string) (*models.Organisation, error)
	GetGitOrganisations(ctx context.Context, page, perPage int, organisation *string) ([]models.GitOrganisatie, models.Pagination, error)
	FindGitOrganisationByURL(ctx context.Context, url string) (*models.GitOrganisatie, error)
	SaveGitOrganisatie(ctx context.Context, gitOrg *models.GitOrganisatie) error
}

type repositoriesRepository struct {
	db *gorm.DB
}

func NewRepositoriesRepository(db *gorm.DB) RepositoriesRepository {
	return &repositoriesRepository{db: db}
}

func (r *repositoriesRepository) SaveRepository(ctx context.Context, repository *models.Repository) error {
	trimmedRepoURL := strings.TrimSpace(repository.Url)
	repository.Url = trimmedRepoURL

	var existing models.Repository
	found := false
	if repository.Id != "" {
		if err := r.db.WithContext(ctx).Where("id = ?", repository.Id).First(&existing).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		} else {
			found = true
		}
	}

	if !found && trimmedRepoURL != "" {
		err := r.db.WithContext(ctx).Where("repository_url = ?", trimmedRepoURL).First(&existing).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err == nil {
			log.Printf("SaveRepository: found existing repository for url %q with id %s", trimmedRepoURL, existing.Id)
			found = true
		}
	}

	// if found && repository.Id != existing.Id {
	// 	return problem.NewBadRequest("Repository already exists; use PUT instead of POST")
	// }

	if found {
		repository.Id = existing.Id
		if repository.CreatedAt.IsZero() {
			repository.CreatedAt = existing.CreatedAt
		}
		if repository.OrganisationID == nil {
			repository.OrganisationID = existing.OrganisationID
		}

		return r.db.WithContext(ctx).Save(repository).Error
	}

	return r.db.WithContext(ctx).Create(repository).Error
}

func (r *repositoriesRepository) GetRepositorys(ctx context.Context, page, perPage int, organisation *string) ([]models.Repository, models.Pagination, error) {
	if page < 1 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	db := r.db.WithContext(ctx)
	db = db.Where("(active IS NULL OR active = ?)", true)

	if organisation != nil && strings.TrimSpace(*organisation) != "" {
		db = db.Where("organisation_id = ?", strings.TrimSpace(*organisation))
	}

	var totalRecords int64
	if err := db.Model(&models.Repository{}).Count(&totalRecords).Error; err != nil {
		return nil, models.Pagination{}, err
	}

	var repositories []models.Repository
	if err := applyRepositoryOrdering(db).Limit(perPage).Preload("Organisation").Offset(offset).Find(&repositories).Error; err != nil {
		return nil, models.Pagination{}, err
	}

	totalPages := int(math.Ceil(float64(totalRecords) / float64(perPage)))
	pagination := models.Pagination{
		CurrentPage:    page,
		RecordsPerPage: perPage,
		TotalPages:     totalPages,
		TotalRecords:   int(totalRecords),
	}

	if page < totalPages {
		next := page + 1
		pagination.Next = &next
	}
	if page > 1 {
		prev := page - 1
		pagination.Previous = &prev
	}

	return repositories, pagination, nil
}

func (r *repositoriesRepository) GetGitOrganisations(ctx context.Context, page, perPage int, organisation *string) ([]models.GitOrganisatie, models.Pagination, error) {
	if page < 1 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	db := r.db.WithContext(ctx)
	if organisation != nil && strings.TrimSpace(*organisation) != "" {
		db = db.Where("organisation_id = ?", strings.TrimSpace(*organisation))
	}

	var totalRecords int64
	if err := db.Model(&models.GitOrganisatie{}).Count(&totalRecords).Error; err != nil {
		return nil, models.Pagination{}, err
	}

	var gitOrganisations []models.GitOrganisatie
	if err := db.Limit(perPage).Preload("Organisation").Offset(offset).Find(&gitOrganisations).Error; err != nil {
		return nil, models.Pagination{}, err
	}

	totalPages := int(math.Ceil(float64(totalRecords) / float64(perPage)))
	pagination := models.Pagination{
		CurrentPage:    page,
		RecordsPerPage: perPage,
		TotalPages:     totalPages,
		TotalRecords:   int(totalRecords),
	}

	if page < totalPages {
		next := page + 1
		pagination.Next = &next
	}
	if page > 1 {
		prev := page - 1
		pagination.Previous = &prev
	}

	return gitOrganisations, pagination, nil
}

func (r *repositoriesRepository) GetRepositoryByID(ctx context.Context, id string) (*models.Repository, error) {
	var api models.Repository
	if err := r.db.WithContext(ctx).Preload("Organisation").First(&api, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &api, nil
}

func (r *repositoriesRepository) SearchRepositorys(ctx context.Context, page, perPage int, organisation *string, query string) ([]models.Repository, models.Pagination, error) {
	trimmed := strings.TrimSpace(query)
	if page < 1 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 20
	}
	if trimmed == "" {
		return []models.Repository{}, models.Pagination{
			CurrentPage:    page,
			RecordsPerPage: perPage,
		}, nil
	}

	base := r.db.WithContext(ctx)
	base = base.Where("(active IS NULL OR active = ?)", true)
	if organisation != nil && strings.TrimSpace(*organisation) != "" {
		base = base.Where("organisation_id = ?", strings.TrimSpace(*organisation))
	}
	var pattern string
	if trimmed != "" {
		pattern = fmt.Sprintf("%%%s%%", strings.ToLower(trimmed))
		base = base.Where("(LOWER(name) LIKE ? OR LOWER(short_description) LIKE ? OR LOWER(long_description) LIKE ?)", pattern, pattern, pattern)
	}

	var totalRecords int64
	if err := base.Model(&models.Repository{}).Count(&totalRecords).Error; err != nil {
		return nil, models.Pagination{}, err
	}

	queryDB := r.db.WithContext(ctx)
	queryDB = queryDB.Where("(active IS NULL OR active = ?)", true)
	if organisation != nil && strings.TrimSpace(*organisation) != "" {
		queryDB = queryDB.Where("organisation_id = ?", strings.TrimSpace(*organisation))
	}
	if pattern != "" {
		queryDB = queryDB.Where("(LOWER(name) LIKE ? OR LOWER(short_description) LIKE ? OR LOWER(long_description) LIKE ?)", pattern, pattern, pattern)
	}

	var repositories []models.Repository
	if err := applyRepositoryOrdering(queryDB).
		Preload("Organisation").
		Offset((page - 1) * perPage).
		Limit(perPage).
		Find(&repositories).Error; err != nil {
		return nil, models.Pagination{}, err
	}

	totalPages := 0
	if totalRecords > 0 {
		totalPages = int(math.Ceil(float64(totalRecords) / float64(perPage)))
	}
	pagination := models.Pagination{
		CurrentPage:    page,
		RecordsPerPage: perPage,
		TotalPages:     totalPages,
		TotalRecords:   int(totalRecords),
	}
	if page < totalPages {
		next := page + 1
		pagination.Next = &next
	}
	if page > 1 && totalPages > 0 {
		prev := page - 1
		pagination.Previous = &prev
	}

	return repositories, pagination, nil
}

func (r *repositoriesRepository) SaveOrganisatie(organisation *models.Organisation) error {
	return r.db.Save(organisation).Error
}

func (r *repositoriesRepository) AllRepositorys(ctx context.Context) ([]models.Repository, error) {
	var repositories []models.Repository
	if err := r.db.WithContext(ctx).Find(&repositories).Error; err != nil {
		return nil, err
	}
	return repositories, nil
}

func (r *repositoriesRepository) GetOrganisations(ctx context.Context) ([]models.Organisation, error) {
	var organisations []models.Organisation
	if err := r.db.WithContext(ctx).Order("label asc").Find(&organisations).Error; err != nil {
		return nil, err
	}
	return organisations, nil
}

func (r *repositoriesRepository) FindOrganisationByURI(ctx context.Context, uri string) (*models.Organisation, error) {
	var org models.Organisation
	if err := r.db.WithContext(ctx).Where("uri = ?", uri).First(&org).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &org, nil
}

func (r *repositoriesRepository) SaveGitOrganisatie(ctx context.Context, gitOrg *models.GitOrganisatie) error {
	return r.db.WithContext(ctx).Save(gitOrg).Error
}

func (r *repositoriesRepository) FindGitOrganisationByURL(ctx context.Context, url string) (*models.GitOrganisatie, error) {
	var gitOrg models.GitOrganisatie
	err := r.db.WithContext(ctx).
		Preload("Organisation").
		Where("url = ?", url).
		First(&gitOrg).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &gitOrg, nil
}

func applyRepositoryOrdering(db *gorm.DB) *gorm.DB {
	return db.Order("(public_code_url IS NOT NULL AND public_code_url <> '') DESC").
		Order("last_activity_at DESC").
		Order("name")
}
