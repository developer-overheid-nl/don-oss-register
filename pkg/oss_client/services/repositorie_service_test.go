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
	listFunc      func(ctx context.Context, page, perPage int, organisation *string, ids *string) ([]models.Repositorie, models.Pagination, error)
	retrieveFunc  func(ctx context.Context, id string) (*models.Repositorie, error)
	searchFunc    func(ctx context.Context, page, perPage int, organisation *string, query string) ([]models.Repositorie, models.Pagination, error)
	saveOrgFunc   func(org *models.Organisation) error
	getOrgFunc    func(ctx context.Context) ([]models.Organisation, int, error)
	findOrgByURIF func(ctx context.Context, uri string) (*models.Organisation, error)
}

func (s *stubRepo) GetRepositories(ctx context.Context, page, perPage int, organisation *string, ids *string) ([]models.Repositorie, models.Pagination, error) {
	if s.listFunc != nil {
		return s.listFunc(ctx, page, perPage, organisation, ids)
	}
	return nil, models.Pagination{}, nil
}

func (s *stubRepo) GetRepositorieByID(ctx context.Context, id string) (*models.Repositorie, error) {
	if s.retrieveFunc != nil {
		return s.retrieveFunc(ctx, id)
	}
	return nil, nil
}

func (s *stubRepo) SaveRepositorie(ctx context.Context, repository *models.Repositorie) error {
	return nil
}

func (s *stubRepo) SearchRepositories(ctx context.Context, page, perPage int, organisation *string, query string) ([]models.Repositorie, models.Pagination, error) {
	if s.searchFunc != nil {
		return s.searchFunc(ctx, page, perPage, organisation, query)
	}
	return []models.Repositorie{}, models.Pagination{}, nil
}

func (s *stubRepo) SaveOrganisatie(org *models.Organisation) error {
	if s.saveOrgFunc != nil {
		return s.saveOrgFunc(org)
	}
	return nil
}

func (s *stubRepo) AllRepositories(ctx context.Context) ([]models.Repositorie, error) {
	return nil, nil
}

func (s *stubRepo) GetOrganisations(ctx context.Context) ([]models.Organisation, int, error) {
	if s.getOrgFunc != nil {
		return s.getOrgFunc(ctx)
	}
	return nil, 0, nil
}

func (s *stubRepo) FindOrganisationByURI(ctx context.Context, uri string) (*models.Organisation, error) {
	if s.findOrgByURIF != nil {
		return s.findOrgByURIF(ctx, uri)
	}
	return nil, gorm.ErrRecordNotFound
}

func TestListRepositories_ReturnsSummaries(t *testing.T) {
	org := &models.Organisation{Uri: "org-1", Label: "Org 1"}
	repo := &stubRepo{
		listFunc: func(ctx context.Context, page, perPage int, organisation *string, ids *string) ([]models.Repositorie, models.Pagination, error) {
			return []models.Repositorie{
				{
					Id:           "repo-1",
					Name:         "Repo One",
					Description:  "desc",
					Organisation: org,
				},
			}, models.Pagination{TotalRecords: 1, CurrentPage: 1, RecordsPerPage: 10}, nil
		},
	}
	svc := services.NewRepositoriesService(repo)

	results, pagination, err := svc.ListRepositories(context.Background(), &models.ListRepositoriesParams{Page: 1, PerPage: 10})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "repo-1", results[0].Id)
	assert.Equal(t, 1, pagination.TotalRecords)
}

func TestRetrieveRepositorie_ReturnsDetail(t *testing.T) {
	repo := &stubRepo{
		retrieveFunc: func(ctx context.Context, id string) (*models.Repositorie, error) {
			return &models.Repositorie{Id: id, Name: "Repo"}, nil
		},
	}
	svc := services.NewRepositoriesService(repo)

	detail, err := svc.RetrieveRepositorie(context.Background(), "repo-1")
	require.NoError(t, err)
	require.NotNil(t, detail)
	assert.Equal(t, "repo-1", detail.Id)
}

func TestRetrieveRepositorie_NotFoundPassesThrough(t *testing.T) {
	repo := &stubRepo{
		retrieveFunc: func(ctx context.Context, id string) (*models.Repositorie, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	svc := services.NewRepositoriesService(repo)

	detail, err := svc.RetrieveRepositorie(context.Background(), "missing")
	assert.Error(t, err)
	assert.Nil(t, detail)
}

func TestSearchRepositories_ReturnsEmptyOnBlankQuery(t *testing.T) {
	repo := &stubRepo{}
	svc := services.NewRepositoriesService(repo)

	results, pagination, err := svc.SearchRepositories(context.Background(), &models.ListRepositoriesSearchParams{Query: "   "})
	require.NoError(t, err)
	assert.Len(t, results, 0)
	assert.Equal(t, 0, pagination.TotalRecords)
}

func TestCreateOrganisation_ValidatesInput(t *testing.T) {
	repo := &stubRepo{}
	svc := services.NewRepositoriesService(repo)

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
	svc := services.NewRepositoriesService(repo)

	org := &models.Organisation{Uri: "https://example.org", Label: "Example"}
	created, err := svc.CreateOrganisation(context.Background(), org)
	require.NoError(t, err)
	assert.Equal(t, saved, created)
}
