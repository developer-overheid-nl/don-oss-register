package repositories

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RepositoriesRepository interface {
	GetRepositorys(ctx context.Context, page, perPage int, organisation *string, ids *string) ([]models.Repository, models.Pagination, error)
	GetRepositoryByID(ctx context.Context, oasUrl string) (*models.Repository, error)
	SaveRepository(ctx context.Context, repository *models.Repository) error
	SearchRepositorys(ctx context.Context, page, perPage int, organisation *string, query string) ([]models.Repository, models.Pagination, error)
	SaveOrganisatie(organisation *models.Organisation) error
	AllRepositorys(ctx context.Context) ([]models.Repository, error)
	GetOrganisations(ctx context.Context, page, perPage int, ids *string) ([]models.Organisation, models.Pagination, error)
	FindOrganisationByURI(ctx context.Context, uri string) (*models.Organisation, error)
	GetGitOrganisations(ctx context.Context, page, perPage int, ids *string) ([]models.GitOrganisatie, models.Pagination, error)
	FindGitOrganisationByOrganisationURI(ctx context.Context, organisationURI string) (*models.GitOrganisatie, error)
	SaveGitOrganisatie(ctx context.Context, gitOrg *models.GitOrganisatie) error
	AddCodeHosting(ctx context.Context, gitOrganisationID string, url string, isGroup *bool) (*models.CodeHosting, error)
}

type repositoriesRepository struct {
	db *gorm.DB
}

func NewRepositoriesRepository(db *gorm.DB) RepositoriesRepository {
	return &repositoriesRepository{db: db}
}

func (r *repositoriesRepository) SaveRepository(ctx context.Context, repository *models.Repository) error {
	//todo upsurt?
	return r.db.Create(repository).Error
}

func (r *repositoriesRepository) GetRepositorys(ctx context.Context, page, perPage int, organisation *string, ids *string) ([]models.Repository, models.Pagination, error) {
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
	if err := db.Model(&models.Repository{}).Count(&totalRecords).Error; err != nil {
		return nil, models.Pagination{}, err
	}

	var repositories []models.Repository
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

func (r *repositoriesRepository) GetGitOrganisations(ctx context.Context, page, perPage int, ids *string) ([]models.GitOrganisatie, models.Pagination, error) {
	offset := (page - 1) * perPage

	db := r.db
	if ids != nil {
		idsSlice := strings.Split(*ids, ",")
		for i := range idsSlice {
			idsSlice[i] = strings.TrimSpace(idsSlice[i])
		}
		db = db.Where("id IN ?", idsSlice)
	}

	var totalRecords int64
	if err := db.Model(&models.GitOrganisatie{}).Count(&totalRecords).Error; err != nil {
		return nil, models.Pagination{}, err
	}

	var gitOrganisations []models.GitOrganisatie
	if err := db.Limit(perPage).Preload("Organisation").Preload("CodeHosting").Offset(offset).Find(&gitOrganisations).Error; err != nil {
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
	if err := r.db.Preload("Organisation").First(&api, "id = ?", id).Error; err != nil {
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
		perPage = 10
	}
	if trimmed == "" {
		return []models.Repository{}, models.Pagination{
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
	if err := base.Model(&models.Repository{}).Count(&totalRecords).Error; err != nil {
		return nil, models.Pagination{}, err
	}

	queryDB := r.db.WithContext(ctx)
	if organisation != nil && strings.TrimSpace(*organisation) != "" {
		queryDB = queryDB.Where("organisation_id = ?", strings.TrimSpace(*organisation))
	}
	if pattern != "" {
		queryDB = queryDB.Where("LOWER(name) LIKE ?", pattern)
	}

	var repositories []models.Repository
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

func (r *repositoriesRepository) AllRepositorys(ctx context.Context) ([]models.Repository, error) {
	var repositories []models.Repository
	if err := r.db.WithContext(ctx).Find(&repositories).Error; err != nil {
		return nil, err
	}
	return repositories, nil
}

func (r *repositoriesRepository) GetOrganisations(ctx context.Context, page, perPage int, ids *string) ([]models.Organisation, models.Pagination, error) {
	offset := (page - 1) * perPage

	db := r.db.WithContext(ctx)
	if ids != nil {
		idsSlice := strings.Split(*ids, ",")
		for i := range idsSlice {
			idsSlice[i] = strings.TrimSpace(idsSlice[i])
		}
		db = db.Where("uri IN ?", idsSlice)
	}

	var totalRecords int64
	if err := db.Model(&models.Organisation{}).Count(&totalRecords).Error; err != nil {
		return nil, models.Pagination{}, err
	}

	var organisations []models.Organisation
	if err := db.Limit(perPage).Offset(offset).Order("label asc").Find(&organisations).Error; err != nil {
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

	return organisations, pagination, nil
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

func (r *repositoriesRepository) FindGitOrganisationByOrganisationURI(ctx context.Context, organisationURI string) (*models.GitOrganisatie, error) {
	var gitOrg models.GitOrganisatie
	err := r.db.WithContext(ctx).
		Preload("Organisation").
		Preload("CodeHosting").
		Where("organisation_id = ?", organisationURI).
		First(&gitOrg).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &gitOrg, nil
}

func (r *repositoriesRepository) SaveGitOrganisatie(ctx context.Context, gitOrg *models.GitOrganisatie) error {
	return r.db.WithContext(ctx).Save(gitOrg).Error
}

func (r *repositoriesRepository) AddCodeHosting(ctx context.Context, gitOrganisationID string, url string, isGroup *bool) (*models.CodeHosting, error) {
	var existing models.CodeHosting
	if err := r.db.WithContext(ctx).Where("url = ?", url).First(&existing).Error; err == nil {
		return &existing, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	groupValue := true
	if isGroup != nil {
		groupValue = *isGroup
	}
	ch := models.CodeHosting{
		ID:          uuid.NewString(),
		URL:         url,
		Group:       &groupValue,
		PublisherID: gitOrganisationID,
	}
	if err := r.db.WithContext(ctx).Create(&ch).Error; err != nil {
		return nil, err
	}
	return &ch, nil
}
