package models

type OrganisationSummary struct {
	Uri   string `gorm:"column:uri;primaryKey" json:"uri"`
	Label string `gorm:"column:label" json:"label"`
}

type Organisation struct {
	Uri   string `gorm:"column:uri;primaryKey" json:"uri"`
	Label string `gorm:"column:label" json:"label"`
}
