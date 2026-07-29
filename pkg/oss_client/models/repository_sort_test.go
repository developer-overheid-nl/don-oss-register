package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRepositorySortDefaultsToTitleAscending(t *testing.T) {
	got, err := ParseRepositorySort(nil, nil)

	require.NoError(t, err)
	assert.Equal(t, RepositorySort{Field: RepositorySortTitle, Order: RepositorySortAscending}, got)
}

func TestParseRepositorySortAppliesDefaultsIndependently(t *testing.T) {
	tests := []struct {
		name      string
		sortBy    string
		sortOrder string
		want      RepositorySort
	}{
		{
			name:      "default field",
			sortOrder: "desc",
			want:      RepositorySort{Field: RepositorySortTitle, Order: RepositorySortDescending},
		},
		{
			name:   "default order",
			sortBy: "lastActivity",
			want:   RepositorySort{Field: RepositorySortLastActivity, Order: RepositorySortAscending},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sortBy, sortOrder *string
			if tt.sortBy != "" {
				sortBy = &tt.sortBy
			}
			if tt.sortOrder != "" {
				sortOrder = &tt.sortOrder
			}
			got, err := ParseRepositorySort(sortBy, sortOrder)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseRepositorySortAcceptsSupportedValues(t *testing.T) {
	tests := []struct {
		name      string
		sortBy    string
		sortOrder string
		want      RepositorySort
	}{
		{
			name:      "title descending",
			sortBy:    "title",
			sortOrder: "desc",
			want:      RepositorySort{Field: RepositorySortTitle, Order: RepositorySortDescending},
		},
		{
			name:      "last activity ascending",
			sortBy:    "lastActivity",
			sortOrder: "asc",
			want:      RepositorySort{Field: RepositorySortLastActivity, Order: RepositorySortAscending},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRepositorySort(&tt.sortBy, &tt.sortOrder)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseRepositorySortRejectsUnsupportedSortBy(t *testing.T) {
	sortBy, sortOrder := "lastCrawled", "asc"
	_, err := ParseRepositorySort(&sortBy, &sortOrder)

	var invalid InvalidRepositorySortError
	require.ErrorAs(t, err, &invalid)
	assert.Equal(t, "sortBy", invalid.Parameter)
	assert.Equal(t, "lastCrawled", invalid.Value)
}

func TestParseRepositorySortRejectsUnsupportedSortOrder(t *testing.T) {
	sortBy, sortOrder := "title", "sideways"
	_, err := ParseRepositorySort(&sortBy, &sortOrder)

	var invalid InvalidRepositorySortError
	require.ErrorAs(t, err, &invalid)
	assert.Equal(t, "sortOrder", invalid.Parameter)
	assert.Equal(t, "sideways", invalid.Value)
}
