package services_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	httpclient "github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/helpers/httpclient"
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
	allRepositoriesFunc func(ctx context.Context) ([]models.Repository, error)
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
	if s.allRepositoriesFunc != nil {
		return s.allRepositoriesFunc(ctx)
	}
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
	archived := true
	query := "forms"
	repo := &stubRepo{
		listFunc: func(ctx context.Context, page, perPage int, p *models.RepositoryFiltersParams) ([]models.Repository, models.Pagination, error) {
			require.Equal(t, &orgURI, p.Organisation)
			require.Equal(t, query, p.Query)
			require.Equal(t, &publicCode, p.PublicCode)
			require.Equal(t, &archived, p.Archived)
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
		Query:              query,
		PublicCode:         &publicCode,
		Archived:           &archived,
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

func TestSearchRepositories_ReturnsEmptyOnBlankQuery(t *testing.T) {
	repo := &stubRepo{}
	svc := services.NewRepositoryService(repo)

	_, _, err := svc.SearchRepositorys(context.Background(), &models.ListRepositorysSearchParams{Query: "   "})
	require.Error(t, err)
}

func TestSearchRepositories_ReturnsSummaries(t *testing.T) {
	orgURI := "https://example.org/org"
	repo := &stubRepo{
		searchFunc: func(ctx context.Context, page, perPage int, organisation *string, query string) ([]models.Repository, models.Pagination, error) {
			require.Equal(t, 3, page)
			require.Equal(t, 7, perPage)
			require.Equal(t, &orgURI, organisation)
			require.Equal(t, "account", query)
			return []models.Repository{{
				Id:           "repo-1",
				Name:         "Account API",
				Organisation: &models.Organisation{Uri: orgURI, Label: "Org"},
			}}, models.Pagination{CurrentPage: 3, RecordsPerPage: 7, TotalRecords: 1}, nil
		},
	}
	svc := services.NewRepositoryService(repo)

	results, pagination, err := svc.SearchRepositorys(context.Background(), &models.ListRepositorysSearchParams{
		Page:         3,
		PerPage:      7,
		Organisation: &orgURI,
		Query:        " account ",
	})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "repo-1", results[0].Id)
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

func TestListGitOrganisations_ReturnsSummaries(t *testing.T) {
	orgURI := "https://example.org/org"
	repo := &stubRepo{
		gitOrgListFunc: func(ctx context.Context, page, perPage int, organisation *string) ([]models.GitOrganisatie, models.Pagination, error) {
			require.Equal(t, 2, page)
			require.Equal(t, 5, perPage)
			require.Equal(t, &orgURI, organisation)
			return []models.GitOrganisatie{
				{Id: "git-1", Url: "https://github.com/example", Organisation: &models.Organisation{Uri: orgURI, Label: "Example"}},
			}, models.Pagination{CurrentPage: 2, RecordsPerPage: 5, TotalRecords: 1}, nil
		},
	}
	svc := services.NewRepositoryService(repo)

	results, pagination, err := svc.ListGitOrganisations(context.Background(), &models.ListGitOrganisationsParams{
		Page:         2,
		PerPage:      5,
		Organisation: &orgURI,
	})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "git-1", results[0].Id)
	assert.Equal(t, "https://github.com/example", results[0].Url)
	assert.Equal(t, 2, pagination.CurrentPage)
}

func TestCreateGitOrganisatie_ValidatesInput(t *testing.T) {
	svc := services.NewRepositoryService(&stubRepo{})

	_, err := svc.CreateGitOrganisatie(context.Background(), models.GitOrganisationInput{
		Url:             "notaurl",
		OrganisationUri: "https://example.org/org",
	})
	require.Error(t, err)

	_, err = svc.CreateGitOrganisatie(context.Background(), models.GitOrganisationInput{
		Url:             "https://github.com/example",
		OrganisationUri: "notaurl",
	})
	require.Error(t, err)
}

func TestCreateGitOrganisatie_ReturnsExistingByURL(t *testing.T) {
	org := &models.Organisation{Uri: "https://example.org/org", Label: "Example"}
	existing := &models.GitOrganisatie{Id: "git-existing", Url: "https://github.com/example", OrganisationID: &org.Uri}
	repo := &stubRepo{
		findOrgByURIF: func(ctx context.Context, uri string) (*models.Organisation, error) {
			return org, nil
		},
		findGitOrgByURLFunc: func(ctx context.Context, url string) (*models.GitOrganisatie, error) {
			return existing, nil
		},
		saveGitOrgFunc: func(ctx context.Context, gitOrg *models.GitOrganisatie) error {
			t.Fatalf("SaveGitOrganisatie should not be called for an existing URL")
			return nil
		},
	}
	svc := services.NewRepositoryService(repo)

	got, err := svc.CreateGitOrganisatie(context.Background(), models.GitOrganisationInput{
		Url:             "https://github.com/example",
		OrganisationUri: org.Uri,
	})
	require.NoError(t, err)
	assert.Equal(t, existing, got)
}

func TestCreateGitOrganisatie_SavesNewOrganisation(t *testing.T) {
	org := &models.Organisation{Uri: "https://example.org/org", Label: "Example"}
	var saved *models.GitOrganisatie
	repo := &stubRepo{
		findOrgByURIF: func(ctx context.Context, uri string) (*models.Organisation, error) {
			return org, nil
		},
		findGitOrgByURLFunc: func(ctx context.Context, url string) (*models.GitOrganisatie, error) {
			return nil, nil
		},
		saveGitOrgFunc: func(ctx context.Context, gitOrg *models.GitOrganisatie) error {
			saved = gitOrg
			return nil
		},
	}
	svc := services.NewRepositoryService(repo)

	got, err := svc.CreateGitOrganisatie(context.Background(), models.GitOrganisationInput{
		Url:             "https://github.com/example",
		OrganisationUri: org.Uri,
	})
	require.NoError(t, err)
	require.NotNil(t, saved)
	assert.Equal(t, saved, got)
	assert.NotEmpty(t, got.Id)
	assert.Equal(t, "https://github.com/example", got.Url)
	assert.Equal(t, &org.Uri, got.OrganisationID)
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

func TestCreateOrganisation_ConflictWhenExistingOrganisationFound(t *testing.T) {
	existing := &models.Organisation{Uri: "https://example.org", Label: "Existing"}
	repo := &stubRepo{
		findOrgByURIF: func(ctx context.Context, uri string) (*models.Organisation, error) {
			return existing, nil
		},
	}
	svc := services.NewRepositoryService(repo)

	_, err := svc.CreateOrganisation(context.Background(), &models.Organisation{
		Uri:   "https://example.org",
		Label: "Example",
	})
	require.Error(t, err)
	var apiErr problem.ProblemJSON
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusConflict, apiErr.Status)
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

func TestListOrganisations_ReturnsSummaries(t *testing.T) {
	repo := &stubRepo{
		getOrgFunc: func(ctx context.Context, page, perPage int) ([]models.Organisation, models.Pagination, error) {
			require.Equal(t, 1, page)
			require.Equal(t, 100, perPage)
			return []models.Organisation{{Uri: "https://example.org", Label: "Example"}}, models.Pagination{TotalRecords: 1}, nil
		},
	}
	svc := services.NewRepositoryService(repo)

	results, pagination, err := svc.ListOrganisations(context.Background(), &models.ListOrganisationsParams{Page: 1, PerPage: 100})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "https://example.org", results[0].Uri)
	assert.Equal(t, "Example", results[0].Label)
	assert.Equal(t, 1, pagination.TotalRecords)
}

func TestUpdateRepository_ValidatesAndUpdatesExistingRepository(t *testing.T) {
	t.Setenv("ENABLE_TYPESENSE", "false")
	org := &models.Organisation{Uri: "https://example.org/new-org", Label: "New Org"}
	existing := &models.Repository{Id: "repo-1", Name: "Old", Url: "https://example.org/old", Active: false}
	var saved *models.Repository
	repo := &stubRepo{
		retrieveFunc: func(ctx context.Context, id string) (*models.Repository, error) {
			assert.Equal(t, "repo-1", id)
			return existing, nil
		},
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

	newURL := "https://example.org/new"
	newName := "New"
	detail, err := svc.UpdateRepository(context.Background(), "repo-1", models.RepositoryInput{
		Url:             &newURL,
		OrganisationUri: &org.Uri,
		Name:            &newName,
	})
	require.NoError(t, err)
	require.NotNil(t, detail)
	require.NotNil(t, saved)
	assert.Equal(t, "repo-1", saved.Id)
	assert.Equal(t, "New", saved.Name)
	assert.Equal(t, "https://example.org/new", saved.Url)
	assert.Equal(t, &org.Uri, saved.OrganisationID)
	assert.True(t, saved.Active)
	assert.Equal(t, "New", detail.Name)
}

func TestUpdateRepository_NotFoundAndInvalidInput(t *testing.T) {
	svc := services.NewRepositoryService(&stubRepo{})

	_, err := svc.UpdateRepository(context.Background(), "", models.RepositoryInput{})
	require.Error(t, err)

	repo := &stubRepo{
		retrieveFunc: func(ctx context.Context, id string) (*models.Repository, error) {
			return nil, nil
		},
	}
	svc = services.NewRepositoryService(repo)
	urlValue := "https://example.org/repo"
	_, err = svc.UpdateRepository(context.Background(), "missing", models.RepositoryInput{Url: &urlValue})
	require.Error(t, err)
	var apiErr problem.ProblemJSON
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusNotFound, apiErr.Status)

	repo = &stubRepo{
		retrieveFunc: func(ctx context.Context, id string) (*models.Repository, error) {
			return &models.Repository{Id: id}, nil
		},
	}
	svc = services.NewRepositoryService(repo)
	badURL := "notaurl"
	_, err = svc.UpdateRepository(context.Background(), "repo-1", models.RepositoryInput{Url: &badURL})
	require.Error(t, err)
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
	assert.Contains(t, keys, "archived")
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
			return &models.RepositoryFilterCounts{PublicCode: n, Archived: n}, nil
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
	assert.Equal(t, true, publiccodeGroup.Value)

	var archivedGroup models.FilterGroup
	for _, g := range groups {
		if g.Key == "archived" {
			archivedGroup = g
		}
	}
	require.NotNil(t, archivedGroup.Count)
	assert.Equal(t, n, *archivedGroup.Count)
	assert.Equal(t, "toggle", archivedGroup.Type)
	assert.Equal(t, false, archivedGroup.Value)
}

func TestGetRepositoryFilters_ToggleValue_TrueWhenActive(t *testing.T) {
	repo := &stubRepo{}
	svc := services.NewRepositoryService(repo)
	trueVal := true

	groups, err := svc.GetRepositoryFilters(context.Background(), &models.RepositoryFiltersParams{PublicCode: &trueVal, Archived: &trueVal})
	require.NoError(t, err)

	for _, g := range groups {
		if g.Key == "publiccode" {
			assert.Equal(t, true, g.Value)
		}
		if g.Key == "archived" {
			assert.Equal(t, true, g.Value)
		}
	}
}

func TestGetRepositoryFilters_PublicCodeFalseReturnsOnlyPublicCodeAndOrganisation(t *testing.T) {
	repo := &stubRepo{}
	svc := services.NewRepositoryService(repo)
	falseVal := false

	groups, err := svc.GetRepositoryFilters(context.Background(), &models.RepositoryFiltersParams{PublicCode: &falseVal})
	require.NoError(t, err)

	keys := make([]string, len(groups))
	for i, g := range groups {
		keys[i] = g.Key
	}
	assert.Equal(t, []string{"publiccode", "archived", "organisation"}, keys)
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
			assert.Equal(t, "addon", g.Options[0].Value)
			assert.False(t, g.Options[0].Selected)
			assert.Equal(t, "library", g.Options[1].Value)
			assert.True(t, g.Options[1].Selected)
		}
	}
}

func TestGetRepositoryFilters_MultiSelectKeepsSelectedOptionWithoutCount(t *testing.T) {
	repo := &stubRepo{
		filterCountsFunc: func(ctx context.Context, p *models.RepositoryFiltersParams) (*models.RepositoryFilterCounts, error) {
			return &models.RepositoryFilterCounts{}, nil
		},
	}
	svc := services.NewRepositoryService(repo)
	p := &models.RepositoryFiltersParams{
		Query:        "bla",
		SoftwareType: []string{"configurationFiles"},
	}

	groups, err := svc.GetRepositoryFilters(context.Background(), p)
	require.NoError(t, err)

	for _, g := range groups {
		if g.Key == "softwareType" {
			require.Len(t, g.Options, 1)
			assert.Equal(t, "configurationFiles", g.Options[0].Value)
			assert.Equal(t, "Configuratiebestanden", g.Options[0].Label)
			assert.Equal(t, 0, g.Options[0].Count)
			assert.True(t, g.Options[0].Selected)
		}
	}
}

func TestGetRepositoryFilters_OrganisationKeepsSelectedOptionWithoutCount(t *testing.T) {
	repo := &stubRepo{
		filterCountsFunc: func(ctx context.Context, p *models.RepositoryFiltersParams) (*models.RepositoryFilterCounts, error) {
			return &models.RepositoryFilterCounts{}, nil
		},
	}
	svc := services.NewRepositoryService(repo)
	org := "https://example.org/org"

	groups, err := svc.GetRepositoryFilters(context.Background(), &models.RepositoryFiltersParams{
		Query:        "bla",
		Organisation: &org,
	})
	require.NoError(t, err)

	for _, g := range groups {
		if g.Key == "organisation" {
			require.Len(t, g.Options, 1)
			assert.Equal(t, org, g.Options[0].Value)
			assert.Equal(t, org, g.Options[0].Label)
			assert.Equal(t, 0, g.Options[0].Count)
			assert.True(t, g.Options[0].Selected)
		}
	}
}

func TestGetRepositoryFilters_OptionsSortedAlphabeticallyByLabel(t *testing.T) {
	repo := &stubRepo{
		filterCountsFunc: func(ctx context.Context, p *models.RepositoryFiltersParams) (*models.RepositoryFilterCounts, error) {
			return &models.RepositoryFilterCounts{
				SoftwareType: []models.FilterCount{
					{Value: "library", Count: 10},
					{Value: "standalone/backend", Count: 20},
					{Value: "addon", Count: 5},
				},
			}, nil
		},
	}
	svc := services.NewRepositoryService(repo)

	groups, err := svc.GetRepositoryFilters(context.Background(), &models.RepositoryFiltersParams{})
	require.NoError(t, err)

	for _, g := range groups {
		if g.Key == "softwareType" {
			require.Len(t, g.Options, 3)
			assert.Equal(t, []string{"Addon / Plugin", "Backend / API", "Library"}, []string{
				g.Options[0].Label,
				g.Options[1].Label,
				g.Options[2].Label,
			})
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

func TestCreateRepository_PreservesManualURLWhenPublicCodeIsProvided(t *testing.T) {
	publicCode := `publiccodeYmlVersion: "0.5.0"
name: Digitale Balie
url: https://git.example.org/upstream/digitale-balie
softwareType: configurationFiles
developmentStatus: stable
platforms:
  - web
description:
  nl:
    shortDescription: Korte beschrijving van de Digitale Balie.
    longDescription: De Digitale Balie maakt dienstverlening persoonlijk met videobellen en ondersteunt gesprekken, verificatie en veilige documentuitwisseling voor burgers en ondernemers binnen gemeentelijke processen.
    features:
      - Videoafspraak
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

	inputURL := "https://git.example.org/custom/digitale-balie"
	isFork := true
	created, err := svc.CreateRepository(context.Background(), models.RepositoryInput{
		Url:             &inputURL,
		OrganisationUri: &org.Uri,
		PublicCodeUrl:   &publicCode,
		IsFork:          &isFork,
	})
	require.NoError(t, err)
	require.NotNil(t, created)
	require.NotNil(t, saved)
	require.NotNil(t, created.PublicCode)
	assert.Equal(t, "https://git.example.org/custom/digitale-balie", saved.Url)
	assert.True(t, saved.IsFork)
	assert.Equal(t, "https://git.example.org/custom/digitale-balie", created.Url)
	assert.Equal(t, models.RepositoryForkTypeTechnicalFork, created.ForkType)
	assert.Equal(t, "https://git.example.org/upstream/digitale-balie", created.PublicCode.Url)
}

func TestPublishAllRepositoriesToTypesense_Disabled(t *testing.T) {
	t.Setenv("ENABLE_TYPESENSE", "false")

	repo := &stubRepo{
		allRepositoriesFunc: func(ctx context.Context) ([]models.Repository, error) {
			t.Fatalf("AllRepositorys should not be called when Typesense is disabled")
			return nil, nil
		},
	}

	service := services.NewRepositoryService(repo)
	err := service.PublishAllRepositoriesToTypesense(context.Background())
	assert.NoError(t, err)
}

func TestPublishAllRepositoriesToTypesense_SendsDocumentsForActiveRepositories(t *testing.T) {
	var calls int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	t.Setenv("TYPESENSE_ENDPOINT", server.URL)
	t.Setenv("TYPESENSE_API_KEY", "secret")
	t.Setenv("TYPESENSE_COLLECTION", "oss-register")
	t.Setenv("ENABLE_TYPESENSE", "true")

	prevClient := httpclient.HTTPClient
	httpclient.HTTPClient = server.Client()
	t.Cleanup(func() { httpclient.HTTPClient = prevClient })

	repo := &stubRepo{
		allRepositoriesFunc: func(ctx context.Context) ([]models.Repository, error) {
			return []models.Repository{
				{Id: "repo-1", Name: "Active repo", Active: true},
				{Id: "repo-2", Name: "Inactive repo", Active: false},
				{Id: "repo-3", Name: "Also active", Active: true},
			}, nil
		},
	}

	service := services.NewRepositoryService(repo)
	err := service.PublishAllRepositoriesToTypesense(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 2, calls)
}
