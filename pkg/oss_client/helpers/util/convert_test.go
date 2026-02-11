package util_test

import (
	"bytes"
	"log"
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

func TestApplyRepositoryInputParsesStandardPublicCodeYAML(t *testing.T) {
	publicCode := validPublicCodeYAML(`
  nl:
    shortDescription: Korte beschrijving van de Digitale Balie.
    longDescription: De Digitale Balie maakt dienstverlening persoonlijk met videobellen en ondersteunt gesprekken, verificatie en veilige documentuitwisseling voor burgers en ondernemers binnen gemeentelijke processen.
    features:
      - Videoafspraak
  en:
    shortDescription: English short description.
    longDescription: The Digital Desk supports personal digital public services through secure video calls and integrated process steps for municipal service delivery workflows while maintaining continuity, accessibility and trust in daily contact moments.
    features:
      - Video appointment
`)

	inputURL := "https://manual.example/repo"
	repo := util.ApplyRepositoryInput(nil, &models.RepositoryInput{
		Url:           &inputURL,
		PublicCodeUrl: strPtr(publicCode),
	})

	assert.Equal(t, "https://manual.example/repo", repo.Url)
	assert.Equal(t, "Digitale Balie", repo.Name)
	assert.Equal(t, "Korte beschrijving van de Digitale Balie.", repo.ShortDescription)
	assert.Equal(t, "De Digitale Balie maakt dienstverlening persoonlijk met videobellen en ondersteunt gesprekken, verificatie en veilige documentuitwisseling voor burgers en ondernemers binnen gemeentelijke processen.", repo.LongDescription)
}

func TestApplyRepositoryInputParsesRegionalLocaleDescription(t *testing.T) {
	publicCode := `publiccodeYmlVersion: "0.5.0"
name: Digitale Balie
url: https://example.org/repo
softwareType: configurationFiles
developmentStatus: stable
platforms:
  - web
description:
  nl-NL:
    shortDescription: Korte beschrijving NL.
    longDescription: Deze regionale Nederlandse beschrijving bevat voldoende tekst om aan de minimale lengte te voldoen en beschrijft helder wat de Digitale Balie voor gemeenten en inwoners betekent.
    features:
      - Videoafspraak
legal:
  license: EUPL-1.2
maintenance:
  type: internal
  contacts:
    - name: Team Digitale Balie
localisation:
  localisationReady: false
  availableLanguages:
    - nl-NL
`

	repo := util.ApplyRepositoryInput(nil, &models.RepositoryInput{
		PublicCodeUrl: strPtr(publicCode),
	})

	assert.Equal(t, "Korte beschrijving NL.", repo.ShortDescription)
	assert.Equal(t, "Deze regionale Nederlandse beschrijving bevat voldoende tekst om aan de minimale lengte te voldoen en beschrijft helder wat de Digitale Balie voor gemeenten en inwoners betekent.", repo.LongDescription)
}

func TestApplyRepositoryInputParsesLegacyVersionWithWarnings(t *testing.T) {
	publicCode := validPublicCodeYAML(`
  nl:
    shortDescription: Korte beschrijving van de Digitale Balie.
    longDescription: De Digitale Balie maakt dienstverlening persoonlijk met videobellen en ondersteunt gesprekken, verificatie en veilige documentuitwisseling voor burgers en ondernemers binnen gemeentelijke processen.
    features:
      - Videoafspraak
`, "0.2")

	inputURL := "https://manual.example/repo"
	repo := util.ApplyRepositoryInput(nil, &models.RepositoryInput{
		Url:           &inputURL,
		PublicCodeUrl: strPtr(publicCode),
	})

	assert.Equal(t, "https://manual.example/repo", repo.Url)
	assert.Equal(t, "Digitale Balie", repo.Name)
	assert.Equal(t, "Korte beschrijving van de Digitale Balie.", repo.ShortDescription)
}

func TestApplyRepositoryInputIgnoresInvalidPublicCodeYAML(t *testing.T) {
	invalid := `
publiccodeYmlVersion: '0.2'
description:
      In maart 2020 startte de gemeente Rotterdam met de ontwikkeling.
  nl:
    shortDescription:
      Korte beschrijving
`

	repo := util.ApplyRepositoryInput(nil, &models.RepositoryInput{
		Url:              strPtr("https://manual.example/repo"),
		Name:             strPtr("Handmatige naam"),
		ShortDescription: strPtr("Handmatige korte omschrijving"),
		PublicCodeUrl:    strPtr(invalid),
	})

	assert.Equal(t, "https://manual.example/repo", repo.Url)
	assert.Equal(t, "Handmatige naam", repo.Name)
	assert.Equal(t, "Handmatige korte omschrijving", repo.ShortDescription)
	assert.Equal(t, "Handmatige korte omschrijving", repo.LongDescription)
}

func TestApplyRepositoryInputLogsPublicCodeParseError(t *testing.T) {
	invalid := `
publiccodeYmlVersion: '0.2'
description:
      In maart 2020 startte de gemeente Rotterdam met de ontwikkeling.
  nl:
    shortDescription:
      Korte beschrijving
`

	var buf bytes.Buffer
	originalWriter := log.Writer()
	originalFlags := log.Flags()
	originalPrefix := log.Prefix()
	log.SetOutput(&buf)
	log.SetFlags(0)
	log.SetPrefix("")
	defer func() {
		log.SetOutput(originalWriter)
		log.SetFlags(originalFlags)
		log.SetPrefix(originalPrefix)
	}()

	repo := util.ApplyRepositoryInput(nil, &models.RepositoryInput{
		Url:              strPtr("https://manual.example/repo"),
		Name:             strPtr("Handmatige naam"),
		ShortDescription: strPtr("Handmatige korte omschrijving"),
		PublicCodeUrl:    strPtr(invalid),
	})

	assert.Equal(t, "https://manual.example/repo", repo.Url)
	assert.Equal(t, "Handmatige naam", repo.Name)
	assert.Equal(t, "Handmatige korte omschrijving", repo.ShortDescription)
	assert.Equal(t, "Handmatige korte omschrijving", repo.LongDescription)
	assert.Contains(t, buf.String(), "publiccode parse validation failed:")
	assert.Contains(t, buf.String(), "yaml: mapping values are not allowed in this context")

	t.Logf("captured parse log:\n%s", buf.String())
}

func TestApplyRepositoryInputFetchesPublicCodeYAMLFromURL(t *testing.T) {
	publicCode := validPublicCodeYAML(`
  nl:
    shortDescription: Korte beschrijving van de Digitale Balie.
    longDescription: De Digitale Balie maakt dienstverlening persoonlijk met videobellen en ondersteunt gesprekken, verificatie en veilige documentuitwisseling voor burgers en ondernemers binnen gemeentelijke processen.
    features:
      - Videoafspraak
`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(publicCode))
	}))
	defer server.Close()

	publicCodeURL := server.URL + "/publiccode.yml"
	inputURL := "https://manual.example/repo"
	repo := util.ApplyRepositoryInput(nil, &models.RepositoryInput{
		Url:           &inputURL,
		PublicCodeUrl: &publicCodeURL,
	})

	assert.Equal(t, publicCodeURL, repo.PublicCodeUrl)
	assert.Equal(t, "https://manual.example/repo", repo.Url)
	assert.Equal(t, "Digitale Balie", repo.Name)
	assert.Equal(t, "Korte beschrijving van de Digitale Balie.", repo.ShortDescription)
}

func validPublicCodeYAML(descriptionBlock string, version ...string) string {
	parsedVersion := "0.5.0"
	if len(version) > 0 && version[0] != "" {
		parsedVersion = version[0]
	}

	return `publiccodeYmlVersion: "` + parsedVersion + `"
name: Digitale Balie
url: https://example.org/repo
softwareType: configurationFiles
developmentStatus: stable
platforms:
  - web
description:
` + descriptionBlock + `
legal:
  license: EUPL-1.2
maintenance:
  type: internal
  contacts:
    - name: Team Digitale Balie
localisation:
  localisationReady: false
  availableLanguages:
    - nl
`
}

func strPtr(val string) *string {
	return &val
}
