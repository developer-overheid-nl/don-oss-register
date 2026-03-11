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
	ShortDescription *string   `json:"shortDescription,omitempty"`
	CreatedAt        time.Time `json:"createdAt" gorm:"column:created_at"`
	LastCrawledAt    time.Time `json:"lastCrawledAt" gorm:"column:last_crawled_at"`
	Name             *string   `json:"name,omitempty"`
	LastActivityAt   time.Time `json:"lastActivityAt,omitempty" gorm:"column:last_activity_at"`
}
type ListRepositorysParams struct {
	Page         int     `query:"page" validate:"omitempty,min=1"`
	PerPage      int     `query:"perPage" validate:"omitempty,min=1,max=100"`
	Organisation *string `query:"organisation"`
	PublicCode   *bool   `query:"publiccode"`
	BaseURL      string
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
