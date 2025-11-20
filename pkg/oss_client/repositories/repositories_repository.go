package repositories

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"gorm.io/gorm"
)

type RepositoriesRepository interface {
	GetRepositories(ctx context.Context, page, perPage int, organisation *string, ids *string) ([]models.Repositorie, models.Pagination, error)
	GetRepositorieByID(ctx context.Context, oasUrl string) (*models.Repositorie, error)
	SaveRepositorie(ctx context.Context, repository *models.Repositorie) error
	SearchRepositories(ctx context.Context, page, perPage int, organisation *string, query string) ([]models.Repositorie, models.Pagination, error)
	SaveOrganisatie(organisation *models.Organisation) error
	AllRepositories(ctx context.Context) ([]models.Repositorie, error)
	GetOrganisations(ctx context.Context) ([]models.Organisation, int, error)
	FindOrganisationByURI(ctx context.Context, uri string) (*models.Organisation, error)
}

type repositoriesRepository struct {
	db *gorm.DB
}

func NewRepositoriesRepository(db *gorm.DB) RepositoriesRepository {
	return &repositoriesRepository{db: db}
}

func (r *repositoriesRepository) SaveRepositorie(ctx context.Context, repository *models.Repositorie) error {
	//todo upsurt?
	return r.db.Create(repository).Error
}

func (r *repositoriesRepository) GetRepositories(ctx context.Context, page, perPage int, organisation *string, ids *string) ([]models.Repositorie, models.Pagination, error) {
	offset := (page - 1) * perPage

	db := r.db
	if organisation != nil && strings.TrimSpace(*organisation) != "" {
		db = db.Where("organisation_id = ?", strings.TrimSpace(*organisation))
	}
	if ids != nil {
		idsSlice := strings.Split(*ids, ",")
		for i := range idsSlice {
			idsSlice[i] = strings.TrimSpace(idsSlice[i])
		}
		db = db.Where("id IN ?", idsSlice)
	}

	var totalRecords int64
	if err := db.Model(&models.Repositorie{}).Count(&totalRecords).Error; err != nil {
		return nil, models.Pagination{}, err
	}

	var repositories []models.Repositorie
	if err := db.Limit(perPage).Preload("Organisation").Offset(offset).Order("name").Find(&repositories).Error; err != nil {
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

func (r *repositoriesRepository) GetRepositorieByID(ctx context.Context, id string) (*models.Repositorie, error) {
	var api models.Repositorie
	if err := r.db.Preload("Organisation").First(&api, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &api, nil
}

func (r *repositoriesRepository) SearchRepositories(ctx context.Context, page, perPage int, organisation *string, query string) ([]models.Repositorie, models.Pagination, error) {
	trimmed := strings.TrimSpace(query)
	if page < 1 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 10
	}
	if trimmed == "" {
		return []models.Repositorie{}, models.Pagination{
			CurrentPage:    page,
			RecordsPerPage: perPage,
		}, nil
	}

	base := r.db.WithContext(ctx)
	if organisation != nil && strings.TrimSpace(*organisation) != "" {
		base = base.Where("organisation_id = ?", strings.TrimSpace(*organisation))
	}
	var pattern string
	if trimmed != "" {
		pattern = fmt.Sprintf("%%%s%%", strings.ToLower(trimmed))
		base = base.Where("LOWER(name) LIKE ?", pattern)
	}

	var totalRecords int64
	if err := base.Model(&models.Repositorie{}).Count(&totalRecords).Error; err != nil {
		return nil, models.Pagination{}, err
	}

	queryDB := r.db.WithContext(ctx)
	if organisation != nil && strings.TrimSpace(*organisation) != "" {
		queryDB = queryDB.Where("organisation_id = ?", strings.TrimSpace(*organisation))
	}
	if pattern != "" {
		queryDB = queryDB.Where("LOWER(name) LIKE ?", pattern)
	}

	var repositories []models.Repositorie
	if err := queryDB.
		Preload("Organisation").
		Order("name").
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

func (r *repositoriesRepository) AllRepositories(ctx context.Context) ([]models.Repositorie, error) {
	var repositories []models.Repositorie
	if err := r.db.WithContext(ctx).Find(&repositories).Error; err != nil {
		return nil, err
	}
	return repositories, nil
}

func (r *repositoriesRepository) GetOrganisations(ctx context.Context) ([]models.Organisation, int, error) {
	var organisations []models.Organisation
	var total int64
	db := r.db.WithContext(ctx)
	if err := db.Model(&models.Organisation{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Order("label asc").Find(&organisations).Error; err != nil {
		return nil, 0, err
	}
	return organisations, int(total), nil
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
