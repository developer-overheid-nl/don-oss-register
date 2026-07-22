package typesense

import (
	"testing"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"github.com/stretchr/testify/assert"
)

func TestEnabledReflectsTypesenseEnvironment(t *testing.T) {
	t.Setenv("TYPESENSE_ENDPOINT", "")
	t.Setenv("TYPESENSE_API_KEY", "")
	assert.False(t, Enabled())

	t.Setenv("TYPESENSE_ENDPOINT", "https://typesense.example.org")
	t.Setenv("TYPESENSE_API_KEY", "secret")
	assert.True(t, Enabled())
}

func TestRepositoryTitleFallsBackToPublicCodeNameAndURL(t *testing.T) {
	assert.Equal(t, "Public Code Name", repositoryTitle(&models.Repository{
		PublicCode: &models.PublicCode{Name: " Public Code Name "},
		Url:        "https://example.org/repo",
	}))
	assert.Equal(t, "https://example.org/repo", repositoryTitle(&models.Repository{
		Url: " https://example.org/repo ",
	}))
}

func TestRepositoryOrganisationLabelFallbacks(t *testing.T) {
	assert.Equal(t, "https://example.org/org", repositoryOrganisationLabel(&models.Repository{
		Organisation: &models.Organisation{Uri: " https://example.org/org "},
	}))

	orgID := " https://example.org/org-id "
	assert.Equal(t, "https://example.org/org-id", repositoryOrganisationLabel(&models.Repository{
		OrganisationID: &orgID,
	}))

	assert.Empty(t, repositoryOrganisationLabel(&models.Repository{}))
}

func TestLabelAndForkTypeFallbacks(t *testing.T) {
	assert.Empty(t, labelWithCode(" ", models.SoftwareTypeLabels))
	assert.Equal(t, "custom", labelWithCode(" custom ", models.SoftwareTypeLabels))
	assert.Equal(t, "Technische fork", forkTypeLabel(models.RepositoryForkTypeTechnicalFork))
	assert.Equal(t, "Variant fork", forkTypeLabel(models.RepositoryForkTypeVariantFork))
	assert.Equal(t, "URL komt niet overeen met publiccode.yml", forkTypeLabel(models.RepositoryForkTypeURLMistake))
	assert.Equal(t, "UNKNOWN", forkTypeLabel(models.RepositoryForkType("UNKNOWN")))
}
