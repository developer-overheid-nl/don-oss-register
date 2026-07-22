package models

type PublicCode struct {
	PubliccodeYmlVersion string                            `json:"publiccodeYmlVersion,omitempty"`
	Name                 string                            `json:"name,omitempty"`
	Url                  string                            `json:"url,omitempty"`
	LandingUrl           string                            `json:"landingURL,omitempty"`
	Platforms            []string                          `json:"platforms,omitempty"`
	DevelopmentStatus    string                            `json:"developmentStatus,omitempty"`
	SoftwareType         string                            `json:"softwareType,omitempty"`
	Legal                *PublicCodeLegal                  `json:"legal,omitempty"`
	Description          map[string]PublicCodeDescription  `json:"description,omitempty"`
	Maintenance          *PublicCodeMaintenance            `json:"maintenance,omitempty"`
	Localisation         *PublicCodeLocalisation           `json:"localisation,omitempty"`
	Organisation         *PublicCodeOrganisation           `json:"organisation,omitempty"`
	DependsOn            *PublicCodeDependsOn              `json:"dependsOn,omitempty"`
	FundedBy             []PublicCodeOrganisationReference `json:"fundedBy,omitempty"`
}

type PublicCodeLegal struct {
	License string `json:"license,omitempty"`
}

type PublicCodeDescription struct {
	ShortDescription string   `json:"shortDescription,omitempty"`
	LongDescription  string   `json:"longDescription,omitempty"`
	Features         []string `json:"features,omitempty"`
}

type PublicCodeMaintenance struct {
	Type        string                 `json:"type,omitempty"`
	Contractors []PublicCodeContractor `json:"contractors,omitempty"`
	Contacts    []PublicCodeContact    `json:"contacts,omitempty"`
}

type PublicCodeContractor struct {
	Name  string `json:"name,omitempty"`
	Until string `json:"until,omitempty"`
}

type PublicCodeContact struct {
	Name string `json:"name,omitempty"`
}

type PublicCodeLocalisation struct {
	LocalisationReady  *bool    `json:"localisationReady,omitempty"`
	AvailableLanguages []string `json:"availableLanguages,omitempty"`
}

type PublicCodeOrganisation struct {
	Uri string `json:"uri,omitempty"`
}

type PublicCodeDependsOn struct {
	Open        []PublicCodeDependency `json:"open,omitempty"`
	Proprietary []PublicCodeDependency `json:"proprietary,omitempty"`
	Hardware    []PublicCodeDependency `json:"hardware,omitempty"`
}

type PublicCodeDependency struct {
	Name string `json:"name,omitempty"`
}

type PublicCodeOrganisationReference struct {
	Name string `json:"name,omitempty"`
}
