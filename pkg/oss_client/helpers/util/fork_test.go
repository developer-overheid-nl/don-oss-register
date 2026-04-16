package util_test

import (
	"testing"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/helpers/util"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"github.com/stretchr/testify/assert"
)

func TestDetectRepositoryForkType(t *testing.T) {
	testCases := map[string]struct {
		repo     models.Repository
		expected models.RepositoryForkType
	}{
		"variant fork when based on another repository": {
			repo: models.Repository{
				Url:             "https://git.example.org/variant/openzaak-brug",
				PublicCode:      &models.PublicCode{Url: "https://git.example.org/variant/openzaak-brug"},
				ForkBasedOnURLs: []string{"https://git.example.org/upstream/openzaak"},
			},
			expected: models.RepositoryForkTypeVariantFork,
		},
		"url mistake when based on itself": {
			repo: models.Repository{
				Url:             "https://git.example.org/team/zds-translator",
				PublicCode:      &models.PublicCode{Url: "https://git.example.org/team/zds-translator"},
				ForkBasedOnURLs: []string{"https://git.example.org/team/zds-translator"},
			},
			expected: models.RepositoryForkTypeURLMistake,
		},
		"technical fork for git fork with upstream publiccode url": {
			repo: models.Repository{
				Url:        "https://git.example.org/fork/don-infra",
				IsFork:     true,
				PublicCode: &models.PublicCode{Url: "https://git.example.org/upstream/don-infra"},
			},
			expected: models.RepositoryForkTypeTechnicalFork,
		},
		"url mistake for manual clone without git fork": {
			repo: models.Repository{
				Url:        "https://github.com/example/manual-clone",
				PublicCode: &models.PublicCode{Url: "https://github.com/example/upstream"},
			},
			expected: models.RepositoryForkTypeURLMistake,
		},
		"git fork without publiccode": {
			repo: models.Repository{
				Url:    "https://git.example.org/fork/git-fork",
				IsFork: true,
			},
			expected: models.RepositoryForkTypeGitFork,
		},
		"git hosting branch urls are normalized": {
			repo: models.Repository{
				Url:        "https://git.example.org/fork/don-infra/tree/digikluster",
				IsFork:     true,
				PublicCode: &models.PublicCode{Url: "https://git.example.org/fork/don-infra"},
			},
			expected: models.RepositoryForkTypeGitFork,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, testCase.expected, util.DetectRepositoryForkType(&testCase.repo))
		})
	}
}
