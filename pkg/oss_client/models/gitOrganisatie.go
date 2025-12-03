package models

type PostGitOrganisatie struct {
	GitOrganisationUrl string `json:"gitOrganisationUrl" binding:"required,url"`
	OrganisationUrl    string `json:"organisationUrl" binding:"required,url"`
}

type GitOrganisatie struct {
	Id                 string        `gorm:"column:id;primaryKey" json:"id"`
	Organisation       *Organisation `json:"organisation,omitempty" gorm:"foreignKey:OrganisationID;references:Uri"`
	OrganisationID     *string       `json:"organisationId,omitempty" gorm:"column:organisation_id"`
	GitOrganisationUrl string        `json:"gitOrganisationUrl" gorm:"column:git_organisation_url;uniqueIndex"`
}

type GitOrganisatieSummary struct {
	Id                 string        `gorm:"column:id;primaryKey" json:"id"`
	Organisation       *Organisation `json:"organisation,omitempty" gorm:"foreignKey:OrganisationID;references:Uri"`
	GitOrganisationUrl string        `json:"gitOrganisationUrl"`
}

type ListGitOrganisationsParams struct {
	Page    int     `query:"page"`
	PerPage int     `query:"perPage"`
	Ids     *string `query:"ids"`
	BaseURL string
}

type ListOrganisationsParams struct {
	Page    int     `query:"page"`
	PerPage int     `query:"perPage"`
	Ids     *string `query:"ids"`
	BaseURL string
}

func (p *ListOrganisationsParams) FilterIDs() *string {
	if p == nil {
		return nil
	}
	return trimPointer(p.Ids)
}
