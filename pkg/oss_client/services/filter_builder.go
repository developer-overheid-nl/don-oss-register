package services

import (
	"sort"
	"strings"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
)

func buildPublicCodeGroup(p *models.RepositoryFiltersParams, counts *models.RepositoryFilterCounts) models.FilterGroup {
	value := p == nil || p.PublicCode == nil || *p.PublicCode
	return models.FilterGroup{
		Key:         "publiccode",
		Label:       "Heeft publiccode.yml",
		Description: "Filter repositories op aanwezigheid van een publiccode.yml bestand.",
		Type:        "toggle",
		Value:       value,
		Count:       &counts.PublicCode,
	}
}

func buildLastActivityGroup(p *models.RepositoryFiltersParams, counts *models.RepositoryFilterCounts) models.FilterGroup {
	var value any
	if p.LastActivityAfter != nil {
		value = *p.LastActivityAfter
	}
	return models.FilterGroup{
		Key:         "lastActivityAfter",
		Label:       "Actief na",
		Description: "Toon repositories die na de opgegeven datum nog activiteit hebben gehad.",
		Type:        "date",
		Value:       value,
		Count:       counts.LastActivityAfter,
	}
}

func buildSoftwareTypeGroup(p *models.RepositoryFiltersParams, counts *models.RepositoryFilterCounts) models.FilterGroup {
	return models.FilterGroup{
		Key:         "softwareType",
		Label:       "Software type",
		Description: "Het type software zoals gedefinieerd in publiccode.yml.",
		Type:        "multi-select",
		Options:     buildMultiSelectOptions(counts.SoftwareType, selectedSet(p.SoftwareType), models.SoftwareTypeLabels),
	}
}

func buildDevelopmentStatusGroup(p *models.RepositoryFiltersParams, counts *models.RepositoryFilterCounts) models.FilterGroup {
	return models.FilterGroup{
		Key:         "developmentStatus",
		Label:       "Ontwikkelstatus",
		Description: "De huidige ontwikkelstatus van de software.",
		Type:        "multi-select",
		Options:     buildMultiSelectOptions(counts.DevelopmentStatus, selectedSet(p.DevelopmentStatus), models.DevelopmentStatusLabels),
	}
}

func buildMaintenanceTypeGroup(p *models.RepositoryFiltersParams, counts *models.RepositoryFilterCounts) models.FilterGroup {
	return models.FilterGroup{
		Key:         "maintenanceType",
		Label:       "Onderhoud",
		Description: "Hoe het onderhoud van de software is georganiseerd.",
		Type:        "multi-select",
		Options:     buildMultiSelectOptions(counts.MaintenanceType, selectedSet(p.MaintenanceType), models.MaintenanceTypeLabels),
	}
}

func buildPlatformsGroup(p *models.RepositoryFiltersParams, counts *models.RepositoryFilterCounts) models.FilterGroup {
	return models.FilterGroup{
		Key:         "platforms",
		Label:       "Platforms",
		Description: "De platforms waarop de software beschikbaar is.",
		Type:        "multi-select",
		Options:     buildMultiSelectOptions(counts.Platforms, selectedSet(p.Platforms), models.PlatformLabels),
	}
}

func buildAvailableLanguagesGroup(p *models.RepositoryFiltersParams, counts *models.RepositoryFilterCounts) models.FilterGroup {
	selected := selectedSet(p.AvailableLanguages)
	options := make([]models.FilterOption, 0, len(counts.AvailableLanguages))
	for _, fc := range counts.AvailableLanguages {
		options = append(options, models.FilterOption{
			Value:    fc.Value,
			Label:    languageLabel(fc.Value),
			Count:    fc.Count,
			Selected: selected[fc.Value],
		})
	}
	options = appendMissingSelectedOptions(options, selected, func(value string) models.FilterOption {
		return models.FilterOption{
			Value:    value,
			Label:    languageLabel(value),
			Selected: true,
		}
	})
	sortFilterOptions(options)
	return models.FilterGroup{
		Key:         "availableLanguages",
		Label:       "Beschikbare talen",
		Description: "De talen waarin de software beschikbaar is.",
		Type:        "multi-select",
		Options:     options,
	}
}

func buildLicenseGroup(p *models.RepositoryFiltersParams, counts *models.RepositoryFilterCounts) models.FilterGroup {
	selected := selectedSet(p.License)
	options := make([]models.FilterOption, 0, len(counts.License))
	for _, fc := range counts.License {
		options = append(options, models.FilterOption{
			Value:    fc.Value,
			Label:    fc.Value,
			Count:    fc.Count,
			Selected: selected[fc.Value],
		})
	}
	options = appendMissingSelectedOptions(options, selected, func(value string) models.FilterOption {
		return models.FilterOption{
			Value:    value,
			Label:    value,
			Selected: true,
		}
	})
	sortFilterOptions(options)
	return models.FilterGroup{
		Key:         "license",
		Label:       "Licentie",
		Description: "De open source licentie van de software (SPDX-identifier).",
		Type:        "multi-select",
		Options:     options,
	}
}

func buildOrganisationGroup(p *models.RepositoryFiltersParams, counts *models.RepositoryFilterCounts) models.FilterGroup {
	activeOrg := ""
	if p.Organisation != nil {
		activeOrg = strings.TrimSpace(*p.Organisation)
	}
	options := make([]models.FilterOption, 0, len(counts.Organisation))
	for _, fc := range counts.Organisation {
		options = append(options, models.FilterOption{
			Value:    fc.Value,
			Label:    fc.Label,
			Count:    fc.Count,
			Selected: activeOrg == fc.Value,
		})
	}
	if activeOrg != "" {
		options = appendMissingSelectedOptions(options, map[string]bool{activeOrg: true}, func(value string) models.FilterOption {
			return models.FilterOption{
				Value:    value,
				Label:    value,
				Selected: true,
			}
		})
	}
	sortFilterOptions(options)
	return models.FilterGroup{
		Key:         "organisation",
		Label:       "Organisatie",
		Description: "De overheidsorganisatie die de repository beheert.",
		Type:        "single-select",
		Options:     options,
	}
}

func buildMultiSelectOptions(counts []models.FilterCount, selected map[string]bool, labels map[string][2]string) []models.FilterOption {
	options := make([]models.FilterOption, 0, len(counts))
	for _, fc := range counts {
		options = append(options, multiSelectOption(fc.Value, fc.Count, selected[fc.Value], labels))
	}
	options = appendMissingSelectedOptions(options, selected, func(value string) models.FilterOption {
		return multiSelectOption(value, 0, true, labels)
	})
	sortFilterOptions(options)
	return options
}

func multiSelectOption(value string, count int, selected bool, labels map[string][2]string) models.FilterOption {
	label := value
	var desc *string
	if meta, ok := labels[value]; ok {
		label = meta[0]
		d := meta[1]
		desc = &d
	}
	return models.FilterOption{
		Value:       value,
		Label:       label,
		Description: desc,
		Count:       count,
		Selected:    selected,
	}
}

func appendMissingSelectedOptions(options []models.FilterOption, selected map[string]bool, build func(string) models.FilterOption) []models.FilterOption {
	seen := make(map[string]bool, len(options))
	for _, option := range options {
		seen[option.Value] = true
	}
	for value, isSelected := range selected {
		if value == "" || !isSelected || seen[value] {
			continue
		}
		options = append(options, build(value))
	}
	return options
}

func languageLabel(value string) string {
	if label, ok := models.LanguageLabels[value]; ok {
		return label
	}
	return value
}

func selectedSet(values []string) map[string]bool {
	m := make(map[string]bool, len(values))
	for _, v := range values {
		m[v] = true
	}
	return m
}

func sortFilterOptions(options []models.FilterOption) {
	sort.Slice(options, func(i, j int) bool {
		left := strings.ToLower(options[i].Label)
		right := strings.ToLower(options[j].Label)
		if left == right {
			return options[i].Value < options[j].Value
		}
		return left < right
	})
}
