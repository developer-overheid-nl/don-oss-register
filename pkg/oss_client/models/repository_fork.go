package models

type RepositoryForkType string

const (
	RepositoryForkTypeTechnicalFork RepositoryForkType = "TECHNICAL_FORK"
	RepositoryForkTypeVariantFork   RepositoryForkType = "VARIANT_FORK"
	RepositoryForkTypeGitFork       RepositoryForkType = "GIT_FORK"
	RepositoryForkTypeURLMistake    RepositoryForkType = "URL_MISTAKE"
)
