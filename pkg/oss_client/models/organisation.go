package models

type OrganisationSummary struct {
	Uri   string `gorm:"column:uri;primaryKey" json:"uri"`
	Label string `gorm:"column:label" json:"label"`
}

type Organisation struct {
	Uri   string `gorm:"column:uri;primaryKey" json:"uri"`
	Label string `gorm:"column:label" json:"label"`
}

type ListOrganisationsParams struct {
	Page         int     `query:"page" validate:"omitempty,min=1"`
	PerPage      int     `query:"perPage" validate:"omitempty,min=1,max=100"`
	Organisation *string `query:"organisation"`
	BaseURL      string
}
