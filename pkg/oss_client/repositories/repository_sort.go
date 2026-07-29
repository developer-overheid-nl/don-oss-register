package repositories

import (
	"sort"
	"strings"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
)

func sortRepositories(repositories []models.Repository, sorting models.RepositorySort) {
	sort.SliceStable(repositories, func(i, j int) bool {
		primary, leftValid, rightValid := compareRepositorySortField(repositories[i], repositories[j], sorting.Field)
		if leftValid != rightValid {
			return leftValid
		}
		if leftValid && primary != 0 {
			if sorting.Order == models.RepositorySortDescending {
				return primary > 0
			}
			return primary < 0
		}

		return compareRepositoryTieBreakers(repositories[i], repositories[j]) < 0
	})
}

func compareRepositorySortField(left, right models.Repository, field models.RepositorySortField) (comparison int, leftValid, rightValid bool) {
	if field == models.RepositorySortLastActivity {
		leftValid = !left.LastActivityAt.IsZero()
		rightValid = !right.LastActivityAt.IsZero()
		if !leftValid || !rightValid {
			return 0, leftValid, rightValid
		}
		return left.LastActivityAt.Compare(right.LastActivityAt), true, true
	}

	return strings.Compare(strings.ToLower(left.Name), strings.ToLower(right.Name)), true, true
}

func compareRepositoryTieBreakers(left, right models.Repository) int {
	if titleComparison := strings.Compare(strings.ToLower(left.Name), strings.ToLower(right.Name)); titleComparison != 0 {
		return titleComparison
	}
	return strings.Compare(left.Id, right.Id)
}
