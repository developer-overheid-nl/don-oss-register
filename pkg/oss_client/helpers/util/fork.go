package util

import (
	"net/url"
	"strings"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
)

func DetectRepositoryForkType(repo *models.Repository) models.RepositoryForkType {
	if repo == nil {
		return ""
	}

	publicCodeURL := ""
	if repo.PublicCode != nil {
		publicCodeURL = repo.PublicCode.Url
	}

	return classifyRepositoryForkType(repo.Url, repo.IsFork, publicCodeURL, repo.ForkBasedOnURLs)
}

func classifyRepositoryForkType(repoURL string, isFork bool, publicCodeURL string, basedOnURLs []string) models.RepositoryForkType {
	current := normalizeRepositoryReference(repoURL)

	hasBasedOn := false
	hasDifferentBasedOn := false
	hasSameBasedOn := false
	for _, candidate := range basedOnURLs {
		normalized := normalizeRepositoryReference(candidate)
		if normalized == "" {
			continue
		}

		hasBasedOn = true
		if normalized == current {
			hasSameBasedOn = true
			continue
		}

		hasDifferentBasedOn = true
	}

	if hasDifferentBasedOn {
		return models.RepositoryForkTypeVariantFork
	}
	if hasBasedOn && hasSameBasedOn {
		return models.RepositoryForkTypeURLMistake
	}

	normalizedPublicCodeURL := normalizeRepositoryReference(publicCodeURL)
	if isFork {
		if normalizedPublicCodeURL == "" || normalizedPublicCodeURL == current {
			return models.RepositoryForkTypeGitFork
		}
		return models.RepositoryForkTypeTechnicalFork
	}

	if normalizedPublicCodeURL != "" && normalizedPublicCodeURL != current {
		return models.RepositoryForkTypeURLMistake
	}

	return ""
}

func normalizeRepositoryReference(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Host == "" {
		return ""
	}

	host := strings.ToLower(parsed.Hostname())
	if port := parsed.Port(); port != "" && !isDefaultPort(parsed.Scheme, port) {
		host += ":" + port
	}

	pathSegments := splitPathSegments(parsed.Path)
	pathSegments = truncateRepositorySegments(pathSegments)
	if len(pathSegments) == 0 {
		return host
	}

	pathSegments[len(pathSegments)-1] = strings.TrimSuffix(pathSegments[len(pathSegments)-1], ".git")
	for i := range pathSegments {
		pathSegments[i] = strings.ToLower(pathSegments[i])
	}

	return host + "/" + strings.Join(pathSegments, "/")
}

func splitPathSegments(path string) []string {
	rawSegments := strings.Split(path, "/")
	segments := make([]string, 0, len(rawSegments))
	for _, segment := range rawSegments {
		trimmed := strings.TrimSpace(segment)
		if trimmed == "" {
			continue
		}
		segments = append(segments, trimmed)
	}
	return segments
}

func truncateRepositorySegments(segments []string) []string {
	trimAt := len(segments)
	for i := 0; i < len(segments); i++ {
		if i < 2 {
			continue
		}

		segment := strings.ToLower(segments[i])
		switch segment {
		case "tree", "blob", "raw", "src", "archive":
			trimAt = i
		case "-":
			if i+1 >= len(segments) {
				continue
			}
			switch strings.ToLower(segments[i+1]) {
			case "tree", "blob", "raw", "archive":
				trimAt = i
			}
		}
		if trimAt != len(segments) {
			break
		}
	}

	return segments[:trimAt]
}

func isDefaultPort(scheme, port string) bool {
	return (scheme == "https" && port == "443") || (scheme == "http" && port == "80")
}
