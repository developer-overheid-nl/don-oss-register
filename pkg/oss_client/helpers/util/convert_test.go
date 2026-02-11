package util_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/helpers/util"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"github.com/stretchr/testify/assert"
)

func TestApplyRepositoryInputSetsTimestamps(t *testing.T) {
	created := time.Date(2024, 2, 1, 10, 0, 0, 0, time.UTC)
	lastCrawled := time.Date(2024, 3, 1, 10, 0, 0, 0, time.UTC)
	lastActivity := time.Date(2024, 3, 5, 8, 0, 0, 0, time.UTC)

	input := models.RepositoryInput{
		Url:            strPtr("https://example.org/repo"),
		CreatedAt:      created,
		LastCrawledAt:  lastCrawled,
		LastActivityAt: lastActivity,
	}

	repo := util.ApplyRepositoryInput(nil, &input)

	assert.Equal(t, created, repo.CreatedAt)
	assert.Equal(t, lastCrawled, repo.LastCrawledAt)
	assert.Equal(t, lastActivity, repo.LastActivityAt)
}

func TestApplyRepositoryInputKeepsExistingTimestampsWhenZero(t *testing.T) {
	created := time.Date(2023, 6, 1, 12, 0, 0, 0, time.UTC)
	lastCrawled := time.Date(2023, 7, 1, 12, 0, 0, 0, time.UTC)
	lastActivity := time.Date(2023, 7, 15, 9, 0, 0, 0, time.UTC)
	existing := &models.Repository{
		CreatedAt:      created,
		LastCrawledAt:  lastCrawled,
		LastActivityAt: lastActivity,
	}

	repo := util.ApplyRepositoryInput(existing, &models.RepositoryInput{})

	assert.Equal(t, created, repo.CreatedAt)
	assert.Equal(t, lastCrawled, repo.LastCrawledAt)
	assert.Equal(t, lastActivity, repo.LastActivityAt)
}

func TestApplyRepositoryInputParsesLocalizedPublicCodeYAML(t *testing.T) {
	publicCode := `
url: https://example.org/repo
name: Digitale Balie
description:
  nl:
    shortDescription: Korte beschrijving
    longDescription: Lange beschrijving
  en:
    shortDescription: Short description
    longDescription: Long description
`

	repo := util.ApplyRepositoryInput(nil, &models.RepositoryInput{
		PublicCodeUrl: strPtr(publicCode),
	})

	assert.Equal(t, "https://example.org/repo", repo.Url)
	assert.Equal(t, "Digitale Balie", repo.Name)
	assert.Equal(t, "Korte beschrijving", repo.ShortDescription)
	assert.Equal(t, "Lange beschrijving", repo.LongDescription)
}

func TestApplyRepositoryInputParsesFlatDescriptionPublicCodeYAML(t *testing.T) {
	publicCode := `
url: https://example.org/repo
name: Digitale Balie
description:
  shortDescription: Korte beschrijving
  longDescription: Lange beschrijving
`

	repo := util.ApplyRepositoryInput(nil, &models.RepositoryInput{
		PublicCodeUrl: strPtr(publicCode),
	})

	assert.Equal(t, "https://example.org/repo", repo.Url)
	assert.Equal(t, "Digitale Balie", repo.Name)
	assert.Equal(t, "Korte beschrijving", repo.ShortDescription)
	assert.Equal(t, "Lange beschrijving", repo.LongDescription)
}

func TestApplyRepositoryInputParsesStringDescriptionPublicCodeYAML(t *testing.T) {
	publicCode := `
url: https://example.org/repo
name: Digitale Balie
description: Beschrijving van de Digitale Balie
`

	repo := util.ApplyRepositoryInput(nil, &models.RepositoryInput{
		PublicCodeUrl: strPtr(publicCode),
	})

	assert.Equal(t, "https://example.org/repo", repo.Url)
	assert.Equal(t, "Digitale Balie", repo.Name)
	assert.Equal(t, "Beschrijving van de Digitale Balie", repo.ShortDescription)
	assert.Equal(t, "Beschrijving van de Digitale Balie", repo.LongDescription)
}

func TestApplyRepositoryInputParsesRegionalLocaleDescriptionPublicCodeYAML(t *testing.T) {
	publicCode := `
url: https://example.org/repo
name: Digitale Balie
description:
  nl-NL:
    shortDescription: Korte beschrijving NL
    longDescription: Lange beschrijving NL
  en:
    shortDescription: Short description EN
    longDescription: Long description EN
`

	repo := util.ApplyRepositoryInput(nil, &models.RepositoryInput{
		PublicCodeUrl: strPtr(publicCode),
	})

	assert.Equal(t, "Korte beschrijving NL", repo.ShortDescription)
	assert.Equal(t, "Lange beschrijving NL", repo.LongDescription)
}

func TestApplyRepositoryInputFetchesPublicCodeYAMLFromURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`
url: https://example.org/repo
name: Digitale Balie
description:
  shortDescription: Korte beschrijving
  longDescription: Lange beschrijving
`))
	}))
	defer server.Close()

	publicCodeURL := server.URL + "/publiccode.yml"
	repo := util.ApplyRepositoryInput(nil, &models.RepositoryInput{
		PublicCodeUrl: &publicCodeURL,
	})

	assert.Equal(t, publicCodeURL, repo.PublicCodeUrl)
	assert.Equal(t, "https://example.org/repo", repo.Url)
	assert.Equal(t, "Digitale Balie", repo.Name)
	assert.Equal(t, "Korte beschrijving", repo.ShortDescription)
	assert.Equal(t, "Lange beschrijving", repo.LongDescription)
}

func strPtr(val string) *string {
	return &val
}
