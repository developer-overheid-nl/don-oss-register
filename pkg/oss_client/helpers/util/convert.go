package util

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"github.com/google/uuid"
	publiccode "github.com/italia/publiccode-parser-go/v5"
)

func ToRepositorySummary(repo *models.Repository) models.RepositorySummary {
	var orgSummary *models.OrganisationSummary
	if repo.Organisation != nil {
		orgSummary = &models.OrganisationSummary{
			Uri:   repo.Organisation.Uri,
			Label: repo.Organisation.Label,
		}
	}
	return models.RepositorySummary{
		Id:               repo.Id,
		Url:              repo.Url,
		Name:             repo.Name,
		ShortDescription: repo.ShortDescription,
		PublicCodeUrl:    repo.PublicCodeUrl,
		CreatedAt:        repo.CreatedAt,
		LastCrawledAt:    repo.LastCrawledAt,
		LastActivityAt:   repo.LastActivityAt,
		Organisation:     orgSummary,
	}
}

func ToRepositoryDetail(repo *models.Repository) *models.RepositoryDetail {
	return &models.RepositoryDetail{
		RepositorySummary:    ToRepositorySummary(repo),
		PublicCode:           repo.PublicCode,
		PublicCodeValidation: repo.PublicCodeValidation,
		LongDescription:      repo.LongDescription,
	}
}

func ToGitOrganisatieSummary(gitOrg *models.GitOrganisatie) models.GitOrganisatieSummary {
	return models.GitOrganisatieSummary{
		Id:           gitOrg.Id,
		Organisation: gitOrg.Organisation,
		Url:          gitOrg.Url,
	}
}

func ApplyRepositoryInput(target *models.Repository, input *models.RepositoryInput) *models.Repository {
	if target == nil {
		target = &models.Repository{
			Id: uuid.NewString(),
		}
	}

	if input == nil {
		return target
	}

	if input.Url != nil {
		target.Url = strings.TrimSpace(*input.Url)
	}
	if !input.CreatedAt.IsZero() {
		target.CreatedAt = input.CreatedAt
	}
	if !input.LastCrawledAt.IsZero() {
		target.LastCrawledAt = input.LastCrawledAt
	}
	if !input.LastActivityAt.IsZero() {
		target.LastActivityAt = input.LastActivityAt
	}
	if input.PublicCodeUrl != nil {
		target.PublicCodeUrl = strings.TrimSpace(*input.PublicCodeUrl)
	}
	if input.Name != nil {
		target.Name = strings.TrimSpace(*input.Name)
	}
	if input.ShortDescription != nil {
		target.ShortDescription = strings.TrimSpace(*input.ShortDescription)
		target.LongDescription = target.ShortDescription
	}

	if input.PublicCodeUrl != nil {
		newURL := strings.TrimSpace(*input.PublicCodeUrl)
		target.PublicCodeUrl = newURL
	}

	if target.PublicCodeUrl != "" {
		content := fetchPublicCodeContent(target.PublicCodeUrl)
		if content != "" {
			hash := fmt.Sprintf("%x", sha256.Sum256([]byte(content)))
			if hash == target.PublicCodeHash {
				return target
			}

			_, name, shortDesc, longDesc, publicCode, validation := parsePublicCodeYAML(content)
			target.PublicCodeHash = hash
			target.PublicCodeValidation = validation
			if publicCode != nil {
				target.PublicCode = publicCode
			}
			if name != "" {
				target.Name = name
			}
			if shortDesc != "" {
				target.ShortDescription = shortDesc
				target.LongDescription = shortDesc
			}
			if longDesc != "" {
				target.LongDescription = longDesc
			}
		}
	}

	return target
}

func fetchPublicCodeContent(urlOrContent string) string {
	if !isLikelyURL(urlOrContent) {
		return urlOrContent
	}
	resp, err := http.Get(urlOrContent)
	if err != nil || resp == nil {
		return ""
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return ""
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return string(body)
}

func parsePublicCodeYAML(raw string) (url, name, shortDescription, longDescription string, details *models.PublicCode, validation *models.PublicCodeValidation) {
	parser, err := publiccode.NewParser(publiccode.ParserConfig{
		DisableExternalChecks: true,
	})
	if err != nil {
		log.Printf("publiccode parser initialization failed: %v", err)
		return "", "", "", "", nil, &models.PublicCodeValidation{
			Valid:       false,
			ValidatedAt: time.Now(),
			Errors:      []models.PublicCodeValidationItem{{Message: err.Error()}},
		}
	}

	parsed, parseErr := parser.ParseStream(strings.NewReader(strings.TrimPrefix(raw, "\ufeff")))
	if parsed == nil {
		msg := "empty parse result"
		if parseErr != nil {
			msg = parseErr.Error()
		}
		log.Printf("publiccode parse failed: %s", msg)
		return "", "", "", "", nil, &models.PublicCodeValidation{
			Valid:       false,
			ValidatedAt: time.Now(),
			Errors:      []models.PublicCodeValidationItem{{Message: msg}},
		}
	}

	validation = buildValidation(parseErr)

	if parseErr != nil {
		if _, ok := parseErr.(publiccode.ValidationResults); !ok {
			log.Printf("publiccode parse failed with non-validation error: %v", parseErr)
			return "", "", "", "", nil, validation
		}
	}

	v0, ok := asPublicCodeV0(parsed)
	if !ok {
		log.Printf("publiccode parse result is not version 0: %T", parsed)
		return "", "", "", "", nil, validation
	}
	details = mapPublicCodeMandatoryFields(v0)

	if v0.URL != nil {
		url = strings.TrimSpace(v0.URL.String())
	}

	name = strings.TrimSpace(v0.Name)

	desc := selectDescription(v0.Description, v0.Localisation.AvailableLanguages)
	if name == "" && desc.LocalisedName != nil {
		name = strings.TrimSpace(*desc.LocalisedName)
	}

	shortDescription = strings.TrimSpace(desc.ShortDescription)
	longDescription = strings.TrimSpace(desc.LongDescription)
	if shortDescription == "" {
		log.Printf("publiccode description does not contain a short description for repository with url %q", url)
	}
	if longDescription == "" {
		log.Printf("publiccode description does not contain a long description for repository with url %q", url)
	}

	return url, name, shortDescription, longDescription, details, validation
}

func buildValidation(parseErr error) *models.PublicCodeValidation {
	v := &models.PublicCodeValidation{Valid: true, ValidatedAt: time.Now()}
	if parseErr == nil {
		return v
	}
	results, ok := parseErr.(publiccode.ValidationResults)
	if !ok {
		v.Valid = false
		v.Errors = []models.PublicCodeValidationItem{{Message: parseErr.Error()}}
		return v
	}
	for _, item := range results {
		switch e := item.(type) {
		case publiccode.ValidationError:
			v.Valid = false
			v.Errors = append(v.Errors, models.PublicCodeValidationItem{
				Field:   e.Key,
				Message: e.Description,
			})
		case publiccode.ValidationWarning:
			v.Warnings = append(v.Warnings, models.PublicCodeValidationItem{
				Field:   e.Key,
				Message: e.Description,
			})
		}
	}
	return v
}

func mapPublicCodeMandatoryFields(v0 publiccode.PublicCodeV0) *models.PublicCode {
	name := strings.TrimSpace(v0.Name)
	if name == "" {
		desc := selectDescription(v0.Description, v0.Localisation.AvailableLanguages)
		if desc.LocalisedName != nil {
			name = strings.TrimSpace(*desc.LocalisedName)
		}
	}

	result := &models.PublicCode{
		PubliccodeYmlVersion: strings.TrimSpace(v0.PubliccodeYamlVersion),
		Name:                 name,
		Platforms:            trimNonEmpty(v0.Platforms),
		DevelopmentStatus:    strings.TrimSpace(v0.DevelopmentStatus),
		SoftwareType:         strings.TrimSpace(v0.SoftwareType),
	}

	if v0.URL != nil {
		result.Url = strings.TrimSpace(v0.URL.String())
	}

	if license := strings.TrimSpace(v0.Legal.License); license != "" {
		result.Legal = &models.PublicCodeLegal{
			License: license,
		}
	}

	if descriptions := mapMandatoryDescriptions(v0.Description); len(descriptions) > 0 {
		result.Description = descriptions
	}

	if maintenance := mapMandatoryMaintenance(v0); maintenance != nil {
		result.Maintenance = maintenance
	}

	if localisation := mapMandatoryLocalisation(v0); localisation != nil {
		result.Localisation = localisation
	}

	if v0.Organisation != nil {
		if uri := strings.TrimSpace(v0.Organisation.URI); uri != "" {
			result.Organisation = &models.PublicCodeOrganisation{
				Uri: uri,
			}
		}
	}

	if v0.DependsOn != nil {
		if dependsOn := mapMandatoryDependsOn(v0.DependsOn.Open, v0.DependsOn.Proprietary, v0.DependsOn.Hardware); dependsOn != nil {
			result.DependsOn = dependsOn
		}
	}

	if fundedBy := mapMandatoryFundedBy(v0.FundedBy); len(fundedBy) > 0 {
		result.FundedBy = fundedBy
	}

	if isEmptyPublicCode(result) {
		return nil
	}

	return result
}

func mapMandatoryDescriptions(descriptions map[string]publiccode.DescV0) map[string]models.PublicCodeDescription {
	if len(descriptions) == 0 {
		return nil
	}

	mapped := make(map[string]models.PublicCodeDescription, len(descriptions))
	for lang, desc := range descriptions {
		item := models.PublicCodeDescription{
			ShortDescription: strings.TrimSpace(desc.ShortDescription),
			LongDescription:  strings.TrimSpace(desc.LongDescription),
			Features:         trimNonEmpty(derefStrings(desc.Features)),
		}
		if item.ShortDescription == "" && item.LongDescription == "" && len(item.Features) == 0 {
			continue
		}
		mapped[lang] = item
	}

	if len(mapped) == 0 {
		return nil
	}

	return mapped
}

func mapMandatoryMaintenance(v0 publiccode.PublicCodeV0) *models.PublicCodeMaintenance {
	maintenance := &models.PublicCodeMaintenance{
		Type: strings.TrimSpace(v0.Maintenance.Type),
	}

	if v0.Maintenance.Contractors != nil {
		maintenance.Contractors = mapMandatoryContractors(*v0.Maintenance.Contractors)
	}
	if v0.Maintenance.Contacts != nil {
		maintenance.Contacts = mapMandatoryContacts(*v0.Maintenance.Contacts)
	}

	if maintenance.Type == "" && len(maintenance.Contractors) == 0 && len(maintenance.Contacts) == 0 {
		return nil
	}

	return maintenance
}

func mapMandatoryContractors(input []publiccode.ContractorV0) []models.PublicCodeContractor {
	if len(input) == 0 {
		return nil
	}

	result := make([]models.PublicCodeContractor, 0, len(input))
	for _, contractor := range input {
		item := models.PublicCodeContractor{
			Name:  strings.TrimSpace(contractor.Name),
			Until: strings.TrimSpace(contractor.Until),
		}
		if item.Name == "" && item.Until == "" {
			continue
		}
		result = append(result, item)
	}

	if len(result) == 0 {
		return nil
	}

	return result
}

func mapMandatoryContacts(input []publiccode.ContactV0) []models.PublicCodeContact {
	if len(input) == 0 {
		return nil
	}

	result := make([]models.PublicCodeContact, 0, len(input))
	for _, contact := range input {
		name := strings.TrimSpace(contact.Name)
		if name == "" {
			continue
		}
		result = append(result, models.PublicCodeContact{Name: name})
	}

	if len(result) == 0 {
		return nil
	}

	return result
}

func mapMandatoryLocalisation(v0 publiccode.PublicCodeV0) *models.PublicCodeLocalisation {
	languages := trimNonEmpty(v0.Localisation.AvailableLanguages)
	if v0.Localisation.LocalisationReady == nil && len(languages) == 0 {
		return nil
	}

	return &models.PublicCodeLocalisation{
		LocalisationReady:  v0.Localisation.LocalisationReady,
		AvailableLanguages: languages,
	}
}

func mapMandatoryDependsOn(
	open *[]publiccode.DependencyV0,
	proprietary *[]publiccode.DependencyV0,
	hardware *[]publiccode.DependencyV0,
) *models.PublicCodeDependsOn {
	result := &models.PublicCodeDependsOn{
		Open:        mapMandatoryDependencies(open),
		Proprietary: mapMandatoryDependencies(proprietary),
		Hardware:    mapMandatoryDependencies(hardware),
	}

	if len(result.Open) == 0 && len(result.Proprietary) == 0 && len(result.Hardware) == 0 {
		return nil
	}

	return result
}

func mapMandatoryDependencies(input *[]publiccode.DependencyV0) []models.PublicCodeDependency {
	if input == nil || len(*input) == 0 {
		return nil
	}

	result := make([]models.PublicCodeDependency, 0, len(*input))
	for _, dependency := range *input {
		name := strings.TrimSpace(dependency.Name)
		if name == "" {
			continue
		}
		result = append(result, models.PublicCodeDependency{Name: name})
	}

	if len(result) == 0 {
		return nil
	}

	return result
}

func mapMandatoryFundedBy(input *[]publiccode.OrganisationV0) []models.PublicCodeOrganisationReference {
	if input == nil || len(*input) == 0 {
		return nil
	}

	result := make([]models.PublicCodeOrganisationReference, 0, len(*input))
	for _, organisation := range *input {
		if organisation.Name == nil {
			continue
		}
		name := strings.TrimSpace(*organisation.Name)
		if name == "" {
			continue
		}
		result = append(result, models.PublicCodeOrganisationReference{Name: name})
	}

	if len(result) == 0 {
		return nil
	}

	return result
}

func derefStrings(items *[]string) []string {
	if items == nil {
		return nil
	}
	return *items
}

func trimNonEmpty(items []string) []string {
	if len(items) == 0 {
		return nil
	}

	result := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}

	if len(result) == 0 {
		return nil
	}

	return result
}

func isEmptyPublicCode(data *models.PublicCode) bool {
	if data == nil {
		return true
	}

	return data.PubliccodeYmlVersion == "" &&
		data.Name == "" &&
		data.Url == "" &&
		len(data.Platforms) == 0 &&
		data.DevelopmentStatus == "" &&
		data.SoftwareType == "" &&
		data.Legal == nil &&
		len(data.Description) == 0 &&
		data.Maintenance == nil &&
		data.Localisation == nil &&
		data.Organisation == nil &&
		data.DependsOn == nil &&
		len(data.FundedBy) == 0
}


func asPublicCodeV0(pc publiccode.PublicCode) (publiccode.PublicCodeV0, bool) {
	switch typed := pc.(type) {
	case publiccode.PublicCodeV0:
		return typed, true
	case *publiccode.PublicCodeV0:
		if typed == nil {
			return publiccode.PublicCodeV0{}, false
		}
		return *typed, true
	default:
		return publiccode.PublicCodeV0{}, false
	}
}

func selectDescription(descriptions map[string]publiccode.DescV0, preferredLocales []string) publiccode.DescV0 {
	if len(descriptions) == 0 {
		return publiccode.DescV0{}
	}

	for _, key := range preferredLocaleKeys(descriptions, preferredLocales) {
		value := descriptions[key]
		if strings.TrimSpace(value.ShortDescription) != "" || strings.TrimSpace(value.LongDescription) != "" {
			return value
		}
	}

	return publiccode.DescV0{}
}

func preferredLocaleKeys(descriptions map[string]publiccode.DescV0, preferredLocales []string) []string {
	keys := make([]string, 0, len(descriptions))
	for key := range descriptions {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	ordered := make([]string, 0, len(keys))
	seen := make(map[string]struct{}, len(keys))
	appendMatches := func(match func(key string) bool) {
		for _, key := range keys {
			lower := strings.ToLower(key)
			if !match(lower) {
				continue
			}
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			ordered = append(ordered, key)
		}
	}

	for _, preferred := range preferredLocales {
		normalized := strings.ToLower(strings.TrimSpace(preferred))
		if normalized == "" {
			continue
		}

		appendMatches(func(key string) bool { return key == normalized })
		appendMatches(func(key string) bool { return strings.HasPrefix(key, normalized+"-") })

		if idx := strings.IndexRune(normalized, '-'); idx > 0 {
			base := normalized[:idx]
			appendMatches(func(key string) bool { return key == base })
			appendMatches(func(key string) bool { return strings.HasPrefix(key, base+"-") })
		}
	}

	appendMatches(func(_ string) bool { return true })

	return ordered
}

func isLikelyURL(val string) bool {
	return strings.HasPrefix(val, "http://") || strings.HasPrefix(val, "https://")
}
