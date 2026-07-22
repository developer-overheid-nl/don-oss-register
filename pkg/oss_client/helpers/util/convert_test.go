package util_test

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/helpers/util"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestRepositoryConversionsIncludeOrganisationAndForkType(t *testing.T) {
	created := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	lastCrawled := time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC)
	lastActivity := time.Date(2024, 1, 3, 10, 0, 0, 0, time.UTC)
	repo := &models.Repository{
		Id:               "repo-1",
		Url:              "https://github.com/example/fork",
		Name:             "Repo",
		ShortDescription: "Short",
		LongDescription:  "Long",
		PublicCodeUrl:    "https://example.org/publiccode.yml",
		CreatedAt:        created,
		LastCrawledAt:    lastCrawled,
		LastActivityAt:   lastActivity,
		Active:           true,
		IsFork:           true,
		Archived:         true,
		Organisation:     &models.Organisation{Uri: "https://example.org/org", Label: "Example"},
		PublicCode:       &models.PublicCode{Name: "Repo"},
	}

	summary := util.ToRepositorySummary(repo)
	assert.Equal(t, "repo-1", summary.Id)
	assert.Equal(t, "Repo", summary.Name)
	assert.Equal(t, "Short", summary.ShortDescription)
	assert.Equal(t, "https://example.org/publiccode.yml", summary.PublicCodeUrl)
	assert.Equal(t, created, summary.CreatedAt)
	assert.Equal(t, lastCrawled, summary.LastCrawledAt)
	assert.Equal(t, lastActivity, summary.LastActivityAt)
	assert.True(t, summary.Archived)
	require.NotNil(t, summary.Organisation)
	assert.Equal(t, "Example", summary.Organisation.Label)
	assert.Equal(t, models.RepositoryForkTypeGitFork, summary.ForkType)

	detail := util.ToRepositoryDetail(repo)
	assert.Equal(t, summary, detail.RepositorySummary)
	assert.Equal(t, "Long", detail.LongDescription)
	assert.Equal(t, repo.PublicCode, detail.PublicCode)
}

func TestToGitOrganisatieSummary(t *testing.T) {
	org := &models.Organisation{Uri: "https://example.org/org", Label: "Example"}
	summary := util.ToGitOrganisatieSummary(&models.GitOrganisatie{
		Id:           "git-1",
		Url:          "https://github.com/example",
		Organisation: org,
	})

	assert.Equal(t, "git-1", summary.Id)
	assert.Equal(t, "https://github.com/example", summary.Url)
	assert.Equal(t, org, summary.Organisation)
}

func TestSetPaginationHeadersBuildsHTTPSLinkHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://api.example.org/v1/repositories?q=test&page=1", nil)
	req.Header.Set("Forwarded-Proto", "https")
	previous := 1
	next := 3
	headers := http.Header{}

	util.SetPaginationHeaders(req, headers.Set, models.Pagination{
		Previous:       &previous,
		Next:           &next,
		CurrentPage:    2,
		RecordsPerPage: 25,
		TotalPages:     4,
		TotalRecords:   88,
	})

	assert.Equal(t, "88", headers.Get("Total-Count"))
	assert.Equal(t, "4", headers.Get("Total-Pages"))
	assert.Equal(t, "25", headers.Get("Per-Page"))
	assert.Equal(t, "2", headers.Get("Current-Page"))
	link := headers.Get("Link")
	assert.Contains(t, link, `<https://api.example.org/v1/repositories?page=1&perPage=25&q=test>; rel="first"`)
	assert.Contains(t, link, `rel="prev"`)
	assert.Contains(t, link, `rel="self"`)
	assert.Contains(t, link, `rel="next"`)
	assert.Contains(t, link, `<https://api.example.org/v1/repositories?page=4&perPage=25&q=test>; rel="last"`)
}

func TestSetPaginationHeadersOmitsLinksWhenThereAreNoPages(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://api.example.org/v1/repositories", nil)
	headers := http.Header{}

	util.SetPaginationHeaders(req, headers.Set, models.Pagination{})

	assert.Equal(t, "0", headers.Get("Total-Count"))
	assert.Empty(t, headers.Get("Link"))
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
	disablePublicCodeValidation(t)

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
	require.NotNil(t, repo.PublicCode)
	assert.Equal(t, "0.5.0", repo.PublicCode.PubliccodeYmlVersion)
	assert.Equal(t, "Digitale Balie", repo.PublicCode.Name)
	assert.Equal(t, "https://example.org/repo", repo.PublicCode.Url)
	assert.Equal(t, "https://service.example.org/digitale-balie", repo.PublicCode.LandingUrl)
	assert.Equal(t, []string{"web"}, repo.PublicCode.Platforms)
	assert.Equal(t, "stable", repo.PublicCode.DevelopmentStatus)
	assert.Equal(t, "configurationFiles", repo.PublicCode.SoftwareType)
	require.NotNil(t, repo.PublicCode.Legal)
	assert.Equal(t, "EUPL-1.2", repo.PublicCode.Legal.License)
	require.NotNil(t, repo.PublicCode.Maintenance)
	assert.Equal(t, "internal", repo.PublicCode.Maintenance.Type)
	require.Len(t, repo.PublicCode.Maintenance.Contacts, 1)
	assert.Equal(t, "Team Digitale Balie", repo.PublicCode.Maintenance.Contacts[0].Name)
	require.NotNil(t, repo.PublicCode.Localisation)
	require.NotNil(t, repo.PublicCode.Localisation.LocalisationReady)
	assert.False(t, *repo.PublicCode.Localisation.LocalisationReady)
	assert.Equal(t, []string{"nl"}, repo.PublicCode.Localisation.AvailableLanguages)
	require.Contains(t, repo.PublicCode.Description, "nl")
	assert.Equal(t, "Korte beschrijving van de Digitale Balie.", repo.PublicCode.Description["nl"].ShortDescription)
	assert.NotEmpty(t, repo.PublicCode.Description["nl"].LongDescription)
	assert.Equal(t, []string{"Videoafspraak"}, repo.PublicCode.Description["nl"].Features)
}

func TestApplyRepositoryInputParsesRegionalLocaleDescription(t *testing.T) {
	disablePublicCodeValidation(t)

	publicCode := `publiccodeYmlVersion: "0.5.0"
name: Digitale Balie
url: https://example.org/repo
landingURL: https://service.example.org/digitale-balie
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

func TestApplyRepositoryInputSetsExplicitForkFlag(t *testing.T) {
	inputURL := "https://git.example.org/custom/frontend"
	isFork := true
	repo := util.ApplyRepositoryInput(nil, &models.RepositoryInput{
		Url:    &inputURL,
		IsFork: &isFork,
	})

	assert.Equal(t, "https://git.example.org/custom/frontend", repo.Url)
	assert.True(t, repo.IsFork)
}

func TestApplyRepositoryInputSetsExplicitArchivedFlag(t *testing.T) {
	inputURL := "https://git.example.org/custom/frontend"
	archived := true
	repo := util.ApplyRepositoryInput(nil, &models.RepositoryInput{
		Url:      &inputURL,
		Archived: &archived,
	})

	assert.Equal(t, "https://git.example.org/custom/frontend", repo.Url)
	assert.True(t, repo.Archived)
}

func TestApplyRepositoryInputStoresBasedOnURLsFromPublicCode(t *testing.T) {
	disablePublicCodeValidation(t)

	inputURL := "https://git.example.org/variant/openzaak-brug"
	publicCode := `publiccodeYmlVersion: "0.5.0"
name: OpenZaakBrug
url: https://git.example.org/variant/openzaak-brug
isBasedOn: https://git.example.org/upstream/openzaak
softwareType: configurationFiles
developmentStatus: stable
platforms:
  - web
description:
  nl:
    shortDescription: Brug voor OpenZaak integraties.
    longDescription: Deze variant van OpenZaak bevat lokale aanpassingen voor de gemeente en legt expliciet vast op welke upstream repository de codebasis gebaseerd is voor beheer en classificatie in het register.
legal:
  license: EUPL-1.2
maintenance:
  type: internal
  contacts:
    - name: Team OpenZaakBrug
localisation:
  localisationReady: true
  availableLanguages:
    - nl
`

	repo := util.ApplyRepositoryInput(nil, &models.RepositoryInput{
		Url:           &inputURL,
		PublicCodeUrl: &publicCode,
	})

	assert.Equal(t, []string{"https://git.example.org/upstream/openzaak"}, repo.ForkBasedOnURLs)
	assert.Equal(t, models.RepositoryForkTypeVariantFork, util.DetectRepositoryForkType(repo))
}

func TestApplyRepositoryInputSelectsDescriptionUsingAvailableLanguages(t *testing.T) {
	disablePublicCodeValidation(t)

	publicCode := `publiccodeYmlVersion: "0.5.0"
name: Service Guichet
url: https://example.org/repo
softwareType: configurationFiles
developmentStatus: stable
platforms:
  - web
description:
  en:
    shortDescription: English short description.
    longDescription: This English long description is present but should not be selected when French is the preferred available language in localisation settings.
    features:
      - English feature
  fr:
    shortDescription: Description courte en francais.
    longDescription: Cette description longue francaise doit etre selectionnee car la langue preferee indiquee dans localisation est le francais.
    features:
      - Fonctionnalite francaise
legal:
  license: EUPL-1.2
maintenance:
  type: internal
  contacts:
    - name: Team Service Guichet
localisation:
  localisationReady: false
  availableLanguages:
    - fr
`

	repo := util.ApplyRepositoryInput(nil, &models.RepositoryInput{
		PublicCodeUrl: strPtr(publicCode),
	})

	assert.Equal(t, "Description courte en francais.", repo.ShortDescription)
	assert.Equal(t, "Cette description longue francaise doit etre selectionnee car la langue preferee indiquee dans localisation est le francais.", repo.LongDescription)
}

func TestApplyRepositoryInputParsesContractMaintenanceAndSpecialMandatoryFields(t *testing.T) {
	disablePublicCodeValidation(t)

	publicCode := `publiccodeYmlVersion: "0.5.0"
name: Zaaksysteem
url: https://example.org/repo
softwareType: standalone/web
developmentStatus: beta
platforms:
  - web
description:
  nl:
    shortDescription: Zaaksysteem voor publieke dienstverlening.
    longDescription: Dit zaaksysteem ondersteunt het volledige proces van publieke dienstverlening, van intake en registratie tot afhandeling en rapportage, en biedt robuuste integraties met bestaande basisregistraties en veilige gegevensuitwisseling.
    features:
      - Zaakafhandeling
legal:
  license: EUPL-1.2
maintenance:
  type: contract
  contractors:
    - name: Acme Maintenance BV
      until: "2027-12-31"
organisation:
  uri: https://example.org/organisations/zaaksysteem
fundedBy:
  - name: Ministerie van Binnenlandse Zaken
dependsOn:
  open:
    - name: PostgreSQL
      versionMin: "14"
  hardware:
    - name: Smartcard Reader
localisation:
  localisationReady: true
  availableLanguages:
    - nl
`

	repo := util.ApplyRepositoryInput(nil, &models.RepositoryInput{
		PublicCodeUrl: strPtr(publicCode),
	})

	require.NotNil(t, repo.PublicCode)
	require.NotNil(t, repo.PublicCode.Maintenance)
	assert.Equal(t, "contract", repo.PublicCode.Maintenance.Type)
	require.Len(t, repo.PublicCode.Maintenance.Contractors, 1)
	assert.Equal(t, "Acme Maintenance BV", repo.PublicCode.Maintenance.Contractors[0].Name)
	assert.Equal(t, "2027-12-31", repo.PublicCode.Maintenance.Contractors[0].Until)
	require.NotNil(t, repo.PublicCode.Organisation)
	assert.Equal(t, "https://example.org/organisations/zaaksysteem", repo.PublicCode.Organisation.Uri)
	require.NotNil(t, repo.PublicCode.DependsOn)
	assert.Equal(t, []models.PublicCodeDependency{{Name: "PostgreSQL"}}, repo.PublicCode.DependsOn.Open)
	assert.Equal(t, []models.PublicCodeDependency{{Name: "Smartcard Reader"}}, repo.PublicCode.DependsOn.Hardware)
	assert.Equal(t, []models.PublicCodeOrganisationReference{{Name: "Ministerie van Binnenlandse Zaken"}}, repo.PublicCode.FundedBy)
}

func TestApplyRepositoryInputParsesLegacyVersionWithWarnings(t *testing.T) {
	disablePublicCodeValidation(t)

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
	disablePublicCodeValidation(t)

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

func TestApplyRepositoryInputUsesPublicCodeNameWhenValidationFails(t *testing.T) {
	validator := &fakePublicCodeValidator{
		err: errors.New("publiccode validation failed"),
	}
	util.SetPublicCodeValidatorForTest(t, validator)

	publicCode := validPublicCodeYAML(`
  nl:
    shortDescription: Korte beschrijving van de Digitale Balie.
    longDescription: De Digitale Balie maakt dienstverlening persoonlijk met videobellen en ondersteunt gesprekken, verificatie en veilige documentuitwisseling voor burgers en ondernemers binnen gemeentelijke processen.
    features:
      - Videoafspraak
`)

	repo := util.ApplyRepositoryInput(nil, &models.RepositoryInput{
		Url:           strPtr("https://manual.example/repo"),
		PublicCodeUrl: strPtr(publicCode),
	})

	assert.Equal(t, "Digitale Balie", repo.Name)
	assert.Empty(t, repo.ShortDescription)
	assert.Nil(t, repo.PublicCode)
	assert.Nil(t, repo.ForkBasedOnURLs)
	assert.Equal(t, 1, validator.calls)
}

func TestApplyRepositoryInputKeepsManualNameWhenValidationFails(t *testing.T) {
	util.SetPublicCodeValidatorForTest(t, &fakePublicCodeValidator{
		err: errors.New("publiccode validation failed"),
	})

	repoURL := "https://github.com/OpenWebconcept/plugin-accessible-docs.git"
	manualName := "Handmatige titel"
	publicCode := validPublicCodeYAML(`
  nl:
    shortDescription: Korte beschrijving van de Digitale Balie.
    longDescription: De Digitale Balie maakt dienstverlening persoonlijk met videobellen en ondersteunt gesprekken, verificatie en veilige documentuitwisseling voor burgers en ondernemers binnen gemeentelijke processen.
    features:
      - Videoafspraak
`)
	repo := util.ApplyRepositoryInput(nil, &models.RepositoryInput{
		Url:           &repoURL,
		Name:          &manualName,
		PublicCodeUrl: &publicCode,
	})

	assert.Equal(t, "Handmatige titel", repo.Name)
	assert.Nil(t, repo.PublicCode)
}

func TestApplyRepositoryInputUsesRepositoryURLSlugWhenValidationFailsWithoutName(t *testing.T) {
	util.SetPublicCodeValidatorForTest(t, &fakePublicCodeValidator{
		err: errors.New("publiccode validation failed"),
	})

	repoURL := "https://github.com/OpenWebconcept/plugin-accessible-docs.git"
	publicCodeURL := "publiccodeYmlVersion: '0.2'\ndescription:\n\tinvalid"
	repo := util.ApplyRepositoryInput(nil, &models.RepositoryInput{
		Url:           &repoURL,
		PublicCodeUrl: &publicCodeURL,
	})

	assert.Equal(t, "plugin-accessible-docs", repo.Name)
	assert.Equal(t, publicCodeURL, repo.PublicCodeUrl)
	assert.Nil(t, repo.PublicCode)
}

func TestApplyRepositoryInputLogsCompactValidationFailure(t *testing.T) {
	var logs bytes.Buffer
	previousOutput := log.Writer()
	previousFlags := log.Flags()
	log.SetOutput(&logs)
	log.SetFlags(0)
	t.Cleanup(func() {
		log.SetOutput(previousOutput)
		log.SetFlags(previousFlags)
	})

	util.SetPublicCodeValidatorForTest(t, &fakePublicCodeValidator{
		err: errors.New(`don-checker publiccode validation failed: exit status 1: Ruleset: publiccode-05
Applied rulesets: https://yml.publiccode.tools/schema/0.5
Diagnostics: 27 (errors 26, warnings 1, info 0, hints 0)

Errors (26)
  1. parser
     message: Invalid symbol
     path: []
     source: https://yml.publiccode.tools/schema/0.5
  2. parser
     message: Invalid symbol
     path: []
     source: https://yml.publiccode.tools/schema/0.5`),
	})

	util.ApplyRepositoryInput(nil, &models.RepositoryInput{
		PublicCodeUrl: strPtr(validPublicCodeYAML(`
  nl:
    shortDescription: Korte beschrijving.
    longDescription: Lange beschrijving.
`)),
	})

	logged := logs.String()
	assert.Contains(t, logged, "publiccode validation failed: Diagnostics: 27 (errors 26, warnings 1, info 0, hints 0); message: Invalid symbol")
	assert.NotContains(t, logged, "Errors (26)")
	assert.NotContains(t, logged, "source: https://yml.publiccode.tools/schema/0.5")
}

func TestApplyRepositoryInputFetchesPublicCodeYAMLFromURL(t *testing.T) {
	disablePublicCodeValidation(t)

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
landingURL: https://service.example.org/digitale-balie
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

func disablePublicCodeValidation(t *testing.T) {
	t.Helper()
	util.SetPublicCodeValidatorForTest(t, &fakePublicCodeValidator{})
}

type fakePublicCodeValidator struct {
	err   error
	calls int
}

func (v *fakePublicCodeValidator) ValidatePublicCode(string) error {
	v.calls++
	return v.err
}
