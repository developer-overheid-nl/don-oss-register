/*
 * API register API v1
 *
 * API van het API register (apis.developer.overheid.nl)
 *
 * API version: 1.0.0
 * Contact: developer.overheid@geonovum.nl
 */

package models

import "strings"

type RepositorySummary struct {
	Id             string               `json:"id" gorm:"column:id;primaryKey"`
	Name           string               `json:"name" gorm:"column:name"`
	Description    string               `json:"description" gorm:"column:description"`
	Organisation   *OrganisationSummary `json:"organisation,omitempty" gorm:"foreignKey:OrganisationID;references:Uri"`
	OrganisationID *string              `json:"organisationId,omitempty" gorm:"column:organisation_id"`
	RepositorieUri string               `json:"repositorieUri" gorm:"column:repositorie_uri"`
	PublicCodeUrl  string               `json:"publicCodeUrl" gorm:"column:public_code_url"`
	CreatedAt      int64                `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt      int64                `json:"updatedAt" gorm:"column:updated_at"`
	Links          *Links               `json:"_links,omitempty"`
}

type RepositorieDetail struct {
	RepositorySummary
}

type Repositorie struct {
	Id             string        `json:"id" gorm:"column:id;primaryKey"`
	Name           string        `json:"name" gorm:"column:name"`
	Description    string        `json:"description" gorm:"column:description"`
	Organisation   *Organisation `json:"organisation,omitempty" gorm:"foreignKey:OrganisationID;references:Uri"`
	OrganisationID *string       `json:"organisationId,omitempty" gorm:"column:organisation_id"`
	RepositorieUri string        `json:"repositorieUri" gorm:"column:repositorie_uri"`
	PublicCodeUrl  string        `json:"publicCodeUrl" gorm:"column:public_code_url"`
	CreatedAt      int64         `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt      int64         `json:"updatedAt" gorm:"column:updated_at"`
}

type ListRepositoriesSearchParams struct {
	Page         int     `query:"page"`
	PerPage      int     `query:"perPage"`
	Organisation *string `query:"organisation"`
	Query        string  `query:"q" binding:"required"`
	BaseURL      string
}

type PostRepositorie struct {
	GitOrganisationUrl string `json:"gitOrganisationUrl" binding:"required,url"`
	OrganisationUrl    string `json:"organisationUrl" binding:"required,url"`
}

type OrganisationSummary struct {
	Uri   string `gorm:"column:uri;primaryKey" json:"uri"`
	Label string `gorm:"column:label" json:"label"`
	Links *Links `json:"_links,omitempty"`
}

type Organisation struct {
	Uri   string `gorm:"column:uri;primaryKey" json:"uri"`
	Label string `gorm:"column:label" json:"label"`
}

type ListRepositoriesParams struct {
	Page         int     `query:"page"`
	PerPage      int     `query:"perPage"`
	Organisation *string `query:"organisation"`
	Ids          *string `query:"ids"`
	BaseURL      string
}

func (p *ListRepositoriesParams) FilterIDs() *string {
	if p == nil {
		return nil
	}
	return trimPointer(p.Ids)
}

func trimPointer(val *string) *string {
	if val == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*val)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

type RepositorieParams struct {
	Id string `path:"id"`
}

// Link representeert een hypermedia‐link
type Link struct {
	Href string `json:"href"`
}
type Links struct {
	First        *Link `json:"first,omitempty"`
	Prev         *Link `json:"prev,omitempty"`
	Self         *Link `json:"self,omitempty"`
	Next         *Link `json:"next,omitempty"`
	Last         *Link `json:"last,omitempty"`
	Repositories *Link `json:"repositories,omitempty"` // link naar de lijst van repositories
}

type Lifecycle struct {
	Status     string `json:"status"`
	Version    string `json:"version"`
	Sunset     string `json:"sunset,omitempty"`
	Deprecated string `json:"deprecated,omitempty"`
}

type Pagination struct {
	Next           *int
	Previous       *int
	CurrentPage    int
	RecordsPerPage int
	TotalPages     int
	TotalRecords   int
}
