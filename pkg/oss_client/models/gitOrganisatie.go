package models

type GitOrganisationInput struct {
	Url             string `json:"url" binding:"required,url"`
	OrganisationUri string `json:"organisationUri" binding:"required,url"`
}

type GitOrganisatie struct {
	Id             string        `gorm:"column:id;primaryKey" json:"id"`
	Organisation   *Organisation `json:"organisation,omitempty" gorm:"foreignKey:OrganisationID;references:Uri"`
	OrganisationID *string       `json:"organisationId,omitempty" gorm:"column:organisation_id"`
	Url            string        `json:"url" gorm:"column:url;uniqueIndex"`
}

type GitOrganisatieSummary struct {
	Id           string        `gorm:"column:id;primaryKey" json:"id"`
	Organisation *Organisation `json:"organisation,omitempty" gorm:"foreignKey:OrganisationID;references:Uri"`
	Url          string        `json:"url"`
}

type ListGitOrganisationsParams struct {
	Page         int     `query:"page" validate:"omitempty,min=1"`
	PerPage      int     `query:"perPage" validate:"omitempty,min=1,max=100"`
	Organisation *string `query:"organisation"`
	BaseURL      string
}

type ListOrganisationsParams struct{}
