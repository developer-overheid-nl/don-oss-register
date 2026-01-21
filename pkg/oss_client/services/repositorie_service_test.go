package services_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	problem "github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/helpers/problem"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type stubRepo struct {
	listFunc            func(ctx context.Context, page, perPage int, organisation *string, publicCode *bool) ([]models.Repository, models.Pagination, error)
	retrieveFunc        func(ctx context.Context, id string) (*models.Repository, error)
	searchFunc          func(ctx context.Context, page, perPage int, organisation *string, query string) ([]models.Repository, models.Pagination, error)
	saveOrgFunc         func(org *models.Organisation) error
	getOrgFunc          func(ctx context.Context) ([]models.Organisation, error)
	gitOrgListFunc      func(ctx context.Context, page, perPage int, organisation *string) ([]models.GitOrganisatie, models.Pagination, error)
	findOrgByURIF       func(ctx context.Context, uri string) (*models.Organisation, error)
	findGitOrgByURLFunc func(ctx context.Context, url string) (*models.GitOrganisatie, error)
	saveGitOrgFunc      func(ctx context.Context, gitOrg *models.GitOrganisatie) error
}

func (s *stubRepo) GetRepositorys(ctx context.Context, page, perPage int, organisation *string, publicCode *bool) ([]models.Repository, models.Pagination, error) {
	if s.listFunc != nil {
		return s.listFunc(ctx, page, perPage, organisation, publicCode)
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

func (s *stubRepo) GetOrganisations(ctx context.Context) ([]models.Organisation, error) {
	if s.getOrgFunc != nil {
		return s.getOrgFunc(ctx)
	}
	return nil, nil
}

func (s *stubRepo) GetGitOrganisations(ctx context.Context, page, perPage int, organisation *string) ([]models.GitOrganisatie, models.Pagination, error) {
	if s.gitOrgListFunc != nil {
		return s.gitOrgListFunc(ctx, page, perPage, organisation)
	}
	return nil, models.Pagination{}, nil
}

func (s *stubRepo) FindOrganisationByURI(ctx context.Context, uri string) (*models.Organisation, error) {
	if s.findOrgByURIF != nil {
		return s.findOrgByURIF(ctx, uri)
	}
	return nil, nil
}

func (s *stubRepo) FindGitOrganisationByURL(ctx context.Context, url string) (*models.GitOrganisatie, error) {
	if s.findGitOrgByURLFunc != nil {
		return s.findGitOrgByURLFunc(ctx, url)
	}
	return nil, nil
}

func (s *stubRepo) SaveGitOrganisatie(ctx context.Context, gitOrg *models.GitOrganisatie) error {
	if s.saveGitOrgFunc != nil {
		return s.saveGitOrgFunc(ctx, gitOrg)
	}
	return nil
}

func TestListRepositories_ReturnsSummaries(t *testing.T) {
	org := &models.Organisation{Uri: "org-1", Label: "Org 1"}
	lastActivity := time.Date(2024, 5, 10, 12, 0, 0, 0, time.UTC)
	repo := &stubRepo{
		listFunc: func(ctx context.Context, page, perPage int, organisation *string, publicCode *bool) ([]models.Repository, models.Pagination, error) {
			return []models.Repository{
				{
					Id:               "repo-1",
					Name:             "Repo One",
					ShortDescription: "desc",
					LongDescription:  "desc",
					Organisation:     org,
					LastActivityAt:   lastActivity,
				},
			}, models.Pagination{TotalRecords: 1, CurrentPage: 1, RecordsPerPage: 10}, nil
		},
	}
	svc := services.NewRepositoryService(repo)

	results, pagination, err := svc.ListRepositorys(context.Background(), &models.ListRepositorysParams{Page: 1, PerPage: 10})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "repo-1", results[0].Id)
	assert.Equal(t, lastActivity, results[0].LastActivityAt)
	assert.Equal(t, 1, pagination.TotalRecords)
}

func TestRetrieveRepository_ReturnsDetail(t *testing.T) {
	lastActivity := time.Date(2024, 5, 10, 12, 0, 0, 0, time.UTC)
	repo := &stubRepo{
		retrieveFunc: func(ctx context.Context, id string) (*models.Repository, error) {
			return &models.Repository{Id: id, Name: "Repo", LastActivityAt: lastActivity}, nil
		},
	}
	svc := services.NewRepositoryService(repo)

	detail, err := svc.RetrieveRepository(context.Background(), "repo-1")
	require.NoError(t, err)
	require.NotNil(t, detail)
	assert.Equal(t, "repo-1", detail.Id)
	assert.Equal(t, lastActivity, detail.LastActivityAt)
}

func TestRetrieveRepository_InvalidIDReturnsBadRequest(t *testing.T) {
	repo := &stubRepo{
		retrieveFunc: func(ctx context.Context, id string) (*models.Repository, error) {
			t.Fatalf("expected GetRepositoryByID not to be called for invalid id")
			return nil, nil
		},
	}
	svc := services.NewRepositoryService(repo)

	_, err := svc.RetrieveRepository(context.Background(), "bad\x00id")
	require.Error(t, err)
	var apiErr problem.ProblemJSON
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusBadRequest, apiErr.Status)
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

	_, _, err := svc.SearchRepositorys(context.Background(), &models.ListRepositorysSearchParams{Query: "   "})
	require.Error(t, err)
}

func TestCreateOrganisation_ValidatesInput(t *testing.T) {
	repo := &stubRepo{}
	svc := services.NewRepositoryService(repo)

	_, err := svc.CreateOrganisation(context.Background(), &models.Organisation{
		Uri:   "notaurl",
		Label: "Label",
	})
	require.Error(t, err)
	var apiErr problem.ProblemJSON
	assert.ErrorAs(t, err, &apiErr)

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
