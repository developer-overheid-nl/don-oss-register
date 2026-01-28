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
	LongDescription string `json:"longDescription,omitempty"`
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
