package util

import (
	"errors"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDonCheckerPublicCodeValidatorAllowsEmptyInput(t *testing.T) {
	validator := donCheckerPublicCodeValidator{}

	require.NoError(t, validator.ValidatePublicCode(""))
	require.NoError(t, validator.ValidatePublicCode("\ufeff  \n\t"))
}

func TestPublicCodeValidationArgs(t *testing.T) {
	assert.Equal(t, []string{
		"validate",
		"--ruleset",
		"publiccode-05",
		"--input",
		"publiccode.yml",
	}, publicCodeValidationArgs("publiccode.yml"))
}

func TestIsExecutableNotFound(t *testing.T) {
	assert.False(t, isExecutableNotFound(nil))
	assert.False(t, isExecutableNotFound(errors.New("plain error")))
	assert.True(t, isExecutableNotFound(&exec.Error{Name: "don-checker", Err: exec.ErrNotFound}))
}

func TestNormalizeRepositoryReferenceVariants(t *testing.T) {
	tests := map[string]string{
		" https://Git.Example.Org:443/Team/Repo.git ":                   "git.example.org/team/repo",
		"http://git.example.org:80/team/repo":                           "git.example.org/team/repo",
		"https://git.example.org:8443/team/repo/blob/main/file.go":      "git.example.org:8443/team/repo",
		"https://git.example.org/group/project/-/tree/main/subfolder":   "git.example.org/group/project",
		"https://git.example.org/group/project/src/branch/main/file.go": "git.example.org/group/project",
		"https://git.example.org":                                       "git.example.org",
		"not a url":                                                     "",
		"":                                                              "",
	}

	for raw, expected := range tests {
		t.Run(raw, func(t *testing.T) {
			assert.Equal(t, expected, normalizeRepositoryReference(raw))
		})
	}
}

func TestTruncateRepositorySegmentsKeepsShortOrTrailingDashPaths(t *testing.T) {
	assert.Equal(t, []string{"team", "repo"}, truncateRepositorySegments([]string{"team", "repo"}))
	assert.Equal(t, []string{"team", "repo", "-"}, truncateRepositorySegments([]string{"team", "repo", "-"}))
}

func TestIsDefaultPort(t *testing.T) {
	assert.True(t, isDefaultPort("https", "443"))
	assert.True(t, isDefaultPort("http", "80"))
	assert.False(t, isDefaultPort("https", "8443"))
	assert.False(t, isDefaultPort("ssh", "22"))
}
