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
	listFunc            func(ctx context.Context, page, perPage int, p *models.RepositoryFiltersParams) ([]models.Repository, models.Pagination, error)
	retrieveFunc        func(ctx context.Context, id string) (*models.Repository, error)
	searchFunc          func(ctx context.Context, page, perPage int, organisation *string, query string) ([]models.Repository, models.Pagination, error)
	saveRepositoryFunc  func(ctx context.Context, repository *models.Repository) error
	saveOrgFunc         func(org *models.Organisation) error
	getOrgFunc          func(ctx context.Context, page, perPage int) ([]models.Organisation, models.Pagination, error)
	gitOrgListFunc      func(ctx context.Context, page, perPage int, organisation *string) ([]models.GitOrganisatie, models.Pagination, error)
	findOrgByURIF       func(ctx context.Context, uri string) (*models.Organisation, error)
	findGitOrgByURLFunc func(ctx context.Context, url string) (*models.GitOrganisatie, error)
	saveGitOrgFunc      func(ctx context.Context, gitOrg *models.GitOrganisatie) error
	filterCountsFunc    func(ctx context.Context, p *models.RepositoryFiltersParams) (*models.RepositoryFilterCounts, error)
}

func (s *stubRepo) GetRepositorys(ctx context.Context, page, perPage int, p *models.RepositoryFiltersParams) ([]models.Repository, models.Pagination, error) {
	if s.listFunc != nil {
		return s.listFunc(ctx, page, perPage, p)
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
	if s.saveRepositoryFunc != nil {
		return s.saveRepositoryFunc(ctx, repository)
	}
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

func (s *stubRepo) GetOrganisations(ctx context.Context, page, perPage int) ([]models.Organisation, models.Pagination, error) {
	if s.getOrgFunc != nil {
		return s.getOrgFunc(ctx, page, perPage)
	}
	return nil, models.Pagination{}, nil
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

func (s *stubRepo) GetRepositoryFilterCounts(ctx context.Context, p *models.RepositoryFiltersParams) (*models.RepositoryFilterCounts, error) {
	if s.filterCountsFunc != nil {
		return s.filterCountsFunc(ctx, p)
	}
	return &models.RepositoryFilterCounts{}, nil
}

func TestListRepositories_ReturnsSummaries(t *testing.T) {
	org := &models.Organisation{Uri: "org-1", Label: "Org 1"}
	lastActivity := time.Date(2024, 5, 10, 12, 0, 0, 0, time.UTC)
	repo := &stubRepo{
		listFunc: func(ctx context.Context, page, perPage int, p *models.RepositoryFiltersParams) ([]models.Repository, models.Pagination, error) {
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

func TestListRepositories_ForwardsAllFilters(t *testing.T) {
	orgURI := "org-1"
	date := "2024-01-01"
	publicCode := true
	repo := &stubRepo{
		listFunc: func(ctx context.Context, page, perPage int, p *models.RepositoryFiltersParams) ([]models.Repository, models.Pagination, error) {
			require.Equal(t, &orgURI, p.Organisation)
			require.Equal(t, &publicCode, p.PublicCode)
			require.Equal(t, &date, p.LastActivityAfter)
			require.Equal(t, []string{"library"}, p.SoftwareType)
			require.Equal(t, []string{"stable"}, p.DevelopmentStatus)
			require.Equal(t, []string{"nl"}, p.AvailableLanguages)
			require.Equal(t, []string{"internal"}, p.MaintenanceType)
			require.Equal(t, []string{"MIT"}, p.License)
			require.Equal(t, []string{"web"}, p.Platforms)
			return []models.Repository{}, models.Pagination{}, nil
		},
	}
	svc := services.NewRepositoryService(repo)

	_, _, err := svc.ListRepositorys(context.Background(), &models.ListRepositorysParams{
		Organisation:       &orgURI,
		PublicCode:         &publicCode,
		LastActivityAfter:  &date,
		SoftwareType:       []string{"library"},
		DevelopmentStatus:  []string{"stable"},
		AvailableLanguages: []string{"nl"},
		MaintenanceType:    []string{"internal"},
		License:            []string{"MIT"},
		Platforms:          []string{"web"},
	})
	require.NoError(t, err)
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

func TestGetRepositoryFilters_ReturnsAllGroups(t *testing.T) {
	repo := &stubRepo{}
	svc := services.NewRepositoryService(repo)

	groups, err := svc.GetRepositoryFilters(context.Background(), &models.RepositoryFiltersParams{})
	require.NoError(t, err)

	keys := make([]string, len(groups))
	for i, g := range groups {
		keys[i] = g.Key
	}
	assert.Contains(t, keys, "publiccode")
	assert.Contains(t, keys, "lastActivityAfter")
	assert.Contains(t, keys, "softwareType")
	assert.Contains(t, keys, "developmentStatus")
	assert.Contains(t, keys, "maintenanceType")
	assert.Contains(t, keys, "platforms")
	assert.Contains(t, keys, "availableLanguages")
	assert.Contains(t, keys, "license")
	assert.Contains(t, keys, "organisation")
}

func TestGetRepositoryFilters_ToggleGroupHasCount(t *testing.T) {
	n := 42
	repo := &stubRepo{
		filterCountsFunc: func(ctx context.Context, p *models.RepositoryFiltersParams) (*models.RepositoryFilterCounts, error) {
			return &models.RepositoryFilterCounts{PublicCode: n}, nil
		},
	}
	svc := services.NewRepositoryService(repo)

	groups, err := svc.GetRepositoryFilters(context.Background(), &models.RepositoryFiltersParams{})
	require.NoError(t, err)

	var publiccodeGroup models.FilterGroup
	for _, g := range groups {
		if g.Key == "publiccode" {
			publiccodeGroup = g
		}
	}
	require.NotNil(t, publiccodeGroup.Count)
	assert.Equal(t, n, *publiccodeGroup.Count)
	assert.Equal(t, "toggle", publiccodeGroup.Type)
	assert.Equal(t, false, publiccodeGroup.Value)
}

func TestGetRepositoryFilters_ToggleValue_TrueWhenActive(t *testing.T) {
	repo := &stubRepo{}
	svc := services.NewRepositoryService(repo)
	trueVal := true

	groups, err := svc.GetRepositoryFilters(context.Background(), &models.RepositoryFiltersParams{PublicCode: &trueVal})
	require.NoError(t, err)

	for _, g := range groups {
		if g.Key == "publiccode" {
			assert.Equal(t, true, g.Value)
		}
	}
}

func TestGetRepositoryFilters_MultiSelectOptionsSelected(t *testing.T) {
	repo := &stubRepo{
		filterCountsFunc: func(ctx context.Context, p *models.RepositoryFiltersParams) (*models.RepositoryFilterCounts, error) {
			return &models.RepositoryFilterCounts{
				SoftwareType: []models.FilterCount{
					{Value: "library", Count: 10},
					{Value: "addon", Count: 5},
				},
			}, nil
		},
	}
	svc := services.NewRepositoryService(repo)
	p := &models.RepositoryFiltersParams{SoftwareType: []string{"library"}}

	groups, err := svc.GetRepositoryFilters(context.Background(), p)
	require.NoError(t, err)

	for _, g := range groups {
		if g.Key == "softwareType" {
			require.Len(t, g.Options, 2)
			assert.True(t, g.Options[0].Selected)
			assert.False(t, g.Options[1].Selected)
		}
	}
}

func TestGetRepositoryFilters_DateGroup_NoCountWhenEmpty(t *testing.T) {
	repo := &stubRepo{}
	svc := services.NewRepositoryService(repo)

	groups, err := svc.GetRepositoryFilters(context.Background(), &models.RepositoryFiltersParams{})
	require.NoError(t, err)

	for _, g := range groups {
		if g.Key == "lastActivityAfter" {
			assert.Equal(t, "date", g.Type)
			assert.Nil(t, g.Value)
			assert.Nil(t, g.Count)
		}
	}
}

func TestCreateRepository_UsesPublicCodeURLAsRepositoryURL(t *testing.T) {
	publicCode := `publiccodeYmlVersion: "0.5.0"
name: Digitale Balie
url: https://example.org/repo-from-publiccode
softwareType: configurationFiles
developmentStatus: stable
platforms:
  - web
description:
  nl:
    shortDescription: Korte beschrijving van de Digitale Balie.
    longDescription: De Digitale Balie maakt dienstverlening persoonlijk met videobellen en ondersteunt gesprekken, verificatie en veilige documentuitwisseling voor burgers en ondernemers binnen gemeentelijke processen.
legal:
  license: EUPL-1.2
maintenance:
  type: internal
  contacts:
    - name: Team Digitale Balie
localisation:
  localisationReady: false
  availableLanguages:
    - nl
`

	org := &models.Organisation{Uri: "https://example.org/organisations/test", Label: "Test Org"}
	var saved *models.Repository
	repo := &stubRepo{
		findOrgByURIF: func(ctx context.Context, uri string) (*models.Organisation, error) {
			assert.Equal(t, org.Uri, uri)
			return org, nil
		},
		saveRepositoryFunc: func(ctx context.Context, repository *models.Repository) error {
			saved = repository
			return nil
		},
	}
	svc := services.NewRepositoryService(repo)

	inputURL := "https://manual.example/repo"
	created, err := svc.CreateRepository(context.Background(), models.RepositoryInput{
		Url:             &inputURL,
		OrganisationUri: &org.Uri,
		PublicCodeUrl:   &publicCode,
	})
	require.NoError(t, err)
	require.NotNil(t, created)
	require.NotNil(t, saved)
	assert.Equal(t, "https://example.org/repo-from-publiccode", saved.Url)
	assert.Equal(t, "https://example.org/repo-from-publiccode", created.Url)
}
