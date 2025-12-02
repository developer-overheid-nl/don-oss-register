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
	"strings"
	"time"
)

type RepositorySummary struct {
	Id             string               `json:"id" gorm:"column:id;primaryKey"`
	Name           string               `json:"name" gorm:"column:name"`
	Description    string               `json:"description" gorm:"column:description"`
	Organisation   *OrganisationSummary `json:"organisation,omitempty" gorm:"foreignKey:OrganisationID;references:Uri"`
	OrganisationID *string              `json:"organisationId,omitempty" gorm:"column:organisation_id"`
	RepositoryUrl  string               `json:"repositoryUrl" gorm:"column:repository_url"`
	PublicCodeUrl  string               `json:"publicCodeUrl" gorm:"column:public_code_url"`
	CreatedAt      time.Time            `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt      time.Time            `json:"updatedAt" gorm:"column:updated_at"`
}

type RepositoryDetail struct {
	RepositorySummary
}

type Repository struct {
	Id               string        `json:"id" gorm:"column:id;primaryKey"`
	Name             string        `json:"name" gorm:"column:name"`
	ShortDescription string        `json:"shortDescription" gorm:"column:short_description"`
	LongDescription  string        `json:"longDescription" gorm:"column:long_description"`
	Organisation     *Organisation `json:"organisation,omitempty" gorm:"foreignKey:OrganisationID;references:Uri"`
	OrganisationID   *string       `json:"organisationId,omitempty" gorm:"column:organisation_id"`
	RepositoryUrl    string        `json:"repositoryUrl" gorm:"column:repository_url"`
	PublicCodeUrl    string        `json:"publicCodeUrl" gorm:"column:public_code_url"`
	CreatedAt        time.Time     `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt        time.Time     `json:"updatedAt" gorm:"column:updated_at"`
	Active           bool          `json:"active" gorm:"column:active"`
}

type ListRepositorysSearchParams struct {
	Page         int     `query:"page"`
	PerPage      int     `query:"perPage"`
	Organisation *string `query:"organisation"`
	Query        string  `query:"q" binding:"required"`
	BaseURL      string
}

type PostRepository struct {
	RepositoryUrl    *string   `json:"repositoryUrl"`
	Name             *string   `json:"name"`
	Description      *string   `json:"description"`
	PubliccodeYmlUrl *string   `json:"publiccodeYmlUrl"`
	Active           bool      `json:"active"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}
type ListRepositorysParams struct {
	Page         int     `query:"page"`
	PerPage      int     `query:"perPage"`
	Organisation *string `query:"organisation"`
	Ids          *string `query:"ids"`
	BaseURL      string
}

func (p *ListRepositorysParams) FilterIDs() *string {
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

type RepositoryParams struct {
	Id string `path:"id"`
}

type Pagination struct {
	Next           *int
	Previous       *int
	CurrentPage    int
	RecordsPerPage int
	TotalPages     int
	TotalRecords   int
}
