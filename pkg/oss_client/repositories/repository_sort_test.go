package repositories

import (
	"testing"
	"time"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"github.com/stretchr/testify/assert"
)

func TestSortRepositoriesOrdersTitlesCaseInsensitively(t *testing.T) {
	tests := []struct {
		name  string
		order models.RepositorySortOrder
		want  []string
	}{
		{
			name:  "ascending",
			order: models.RepositorySortAscending,
			want:  []string{"alpha-lower", "alpha-upper", "beta", "zulu"},
		},
		{
			name:  "descending",
			order: models.RepositorySortDescending,
			want:  []string{"zulu", "beta", "alpha-lower", "alpha-upper"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repositories := []models.Repository{
				{Id: "zulu", Name: "Zulu"},
				{Id: "alpha-upper", Name: "Alpha"},
				{Id: "beta", Name: "beta"},
				{Id: "alpha-lower", Name: "alpha"},
			}

			sortRepositories(repositories, models.RepositorySort{Field: models.RepositorySortTitle, Order: tt.order})

			assert.Equal(t, tt.want, repositoryIDs(repositories))
		})
	}
}

func TestSortRepositoriesOrdersLastActivityWithMissingLast(t *testing.T) {
	oldActivity := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	newActivity := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name  string
		order models.RepositorySortOrder
		want  []string
	}{
		{
			name:  "ascending",
			order: models.RepositorySortAscending,
			want:  []string{"old", "new", "missing"},
		},
		{
			name:  "descending",
			order: models.RepositorySortDescending,
			want:  []string{"new", "old", "missing"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repositories := []models.Repository{
				{Id: "missing", Name: "Aardvark"},
				{Id: "old", Name: "Old", LastActivityAt: oldActivity},
				{Id: "new", Name: "New", LastActivityAt: newActivity},
			}

			sortRepositories(repositories, models.RepositorySort{Field: models.RepositorySortLastActivity, Order: tt.order})

			assert.Equal(t, tt.want, repositoryIDs(repositories))
		})
	}
}

func TestSortRepositoriesUsesTitleThenIDAsTieBreakers(t *testing.T) {
	activity := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	repositories := []models.Repository{
		{Id: "same-b", Name: "same", LastActivityAt: activity},
		{Id: "zulu", Name: "Zulu", LastActivityAt: activity},
		{Id: "same-a", Name: "Same", LastActivityAt: activity},
	}

	sortRepositories(repositories, models.RepositorySort{Field: models.RepositorySortLastActivity, Order: models.RepositorySortDescending})

	assert.Equal(t, []string{"same-a", "same-b", "zulu"}, repositoryIDs(repositories))
}

func repositoryIDs(repositories []models.Repository) []string {
	ids := make([]string, len(repositories))
	for i := range repositories {
		ids[i] = repositories[i].Id
	}
	return ids
}
