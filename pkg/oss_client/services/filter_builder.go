package services

import (
	"strings"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
)

func buildPublicCodeGroup(p *models.RepositoryFiltersParams, counts *models.RepositoryFilterCounts) models.FilterGroup {
	value := p.PublicCode != nil && *p.PublicCode
	return models.FilterGroup{
		Key:         "publiccode",
		Label:       "Heeft publiccode.yml",
		Description: "Toon alleen repositories met een publiccode.yml bestand.",
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
		label := fc.Value
		if name, ok := models.LanguageLabels[fc.Value]; ok {
			label = name
		}
		options = append(options, models.FilterOption{
			Value:    fc.Value,
			Label:    label,
			Count:    fc.Count,
			Selected: selected[fc.Value],
		})
	}
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
		label := fc.Value
		var desc *string
		if meta, ok := labels[fc.Value]; ok {
			label = meta[0]
			d := meta[1]
			desc = &d
		}
		options = append(options, models.FilterOption{
			Value:       fc.Value,
			Label:       label,
			Description: desc,
			Count:       fc.Count,
			Selected:    selected[fc.Value],
		})
	}
	return options
}

func selectedSet(values []string) map[string]bool {
	m := make(map[string]bool, len(values))
	for _, v := range values {
		m[v] = true
	}
	return m
}
