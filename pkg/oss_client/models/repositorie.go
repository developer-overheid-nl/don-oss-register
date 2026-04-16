/*
 * API register API v1
 *
 * API van het API register (apis.developer.overheid.nl)
 *
 * API version: 1.0.0
 * Contact: developer.overheid@geonovum.nl
 */

package models

import (
	"fmt"
	"time"
)

type PublicCode struct {
	PubliccodeYmlVersion string                            `json:"publiccodeYmlVersion,omitempty"`
	Name                 string                            `json:"name,omitempty"`
	Url                  string                            `json:"url,omitempty"`
	Platforms            []string                          `json:"platforms,omitempty"`
	DevelopmentStatus    string                            `json:"developmentStatus,omitempty"`
	SoftwareType         string                            `json:"softwareType,omitempty"`
	Legal                *PublicCodeLegal                  `json:"legal,omitempty"`
	Description          map[string]PublicCodeDescription  `json:"description,omitempty"`
	Maintenance          *PublicCodeMaintenance            `json:"maintenance,omitempty"`
	Localisation         *PublicCodeLocalisation           `json:"localisation,omitempty"`
	Organisation         *PublicCodeOrganisation           `json:"organisation,omitempty"`
	DependsOn            *PublicCodeDependsOn              `json:"dependsOn,omitempty"`
	FundedBy             []PublicCodeOrganisationReference `json:"fundedBy,omitempty"`
}

type PublicCodeLegal struct {
	License string `json:"license,omitempty"`
}

type PublicCodeDescription struct {
	ShortDescription string   `json:"shortDescription,omitempty"`
	LongDescription  string   `json:"longDescription,omitempty"`
	Features         []string `json:"features,omitempty"`
}

type PublicCodeMaintenance struct {
	Type        string                 `json:"type,omitempty"`
	Contractors []PublicCodeContractor `json:"contractors,omitempty"`
	Contacts    []PublicCodeContact    `json:"contacts,omitempty"`
}

type PublicCodeContractor struct {
	Name  string `json:"name,omitempty"`
	Until string `json:"until,omitempty"`
}

type PublicCodeContact struct {
	Name string `json:"name,omitempty"`
}

type PublicCodeLocalisation struct {
	LocalisationReady  *bool    `json:"localisationReady,omitempty"`
	AvailableLanguages []string `json:"availableLanguages,omitempty"`
}

type PublicCodeOrganisation struct {
	Uri string `json:"uri,omitempty"`
}

type PublicCodeDependsOn struct {
	Open        []PublicCodeDependency `json:"open,omitempty"`
	Proprietary []PublicCodeDependency `json:"proprietary,omitempty"`
	Hardware    []PublicCodeDependency `json:"hardware,omitempty"`
}

type PublicCodeDependency struct {
	Name string `json:"name,omitempty"`
}

type PublicCodeOrganisationReference struct {
	Name string `json:"name,omitempty"`
}

type RepositorySummary struct {
	Id               string               `json:"id" gorm:"column:id;primaryKey"`
	Url              string               `json:"url" gorm:"column:repository_url"`
	ForkType         RepositoryForkType   `json:"forkType,omitempty" gorm:"-"`
	Organisation     *OrganisationSummary `json:"organisation,omitempty" gorm:"foreignKey:OrganisationID;references:Uri"`
	OrganisationID   *string              `json:"-" gorm:"column:organisation_id"`
	PublicCodeUrl    string               `json:"publicCodeUrl,omitempty" gorm:"column:public_code_url"`
	ShortDescription string               `json:"shortDescription,omitempty" gorm:"column:short_description"`
	Name             string               `json:"name,omitempty" gorm:"column:name"`
	CreatedAt        time.Time            `json:"createdAt" gorm:"column:created_at"`
	LastCrawledAt    time.Time            `json:"lastCrawledAt" gorm:"column:last_crawled_at"`
	LastActivityAt   time.Time            `json:"lastActivityAt,omitempty" gorm:"column:last_activity_at"`
}

type RepositoryDetail struct {
	RepositorySummary
	PublicCode      *PublicCode `json:"publicCode,omitempty"`
	LongDescription string      `json:"longDescription,omitempty"`
}

type Repository struct {
	Id               string        `json:"id" gorm:"column:id;primaryKey"`
	Name             string        `json:"name" gorm:"column:name"`
	ShortDescription string        `json:"shortDescription" gorm:"column:short_description"`
	LongDescription  string        `json:"longDescription,omitempty" gorm:"column:long_description"`
	Organisation     *Organisation `json:"-" gorm:"foreignKey:OrganisationID;references:Uri"`
	OrganisationID   *string       `json:"-" gorm:"column:organisation_id"`
	Url              string        `json:"url" gorm:"column:repository_url"`
	IsFork           bool          `json:"-" gorm:"column:is_fork"`
	ForkBasedOnURLs  []string      `json:"-" gorm:"column:fork_based_on_urls;serializer:json"`
	PublicCodeUrl    string        `json:"publicCodeUrl,omitempty" gorm:"column:public_code_url"`
	PublicCode       *PublicCode   `json:"publicCode,omitempty" gorm:"column:public_code_data;serializer:json"`
	CreatedAt        time.Time     `json:"createdAt" gorm:"column:created_at"`
	LastCrawledAt    time.Time     `json:"lastCrawledAt" gorm:"column:last_crawled_at"`
	LastActivityAt   time.Time     `json:"lastActivityAt,omitempty" gorm:"column:last_activity_at"`
	Active           bool          `json:"-" gorm:"column:active"`
}

type ListRepositorysSearchParams struct {
	Page         int     `query:"page" validate:"omitempty,min=1"`
	PerPage      int     `query:"perPage" validate:"omitempty,min=1,max=100"`
	Organisation *string `query:"organisation"`
	Query        string  `query:"q" binding:"required"`
	BaseURL      string
}

type RepositoryInput struct {
	Url              *string   `json:"url" binding:"required,url"`
	OrganisationUri  *string   `json:"organisationUri" binding:"required,url"`
	PublicCodeUrl    *string   `json:"publicCodeUrl" binding:"omitempty,url"`
	IsFork           *bool     `json:"isFork,omitempty"`
	ShortDescription *string   `json:"shortDescription,omitempty"`
	CreatedAt        time.Time `json:"createdAt" gorm:"column:created_at"`
	LastCrawledAt    time.Time `json:"lastCrawledAt" gorm:"column:last_crawled_at"`
	Name             *string   `json:"name,omitempty"`
	LastActivityAt   time.Time `json:"lastActivityAt,omitempty" gorm:"column:last_activity_at"`
}
type ListRepositorysParams struct {
	Page               int      `query:"page" validate:"omitempty,min=1"`
	PerPage            int      `query:"perPage" validate:"omitempty,min=1,max=100"`
	Organisation       *string  `query:"organisation"`
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

type RepositoryParams struct {
	Id string `path:"id"`
}

type UpdateRepositoryRequest struct {
	RepositoryParams
	RepositoryInput
}

type Pagination struct {
	Next           *int
	Previous       *int
	CurrentPage    int
	RecordsPerPage int
	TotalPages     int
	TotalRecords   int
}

type FilterOption struct {
	Value       string  `json:"value"`
	Label       string  `json:"label"`
	Description *string `json:"description"`
	Count       int     `json:"count"`
	Selected    bool    `json:"selected"`
}

type FilterGroup struct {
	Key         string         `json:"key"`
	Label       string         `json:"label"`
	Description string         `json:"description"`
	Type        string         `json:"type"`
	Value       any            `json:"value,omitempty"`
	Count       *int           `json:"count,omitempty"`
	Options     []FilterOption `json:"options,omitempty"`
}

func (f FilterGroup) Validate() error {
	switch f.Type {
	case "toggle":
		if _, ok := f.Value.(bool); !ok {
			return fmt.Errorf("filter %q: toggle value must be bool, got %T", f.Key, f.Value)
		}
	case "date":
		if f.Value != nil {
			if _, ok := f.Value.(string); !ok {
				return fmt.Errorf("filter %q: date value must be string, got %T", f.Key, f.Value)
			}
		}
	}
	return nil
}

type FilterCount struct {
	Value string
	Count int
}

type OrgFilterCount struct {
	Value string
	Label string
	Count int
}

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
	PublicCode         *bool    `query:"publiccode"`
	LastActivityAfter  *string  `query:"lastActivityAfter"`
	SoftwareType       []string `query:"softwareType"`
	DevelopmentStatus  []string `query:"developmentStatus"`
	AvailableLanguages []string `query:"availableLanguages"`
	MaintenanceType    []string `query:"maintenanceType"`
	License            []string `query:"license"`
	Platforms          []string `query:"platforms"`
}
