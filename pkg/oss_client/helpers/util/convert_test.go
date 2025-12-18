package util_test

import (
	"testing"
	"time"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/helpers/util"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"github.com/stretchr/testify/assert"
)

func TestApplyRepositoryInputSetsTimestamps(t *testing.T) {
	created := time.Date(2024, 2, 1, 10, 0, 0, 0, time.UTC)
	updated := time.Date(2024, 3, 1, 10, 0, 0, 0, time.UTC)
	lastActivity := time.Date(2024, 3, 5, 8, 0, 0, 0, time.UTC)

	input := models.RepositoryInput{
		Url:          strPtr("https://example.org/repo"),
		CreatedAt:    created,
		UpdatedAt:    updated,
		LastActivity: lastActivity,
	}

	repo := util.ApplyRepositoryInput(nil, &input)

	assert.Equal(t, created, repo.CreatedAt)
	assert.Equal(t, updated, repo.UpdatedAt)
	assert.Equal(t, lastActivity, repo.LastActivity)
}

func TestApplyRepositoryInputKeepsExistingTimestampsWhenZero(t *testing.T) {
	created := time.Date(2023, 6, 1, 12, 0, 0, 0, time.UTC)
	updated := time.Date(2023, 7, 1, 12, 0, 0, 0, time.UTC)
	lastActivity := time.Date(2023, 7, 15, 9, 0, 0, 0, time.UTC)
	existing := &models.Repository{
		CreatedAt:    created,
		UpdatedAt:    updated,
		LastActivity: lastActivity,
	}

	repo := util.ApplyRepositoryInput(existing, &models.RepositoryInput{})

	assert.Equal(t, created, repo.CreatedAt)
	assert.Equal(t, updated, repo.UpdatedAt)
	assert.Equal(t, lastActivity, repo.LastActivity)
}

func strPtr(val string) *string {
	return &val
}
