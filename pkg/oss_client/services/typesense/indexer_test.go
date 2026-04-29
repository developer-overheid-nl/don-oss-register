package typesense_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	httpclient "github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/helpers/httpclient"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/services/typesense"
)

func TestPublishRepository_Disabled(t *testing.T) {
	t.Setenv("TYPESENSE_ENDPOINT", "")
	t.Setenv("TYPESENSE_API_KEY", "")
	t.Setenv("TYPESENSE_COLLECTION", "")

	err := typesense.PublishRepository(context.Background(), &models.Repository{Id: "repo-1"})
	if !errors.Is(err, typesense.ErrDisabled) {
		t.Fatalf("expected ErrDisabled, got %v", err)
	}
}

func TestPublishRepository_SendsDocument(t *testing.T) {
	var capturedBody []byte
	var capturedPath, capturedAction, capturedKey string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		capturedAction = r.URL.Query().Get("action")
		capturedKey = r.Header.Get("X-TYPESENSE-API-KEY")
		defer func() {
			if err := r.Body.Close(); err != nil {
				t.Errorf("failed to close request body: %v", err)
			}
		}()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		capturedBody = body
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	t.Setenv("TYPESENSE_ENDPOINT", server.URL)
	t.Setenv("TYPESENSE_API_KEY", "secret")
	t.Setenv("TYPESENSE_COLLECTION", "oss-register")
	t.Setenv("TYPESENSE_DETAIL_BASE_URL", "https://oss.developer.overheid.nl/repositories")
	t.Setenv("TYPESENSE_LANGUAGE", "nl")
	t.Setenv("TYPESENSE_ITEM_PRIORITY", "5")
	t.Setenv("TYPESENSE_DEFAULT_TAGS", "oss-register,repository")

	prevClient := httpclient.HTTPClient
	httpclient.HTTPClient = server.Client()
	t.Cleanup(func() {
		httpclient.HTTPClient = prevClient
	})

	localisationReady := true
	repository := &models.Repository{
		Id:               "repo-1",
		Name:             "Mijn Repository",
		Url:              "https://github.com/example/my-repo",
		ShortDescription: "Korte beschrijving",
		LongDescription:  "Lange beschrijving voor de zoekindex.",
		Organisation: &models.Organisation{
			Label: "Ministerie van Test",
			Uri:   "https://organisaties.example.com/min-test",
		},
		PublicCodeUrl: "https://github.com/example/my-repo/blob/main/publiccode.yml",
		PublicCode: &models.PublicCode{
			Name:              "Mijn Repository",
			SoftwareType:      "standalone/web",
			DevelopmentStatus: "stable",
			Platforms:         []string{"web", "linux"},
			Legal:             &models.PublicCodeLegal{License: "EUPL-1.2"},
			Localisation: &models.PublicCodeLocalisation{
				LocalisationReady:  &localisationReady,
				AvailableLanguages: []string{"nl", "en"},
			},
			Maintenance: &models.PublicCodeMaintenance{
				Type: "internal",
				Contacts: []models.PublicCodeContact{
					{Name: "Team OSS"},
				},
			},
			Description: map[string]models.PublicCodeDescription{
				"nl": {
					ShortDescription: "Nog een korte beschrijving",
					LongDescription:  "Deze repository ondersteunt overheidsorganisaties.",
					Features:         []string{"Zoeken", "Metadata"},
				},
			},
			FundedBy: []models.PublicCodeOrganisationReference{
				{Name: "Ministerie van Test"},
			},
		},
	}

	if err := typesense.PublishRepository(context.Background(), repository); err != nil {
		t.Fatalf("PublishRepository returned error: %v", err)
	}

	if capturedPath != "/collections/oss-register/documents" {
		t.Fatalf("unexpected path %q", capturedPath)
	}
	if capturedAction != "upsert" {
		t.Fatalf("expected action=upsert, got %q", capturedAction)
	}
	if capturedKey != "secret" {
		t.Fatalf("expected api key %q, got %q", "secret", capturedKey)
	}

	var doc map[string]any
	if err := json.Unmarshal(capturedBody, &doc); err != nil {
		t.Fatalf("failed to parse payload: %v", err)
	}

	wantURL := "https://oss.developer.overheid.nl/repositories/repo-1"
	if got := doc["url"]; got != wantURL {
		t.Fatalf("unexpected url: %v", got)
	}
	if got := doc["url_without_anchor"]; got != wantURL {
		t.Fatalf("unexpected url_without_anchor: %v", got)
	}
	if doc["anchor"] != nil {
		t.Fatalf("expected anchor to be nil")
	}
	if got := doc["hierarchy.lvl0"]; got != "Mijn Repository" {
		t.Fatalf("unexpected lvl0: %v", got)
	}
	if got := doc["hierarchy.lvl1"]; got != "Ministerie van Test" {
		t.Fatalf("unexpected lvl1: %v", got)
	}
	if got := doc["hierarchy.lvl2"]; got != "Web applicatie (standalone/web)" {
		t.Fatalf("unexpected lvl2: %v", got)
	}
	if got := doc["hierarchy.lvl3"]; got != "Stabiel (stable)" {
		t.Fatalf("unexpected lvl3: %v", got)
	}
	if got := doc["hierarchy.lvl4"]; got != "EUPL-1.2" {
		t.Fatalf("unexpected lvl4: %v", got)
	}
	if got := doc["language"]; got != "nl" {
		t.Fatalf("unexpected language: %v", got)
	}
	if got := doc["item_priority"]; int(got.(float64)) != 5 {
		t.Fatalf("unexpected item_priority: %v", got)
	}

	content, ok := doc["content"].(string)
	if !ok || !strings.Contains(content, "Naam: Mijn Repository") {
		t.Fatalf("content missing name: %v", doc["content"])
	}
	if !strings.Contains(content, "Repository URL: https://github.com/example/my-repo") {
		t.Fatalf("content missing repository url: %v", doc["content"])
	}
	if !strings.Contains(content, "Publiccode URL: https://github.com/example/my-repo/blob/main/publiccode.yml") {
		t.Fatalf("content missing publiccode url: %v", content)
	}
	if !strings.Contains(content, "Organisatie: Ministerie van Test") {
		t.Fatalf("content missing organisation: %v", content)
	}
	if !strings.Contains(content, "Licentie: EUPL-1.2") {
		t.Fatalf("content missing license: %v", content)
	}
	if !strings.Contains(content, "Softwaretype: Web applicatie (standalone/web)") {
		t.Fatalf("content missing readable software type: %v", content)
	}
	if !strings.Contains(content, "Ontwikkelstatus: Stabiel (stable)") {
		t.Fatalf("content missing readable development status: %v", content)
	}
	if !strings.Contains(content, "Onderhoud: Intern (internal)") {
		t.Fatalf("content missing readable maintenance type: %v", content)
	}
	if !strings.Contains(content, "Features: Zoeken, Metadata") {
		t.Fatalf("content missing features: %v", content)
	}

	rawTags, ok := doc["tags"].([]any)
	if !ok {
		t.Fatalf("tags missing or wrong type: %T", doc["tags"])
	}

	gotTags := make([]string, 0, len(rawTags))
	for _, v := range rawTags {
		gotTags = append(gotTags, v.(string))
	}

	wantTags := []string{
		"oss-register",
		"repository",
		"repository-id:repo-1",
		"Ministerie van Test",
		"https://organisaties.example.com/min-test",
		"publiccode",
		"softwareType:standalone/web",
		"developmentStatus:stable",
		"license:EUPL-1.2",
		"platform:web",
		"platform:linux",
		"language:nl",
		"language:en",
	}
	if len(gotTags) != len(wantTags) {
		t.Fatalf("unexpected tag count: %v", gotTags)
	}
	for i, want := range wantTags {
		if gotTags[i] != want {
			t.Fatalf("unexpected tag at position %d: want %q got %q", i, want, gotTags[i])
		}
	}
}

func TestPublishRepository_UsesReadableFallbackForRepositoryWithoutPublicCode(t *testing.T) {
	var capturedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := r.Body.Close(); err != nil {
				t.Errorf("failed to close request body: %v", err)
			}
		}()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		capturedBody = body
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	t.Setenv("TYPESENSE_ENDPOINT", server.URL)
	t.Setenv("TYPESENSE_API_KEY", "secret")
	t.Setenv("TYPESENSE_COLLECTION", "oss-register")
	t.Setenv("TYPESENSE_DETAIL_BASE_URL", "https://oss.developer.overheid.nl/repositories")

	prevClient := httpclient.HTTPClient
	httpclient.HTTPClient = server.Client()
	t.Cleanup(func() {
		httpclient.HTTPClient = prevClient
	})

	repository := &models.Repository{
		Id:               "6114c7cb-7b60-4905-8946-31e1878d74b7",
		Name:             "sdmxdata",
		Url:              "https://github.com/statistiekcbs/sdmxdata.git",
		ShortDescription: "R package based on SDMX REST v2.1",
		IsFork:           true,
		Organisation: &models.Organisation{
			Label: "Centraal Bureau voor de Statistiek",
			Uri:   "https://identifier.overheid.nl/tooi/id/zbo/zb000193",
		},
	}

	if err := typesense.PublishRepository(context.Background(), repository); err != nil {
		t.Fatalf("PublishRepository returned error: %v", err)
	}

	var doc map[string]any
	if err := json.Unmarshal(capturedBody, &doc); err != nil {
		t.Fatalf("failed to parse payload: %v", err)
	}

	if got := doc["hierarchy.lvl0"]; got != "sdmxdata" {
		t.Fatalf("unexpected lvl0: %v", got)
	}
	if got := doc["hierarchy.lvl1"]; got != "Centraal Bureau voor de Statistiek" {
		t.Fatalf("unexpected lvl1: %v", got)
	}
	if got := doc["hierarchy.lvl2"]; got != "Repository" {
		t.Fatalf("unexpected lvl2 fallback: %v", got)
	}
	if got := doc["hierarchy.lvl3"]; got != "Git fork" {
		t.Fatalf("unexpected lvl3 fallback: %v", got)
	}

	content, ok := doc["content"].(string)
	if !ok {
		t.Fatalf("content missing or wrong type: %T", doc["content"])
	}
	if strings.Contains(content, "ForkType: GIT_FORK") {
		t.Fatalf("content still contains raw fork type: %v", content)
	}
	if !strings.Contains(content, "Naam: sdmxdata") {
		t.Fatalf("content missing name: %v", content)
	}
	if !strings.Contains(content, "Beschrijving: R package based on SDMX REST v2.1") {
		t.Fatalf("content missing description: %v", content)
	}
	if !strings.Contains(content, "Repository URL: https://github.com/statistiekcbs/sdmxdata.git") {
		t.Fatalf("content missing repository url: %v", content)
	}
	if !strings.Contains(content, "Repositorytype: Git fork") {
		t.Fatalf("content missing readable fork type: %v", content)
	}
}
