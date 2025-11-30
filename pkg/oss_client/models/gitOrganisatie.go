package models

type PostGitOrganisatie struct {
	GitOrganisationUrl string `json:"gitOrganisationUrl" binding:"required,url"`
	OrganisationUrl    string `json:"organisationUrl" binding:"required,url"`
}

type GitOrganisatie struct {
	Id             string        `gorm:"column:id;primaryKey" json:"id"`
	Organisation   *Organisation `json:"organisation,omitempty" gorm:"foreignKey:OrganisationID;references:Uri"`
	OrganisationID *string       `json:"organisationId,omitempty" gorm:"column:organisation_id"`
	CodeHosting    []CodeHosting `json:"codeHosting" gorm:"foreignKey:PublisherID;references:Id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type GitOrganisatieSummary struct {
	Id           string        `gorm:"column:id;primaryKey" json:"id"`
	Organisation *Organisation `json:"organisation,omitempty" gorm:"foreignKey:OrganisationID;references:Uri"`
	CodeHosting  []CodeHosting `json:"codeHosting" gorm:"foreignKey:PublisherID;references:Id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type CodeHosting struct {
	ID          string `json:"-" gorm:"primaryKey"`
	URL         string `json:"url" gorm:"not null;uniqueIndex"`
	Group       *bool  `json:"group" gorm:"default:true;not null"`
	PublisherID string `json:"-" gorm:"column:publisher_id"`
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
