package models_test

import (
	"testing"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListRepositorysParamsRepositoryFiltersCopiesAllFilterValues(t *testing.T) {
	org := "https://example.org/org"
	publicCode := false
	lastActivityAfter := "2024-01-01"
	params := &models.ListRepositorysParams{
		Organisation:       &org,
		Query:              "zaak",
		PublicCode:         &publicCode,
		LastActivityAfter:  &lastActivityAfter,
		SoftwareType:       []string{"library"},
		DevelopmentStatus:  []string{"stable"},
		AvailableLanguages: []string{"nl"},
		MaintenanceType:    []string{"internal"},
		License:            []string{"EUPL-1.2"},
		Platforms:          []string{"web"},
	}

	filters := params.RepositoryFilters()

	require.NotNil(t, filters)
	assert.Equal(t, params.Organisation, filters.Organisation)
	assert.Equal(t, params.Query, filters.Query)
	assert.Equal(t, params.PublicCode, filters.PublicCode)
	assert.Equal(t, params.LastActivityAfter, filters.LastActivityAfter)
	assert.Equal(t, params.SoftwareType, filters.SoftwareType)
	assert.Equal(t, params.DevelopmentStatus, filters.DevelopmentStatus)
	assert.Equal(t, params.AvailableLanguages, filters.AvailableLanguages)
	assert.Equal(t, params.MaintenanceType, filters.MaintenanceType)
	assert.Equal(t, params.License, filters.License)
	assert.Equal(t, params.Platforms, filters.Platforms)

	params.SoftwareType[0] = "changed"
	params.DevelopmentStatus[0] = "changed"
	params.AvailableLanguages[0] = "changed"
	params.MaintenanceType[0] = "changed"
	params.License[0] = "changed"
	params.Platforms[0] = "changed"

	assert.Equal(t, []string{"library"}, filters.SoftwareType)
	assert.Equal(t, []string{"stable"}, filters.DevelopmentStatus)
	assert.Equal(t, []string{"nl"}, filters.AvailableLanguages)
	assert.Equal(t, []string{"internal"}, filters.MaintenanceType)
	assert.Equal(t, []string{"EUPL-1.2"}, filters.License)
	assert.Equal(t, []string{"web"}, filters.Platforms)
}

func TestListRepositorysParamsRepositoryFiltersHandlesNilReceiver(t *testing.T) {
	var params *models.ListRepositorysParams

	filters := params.RepositoryFilters()

	require.NotNil(t, filters)
	assert.Empty(t, filters.Query)
	assert.Nil(t, filters.Organisation)
}
