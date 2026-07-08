package models

import (
	commonfilters "github.com/developer-overheid-nl/don-register-common/filters"
	commonpagination "github.com/developer-overheid-nl/don-register-common/pagination"
)

type ListRepositorysSearchParams struct {
	Page         int     `query:"page" validate:"omitempty,min=1"`
	PerPage      int     `query:"perPage" validate:"omitempty,min=1,max=100"`
	Organisation *string `query:"organisation"`
	Query        string  `query:"q" binding:"required"`
	BaseURL      string
}

type ListRepositorysParams struct {
	Page               int      `query:"page" validate:"omitempty,min=1"`
	PerPage            int      `query:"perPage" validate:"omitempty,min=1,max=100"`
	Organisation       *string  `query:"organisation"`
	Query              string   `query:"q"`
	PublicCode         *bool    `query:"publiccode"`
	LastActivityAfter  *string  `query:"lastActivityAfter"`
	SoftwareType       []string `query:"softwareType"`
	DevelopmentStatus  []string `query:"developmentStatus"`
	AvailableLanguages []string `query:"availableLanguages"`
	MaintenanceType    []string `query:"maintenanceType"`
	License            []string `query:"license"`
	Platforms          []string `query:"platforms"`
	BaseURL            string
}

func (p *ListRepositorysParams) RepositoryFilters() *RepositoryFiltersParams {
	if p == nil {
		return &RepositoryFiltersParams{}
	}
	return &RepositoryFiltersParams{
		Organisation:       p.Organisation,
		Query:              p.Query,
		PublicCode:         p.PublicCode,
		LastActivityAfter:  p.LastActivityAfter,
		SoftwareType:       append([]string(nil), p.SoftwareType...),
		DevelopmentStatus:  append([]string(nil), p.DevelopmentStatus...),
		AvailableLanguages: append([]string(nil), p.AvailableLanguages...),
		MaintenanceType:    append([]string(nil), p.MaintenanceType...),
		License:            append([]string(nil), p.License...),
		Platforms:          append([]string(nil), p.Platforms...),
	}
}

type Pagination = commonpagination.Pagination

type FilterOption = commonfilters.FilterOption

type FilterGroup = commonfilters.FilterGroup

type FilterCount = commonfilters.FilterCount

type OrgFilterCount = commonfilters.FilterCount

type RepositoryFilterCounts struct {
	PublicCode         int
	LastActivityAfter  *int
	SoftwareType       []FilterCount
	DevelopmentStatus  []FilterCount
	MaintenanceType    []FilterCount
	License            []FilterCount
	Platforms          []FilterCount
	AvailableLanguages []FilterCount
	Organisation       []OrgFilterCount
}

type RepositoryFiltersParams struct {
	Organisation       *string  `query:"organisation"`
	Query              string   `query:"q"`
	PublicCode         *bool    `query:"publiccode"`
	LastActivityAfter  *string  `query:"lastActivityAfter"`
	SoftwareType       []string `query:"softwareType"`
	DevelopmentStatus  []string `query:"developmentStatus"`
	AvailableLanguages []string `query:"availableLanguages"`
	MaintenanceType    []string `query:"maintenanceType"`
	License            []string `query:"license"`
	Platforms          []string `query:"platforms"`
}
