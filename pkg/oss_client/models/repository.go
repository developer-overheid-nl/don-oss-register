package models

import "time"

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
	Archived         bool                 `json:"archived"`
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
	IsFork           bool          `json:"-" gorm:"column:is_fork;default:false"`
	ForkBasedOnURLs  []string      `json:"-" gorm:"column:fork_based_on_urls;serializer:json"`
	Archived         bool          `json:"archived" gorm:"column:archived;default:false"`
	PublicCodeUrl    string        `json:"publicCodeUrl,omitempty" gorm:"column:public_code_url"`
	PublicCode       *PublicCode   `json:"publicCode,omitempty" gorm:"column:public_code_data;serializer:json"`
	CreatedAt        time.Time     `json:"createdAt" gorm:"column:created_at"`
	LastCrawledAt    time.Time     `json:"lastCrawledAt" gorm:"column:last_crawled_at"`
	LastActivityAt   time.Time     `json:"lastActivityAt,omitempty" gorm:"column:last_activity_at"`
	Active           bool          `json:"-" gorm:"column:active"`
}

type RepositoryInput struct {
	Url              *string   `json:"url" binding:"required,url"`
	OrganisationUri  *string   `json:"organisationUri" binding:"required,url"`
	PublicCodeUrl    *string   `json:"publicCodeUrl" binding:"omitempty,url"`
	IsFork           *bool     `json:"isFork,omitempty"`
	Archived         *bool     `json:"archived,omitempty"`
	ShortDescription *string   `json:"shortDescription,omitempty"`
	CreatedAt        time.Time `json:"createdAt" gorm:"column:created_at"`
	LastCrawledAt    time.Time `json:"lastCrawledAt" gorm:"column:last_crawled_at"`
	Name             *string   `json:"name,omitempty"`
	LastActivityAt   time.Time `json:"lastActivityAt,omitempty" gorm:"column:last_activity_at"`
}

type RepositoryParams struct {
	Id string `path:"id"`
}

type UpdateRepositoryRequest struct {
	RepositoryParams
	RepositoryInput
}
