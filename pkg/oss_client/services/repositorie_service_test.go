package services_test

import (
	"context"
	"testing"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type stubRepo struct {
	listFunc            func(ctx context.Context, page, perPage int, organisation *string, ids *string) ([]models.Repository, models.Pagination, error)
	retrieveFunc        func(ctx context.Context, id string) (*models.Repository, error)
	searchFunc          func(ctx context.Context, page, perPage int, organisation *string, query string) ([]models.Repository, models.Pagination, error)
	saveOrgFunc         func(org *models.Organisation) error
	getOrgFunc          func(ctx context.Context, page, perPage int, ids *string) ([]models.Organisation, models.Pagination, error)
	gitOrgListFunc      func(ctx context.Context, page, perPage int, ids *string) ([]models.GitOrganisatie, models.Pagination, error)
	findOrgByURIF       func(ctx context.Context, uri string) (*models.Organisation, error)
	findGitOrgByOrgFunc func(ctx context.Context, organisationURI string) (*models.GitOrganisatie, error)
	saveGitOrgFunc      func(ctx context.Context, gitOrg *models.GitOrganisatie) error
	addCodeHostingFunc  func(ctx context.Context, gitOrgID, url string, isGroup *bool) (*models.CodeHosting, error)
}

func (s *stubRepo) GetRepositorys(ctx context.Context, page, perPage int, organisation *string, ids *string) ([]models.Repository, models.Pagination, error) {
	if s.listFunc != nil {
		return s.listFunc(ctx, page, perPage, organisation, ids)
	}
	return nil, models.Pagination{}, nil
}

func (s *stubRepo) GetRepositoryByID(ctx context.Context, id string) (*models.Repository, error) {
	if s.retrieveFunc != nil {
		return s.retrieveFunc(ctx, id)
	}
	return nil, nil
}

func (s *stubRepo) SaveRepository(ctx context.Context, repository *models.Repository) error {
	return nil
}

func (s *stubRepo) SearchRepositorys(ctx context.Context, page, perPage int, organisation *string, query string) ([]models.Repository, models.Pagination, error) {
	if s.searchFunc != nil {
		return s.searchFunc(ctx, page, perPage, organisation, query)
	}
	return []models.Repository{}, models.Pagination{}, nil
}

func (s *stubRepo) SaveOrganisatie(org *models.Organisation) error {
	if s.saveOrgFunc != nil {
		return s.saveOrgFunc(org)
	}
	return nil
}

func (s *stubRepo) AllRepositorys(ctx context.Context) ([]models.Repository, error) {
	return nil, nil
}

func (s *stubRepo) GetOrganisations(ctx context.Context, page, perPage int, ids *string) ([]models.Organisation, models.Pagination, error) {
	if s.getOrgFunc != nil {
		return s.getOrgFunc(ctx, page, perPage, ids)
	}
	return nil, models.Pagination{}, nil
}

func (s *stubRepo) GetGitOrganisations(ctx context.Context, page, perPage int, ids *string) ([]models.GitOrganisatie, models.Pagination, error) {
	if s.gitOrgListFunc != nil {
		return s.gitOrgListFunc(ctx, page, perPage, ids)
	}
	return nil, models.Pagination{}, nil
}

func (s *stubRepo) FindOrganisationByURI(ctx context.Context, uri string) (*models.Organisation, error) {
	if s.findOrgByURIF != nil {
		return s.findOrgByURIF(ctx, uri)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubRepo) FindGitOrganisationByOrganisationURI(ctx context.Context, organisationURI string) (*models.GitOrganisatie, error) {
	if s.findGitOrgByOrgFunc != nil {
		return s.findGitOrgByOrgFunc(ctx, organisationURI)
	}
	return nil, nil
}

func (s *stubRepo) SaveGitOrganisatie(ctx context.Context, gitOrg *models.GitOrganisatie) error {
	if s.saveGitOrgFunc != nil {
		return s.saveGitOrgFunc(ctx, gitOrg)
	}
	return nil
}

func (s *stubRepo) AddCodeHosting(ctx context.Context, gitOrgID, url string, isGroup *bool) (*models.CodeHosting, error) {
	if s.addCodeHostingFunc != nil {
		return s.addCodeHostingFunc(ctx, gitOrgID, url, isGroup)
	}
	return nil, nil
}

func TestListRepositories_ReturnsSummaries(t *testing.T) {
	org := &models.Organisation{Uri: "org-1", Label: "Org 1"}
	repo := &stubRepo{
		listFunc: func(ctx context.Context, page, perPage int, organisation *string, ids *string) ([]models.Repository, models.Pagination, error) {
			return []models.Repository{
				{
					Id:           "repo-1",
					Name:         "Repo One",
					Description:  "desc",
					Organisation: org,
				},
			}, models.Pagination{TotalRecords: 1, CurrentPage: 1, RecordsPerPage: 10}, nil
		},
	}
	svc := services.NewRepositoryService(repo)

	results, pagination, err := svc.ListRepositorys(context.Background(), &models.ListRepositorysParams{Page: 1, PerPage: 10})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "repo-1", results[0].Id)
	assert.Equal(t, 1, pagination.TotalRecords)
}

func TestRetrieveRepository_ReturnsDetail(t *testing.T) {
	repo := &stubRepo{
		retrieveFunc: func(ctx context.Context, id string) (*models.Repository, error) {
			return &models.Repository{Id: id, Name: "Repo"}, nil
		},
	}
	svc := services.NewRepositoryService(repo)

	detail, err := svc.RetrieveRepository(context.Background(), "repo-1")
	require.NoError(t, err)
	require.NotNil(t, detail)
	assert.Equal(t, "repo-1", detail.Id)
}

func TestRetrieveRepository_NotFoundPassesThrough(t *testing.T) {
	repo := &stubRepo{
		retrieveFunc: func(ctx context.Context, id string) (*models.Repository, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	svc := services.NewRepositoryService(repo)

	detail, err := svc.RetrieveRepository(context.Background(), "missing")
	assert.Error(t, err)
	assert.Nil(t, detail)
}

func TestSearchRepositories_ReturnsEmptyOnBlankQuery(t *testing.T) {
	repo := &stubRepo{}
	svc := services.NewRepositoryService(repo)

	results, pagination, err := svc.SearchRepositorys(context.Background(), &models.ListRepositorysSearchParams{Query: "   "})
	require.NoError(t, err)
	assert.Len(t, results, 0)
	assert.Equal(t, 0, pagination.TotalRecords)
}

func TestCreateOrganisation_ValidatesInput(t *testing.T) {
	repo := &stubRepo{}
	svc := services.NewRepositoryService(repo)

	_, err := svc.CreateOrganisation(context.Background(), &models.Organisation{
		Uri:   "notaurl",
		Label: "Label",
	})
	require.Error(t, err)
	_, isAPIError := err.(interface{ Error() string })
	assert.True(t, isAPIError)

	_, err = svc.CreateOrganisation(context.Background(), &models.Organisation{
		Uri:   "https://example.org",
		Label: " ",
	})
	require.Error(t, err)
}

func TestCreateOrganisation_Saves(t *testing.T) {
	var saved *models.Organisation
	repo := &stubRepo{
		saveOrgFunc: func(org *models.Organisation) error {
			saved = org
			return nil
		},
	}
	svc := services.NewRepositoryService(repo)

	org := &models.Organisation{Uri: "https://example.org", Label: "Example"}
	created, err := svc.CreateOrganisation(context.Background(), org)
	require.NoError(t, err)
	assert.Equal(t, saved, created)
}
