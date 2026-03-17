package repositories

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"gorm.io/gorm"
)

type RepositoriesRepository interface {
	GetRepositorys(ctx context.Context, page, perPage int, p *models.RepositoryFiltersParams) ([]models.Repository, models.Pagination, error)
	GetRepositoryByID(ctx context.Context, oasUrl string) (*models.Repository, error)
	SaveRepository(ctx context.Context, repository *models.Repository) error
	SearchRepositorys(ctx context.Context, page, perPage int, organisation *string, query string) ([]models.Repository, models.Pagination, error)
	SaveOrganisatie(organisation *models.Organisation) error
	AllRepositorys(ctx context.Context) ([]models.Repository, error)
	GetOrganisations(ctx context.Context, page, perPage int) ([]models.Organisation, models.Pagination, error)
	FindOrganisationByURI(ctx context.Context, uri string) (*models.Organisation, error)
	GetGitOrganisations(ctx context.Context, page, perPage int, organisation *string) ([]models.GitOrganisatie, models.Pagination, error)
	FindGitOrganisationByURL(ctx context.Context, url string) (*models.GitOrganisatie, error)
	SaveGitOrganisatie(ctx context.Context, gitOrg *models.GitOrganisatie) error
	GetRepositoryFilterCounts(ctx context.Context, p *models.RepositoryFiltersParams) (*models.RepositoryFilterCounts, error)
}

type repositoriesRepository struct {
	db *gorm.DB
}

func NewRepositoriesRepository(db *gorm.DB) RepositoriesRepository {
	return &repositoriesRepository{db: db}
}

func (r *repositoriesRepository) SaveRepository(ctx context.Context, repository *models.Repository) error {
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

	if !found && repository.Url != "" {
		err := r.db.WithContext(ctx).Where("repository_url = ?", repository.Url).First(&existing).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err == nil {
			log.Printf("SaveRepository: found existing repository for url %q with id %s", repository.Url, existing.Id)
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

func (r *repositoriesRepository) GetRepositorys(ctx context.Context, page, perPage int, p *models.RepositoryFiltersParams) ([]models.Repository, models.Pagination, error) {
	if page < 1 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 20
	}
	if p == nil {
		p = &models.RepositoryFiltersParams{}
	}
	if p.LastActivityAfter != nil && strings.TrimSpace(*p.LastActivityAfter) != "" {
		if _, err := time.Parse("2006-01-02", *p.LastActivityAfter); err != nil {
			return nil, models.Pagination{}, fmt.Errorf("invalid lastActivityAfter format, expected YYYY-MM-DD: %w", err)
		}
	}

	var repositories []models.Repository
	if err := applyRepositoryOrdering(
		r.db.WithContext(ctx).Where("(active IS NULL OR active = ?)", true),
	).Preload("Organisation").Find(&repositories).Error; err != nil {
		return nil, models.Pagination{}, err
	}

	filtered := make([]models.Repository, 0, len(repositories))
	for _, repo := range repositories {
		if repoMatchesFilters(repo, p, "") {
			filtered = append(filtered, repo)
		}
	}

	totalRecords := len(filtered)
	totalPages := 0
	if totalRecords > 0 {
		totalPages = int(math.Ceil(float64(totalRecords) / float64(perPage)))
	}
	pagination := models.Pagination{
		CurrentPage:    page,
		RecordsPerPage: perPage,
		TotalPages:     totalPages,
		TotalRecords:   totalRecords,
	}

	if page < totalPages {
		next := page + 1
		pagination.Next = &next
	}
	if page > 1 {
		prev := page - 1
		pagination.Previous = &prev
	}

	offset := (page - 1) * perPage
	if offset >= totalRecords {
		return []models.Repository{}, pagination, nil
	}

	end := offset + perPage
	if end > totalRecords {
		end = totalRecords
	}

	return filtered[offset:end], pagination, nil
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

func (r *repositoriesRepository) GetOrganisations(ctx context.Context, page, perPage int) ([]models.Organisation, models.Pagination, error) {
	if page < 1 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	var organisations []models.Organisation
	if err := r.db.WithContext(ctx).Order("label asc").Offset(offset).Limit(perPage).Find(&organisations).Error; err != nil {
		return nil, models.Pagination{}, err
	}

	var totalRecords int64
	if err := r.db.Model(&models.Organisation{}).Count(&totalRecords).Error; err != nil {
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

func (r *repositoriesRepository) GetRepositoryFilterCounts(ctx context.Context, p *models.RepositoryFiltersParams) (*models.RepositoryFilterCounts, error) {
	if p == nil {
		p = &models.RepositoryFiltersParams{}
	}
	var allRepos []models.Repository
	if err := r.db.WithContext(ctx).
		Where("(active IS NULL OR active = ?)", true).
		Preload("Organisation").
		Find(&allRepos).Error; err != nil {
		return nil, err
	}

	result := &models.RepositoryFilterCounts{}

	result.PublicCode = countRepos(allRepos, p, "publiccode", func(repo models.Repository) bool {
		return repo.PublicCodeUrl != ""
	})

	if p.LastActivityAfter != nil && *p.LastActivityAfter != "" {
		date, err := time.Parse("2006-01-02", *p.LastActivityAfter)
		if err != nil {
			return nil, fmt.Errorf("invalid lastActivityAfter format, expected YYYY-MM-DD: %w", err)
		}
		n := countRepos(allRepos, p, "lastActivityAfter", func(repo models.Repository) bool {
			return !repo.LastActivityAt.Before(date)
		})
		result.LastActivityAfter = &n
	}

	result.SoftwareType = countByField(allRepos, p, "softwareType", func(repo models.Repository) string {
		if repo.PublicCode == nil {
			return ""
		}
		return repo.PublicCode.SoftwareType
	})

	result.DevelopmentStatus = countByField(allRepos, p, "developmentStatus", func(repo models.Repository) string {
		if repo.PublicCode == nil {
			return ""
		}
		return repo.PublicCode.DevelopmentStatus
	})

	result.MaintenanceType = countByField(allRepos, p, "maintenanceType", func(repo models.Repository) string {
		if repo.PublicCode == nil || repo.PublicCode.Maintenance == nil {
			return ""
		}
		return repo.PublicCode.Maintenance.Type
	})

	result.License = countByField(allRepos, p, "license", func(repo models.Repository) string {
		if repo.PublicCode == nil || repo.PublicCode.Legal == nil {
			return ""
		}
		return repo.PublicCode.Legal.License
	})

	result.Platforms = countByArrayField(allRepos, p, "platforms", func(repo models.Repository) []string {
		if repo.PublicCode == nil {
			return nil
		}
		return repo.PublicCode.Platforms
	})

	result.AvailableLanguages = countByArrayField(allRepos, p, "availableLanguages", func(repo models.Repository) []string {
		if repo.PublicCode == nil || repo.PublicCode.Localisation == nil {
			return nil
		}
		return repo.PublicCode.Localisation.AvailableLanguages
	})

	orgCounts := make(map[string]*models.OrgFilterCount)
	for _, repo := range allRepos {
		if !repoMatchesFilters(repo, p, "organisation") {
			continue
		}
		if repo.OrganisationID == nil || *repo.OrganisationID == "" {
			continue
		}
		orgID := *repo.OrganisationID
		if _, ok := orgCounts[orgID]; !ok {
			label := orgID
			if repo.Organisation != nil {
				label = repo.Organisation.Label
			}
			orgCounts[orgID] = &models.OrgFilterCount{Value: orgID, Label: label}
		}
		orgCounts[orgID].Count++
	}
	for _, fc := range orgCounts {
		result.Organisation = append(result.Organisation, *fc)
	}
	sort.Slice(result.Organisation, func(i, j int) bool {
		return result.Organisation[i].Count > result.Organisation[j].Count
	})

	return result, nil
}

func countRepos(repos []models.Repository, p *models.RepositoryFiltersParams, exclude string, match func(models.Repository) bool) int {
	count := 0
	for _, repo := range repos {
		if repoMatchesFilters(repo, p, exclude) && match(repo) {
			count++
		}
	}
	return count
}

func countByField(repos []models.Repository, p *models.RepositoryFiltersParams, exclude string, getValue func(models.Repository) string) []models.FilterCount {
	counts := make(map[string]int)
	for _, repo := range repos {
		if !repoMatchesFilters(repo, p, exclude) {
			continue
		}
		if val := getValue(repo); val != "" {
			counts[val]++
		}
	}
	result := make([]models.FilterCount, 0, len(counts))
	for val, count := range counts {
		result = append(result, models.FilterCount{Value: val, Count: count})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Count > result[j].Count })
	return result
}

func countByArrayField(repos []models.Repository, p *models.RepositoryFiltersParams, exclude string, getValues func(models.Repository) []string) []models.FilterCount {
	counts := make(map[string]int)
	for _, repo := range repos {
		if !repoMatchesFilters(repo, p, exclude) {
			continue
		}
		for _, val := range getValues(repo) {
			if val != "" {
				counts[val]++
			}
		}
	}
	result := make([]models.FilterCount, 0, len(counts))
	for val, count := range counts {
		result = append(result, models.FilterCount{Value: val, Count: count})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Count > result[j].Count })
	return result
}

func repoMatchesFilters(repo models.Repository, p *models.RepositoryFiltersParams, exclude string) bool {
	if p == nil {
		return true
	}
	if exclude != "organisation" && p.Organisation != nil && strings.TrimSpace(*p.Organisation) != "" {
		if repo.OrganisationID == nil || *repo.OrganisationID != strings.TrimSpace(*p.Organisation) {
			return false
		}
	}
	if exclude != "publiccode" && p.PublicCode != nil {
		if *p.PublicCode {
			// Require repositories that have a publiccode URL
			if repo.PublicCodeUrl == "" {
				return false
			}
		} else {
			// Require repositories that do NOT have a publiccode URL
			if repo.PublicCodeUrl != "" {
				return false
			}
		}
	}
	if exclude != "lastActivityAfter" && p.LastActivityAfter != nil && *p.LastActivityAfter != "" {
		if date, err := time.Parse("2006-01-02", *p.LastActivityAfter); err == nil {
			if repo.LastActivityAt.Before(date) {
				return false
			}
		}
	}
	if exclude != "softwareType" && len(p.SoftwareType) > 0 {
		st := ""
		if repo.PublicCode != nil {
			st = repo.PublicCode.SoftwareType
		}
		if !containsStr(p.SoftwareType, st) {
			return false
		}
	}
	if exclude != "developmentStatus" && len(p.DevelopmentStatus) > 0 {
		ds := ""
		if repo.PublicCode != nil {
			ds = repo.PublicCode.DevelopmentStatus
		}
		if !containsStr(p.DevelopmentStatus, ds) {
			return false
		}
	}
	if exclude != "maintenanceType" && len(p.MaintenanceType) > 0 {
		mt := ""
		if repo.PublicCode != nil && repo.PublicCode.Maintenance != nil {
			mt = repo.PublicCode.Maintenance.Type
		}
		if !containsStr(p.MaintenanceType, mt) {
			return false
		}
	}
	if exclude != "license" && len(p.License) > 0 {
		lic := ""
		if repo.PublicCode != nil && repo.PublicCode.Legal != nil {
			lic = repo.PublicCode.Legal.License
		}
		if !containsStr(p.License, lic) {
			return false
		}
	}
	if exclude != "platforms" && len(p.Platforms) > 0 {
		var repoPlatforms []string
		if repo.PublicCode != nil {
			repoPlatforms = repo.PublicCode.Platforms
		}
		for _, platform := range p.Platforms {
			if !containsStr(repoPlatforms, platform) {
				return false
			}
		}
	}
	if exclude != "availableLanguages" && len(p.AvailableLanguages) > 0 {
		var repoLangs []string
		if repo.PublicCode != nil && repo.PublicCode.Localisation != nil {
			repoLangs = repo.PublicCode.Localisation.AvailableLanguages
		}
		for _, lang := range p.AvailableLanguages {
			if !containsStr(repoLangs, lang) {
				return false
			}
		}
	}
	return true
}

func containsStr(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}
